import React, { useState, useEffect, useRef, useCallback } from "react";
import { Terminal as XTerm } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";
import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime";
import * as AppBackend from "../../wailsjs/go/backend/App";
import {
  Edit2,
  Edit3,
  Download,
  Trash2,
  RefreshCw,
  ExternalLink,
  ArrowLeft,
  Check,
  Copy,
  Clipboard,
  Settings,
  ZoomIn,
  ZoomOut,
  X,
  Folder,
  File,
} from "lucide-react";
import SSHToolbar from "./ssh/SSHToolbar";
import SSHFileExplorer from "./ssh/SSHFileExplorer";
import { handleActionKey } from "../utils/a11y";

interface ResourceLineChartProps {
  /** Sequential historical data of resource metrics containing CPU, Memory, and Disk ratios */
  data: Array<{ cpu: number | null; mem: number | null; disk: number | null }>;
  /** Metric type to parse and plot */
  metric: "cpu" | "mem" | "disk";
  /** Stroke color of the plotted sparkline */
  color: string;
  /** Background area gradient or fill color of the sparkline */
  fillColor: string;
}

/**
 * ResourceLineChart Component
 *
 * Renders a lightweight SVG-based real-time line and area sparkline chart.
 * Automatically aligns historical datapoints and renders dotted grid guidelines.
 */
const ResourceLineChart: React.FC<ResourceLineChartProps> = ({ data, metric, color, fillColor }) => {
  const width = 120;
  const height = 30;
  const pointsCount = 30;

  // Ensure chart has exactly pointsCount items by padding missing slots with null values
  const paddedData = [...new Array(pointsCount).fill({ cpu: null, mem: null, disk: null }), ...data].slice(-pointsCount);

  const getX = (index: number) => {
    return (index / (pointsCount - 1)) * width;
  };

  const getY = (val: number | null) => {
    if (val === null) return height;
    return height - 1 - (val / 100) * (height - 2);
  };

  let linePath = "";
  let areaPath = "";
  let currentSegment: Array<[number, number]> = [];
  const segments: Array<Array<[number, number]>> = [];

  paddedData.forEach((d, i) => {
    const val = d[metric];
    if (val !== null) {
      currentSegment.push([getX(i), getY(val)]);
    } else if (currentSegment.length > 0) {
      segments.push(currentSegment);
      currentSegment = [];
    }
  });
  if (currentSegment.length > 0) {
    segments.push(currentSegment);
  }

  // Draw sparkline paths
  segments.forEach((seg) => {
    if (seg.length > 0) {
      let segLine = `M ${seg[0][0]} ${seg[0][1]}`;
      for (let j = 1; j < seg.length; j++) {
        segLine += ` L ${seg[j][0]} ${seg[j][1]}`;
      }
      linePath += " " + segLine;
    }
  });

  // Draw area path underneath the sparkline
  segments.forEach((seg) => {
    if (seg.length > 0) {
      let segArea = `M ${seg[0][0]} ${height} L ${seg[0][0]} ${seg[0][1]}`;
      for (let j = 1; j < seg.length; j++) {
        segArea += ` L ${seg[j][0]} ${seg[j][1]}`;
      }
      segArea += ` L ${seg.at(-1)[0]} ${height} Z`;
      areaPath += " " + segArea;
    }
  });

  return (
    <svg width={width} height={height} className="overflow-hidden border border-mui-grey-200 dark:border-white/10 rounded bg-white dark:bg-mui-grey-950">
      {/* Horizontal grid lines */}
      <line x1={0} y1={height / 3} x2={width} y2={height / 3} stroke="rgba(156, 163, 175, 0.12)" strokeDasharray="1,1" />
      <line x1={0} y1={(2 * height) / 3} x2={width} y2={(2 * height) / 3} stroke="rgba(156, 163, 175, 0.12)" strokeDasharray="1,1" />

      {/* Vertical grid lines */}
      <line x1={width / 4} y1={0} x2={width / 4} y2={height} stroke="rgba(156, 163, 175, 0.12)" strokeDasharray="1,1" />
      <line x1={width / 2} y1={0} x2={width / 2} y2={height} stroke="rgba(156, 163, 175, 0.12)" strokeDasharray="1,1" />
      <line x1={(3 * width) / 4} y1={0} x2={(3 * width) / 4} y2={height} stroke="rgba(156, 163, 175, 0.12)" strokeDasharray="1,1" />

      {areaPath && <path d={areaPath} fill={fillColor} />}
      {linePath && <path d={linePath} fill="none" stroke={color} strokeWidth={1.5} />}
    </svg>
  );
};

interface SSHSessionViewProps {
  /** The current active SSH/WSL session profile */
  session: any;
  /** Close callback to terminate active connection panel */
  onClose: () => void;
  /** Toast callback to show contextual alert popups */
  addToast: (
    title: string,
    message: string,
    type?: "info" | "success" | "warn" | "error",
  ) => void;
  /** Focus indicator indicating if this terminal tab is currently visible */
  isActive: boolean;
  /** Layout theme selector ('light' or 'dark') */
  theme?: string;
  /** Opens global settings dialog corresponding to the selected category */
  onOpenSettings: (category: string) => void;
}

/**
 * SSHSessionView Component
 *
 * Implements the interactive terminal session view. Hooks up an xterm.js instance
 * directly to backend pseudo-terminals (PTY) and coordinates remote sftp file explorer access.
 */
const SSHSessionView: React.FC<SSHSessionViewProps> = ({
  session,
  onClose,
  addToast,
  isActive,
  theme,
  onOpenSettings,
}) => {
  // --- Refs ---
  const terminalRef = useRef<HTMLDivElement>(null);
  const xterm = useRef<XTerm | null>(null);
  const fitAddon = useRef(new FitAddon());
  const contextMenuRef = useRef<HTMLDivElement>(null);
  const currentPathRef = useRef("");
  const isFetchingUsageRef = useRef(false);

  // --- State Hooks ---
  const [connecting, setConnecting] = useState(true);
  const [isNewFileModalOpen, setIsNewFileModalOpen] = useState(false);
  const [isNewFolderModalOpen, setIsNewFolderModalOpen] = useState(false);
  const [newFileName, setNewFileName] = useState("");
  const [newFolderName, setNewFolderName] = useState("");
  const [selectedExtension, setSelectedExtension] = useState("");
  const [createTargetDir, setCreateTargetDir] = useState("");
  const [remotePath, setRemotePath] = useState("");
  const [editingPath, setEditingPath] = useState("");
  const [files, setFiles] = useState<any[]>([]);
  const [loadingFiles, setLoadingFiles] = useState(false);
  const [explorerVisible, setExplorerVisible] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [fileContextMenu, setFileContextMenu] = useState<any>(null);
  const [showHiddenFiles, setShowHiddenFiles] = useState(true);
  const [explorerContextMenu, setExplorerContextMenu] = useState<any>(null);
  const [terminalContextMenu, setTerminalContextMenu] = useState<any>(null);
  const [sortConfig, setSortConfig] = useState<{
    key: string;
    direction: "asc" | "desc";
  }>({ key: "name", direction: "asc" });

  const [isZoomEnabled, setIsZoomEnabled] = useState<boolean>(() => {
    return localStorage.getItem('ostenia_ssh_zoom_enabled') !== 'false';
  });
  const [zoomFontSize, setZoomFontSize] = useState<number>(14);

  const handleZoomIn = () => {
    setZoomFontSize((prev) => Math.min(prev + 2, 40));
  };

  const handleZoomOut = () => {
    setZoomFontSize((prev) => Math.max(prev - 2, 8));
  };

  const handleZoomReset = () => {
    setZoomFontSize(14);
  };

  const handlersRef = useRef({ handleZoomIn, handleZoomOut, handleZoomReset });
  useEffect(() => {
    handlersRef.current = { handleZoomIn, handleZoomOut, handleZoomReset };
  });

  useEffect(() => {
    const handleZoomSettingsChanged = () => {
      setIsZoomEnabled(localStorage.getItem('ostenia_ssh_zoom_enabled') !== 'false');
    };

    window.addEventListener('ostenia_ssh_zoom_settings_changed', handleZoomSettingsChanged);
    return () => {
      window.removeEventListener('ostenia_ssh_zoom_settings_changed', handleZoomSettingsChanged);
    };
  }, []);

  useEffect(() => {
    if (xterm.current) {
      xterm.current.options.fontSize = zoomFontSize;
      setTimeout(performFit, 50);
    }
  }, [zoomFontSize]);

  const [resourceUsage, setResourceUsage] = useState<{
    cpu: number;
    mem: number;
    memTotal: number;
    memUsed: number;
    disk: number;
    diskTotal: number;
    diskUsed: number;
  } | null>(null);
  const [monitorInterval, setMonitorInterval] = useState<number>(() => {
    const val = Number.parseInt(localStorage.getItem('ostenia_ssh_monitor_interval') || '3', 10);
    return Number.isNaN(val) || val < 1 ? 3 : val;
  });
  const [isMonitoringEnabled, setIsMonitoringEnabled] = useState<boolean>(() => {
    return localStorage.getItem('ostenia_ssh_monitor_enabled') !== 'false';
  });
  const [hoveredMetric, setHoveredMetric] = useState<"cpu" | "mem" | "disk" | null>(null);
  const [displayMode, setDisplayMode] = useState<string>(() => {
    return localStorage.getItem('ostenia_ssh_monitor_display_mode') || 'tooltip';
  });
  const [history, setHistory] = useState<Array<{
    cpu: number | null;
    mem: number | null;
    disk: number | null;
  }>>([]);

  // Load and apply SSH monitoring parameters from client preferences on mount/modification
  useEffect(() => {
    const handleSettingsChanged = () => {
      setIsMonitoringEnabled(localStorage.getItem('ostenia_ssh_monitor_enabled') !== 'false');
      const val = Number.parseInt(localStorage.getItem('ostenia_ssh_monitor_interval') || '3', 10);
      setMonitorInterval(Number.isNaN(val) || val < 1 ? 3 : val);
      setDisplayMode(localStorage.getItem('ostenia_ssh_monitor_display_mode') || 'tooltip');
    };

    window.addEventListener('ostenia_ssh_monitor_settings_changed', handleSettingsChanged);
    handleSettingsChanged();

    return () => {
      window.removeEventListener('ostenia_ssh_monitor_settings_changed', handleSettingsChanged);
    };
  }, []);

  // Set up background resource utilization queries
  useEffect(() => {
    if (!isMonitoringEnabled || connecting) {
      setResourceUsage(null);
      return;
    }

    let isMounted = true;
    const fetchUsage = async () => {
      if (isFetchingUsageRef.current) return;
      isFetchingUsageRef.current = true;
      try {
        const usage = await AppBackend.GetSSHResourceUsage(session.id);
        if (isMounted) {
          setResourceUsage(usage);
          setHistory((prev) => [...prev, { cpu: usage.cpu, mem: usage.mem, disk: usage.disk }].slice(-30));
        }
      } catch (e) {
        console.error("Failed to fetch SSH resource usage", e);
        if (isMounted) {
          setResourceUsage(null);
          setHistory((prev) => [...prev, { cpu: null, mem: null, disk: null }].slice(-30));
        }
      } finally {
        isFetchingUsageRef.current = false;
      }
    };

    // Initial fetch instantly
    fetchUsage();

    const intervalId = setInterval(fetchUsage, monitorInterval * 1000);

    return () => {
      isMounted = false;
      clearInterval(intervalId);
    };
  }, [session.id, monitorInterval, isMonitoringEnabled, connecting]);

  useEffect(() => {
    setEditingPath(remotePath);
  }, [remotePath]);

  /**
   * Helper utility to convert raw file sizes into readable units.
   */
  const formatSize = (bytes: number) => {
    if (!bytes) return "-";
    const units = ["B", "KB", "MB", "GB", "TB"];
    let size = bytes;
    let unitIndex = 0;
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }
    return `${size.toFixed(1)} ${units[unitIndex]}`;
  };

  // Close context menus when clicking outside
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (contextMenuRef.current?.contains(e.target as Node)) {
        return;
      }
      setFileContextMenu(null);
      setExplorerContextMenu(null);
      setTerminalContextMenu(null);
    };
    window.addEventListener("click", handleClick);
    return () => window.removeEventListener("click", handleClick);
  }, []);

  /**
   * Triggers the custom contextual action menu on remote files/folders.
   * Adjusts top Y position upwards near viewport limits to prevent layout clipping.
   */
  const handleFileContextMenu = (e: React.MouseEvent, file: any) => {
    e.preventDefault();
    e.stopPropagation();

    const menuHeight = 180;
    let y = e.clientY;
    if (y + menuHeight > window.innerHeight) {
      y = Math.max(10, y - menuHeight);
    }

    setFileContextMenu({
      x: e.clientX,
      y: y,
      file: file,
    });
    setExplorerContextMenu(null);
    setTerminalContextMenu(null);
  };

  /**
   * Triggers the custom background action menu for the SFTP explorer.
   * Adjusts Y positioning near bottom edges.
   */
  const handleExplorerContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();

    const menuHeight = 130;
    let y = e.clientY;
    if (y + menuHeight > window.innerHeight) {
      y = Math.max(10, y - menuHeight);
    }

    setExplorerContextMenu({
      x: e.clientX,
      y: y,
    });
    setFileContextMenu(null);
    setTerminalContextMenu(null);
  };

  /**
   * Triggers the custom terminal utilities action menu (Copy/Paste/Refresh layout).
   * Adjusts Y positioning near bottom edges.
   */
  const handleTerminalContextMenu = (e: React.MouseEvent) => {
    e.preventDefault();

    const menuHeight = 100;
    let y = e.clientY;
    if (y + menuHeight > window.innerHeight) {
      y = Math.max(10, y - menuHeight);
    }

    setTerminalContextMenu({
      x: e.clientX,
      y: y,
    });
    setFileContextMenu(null);
    setExplorerContextMenu(null);
  };

  /**
   * Reads selected text segments from the xterm.js instance and saves to clipboard.
   */
  const handleTerminalCopy = async () => {
    if (xterm.current) {
      const selection = xterm.current.getSelection();
      if (selection) {
        await navigator.clipboard.writeText(selection);
        addToast("SSH", "Text copied to clipboard", "success");
      }
    }
  };

  /**
   * Reads text from system clipboard and emits directly to active PTY channel.
   */
  const handleTerminalPaste = async () => {
    try {
      const text = await navigator.clipboard.readText();
      if (text) {
        AppBackend.SendSSHInput(session.id, text);
      }
    } catch (e: any) {
      addToast("Error", "Clipboard permission denied", "error");
    }
  };

  const handleTerminalRefresh = () => {
    performFit();
  };

  /**
   * Forces re-fit on xterm grid layout matching visible shell dimensions.
   */
  const performFit = () => {
    if (!terminalRef.current || !xterm.current || !isActive) return;
    if (terminalRef.current.offsetParent === null) return;

    try {
      fitAddon.current?.fit();
      const dims = fitAddon.current?.proposeDimensions();

      if (dims && dims.cols >= 20 && dims.rows >= 2) {
        const safeCols = Math.max(dims.cols, 120);
        const safeRows = Math.max(dims.rows, 24);
        AppBackend.ResizeSSHTerminal(session.id, safeCols, safeRows);
        setTimeout(() => xterm.current?.focus(), 50);
      }
    } catch (e) {
      console.error("Fit error:", e);
    }
  };

  // Adjust xterm theme configuration dynamically when theme options change
  useEffect(() => {
    if (xterm.current) {
      const newTheme =
        theme === "dark"
          ? {
              background: "#121212",
              foreground: "#eeeeee",
              cursor: "#2196f3",
              selectionBackground: "rgba(33, 150, 243, 0.3)",
            }
          : {
              background: "#fafafa",
              foreground: "#212121",
              cursor: "#1976d2",
              selectionBackground: "rgba(25, 118, 210, 0.2)",
            };
      xterm.current.options.theme = newTheme;
    }
  }, [theme]);

  // Handle lazy PTY alignment upon focused view toggles
  useEffect(() => {
    if (isActive) {
      setTimeout(performFit, 100);
      setTimeout(performFit, 600);
      setTimeout(performFit, 1500);
    }
  }, [isActive]);

  // Hook up full xterm.js instance and register Wails event listeners
  useEffect(() => {
    xterm.current = new XTerm({
      cursorBlink: true,
      convertEol: true,
      fontSize: zoomFontSize,
      lineHeight: 1.2,
      fontFamily:
        'JetBrains Mono, Menlo, Monaco, Consolas, "Courier New", monospace',
      theme:
        theme === "dark"
          ? {
              background: "#121212",
              foreground: "#eeeeee",
              cursor: "#2196f3",
              selectionBackground: "rgba(33, 150, 243, 0.3)",
            }
          : {
              background: "#fafafa",
              foreground: "#212121",
              cursor: "#1976d2",
              selectionBackground: "rgba(25, 118, 210, 0.2)",
            },
      allowProposedApi: true,
      scrollback: 10000,
      macOptionIsMeta: true,
    });

    xterm.current.loadAddon(fitAddon.current);
    if (terminalRef.current) {
      xterm.current.open(terminalRef.current);

      // Custom key event handler for zoom keyboard shortcuts (Ctrl + Plus, Ctrl + Minus, Ctrl + 0)
      xterm.current.attachCustomKeyEventHandler((event: KeyboardEvent) => {
        if (event.type === "keydown") {
          const isCtrlOrCmd = event.ctrlKey || event.metaKey;
          if (isCtrlOrCmd) {
            const enabled = localStorage.getItem('ostenia_ssh_zoom_enabled') !== 'false';
            if (enabled) {
              if (event.key === "=" || event.key === "+") {
                handlersRef.current.handleZoomIn();
                return false;
              }
              if (event.key === "-") {
                handlersRef.current.handleZoomOut();
                return false;
              }
              if (event.key === "0") {
                handlersRef.current.handleZoomReset();
                return false;
              }
            }
          }
        }
        return true;
      });
    }

    let resizeTimeout: any;
    const handleWindowResize = () => {
      clearTimeout(resizeTimeout);
      resizeTimeout = setTimeout(performFit, 200);
    };

    window.addEventListener("resize", handleWindowResize);
    const ro = new ResizeObserver(() => {
      clearTimeout(resizeTimeout);
      resizeTimeout = setTimeout(performFit, 250);
    });

    if (terminalRef.current) ro.observe(terminalRef.current);
    setTimeout(performFit, 500);

    // Forward local user keystrokes straight to back-end shell stream
    xterm.current.onData((data) => AppBackend.SendSSHInput(session.id, data));
    xterm.current.onTitleChange((title) => {
      if (title.includes(":")) {
        const parts = title.split(":");
        const potentialPath = parts.at(-1).trim();
        if (potentialPath.startsWith("/") || potentialPath.startsWith("~")) {
          syncExplorer(potentialPath);
        }
      }
    });

    const handleOutput = (event: any) => {
      if (event.sessionId === session.id) xterm.current?.write(event.data);
    };

    const handlePathChange = (event: any) => {
      if (event.sessionId === session.id) {
        setRemotePath((prev) => {
          if (event.path !== prev) loadRemoteFiles(event.path, false, true);
          return event.path;
        });
      }
    };

    const handleDisconnect = (id: any) => {
      if (id === session.id) {
        xterm.current?.write(
          "\r\n\x1b[31mDisconnected from server.\x1b[0m\r\n",
        );
        addToast("SSH", "Disconnected from " + session.name, "warn");
      }
    };

    EventsOn("ssh_output", handleOutput);
    EventsOn("ssh_path_changed", handlePathChange);
    EventsOn("ssh_disconnected", handleDisconnect);
    connectSSH();

    return () => {
      (EventsOff as any)("ssh_output", handleOutput);
      (EventsOff as any)("ssh_path_changed", handlePathChange);
      (EventsOff as any)("ssh_disconnected", handleDisconnect);
      if (ro) ro.disconnect();
      window.removeEventListener("resize", handleWindowResize);
      if (xterm.current) xterm.current.dispose();
    };
  }, [session.id]);

  /**
   * Registers a fresh connection with the Go backend
   */
  const connectSSH = async () => {
    if (!xterm.current) return;
    setConnecting(true);
    xterm.current.write(`Connecting to ${session.host}...\r\n`);
    try {
      await AppBackend.ConnectSSH(session);
      setConnecting(false);
      xterm.current.write("\x1b[32mConnected successfully.\x1b[0m\r\n\r\n");
      performFit();
      loadRemoteFiles("", false, true);
    } catch (err) {
      setConnecting(false);
      xterm.current.write(`\x1b[31mConnection failed: ${err}\x1b[0m\r\n`);
      addToast("Error", "SSH connection failed: " + err, "error");
    }
  };

  /**
   * Disconnects current session and starts connection flow afresh.
   */
  const handleReconnect = async () => {
    try {
      await AppBackend.DisconnectSSH(session.id);
    } catch (e) {
      console.error("Disconnect error on reconnect:", e);
    }
    xterm.current?.clear();
    connectSSH();
  };

  /**
   * Helper to handle file loading errors separately to minimize Cognitive Complexity.
   */
  const handleLoadFilesError = (err: any, isManualEntry: boolean, isAutoSync: boolean) => {
    if (isManualEntry) {
      addToast("Navigation", "Directory not available", "error");
      setEditingPath(remotePath);
      return;
    }

    if (isAutoSync) {
      // Suppress Toast errors during background/automatic sync silently
      return;
    }

    const errStr = String(err).toLowerCase();
    if (
      errStr.includes("eof") ||
      errStr.includes("session not found") ||
      errStr.includes("session not connected") ||
      errStr.includes("sftp not connected")
    ) {
      return;
    }

    addToast("Explorer", "Failed to list files: " + err, "error");
  };

  /**
   * Queries directory listings from remote target via SFTP.
   *
   * @param path Target remote directory path to load.
   * @param isManualEntry Boolean indicating if navigation was explicitly typed by the user.
   * @param isAutoSync Boolean flag indicating automatic sync events. Auto-sync silently absorbs
   *                   any permissions/access/EOF errors to prevent annoying toast interruptions
   *                   while the user is busy interacting with the terminal.
   */
  const loadRemoteFiles = async (path: string, isManualEntry = false, isAutoSync = false) => {
    setLoadingFiles(true);
    try {
      let targetPath = path;
      if (targetPath === "") {
        const current = await AppBackend.GetRemoteCurrentPath(session.id);
        if (current) {
          targetPath = current;
        }
      }
      const result = await AppBackend.GetRemoteFiles(session.id, targetPath);
      if (result === null && isManualEntry)
        throw new Error("Directory not available");
      setFiles(result || []);
      setRemotePath(targetPath);
      currentPathRef.current = targetPath;
    } catch (err: any) {
      handleLoadFilesError(err, isManualEntry, isAutoSync);
    } finally {
      setLoadingFiles(false);
    }
  };

  /**
   * Synchronizes the remote sftp explorer view with terminal working directories.
   */
  const syncExplorer = async (forcedPath: string | null = null) => {
    try {
      let current =
        forcedPath || (await AppBackend.GetRemoteCurrentPath(session.id));
      if (!current) return;
      let normalized = current.trim();
      if (normalized.length > 1 && normalized.endsWith("/"))
        normalized = normalized.slice(0, -1);
      if (normalized !== currentPathRef.current) loadRemoteFiles(normalized, false, true);
    } catch (e) {}
  };

  /**
   * Trigger folder traversal or edits the file locally.
   */
  const handleFileDoubleClick = (file: any) => {
    if (file.isDir) {
      const current = remotePath || "/";
      const newPath = current.endsWith("/")
        ? `${current}${file.name}`
        : `${current}/${file.name}`;
      loadRemoteFiles(newPath);
      AppBackend.SendSSHInput(session.id, `cd "${newPath}"\r`);
    } else {
      handleEdit(file);
    }
  };

  /**
   * Navigates up to the parent directory path level.
   */
  const navigateUp = () => {
    if (remotePath === "/" || remotePath === "") return;
    const normalized = remotePath.endsWith("/")
      ? remotePath.slice(0, -1)
      : remotePath;
    const parts = normalized.split("/");
    parts.pop();
    const newPath = parts.join("/") || "/";
    loadRemoteFiles(newPath);
    AppBackend.SendSSHInput(session.id, `cd "${newPath}"\r`);
  };

  /**
   * Downloads a remote file using native save dialog prompts.
   */
  const handleDownload = async (file: any) => {
    try {
      await AppBackend.DownloadRemoteFile(
        session.id,
        `${remotePath}/${file.name}`,
      );
      addToast("Success", "File download started", "success");
    } catch (err: any) {
      addToast("Error", "Download failed: " + err, "error");
    }
  };

  /**
   * Uploads a file via file picker prompts and updates listing.
   */
  const handleUpload = async () => {
    try {
      await AppBackend.UploadRemoteFile(session.id, remotePath);
      addToast("Success", "File uploaded successfully", "success");
      loadRemoteFiles(remotePath);
    } catch (err: any) {
      addToast("Error", "Upload failed: " + err, "error");
    }
  };

  /**
   * Initiates local editing wrapper for the selected file.
   */
  const handleEdit = async (file: any) => {
    try {
      addToast("Editor", "Opening " + file.name + "...", "info");
      await AppBackend.EditRemoteFile(session.id, `${remotePath}/${file.name}`);
      addToast("Success", "File saved and uploaded", "success");
      loadRemoteFiles(remotePath);
    } catch (err: any) {
      addToast("Error", "Edit failed: " + err, "error");
    }
  };

  /**
   * Deletes target files/folders recursively on the remote machine.
   */
  const handleDelete = async (file: any) => {
    if (confirm(`Are you sure you want to delete ${file.name}?`)) {
      try {
        await AppBackend.ExecuteSFTPAction(
          session.id,
          "delete",
          `${remotePath}/${file.name}`,
          "",
        );
        addToast("Success", "Deleted " + file.name, "success");
        loadRemoteFiles(remotePath);
      } catch (err: any) {
        addToast("Error", "Delete failed: " + err, "error");
      }
    }
  };

  // Perform search, hidden filters, and sort order calculation over files list
  const sortedFiles = [...files]
    .filter((f) => {
      const matchesSearch = f.name.toLowerCase().includes(searchQuery.toLowerCase());
      if (!showHiddenFiles) {
        return matchesSearch && !f.name.startsWith(".");
      }
      return matchesSearch;
    })
    .sort((a, b) => {
      if (a.isDir !== b.isDir) return b.isDir ? 1 : -1;
      let comparison =
        sortConfig.key === "name"
          ? a.name.localeCompare(b.name)
          : (a.size || 0) - (b.size || 0);
      return sortConfig.direction === "asc" ? comparison : -comparison;
    });

  const toggleSort = (key: string) => {
    setSortConfig((prev) => ({
      key,
      direction: prev.key === key && prev.direction === "asc" ? "desc" : "asc",
    }));
  };

  const openNewFileModal = (targetDir: string) => {
    setCreateTargetDir(targetDir || remotePath || "/");
    setNewFileName("");
    setSelectedExtension("");
    setIsNewFileModalOpen(true);
  };

  const openNewFolderModal = (targetDir: string) => {
    setCreateTargetDir(targetDir || remotePath || "/");
    setNewFolderName("");
    setIsNewFolderModalOpen(true);
  };

  const handleCreateFileSubmit = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const name = newFileName.trim();
    if (!name) {
      addToast("Error", "File name cannot be empty", "error");
      return;
    }

    let finalName = name;
    if (selectedExtension && !name.toLowerCase().endsWith(selectedExtension.toLowerCase())) {
      finalName = name + selectedExtension;
    }

    const dir = createTargetDir.endsWith("/") ? createTargetDir : `${createTargetDir}/`;
    const finalPath = `${dir}${finalName}`;

    try {
      await AppBackend.ExecuteSFTPAction(session.id, "create_file", finalPath, "");
      addToast("Success", `File "${finalName}" created successfully`, "success");
      setIsNewFileModalOpen(false);
      loadRemoteFiles(remotePath);
    } catch (err: any) {
      addToast("Error", err.toString(), "error");
    }
  };

  const handleCreateFolderSubmit = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const name = newFolderName.trim();
    if (!name) {
      addToast("Error", "Folder name cannot be empty", "error");
      return;
    }

    const dir = createTargetDir.endsWith("/") ? createTargetDir : `${createTargetDir}/`;
    const finalPath = `${dir}${name}`;

    try {
      await AppBackend.ExecuteSFTPAction(session.id, "mkdir", finalPath, "");
      addToast("Success", `Folder "${name}" created successfully`, "success");
      setIsNewFolderModalOpen(false);
      loadRemoteFiles(remotePath);
    } catch (err: any) {
      addToast("Error", err.toString(), "error");
    }
  };

  const handleManualNavigation = (path: string) => {
    const p = path.trim();
    if (p && p !== remotePath) {
      loadRemoteFiles(p, true).then(() => {
        AppBackend.SendSSHInput(session.id, `cd "${p}"\r`);
      });
    }
  };

  return (
    <div className="flex flex-col h-full bg-white dark:bg-mui-dark-bg overflow-hidden border-t border-mui-grey-200 dark:border-white/5">
      <SSHToolbar
        explorerVisible={explorerVisible}
        setExplorerVisible={setExplorerVisible}
        onReconnect={handleReconnect}
        connecting={connecting}
      />

      <div className="flex-1 flex overflow-hidden">
        {explorerVisible && (
          <SSHFileExplorer
            remotePath={remotePath}
            editingPath={editingPath}
            setEditingPath={setEditingPath}
            onNavigateUp={navigateUp}
            onSync={syncExplorer}
            searchQuery={searchQuery}
            setSearchQuery={setSearchQuery}
            onUpload={handleUpload}
            onNewFolder={() => openNewFolderModal(remotePath)}
            loadingFiles={loadingFiles}
            sortedFiles={sortedFiles}
            onFileDoubleClick={handleFileDoubleClick}
            onFileContextMenu={handleFileContextMenu}
            onExplorerContextMenu={handleExplorerContextMenu}
            formatSize={formatSize}
            toggleSort={toggleSort}
            sortConfig={sortConfig}
            onManualNavigation={handleManualNavigation}
          />
        )}

        {/* --- Interactive Terminal Container --- */}
        <div className="flex-1 bg-white dark:bg-mui-dark-bg relative overflow-hidden">
          <div
            ref={terminalRef}
            className="absolute inset-0 px-2 pt-2"
            onContextMenu={handleTerminalContextMenu}
          />
          {connecting && (
            <div className="absolute inset-0 bg-white dark:bg-mui-dark-bg flex items-center justify-center">
              <div className="flex items-center gap-3">
                <RefreshCw
                  className="animate-spin text-mui-blue-600 dark:text-mui-blue-500"
                  size={18}
                />
                <span className="text-mui-grey-600 dark:text-mui-grey-400 text-xs font-bold uppercase tracking-widest">Connecting...</span>
              </div>
            </div>
          )}

          {isZoomEnabled && (
            <div className="absolute bottom-4 right-4 z-10 flex flex-col rounded-lg shadow-lg border border-mui-grey-200 dark:border-white/10 bg-white/90 dark:bg-mui-grey-900/90 overflow-hidden divide-y divide-mui-grey-200 dark:divide-white/10">
              <button
                type="button"
                onClick={() => handleZoomIn()}
                className="p-2 text-mui-grey-600 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 transition-colors focus:outline-none flex items-center justify-center"
                title="Zoom In (Ctrl + Plus)"
              >
                <ZoomIn size={16} />
              </button>
              <button
                type="button"
                onClick={() => handleZoomReset()}
                className="p-2 text-mui-grey-600 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 transition-colors focus:outline-none flex items-center justify-center"
                title="Reset Zoom (Ctrl + 0)"
              >
                <RefreshCw size={14} className="mx-auto" />
              </button>
              <button
                type="button"
                onClick={() => handleZoomOut()}
                className="p-2 text-mui-grey-600 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 transition-colors focus:outline-none flex items-center justify-center"
                title="Zoom Out (Ctrl + Minus)"
              >
                <ZoomOut size={16} />
              </button>
            </div>
          )}
        </div>
      </div>

      {/* --- Bottom Resource Usage Monitoring Bar --- */}
      <div className="h-12 flex items-center justify-between px-3 bg-mui-grey-50 dark:bg-mui-grey-900 border-t border-mui-grey-200 dark:border-white/5 shrink-0 relative select-none">
        <div className="flex items-center gap-4 text-xs">
          <div className="relative">
            <button
              type="button"
              onClick={() => onOpenSettings("ssh")}
              className="p-1 rounded text-mui-grey-500 dark:text-mui-grey-400 hover:text-mui-blue-600 dark:hover:text-white transition-colors"
              title="Monitoring Settings"
            >
              <Settings size={14} />
            </button>
          </div>

          {isMonitoringEnabled ? (
            <div className="flex items-center gap-6 text-[10px] font-bold text-mui-grey-600 dark:text-mui-grey-400">
              {/* CPU Metric Container */}
              <div
                className="relative py-1 cursor-help group flex items-center gap-2"
                onMouseEnter={() => setHoveredMetric("cpu")}
                onMouseLeave={() => setHoveredMetric(null)}
              >
                <span className="min-w-[55px]">CPU: {resourceUsage ? `${resourceUsage.cpu.toFixed(0)}%` : "—"}</span>
                {displayMode === "tooltip" && hoveredMetric === "cpu" && (
                  <div className="absolute bottom-7 left-0 z-50 bg-white dark:bg-mui-grey-850 p-2 rounded shadow-lg border border-mui-grey-200 dark:border-white/10 flex flex-col gap-1 items-center animate-in fade-in duration-100 min-w-[130px]">
                    <span className="text-[9px] uppercase tracking-wider text-mui-grey-500 dark:text-mui-grey-400">CPU Usage History</span>
                    <ResourceLineChart data={history} metric="cpu" color="#2196f3" fillColor="rgba(33, 150, 243, 0.15)" />
                  </div>
                )}
                {displayMode !== "tooltip" && (displayMode === "always" || (displayMode === "hover-inline" && hoveredMetric === "cpu")) && (
                  <div className="flex items-center shrink-0">
                    <ResourceLineChart data={history} metric="cpu" color="#2196f3" fillColor="rgba(33, 150, 243, 0.15)" />
                  </div>
                )}
              </div>

              {/* RAM Metric Container */}
              <div
                className="relative py-1 cursor-help group flex items-center gap-2"
                onMouseEnter={() => setHoveredMetric("mem")}
                onMouseLeave={() => setHoveredMetric(null)}
              >
                <span className="min-w-[130px]">RAM: {resourceUsage ? `${(resourceUsage.memUsed / 1024).toFixed(1)} GB / ${(resourceUsage.memTotal / 1024).toFixed(1)} GB (${resourceUsage.mem.toFixed(0)}%)` : "—"}</span>
                {displayMode === "tooltip" && hoveredMetric === "mem" && (
                  <div className="absolute bottom-7 left-0 z-50 bg-white dark:bg-mui-grey-850 p-2 rounded shadow-lg border border-mui-grey-200 dark:border-white/10 flex flex-col gap-1 items-center animate-in fade-in duration-100 min-w-[130px]">
                    <span className="text-[9px] uppercase tracking-wider text-mui-grey-500 dark:text-mui-grey-400">RAM Usage History</span>
                    <ResourceLineChart data={history} metric="mem" color="#9c27b0" fillColor="rgba(156, 39, 176, 0.15)" />
                  </div>
                )}
                {displayMode !== "tooltip" && (displayMode === "always" || (displayMode === "hover-inline" && hoveredMetric === "mem")) && (
                  <div className="flex items-center shrink-0">
                    <ResourceLineChart data={history} metric="mem" color="#9c27b0" fillColor="rgba(156, 39, 176, 0.15)" />
                  </div>
                )}
              </div>

              {/* Disk Metric Container */}
              <div
                className="relative py-1 cursor-help group flex items-center gap-2"
                onMouseEnter={() => setHoveredMetric("disk")}
                onMouseLeave={() => setHoveredMetric(null)}
              >
                <span className="min-w-[140px]">DISK: {resourceUsage ? `${(resourceUsage.diskUsed / 1024).toFixed(1)} GB / ${(resourceUsage.diskTotal / 1024).toFixed(1)} GB (${resourceUsage.disk.toFixed(0)}%)` : "—"}</span>
                {displayMode === "tooltip" && hoveredMetric === "disk" && (
                  <div className="absolute bottom-7 left-0 z-50 bg-white dark:bg-mui-grey-850 p-2 rounded shadow-lg border border-mui-grey-200 dark:border-white/10 flex flex-col gap-1 items-center animate-in fade-in duration-100 min-w-[130px]">
                    <span className="text-[9px] uppercase tracking-wider text-mui-grey-500 dark:text-mui-grey-400">Disk Usage History</span>
                    <ResourceLineChart data={history} metric="disk" color="#009688" fillColor="rgba(0, 150, 136, 0.15)" />
                  </div>
                )}
                {displayMode !== "tooltip" && (displayMode === "always" || (displayMode === "hover-inline" && hoveredMetric === "disk")) && (
                  <div className="flex items-center shrink-0">
                    <ResourceLineChart data={history} metric="disk" color="#009688" fillColor="rgba(0, 150, 136, 0.15)" />
                  </div>
                )}
              </div>
            </div>
          ) : (
            <span className="text-[10px] text-mui-grey-400 uppercase tracking-wider font-bold">Monitoring disabled</span>
          )}
        </div>
      </div>

      {/* --- File/Folder Item Right-Click Context Menu --- */}
      {fileContextMenu && (
        <div
          ref={contextMenuRef}
          className="fixed z-50 bg-white dark:bg-mui-grey-800 shadow-xl border border-mui-grey-200 dark:border-white/10 rounded-lg py-1 min-w-[140px] animate-in fade-in zoom-in-95 duration-100 cursor-default p-0"
          style={{ top: fileContextMenu.y, left: fileContextMenu.x }}
        >
          {fileContextMenu.file?.isDir && (
            <>
              <button
                type="button"
                onClick={() => {
                  const targetDir = `${remotePath}/${fileContextMenu.file.name}`;
                  openNewFileModal(targetDir);
                  setFileContextMenu(null);
                }}
                className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
              >
                <File size={14} /> New File
              </button>

              <button
                type="button"
                onClick={() => {
                  const targetDir = `${remotePath}/${fileContextMenu.file.name}`;
                  openNewFolderModal(targetDir);
                  setFileContextMenu(null);
                }}
                className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
              >
                <Folder size={14} /> New Folder
              </button>

              <div className="h-px bg-mui-grey-100 dark:bg-white/5 my-1" />
            </>
          )}

          <button
            type="button"
            onClick={async () => {
              const name = prompt("Rename:", fileContextMenu.file.name);
              if (name && name !== fileContextMenu.file.name) {
                try {
                  await AppBackend.ExecuteSFTPAction(
                    session.id,
                    "rename",
                    `${remotePath}/${fileContextMenu.file.name}`,
                    `${remotePath}/${name}`,
                  );
                  loadRemoteFiles(remotePath);
                } catch (err: any) {
                  addToast("Error", err.toString(), "error");
                }
              }
              setFileContextMenu(null);
            }}
            className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
          >
            <Edit2 size={14} /> Rename
          </button>

          {!fileContextMenu?.file?.isDir && (
            <>
              <button
                type="button"
                onClick={() => {
                  handleEdit(fileContextMenu.file);
                  setFileContextMenu(null);
                }}
                className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
              >
                <Edit3 size={14} /> Edit File
              </button>
              {(() => {
                const name = fileContextMenu?.file?.name?.toLowerCase() || "";
                const isArchive =
                  name.endsWith(".zip") ||
                  name.endsWith(".tar") ||
                  name.endsWith(".gz") ||
                  name.endsWith(".7z") ||
                  name.endsWith(".rar") ||
                  name.endsWith(".bz2") ||
                  name.endsWith(".xz");
                if (!isArchive) {
                  return (
                    <button
                      type="button"
                      onClick={() => {
                        onOpenSettings("config");
                        setFileContextMenu(null);
                      }}
                      className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
                    >
                      <ExternalLink size={14} /> Open With
                    </button>
                  );
                }
                return null;
              })()}
              <button
                type="button"
                onClick={() => {
                  handleDownload(fileContextMenu.file);
                  setFileContextMenu(null);
                }}
                className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
              >
                <Download size={14} /> Download
              </button>
            </>
          )}
          <div className="h-px bg-mui-grey-100 dark:bg-white/5 my-1" />
          <button
            type="button"
            onClick={() => {
              handleDelete(fileContextMenu.file);
              setFileContextMenu(null);
            }}
            className="w-full px-4 py-2 text-left text-[11px] font-bold text-red-500 hover:bg-red-50 dark:hover:bg-red-500/10 flex items-center gap-2 transition-all"
          >
            <Trash2 size={14} /> Delete
          </button>
        </div>
      )}

      {/* --- File Explorer Background Right-Click Context Menu --- */}
      {explorerContextMenu && (
        <div
          ref={contextMenuRef}
          className="fixed z-50 bg-white dark:bg-mui-grey-800 shadow-xl border border-mui-grey-200 dark:border-white/10 rounded-lg py-1 min-w-[170px] animate-in fade-in zoom-in-95 duration-100 cursor-default p-0"
          style={{ top: explorerContextMenu.y, left: explorerContextMenu.x }}
        >
          <button
            type="button"
            disabled={remotePath === "/" || remotePath === ""}
            onClick={() => {
              navigateUp();
              setExplorerContextMenu(null);
            }}
            className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2 disabled:opacity-50 disabled:hover:bg-transparent"
          >
            <ArrowLeft size={14} /> Back
          </button>

          <button
            type="button"
            onClick={() => {
              loadRemoteFiles(remotePath);
              setExplorerContextMenu(null);
            }}
            className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
          >
            <RefreshCw size={14} /> Refresh
          </button>

          <div className="h-px bg-mui-grey-100 dark:bg-white/5 my-1" />

          <button
            type="button"
            onClick={() => {
              openNewFileModal(remotePath);
              setExplorerContextMenu(null);
            }}
            className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
          >
            <File size={14} /> New File
          </button>

          <button
            type="button"
            onClick={() => {
              openNewFolderModal(remotePath);
              setExplorerContextMenu(null);
            }}
            className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
          >
            <Folder size={14} /> New Folder
          </button>

          <div className="h-px bg-mui-grey-100 dark:bg-white/5 my-1" />

          <button
            type="button"
            onClick={() => {
              setShowHiddenFiles(!showHiddenFiles);
              setExplorerContextMenu(null);
            }}
            className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
          >
            <span className="w-3.5 flex items-center justify-center">
              {showHiddenFiles && <Check size={14} />}
            </span>
            View hidden files/folder
          </button>
        </div>
      )}

      {/* --- Terminal Right-Click Context Menu --- */}
      {terminalContextMenu && (
        <div
          ref={contextMenuRef}
          className="fixed z-50 bg-white dark:bg-mui-grey-800 shadow-xl border border-mui-grey-200 dark:border-white/10 rounded-lg py-1 min-w-[180px] animate-in fade-in zoom-in-95 duration-100 cursor-default p-0"
          style={{ top: terminalContextMenu.y, left: terminalContextMenu.x }}
        >
          <button
            type="button"
            onClick={() => {
              handleTerminalCopy();
              setTerminalContextMenu(null);
            }}
            className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
          >
            <Copy size={14} /> Copy
          </button>

          <button
            type="button"
            onClick={() => {
              handleTerminalPaste();
              setTerminalContextMenu(null);
            }}
            className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
          >
            <Clipboard size={14} /> Paste
          </button>

          <button
            type="button"
            onClick={() => {
              handleTerminalRefresh();
              setTerminalContextMenu(null);
            }}
            className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
          >
            <RefreshCw size={14} /> Refresh
          </button>

          <div className="h-px bg-mui-grey-100 dark:bg-white/5 my-1" />

          <button
            type="button"
            onClick={() => {
              openNewFileModal(remotePath);
              setTerminalContextMenu(null);
            }}
            className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
          >
            <File size={14} /> New File
          </button>

          <button
            type="button"
            onClick={() => {
              openNewFolderModal(remotePath);
              setTerminalContextMenu(null);
            }}
            className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
          >
            <Folder size={14} /> New Folder
          </button>
        </div>
      )}

      {/* --- Create New File Modal --- */}
      {isNewFileModalOpen && (
        <div className="fixed inset-0 z-[1000] flex items-center justify-center p-4 sm:p-6">
          {/* Backdrop */}
          <button
            type="button"
            className="absolute inset-0 bg-slate-900/60 animate-in fade-in duration-300 w-full h-full border-none p-0 cursor-default focus:outline-none"
            onClick={() => setIsNewFileModalOpen(false)}
            onKeyDown={handleActionKey(() => setIsNewFileModalOpen(false))}
          />

          {/* Modal */}
          <div className="relative w-full max-w-md bg-white dark:bg-slate-900 rounded-sm shadow-2xl border border-slate-200 dark:border-white/10 overflow-hidden animate-in zoom-in-95 fade-in duration-200 text-left">
            {/* Header */}
            <div className="px-6 py-4 border-b border-slate-100 dark:border-white/5 flex items-center justify-between bg-slate-50 dark:bg-white/5">
              <h3 className="text-sm font-black text-slate-900 dark:text-white uppercase italic tracking-tighter flex items-center gap-2">
                <File size={16} className="text-blue-500" />
                Create New File
              </h3>
              <button
                type="button"
                onClick={() => setIsNewFileModalOpen(false)}
                className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors"
              >
                <X size={18} />
              </button>
            </div>

            {/* Content */}
            <div className="px-6 py-6 space-y-4">
              <div className="space-y-1.5">
                <label className="text-[10px] font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest block">
                  File Name
                </label>
                <input
                  type="text"
                  className="w-full bg-mui-grey-50 dark:bg-mui-grey-850 border border-mui-grey-200 dark:border-mui-grey-700 rounded py-2 px-3 text-xs text-mui-grey-700 dark:text-mui-grey-300 outline-none focus:border-mui-blue-500 transition-all font-mono"
                  placeholder="e.g. index"
                  value={newFileName}
                  onChange={(e) => setNewFileName(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      handleCreateFileSubmit();
                    } else if (e.key === "Escape") {
                      setIsNewFileModalOpen(false);
                    }
                  }}
                  autoFocus
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-[10px] font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest block">
                  Extension
                </label>
                <select
                  className="w-full bg-mui-grey-50 dark:bg-mui-grey-850 border border-mui-grey-200 dark:border-mui-grey-700 rounded py-2 px-3 text-xs text-mui-grey-700 dark:text-mui-grey-300 outline-none focus:border-mui-blue-500 transition-all font-sans"
                  value={selectedExtension}
                  onChange={(e) => setSelectedExtension(e.target.value)}
                >
                  <option value="">All Files (*.*)</option>
                  <option value=".txt">Text File (*.txt)</option>
                  <option value=".html">HTML Document (*.html)</option>
                  <option value=".css">CSS Stylesheet (*.css)</option>
                  <option value=".js">JavaScript Script (*.js)</option>
                  <option value=".ts">TypeScript Script (*.ts)</option>
                  <option value=".json">JSON Config (*.json)</option>
                  <option value=".php">PHP Script (*.php)</option>
                  <option value=".py">Python Script (*.py)</option>
                  <option value=".sh">Shell Script (*.sh)</option>
                  <option value=".sql">SQL Query (*.sql)</option>
                  <option value=".md">Markdown Document (*.md)</option>
                </select>
              </div>

              <p className="text-[9px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-widest leading-relaxed">
                Target Dir: <span className="font-mono text-slate-500 dark:text-slate-300 lowercase">{createTargetDir}</span>
              </p>
            </div>

            {/* Footer */}
            <div className="px-6 py-4 bg-slate-50 dark:bg-white/5 flex items-center justify-end gap-3">
              <button
                type="button"
                onClick={() => setIsNewFileModalOpen(false)}
                className="px-4 py-2 rounded-sm text-[10px] font-black uppercase tracking-widest text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white transition-all"
              >
                Cancel
              </button>
              <button
                type="button"
                data-testid="create-file-btn"
                onClick={() => handleCreateFileSubmit()}
                className="px-5 py-2 rounded-sm text-[10px] font-black uppercase tracking-widest text-white shadow-lg transition-all hover:scale-105 active:scale-95 bg-blue-600 hover:bg-blue-500 shadow-blue-500/20"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}

      {/* --- Create New Folder Modal --- */}
      {isNewFolderModalOpen && (
        <div className="fixed inset-0 z-[1000] flex items-center justify-center p-4 sm:p-6">
          {/* Backdrop */}
          <button
            type="button"
            className="absolute inset-0 bg-slate-900/60 animate-in fade-in duration-300 w-full h-full border-none p-0 cursor-default focus:outline-none"
            onClick={() => setIsNewFolderModalOpen(false)}
            onKeyDown={handleActionKey(() => setIsNewFolderModalOpen(false))}
          />

          {/* Modal */}
          <div className="relative w-full max-w-md bg-white dark:bg-slate-900 rounded-sm shadow-2xl border border-slate-200 dark:border-white/10 overflow-hidden animate-in zoom-in-95 fade-in duration-200 text-left">
            {/* Header */}
            <div className="px-6 py-4 border-b border-slate-100 dark:border-white/5 flex items-center justify-between bg-slate-50 dark:bg-white/5">
              <h3 className="text-sm font-black text-slate-900 dark:text-white uppercase italic tracking-tighter flex items-center gap-2">
                <Folder size={16} className="text-blue-500" />
                Create New Folder
              </h3>
              <button
                type="button"
                onClick={() => setIsNewFolderModalOpen(false)}
                className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors"
              >
                <X size={18} />
              </button>
            </div>

            {/* Content */}
            <div className="px-6 py-6 space-y-4">
              <div className="space-y-1.5">
                <label className="text-[10px] font-bold text-slate-500 dark:text-slate-400 uppercase tracking-widest block">
                  Folder Name
                </label>
                <input
                  type="text"
                  className="w-full bg-mui-grey-50 dark:bg-mui-grey-850 border border-mui-grey-200 dark:border-mui-grey-700 rounded py-2 px-3 text-xs text-mui-grey-700 dark:text-mui-grey-300 outline-none focus:border-mui-blue-500 transition-all font-mono"
                  placeholder="e.g. assets"
                  value={newFolderName}
                  onChange={(e) => setNewFolderName(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      handleCreateFolderSubmit();
                    } else if (e.key === "Escape") {
                      setIsNewFolderModalOpen(false);
                    }
                  }}
                  autoFocus
                />
              </div>

              <p className="text-[9px] font-bold text-slate-400 dark:text-slate-500 uppercase tracking-widest leading-relaxed">
                Target Dir: <span className="font-mono text-slate-500 dark:text-slate-300 lowercase">{createTargetDir}</span>
              </p>
            </div>

            {/* Footer */}
            <div className="px-6 py-4 bg-slate-50 dark:bg-white/5 flex items-center justify-end gap-3">
              <button
                type="button"
                onClick={() => setIsNewFolderModalOpen(false)}
                className="px-4 py-2 rounded-sm text-[10px] font-black uppercase tracking-widest text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white transition-all"
              >
                Cancel
              </button>
              <button
                type="button"
                data-testid="create-folder-btn"
                onClick={() => handleCreateFolderSubmit()}
                className="px-5 py-2 rounded-sm text-[10px] font-black uppercase tracking-widest text-white shadow-lg transition-all hover:scale-105 active:scale-95 bg-blue-600 hover:bg-blue-500 shadow-blue-500/20"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default SSHSessionView;
