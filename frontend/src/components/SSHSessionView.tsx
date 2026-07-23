import React, { useState, useEffect, useRef } from "react";
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
} from "lucide-react";
import SSHToolbar from "./ssh/SSHToolbar";
import SSHFileExplorer from "./ssh/SSHFileExplorer";
import { handleActionKey } from "../utils/a11y";

interface SSHSessionViewProps {
  session: any;
  onClose: () => void;
  addToast: (
    title: string,
    message: string,
    type?: "info" | "success" | "warn" | "error",
  ) => void;
  isActive: boolean;
  theme?: string;
  onOpenSettings: (category: string) => void;
}

const SSHSessionView: React.FC<SSHSessionViewProps> = ({
  session,
  onClose,
  addToast,
  isActive,
  theme,
  onOpenSettings,
}) => {
  const terminalRef = useRef<HTMLDivElement>(null);
  const xterm = useRef<XTerm | null>(null);
  const fitAddon = useRef(new FitAddon());
  const contextMenuRef = useRef<HTMLDivElement>(null);
  const currentPathRef = useRef("");
  const [connecting, setConnecting] = useState(true);
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

  useEffect(() => {
    setEditingPath(remotePath);
  }, [remotePath]);

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

  const handleTerminalCopy = async () => {
    if (xterm.current) {
      const selection = xterm.current.getSelection();
      if (selection) {
        await navigator.clipboard.writeText(selection);
        addToast("SSH", "Text copied to clipboard", "success");
      }
    }
  };

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

  useEffect(() => {
    if (isActive) {
      setTimeout(performFit, 100);
      setTimeout(performFit, 600);
      setTimeout(performFit, 1500);
    }
  }, [isActive]);

  useEffect(() => {
    xterm.current = new XTerm({
      cursorBlink: true,
      convertEol: true,
      fontSize: 14,
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
    if (terminalRef.current) xterm.current.open(terminalRef.current);

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

    xterm.current.onData((data) => AppBackend.SendSSHInput(session.id, data));
    xterm.current.onTitleChange((title) => {
      if (title.includes(":")) {
        const parts = title.split(":");
        const potentialPath = parts[parts.length - 1].trim();
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
          if (event.path !== prev) loadRemoteFiles(event.path);
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

  const connectSSH = async () => {
    if (!xterm.current) return;
    setConnecting(true);
    xterm.current.write(`Connecting to ${session.host}...\r\n`);
    try {
      await AppBackend.ConnectSSH(session);
      setConnecting(false);
      xterm.current.write("\x1b[32mConnected successfully.\x1b[0m\r\n\r\n");
      performFit();
      loadRemoteFiles("");
    } catch (err) {
      setConnecting(false);
      xterm.current.write(`\x1b[31mConnection failed: ${err}\x1b[0m\r\n`);
      addToast("Error", "SSH connection failed: " + err, "error");
    }
  };

  const loadRemoteFiles = async (path: string, isManualEntry = false) => {
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
      if (isManualEntry) {
        addToast("Navigation", "Directory not available", "error");
        setEditingPath(remotePath);
      } else {
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
      }
    } finally {
      setLoadingFiles(false);
    }
  };

  const syncExplorer = async (forcedPath: string | null = null) => {
    try {
      let current =
        forcedPath || (await AppBackend.GetRemoteCurrentPath(session.id));
      if (!current) return;
      let normalized = current.trim();
      if (normalized.length > 1 && normalized.endsWith("/"))
        normalized = normalized.substring(0, normalized.length - 1);
      if (normalized !== currentPathRef.current) loadRemoteFiles(normalized);
    } catch (e) {}
  };

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

  const handleUpload = async () => {
    try {
      await AppBackend.UploadRemoteFile(session.id, remotePath);
      addToast("Success", "File uploaded successfully", "success");
      loadRemoteFiles(remotePath);
    } catch (err: any) {
      addToast("Error", "Upload failed: " + err, "error");
    }
  };

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
        onFit={performFit}
        onReconnect={connectSSH}
        onClose={onClose}
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
            onNewFolder={async () => {
              const name = prompt("Folder Name:");
              if (name) {
                try {
                  await AppBackend.ExecuteSFTPAction(
                    session.id,
                    "mkdir",
                    `${remotePath}/${name}`,
                    "",
                  );
                  loadRemoteFiles(remotePath);
                } catch (e: any) {
                  addToast("Error", e.toString(), "error");
                }
              }
            }}
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
                <span className="text-mui-grey-600 dark:text-mui-grey-400 text-xs font-bold uppercase tracking-widest">
                  Connecting...
                </span>
              </div>
            </div>
          )}
        </div>
      </div>

      {fileContextMenu && (
        <div
          ref={contextMenuRef}
          className="fixed z-50 bg-white dark:bg-mui-grey-800 shadow-xl border border-mui-grey-200 dark:border-white/10 rounded-lg py-1 min-w-[140px] animate-in fade-in zoom-in-95 duration-100 cursor-default p-0"
          style={{ top: fileContextMenu.y, left: fileContextMenu.x }}
        >
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
        </div>
      )}
    </div>
  );
};

export default SSHSessionView;
