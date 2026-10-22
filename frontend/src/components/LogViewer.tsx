import React from "react";
import { List, Copy as CopyIcon, Check, ArrowUp, CheckSquare, X, FolderOpen, Trash2 } from "lucide-react";
import * as AppBackend from "@/wailsjs/go/backend/App";
import { cn } from "@/utils/utils";


import type {
  CallerInfo,
  LogId,
  LogLevel,
  LogViewMode,
  LogEntry,
  PaginationItem,
} from "@/types/logs";
import {
  getPaginationRange,
  getLogLevel,
  formatCompleteLog,
} from "@/utils/logFormatters";

export { getLogLevel, formatCompleteLog, getPaginationRange };
export type {
  CallerInfo,
  LogId,
  LogLevel,
  LogViewMode,
  LogEntry,
  PaginationItem,
};

interface LogViewerProps {
  readonly logs: LogEntry[];
  readonly isActive?: boolean;
  readonly onClearLogs?: () => void;
}

interface CopyLogButtonProps {
  readonly isCopied: boolean;
  readonly onCopy: (event: React.MouseEvent) => void;
  readonly title: string;
}

interface LogCheckboxProps {
  readonly isSelected: boolean;
  readonly onToggleSelect: (isShiftKey: boolean) => void;
  readonly className?: string;
}

interface LogRowWrapperProps {
  readonly id?: string;
  readonly log: LogEntry;
  readonly isSelected: boolean;
  readonly isSelectMode: boolean;
  readonly onRowContextMenu: (event: React.MouseEvent, log: LogEntry) => void;
  readonly onToggleSelect: (id: LogId, isShiftKey: boolean) => void;
  readonly onClick?: (event: React.MouseEvent, log: LogEntry) => void;
  readonly className: string;
  readonly children: React.ReactNode;
}


const CopyLogButton = ({ isCopied, onCopy, title }: Readonly<CopyLogButtonProps>) => {
  return (
    <button
      type="button"
      onClick={(event) => {
        event.stopPropagation();
        onCopy(event);
      }}
      className={cn(
        "shrink-0 p-1 rounded border transition-all relative",
        isCopied
          ? "text-emerald-500 border-emerald-500/30 bg-emerald-500/10"
          : "text-slate-400 border-transparent hover:border-slate-200 hover:bg-slate-100 dark:hover:border-slate-700 dark:hover:bg-slate-800"
      )}
      title={title}
    >
      {isCopied ? <Check size={12} /> : <CopyIcon size={12} />}
    </button>
  );
};

const LogCheckbox = ({ isSelected, onToggleSelect, className }: Readonly<LogCheckboxProps>) => {
  return (
    <input
      type="checkbox"
      checked={isSelected}
      onClick={(event) => event.stopPropagation()}
      onChange={(event) => {
        const nativeEvent = event.nativeEvent as MouseEvent;
        onToggleSelect(nativeEvent.shiftKey);
      }}
      className={cn("shrink-0 accent-blue-600 rounded cursor-pointer", className)}
    />
  );
};

const LogRowWrapper = ({
  id,
  log,
  isSelectMode,
  onRowContextMenu,
  onToggleSelect,
  className,
  children,
}: Readonly<LogRowWrapperProps>) => {
  if (isSelectMode) {
    return (
      <button
        id={id}
        type="button"
        onContextMenu={(event) => onRowContextMenu(event, log)}
        onClick={(event) => onToggleSelect(log.id, event.shiftKey)}
        className={cn(className, "w-full text-left")}
      >
        {children}
      </button>
    );
  }

  return (
    <div
      id={id}
      onContextMenu={(event) => onRowContextMenu(event, log)}
      className={className}
    >
      {children}
    </div>
  );
};


function LogViewer({ logs, isActive = false, onClearLogs }: Readonly<LogViewerProps>) {
  const [viewMode, setViewMode] = React.useState<LogViewMode>("simple");
  const [copiedId, setCopiedId] = React.useState<LogId | null>(null);
  const [showBatchCopiedTooltip, setShowBatchCopiedTooltip] = React.useState(false);

  // Selection states (Gmail style with Shift+Click)
  const [isSelectMode, setIsSelectMode] = React.useState(false);
  const [selectedIds, setSelectedIds] = React.useState<Set<LogId>>(new Set());
  const lastSelectedIdRef = React.useRef<LogId | null>(null);

  // Context menu state & ref
  const [contextMenu, setContextMenu] = React.useState<{ x: number; y: number; targetLog: LogEntry } | null>(null);
  const menuRef = React.useRef<HTMLDivElement>(null);

  // Pagination states
  const [currentPage, setCurrentPage] = React.useState(1);
  const [pageSize, setPageSize] = React.useState(20);
  const [highlightedLogId, setHighlightedLogId] = React.useState<LogId | null>(null);

  const handleSimpleLogClick = (log: LogEntry) => {
    const logIndex = displayedLogs.findIndex(logEntry => logEntry.id === log.id);
    if (logIndex !== -1) {
      const targetPage = Math.floor(logIndex / pageSize) + 1;
      setCurrentPage(targetPage);
    }
    setViewMode("complete");
    setHighlightedLogId(log.id);
    setTimeout(() => {
      const el = document.getElementById(`log-card-${log.id}`);
      if (el) {
        el.scrollIntoView({ behavior: "smooth", block: "center" });
      }
    }, 50);
    setTimeout(() => {
      setHighlightedLogId(prev => (prev === log.id ? null : prev));
    }, 2500);
  };

  // Buffering states for real-time logs when active
  const [displayedLogs, setDisplayedLogs] = React.useState<LogEntry[]>([]);
  const [pendingLogs, setPendingLogs] = React.useState<LogEntry[]>([]);
  const [isHovered, setIsHovered] = React.useState(false);

  const containerRef = React.useRef<HTMLDivElement>(null);

  // Close context menu on outside click or scroll
  React.useEffect(() => {
    const handleCloseMenu = (event: MouseEvent) => {
      if (menuRef.current?.contains(event.target as Node)) {
        return;
      }
      setContextMenu(null);
    };
    window.addEventListener("click", handleCloseMenu);
    window.addEventListener("scroll", handleCloseMenu, true);
    return () => {
      window.removeEventListener("click", handleCloseMenu);
      window.removeEventListener("scroll", handleCloseMenu, true);
    };
  }, []);

  // Sync displayedLogs with incoming logs, buffering if active
  React.useEffect(() => {
    if (logs.length === 0) {
      setDisplayedLogs([]);
      setPendingLogs([]);
      return;
    }

    if (!isActive) {
      setDisplayedLogs(logs);
      setPendingLogs([]);
      return;
    }

    const displayedIds = new Set(displayedLogs.map(logEntry => logEntry.id));

    if (displayedLogs.length === 0) {
      setDisplayedLogs(logs);
      setPendingLogs([]);
      return;
    }

    // Check for a complete reload or clear (no overlap between previous displayed and new logs)
    const hasOverlap = logs.some(logEntry => displayedIds.has(logEntry.id));
    if (!hasOverlap) {
      setDisplayedLogs(logs);
      setPendingLogs([]);
      return;
    }

    // Find new logs that are in `logs` but not in `displayedLogs`
    const newLogs = logs.filter(logEntry => !displayedIds.has(logEntry.id));

    if (newLogs.length > 0) {
      const pendingIds = new Set(pendingLogs.map(logEntry => logEntry.id));
      const actuallyNew = newLogs.filter(logEntry => !pendingIds.has(logEntry.id));

      if (actuallyNew.length > 0) {
        setPendingLogs(() => {
          // Keep the exact same descending order as in the logs array
          return logs.filter(logEntry => !displayedIds.has(logEntry.id));
        });
      }
    } else {
      // If logs decreased or some were removed, filter them accordingly
      const logIds = new Set(logs.map(logEntry => logEntry.id));
      const stillValidDisplayed = displayedLogs.filter(logEntry => logIds.has(logEntry.id));
      if (stillValidDisplayed.length !== displayedLogs.length) {
        setDisplayedLogs(stillValidDisplayed);
      }
      const stillValidPending = pendingLogs.filter(logEntry => logIds.has(logEntry.id));
      if (stillValidPending.length !== pendingLogs.length) {
        setPendingLogs(stillValidPending);
      }
    }
  }, [logs, isActive]);

  const handleReleasePending = () => {
    setDisplayedLogs(logs);
    setPendingLogs([]);
    setCurrentPage(1);
    if (containerRef.current) {
      if (typeof containerRef.current.scrollTo === "function") {
        containerRef.current.scrollTo({ top: 0, behavior: "smooth" });
      } else {
        containerRef.current.scrollTop = 0;
      }
    }
  };

  const getSimpleLogColorClass = (type?: string, isServiceLog?: boolean) => {
    if (isServiceLog) {
      return "text-blue-600 dark:text-blue-400 font-medium";
    }
    const lowerType = (type || "").toLowerCase();
    if (lowerType === "error" || lowerType === "err") {
      return "text-rose-600 dark:text-rose-400 font-medium";
    }
    if (lowerType === "warn" || lowerType === "wrn" || lowerType === "warning") {
      return "text-amber-600 dark:text-amber-400 font-medium";
    }
    if (lowerType === "success") {
      return "text-emerald-600 dark:text-emerald-400 font-medium";
    }
    return "text-slate-700 dark:text-slate-300";
  };

  const getLogColorClass = (msg: string, type?: string) => {
    const lowerMsg = (msg || "").toLowerCase();
    const lowerType = (type || "").toLowerCase();

    if (
      lowerType === "error" ||
      lowerType === "err" ||
      lowerMsg.includes("err") ||
      lowerMsg.includes("error") ||
      lowerMsg.includes("failed") ||
      lowerMsg.includes("crash") ||
      lowerMsg.includes("fatal") ||
      lowerMsg.includes("panic") ||
      lowerMsg.includes("exception")
    ) {
      return "text-rose-500 dark:text-rose-400 font-bold";
    }
    if (
      lowerMsg.includes("success") ||
      lowerMsg.includes("ready") ||
      lowerMsg.includes("completed")
    ) {
      return "text-emerald-500 dark:text-emerald-400 font-bold";
    }
    if (lowerType === "warn" || lowerType === "wrn" || lowerMsg.includes("[wrn]")) {
      return "text-amber-500 dark:text-amber-400 font-bold";
    }
    return "text-slate-600 dark:text-slate-400";
  };

  const copyToClipboardText = (text: string, callback?: () => void) => {
    if (navigator.clipboard?.writeText) {
      void navigator.clipboard.writeText(text).then(() => {
        if (callback) callback();
      }).catch((err) => {
        console.error("Failed to copy text:", err);
      });
    } else {
      try {
        const textArea = document.createElement("textarea");
        textArea.value = text;
        document.body.appendChild(textArea);
        textArea.select();
        (document as any).execCommand("copy");
        textArea.remove();
        if (callback) callback();
      } catch (err) {
        console.error("Failed to copy using fallback:", err);
      }
    }
  };

  const handleCopyLog = (log: LogEntry) => {
    const copyText = formatCompleteLog(log);

    copyToClipboardText(copyText, () => {
      setCopiedId(log.id);
      setTimeout(() => {
        setCopiedId(null);
      }, 2000);
    });
  };

  // Slice list for pagination. Array contains newest logs at index 0.
  const totalPages = Math.ceil(displayedLogs.length / pageSize) || 1;
  const safeCurrentPage = Math.min(currentPage, totalPages);
  const startIndex = (safeCurrentPage - 1) * pageSize;
  const endIndex = startIndex + pageSize;
  const paginatedLogs = displayedLogs.slice(startIndex, endIndex);

  // Selection Helper Methods with Shift+Click Range Selection support
  const toggleSelectLog = (id: LogId, isShiftKey: boolean = false) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);

      if (isShiftKey && lastSelectedIdRef.current !== null) {
        const lastId = lastSelectedIdRef.current;
        const lastIndex = paginatedLogs.findIndex((logEntry) => logEntry.id === lastId);
        const currentIndex = paginatedLogs.findIndex((logEntry) => logEntry.id === id);

        if (lastIndex !== -1 && currentIndex !== -1) {
          const start = Math.min(lastIndex, currentIndex);
          const end = Math.max(lastIndex, currentIndex);
          const rangeLogs = paginatedLogs.slice(start, end + 1);

          rangeLogs.forEach((logEntry) => next.add(logEntry.id));
          lastSelectedIdRef.current = id;
          return next;
        }
      }

      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }

      lastSelectedIdRef.current = id;
      return next;
    });
  };

  const isCurrentPageAllSelected = paginatedLogs.length > 0 && paginatedLogs.every(log => selectedIds.has(log.id));

  const handleSelectAllCurrentPage = () => {
    if (isCurrentPageAllSelected) {
      // Unselect all on current page
      setSelectedIds((prev) => {
        const next = new Set(prev);
        paginatedLogs.forEach(log => next.delete(log.id));
        return next;
      });
    } else {
      // Select all on current page
      setSelectedIds((prev) => {
        const next = new Set(prev);
        paginatedLogs.forEach(log => next.add(log.id));
        return next;
      });
    }
  };

  const handleCancelSelection = () => {
    setSelectedIds(new Set());
    setIsSelectMode(false);
    setContextMenu(null);
    lastSelectedIdRef.current = null;
  };

  const copyFormattedSelectedLogs = () => {
    // Preserve log order as displayed on current page / list
    const selectedLogs = displayedLogs.filter(log => selectedIds.has(log.id));
    if (selectedLogs.length === 0) return;

    // Format all selected logs as Complete Log version
    const copyText = selectedLogs.map(log => formatCompleteLog(log)).join("\n\n---\n\n");

    copyToClipboardText(copyText, () => {
      setContextMenu(null);
      setShowBatchCopiedTooltip(true);
      setTimeout(() => {
        setShowBatchCopiedTooltip(false);
      }, 2000);
    });
  };

  const handleRowContextMenu = (event: React.MouseEvent, log: LogEntry) => {
    event.preventDefault();
    event.stopPropagation();
    if (!isSelectMode) {
      setContextMenu({ x: event.clientX, y: event.clientY, targetLog: log });
    } else {
      // If current log is not selected yet, select it
      if (!selectedIds.has(log.id)) {
        setSelectedIds(prev => new Set(prev).add(log.id));
        lastSelectedIdRef.current = log.id;
      }
      setContextMenu({ x: event.clientX, y: event.clientY, targetLog: log });
    }
  };

  const handleContextMenuSelectLogOption = (event: React.MouseEvent) => {
    event.stopPropagation();
    if (contextMenu) {
      const targetId = contextMenu.targetLog.id;
      lastSelectedIdRef.current = targetId;
      setSelectedIds(new Set([targetId]));
      setIsSelectMode(true);
      setContextMenu(null);
    }
  };

  const renderLogContent = () => {
    if (displayedLogs.length === 0) {
      return (
        <div className="h-full flex flex-col items-center justify-center text-slate-400 dark:text-slate-600 gap-2 opacity-50">
          <List size={32} strokeWidth={1} />
          <p className="text-[10px] font-bold uppercase tracking-widest italic">
            No activity recorded yet...
          </p>
        </div>
      );
    }

    if (viewMode === "simple") {
      return (
        <div className="flex flex-col min-h-full space-y-1">
          {paginatedLogs.map((log) => {
            const isCopied = copiedId === log.id;
            const isSelected = selectedIds.has(log.id);
            const cleanText = log.cleanMsg || log.msg || "";
            const colorClass = getSimpleLogColorClass(log.type, log.isServiceLog);

            return (
              <LogRowWrapper
                key={log.id}
                id={`simple-log-${log.id}`}
                log={log}
                isSelected={isSelected}
                isSelectMode={isSelectMode}
                onRowContextMenu={handleRowContextMenu}
                onToggleSelect={toggleSelectLog}
                className={cn(
                  "flex items-center justify-between gap-3 group h-7 py-1 px-3 border rounded transition-all select-none overflow-hidden",
                  isSelected
                    ? "bg-blue-50/80 dark:bg-blue-950/30 border-blue-200 dark:border-blue-800/50"
                    : "border-transparent hover:border-slate-200 dark:hover:border-white/10 hover:bg-slate-100/70 dark:hover:bg-slate-800/50"
                )}
              >
                {isSelectMode ? (
                  <div className="flex items-center gap-2.5 min-w-0 flex-1 overflow-hidden">
                    <LogCheckbox
                      isSelected={isSelected}
                      onToggleSelect={(isShift) => toggleSelectLog(log.id, isShift)}
                      className="self-center shrink-0"
                    />
                    <span className="text-slate-400 dark:text-slate-500 font-mono text-[9px] select-none shrink-0">
                      [{log.time}]
                    </span>
                    <span
                      title={cleanText}
                      className={cn(
                        "truncate whitespace-nowrap overflow-hidden text-ellipsis text-[10px] min-w-0 flex-1 font-mono",
                        colorClass
                      )}
                    >
                      {cleanText}
                    </span>
                  </div>
                ) : (
                  <>
                    <button
                      type="button"
                      onClick={() => handleSimpleLogClick(log)}
                      className="flex items-center gap-2.5 min-w-0 flex-1 overflow-hidden bg-transparent border-none p-0 text-left cursor-pointer outline-none focus:ring-1 focus:ring-blue-500/40 rounded-sm"
                    >
                      <span className="text-slate-400 dark:text-slate-500 font-mono text-[9px] select-none shrink-0">
                        [{log.time}]
                      </span>
                      <span
                        title={cleanText}
                        className={cn(
                          "truncate whitespace-nowrap overflow-hidden text-ellipsis text-[10px] min-w-0 flex-1 font-mono",
                          colorClass
                        )}
                      >
                        {cleanText}
                      </span>
                    </button>
                    <CopyLogButton
                      isCopied={isCopied}
                      onCopy={(event) => {
                        event.preventDefault();
                        handleCopyLog(log);
                      }}
                      title="Copy Log"
                    />
                  </>
                )}
              </LogRowWrapper>
            );
          })}
        </div>
      );
    }

    return (
      <div className="flex flex-col min-h-full space-y-3">
        {paginatedLogs.map((log) => {
          const isCopied = copiedId === log.id;
          const isSelected = selectedIds.has(log.id);
          const level = getLogLevel(log);
          const isService = log.isServiceLog;

          const badgeClasses = {
            SERVICE: "bg-blue-100 text-blue-800 dark:bg-blue-950/40 dark:text-blue-300 border border-blue-200 dark:border-blue-800/50",
            ERROR: "bg-rose-100 text-rose-800 dark:bg-rose-950/40 dark:text-rose-300 border border-rose-200 dark:border-rose-800/50 font-bold",
            WARN: "bg-amber-100 text-amber-800 dark:bg-amber-950/40 dark:text-amber-300 border border-amber-200 dark:border-amber-800/50 font-bold",
            INFO: "bg-slate-100 text-slate-800 dark:bg-slate-800/60 dark:text-slate-300 border border-slate-200 dark:border-slate-700/50"
          };

          const currentBadgeClass = badgeClasses[level] || badgeClasses.INFO;

          let cardHighlightClass =
            "bg-slate-50/50 dark:bg-slate-900/20 border-slate-100 dark:border-slate-800 hover:border-slate-200 dark:hover:border-slate-700";
          if (highlightedLogId === log.id) {
            cardHighlightClass =
              "ring-2 ring-blue-500 border-blue-400 dark:border-blue-500 bg-blue-50/70 dark:bg-blue-950/40 shadow-md";
          } else if (isSelected) {
            cardHighlightClass =
              "bg-blue-50/80 dark:bg-blue-950/30 border-blue-200 dark:border-blue-800/50";
          }

          return (
            <LogRowWrapper
              key={log.id}
              id={`log-card-${log.id}`}
              log={log}
              isSelected={isSelected}
              isSelectMode={isSelectMode}
              onRowContextMenu={handleRowContextMenu}
              onToggleSelect={toggleSelectLog}
              className={cn(
                "p-4 rounded-lg border text-[10px] transition-all shadow-sm flex flex-col gap-3 select-none",
                cardHighlightClass
              )}
            >
              {/* Top line with metadata */}
              <div className="flex items-center justify-between gap-4">
                <div className="flex flex-wrap items-center gap-2">
                  {isSelectMode && (
                    <LogCheckbox
                      isSelected={isSelected}
                      onToggleSelect={(isShift) => toggleSelectLog(log.id, isShift)}
                      className="mr-1"
                    />
                  )}
                  <span className="text-slate-400 dark:text-slate-500 font-bold tracking-tight">
                    [{log.time}]
                  </span>
                  <span className={cn("px-2 py-0.5 rounded text-[8px] font-bold uppercase tracking-wider", currentBadgeClass)}>
                    {level}
                  </span>
                  {!isService && log.caller && (
                    <div className="flex flex-wrap items-center gap-1 text-slate-500 dark:text-slate-400 text-[9px] bg-slate-100 dark:bg-slate-800/40 px-2 py-0.5 rounded">
                      <span className="font-semibold text-slate-700 dark:text-slate-300">
                        {log.caller.functionName}()
                      </span>
                      <span className="text-slate-400 dark:text-slate-500">
                        in
                      </span>
                      <span className="font-mono text-slate-600 dark:text-slate-400 font-bold">
                        {log.caller.fileName}:{log.caller.line}:{log.caller.column}
                      </span>
                    </div>
                  )}
                </div>

                {!isSelectMode && (
                  <CopyLogButton
                    isCopied={isCopied}
                    onCopy={(event) => {
                      event.preventDefault();
                      handleCopyLog(log);
                    }}
                    title="Copy Laravel-Style Log"
                  />
                )}
              </div>

              {/* Message details */}
              <div className={cn(
                "font-semibold leading-relaxed break-all text-xs",
                getLogColorClass(log.cleanMsg || log.msg, log.type)
              )}>
                {log.cleanMsg || log.msg}
              </div>

              {/* Expandable Stack Trace */}
              {!isService && log.rawStack && (
                <details className="group/details text-[10px] text-slate-500 dark:text-slate-400">
                  <summary className="cursor-pointer font-bold select-none hover:text-slate-800 dark:hover:text-slate-200 flex items-center gap-1 outline-none">
                    <span className="inline-block transition-transform duration-100 group-open/details:rotate-90">â–¶</span>
                    {" "}View Stack Trace
                  </summary>
                  <pre className="bg-black text-rose-400/90 dark:text-rose-300 p-4 rounded-md overflow-x-auto text-[9px] mt-2 font-mono border border-slate-800 leading-normal whitespace-pre">
                    {log.rawStack}
                  </pre>
                </details>
              )}
            </LogRowWrapper>
          );
        })}
      </div>
    );
  };

  return (
    <div className="flex flex-col h-full animate-in fade-in slide-in-from-bottom-2 duration-500 bg-white dark:bg-[#0f172a] relative">
      {/* Custom Context Menu */}
      {contextMenu && (
        <div
          ref={menuRef}
          style={{ top: `${contextMenu.y}px`, left: `${contextMenu.x}px` }}
          className="fixed z-50 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md shadow-xl py-1 text-[11px] font-sans min-w-[140px] animate-in fade-in duration-100"
        >
          {isSelectMode ? (
            <button
              type="button"
              onClick={(event) => {
                event.stopPropagation();
                copyFormattedSelectedLogs();
              }}
              disabled={selectedIds.size === 0}
              className="w-full text-left px-3 py-1.5 flex items-center gap-2 hover:bg-slate-100 dark:hover:bg-slate-700/60 text-slate-700 dark:text-slate-200 disabled:opacity-40 disabled:cursor-not-allowed font-semibold"
            >
              <CopyIcon size={13} />
              <span>Copy ({selectedIds.size})</span>
            </button>
          ) : (
            <button
              type="button"
              onClick={handleContextMenuSelectLogOption}
              className="w-full text-left px-3 py-1.5 flex items-center gap-2 hover:bg-slate-100 dark:hover:bg-slate-700/60 text-slate-700 dark:text-slate-200 font-semibold"
            >
              <CheckSquare size={13} />
              <span>Select Log</span>
            </button>
          )}

          <button
            type="button"
            onClick={(event) => {
              event.stopPropagation();
              handleCancelSelection();
            }}
            className="w-full text-left px-3 py-1.5 flex items-center gap-2 hover:bg-slate-100 dark:hover:bg-slate-700/60 text-rose-600 dark:text-rose-400 font-semibold"
          >
            <X size={13} />
            <span>Cancel</span>
          </button>
        </div>
      )}

      {/* Header Area */}
      <div className="shrink-0 p-6 border-b border-slate-200 dark:border-white/5 bg-white/50 dark:bg-slate-900/40 flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="font-black text-slate-900 dark:text-white uppercase italic tracking-tighter text-sm">
              System Activity Logs
            </h3>
            <p className="text-[9px] text-slate-400 uppercase tracking-widest font-bold">
              Real-time application monitoring
            </p>
          </div>

          {/* Select Button Toggle */}
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={() => {
                if (AppBackend && (AppBackend as any).OpenLogsFolder) {
                  try {
                    const promise = (AppBackend as any).OpenLogsFolder();
                    if (promise && typeof promise.catch === "function") {
                      promise.catch(() => {
                        // Ignored: opening folder failed
                      });
                    }
                  } catch {
                    // Ignored: failure invoking backend open logs folder
                  }
                }
              }}
              title="Open logs folder in AppData"
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-[10px] font-bold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 transition-all uppercase tracking-wider border border-slate-200 dark:border-white/10"
            >
              <FolderOpen size={13} />
              <span>Open Folder</span>
            </button>
            {onClearLogs && (
              <button
                type="button"
                onClick={onClearLogs}
                title="Clear all logs"
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-[10px] font-bold bg-slate-100 hover:bg-rose-100 dark:bg-slate-800 dark:hover:bg-rose-950/40 text-slate-700 hover:text-rose-600 dark:text-slate-300 dark:hover:text-rose-400 transition-all uppercase tracking-wider border border-slate-200 dark:border-white/10 hover:border-rose-300 dark:hover:border-rose-800/40"
              >
                <Trash2 size={13} />
                <span>Clear</span>
              </button>
            )}
            {!isSelectMode ? (
              <button
                type="button"
                onClick={() => setIsSelectMode(true)}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-[10px] font-bold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 transition-all uppercase tracking-wider border border-slate-200 dark:border-white/10"
              >
                <CheckSquare size={13} />
                <span>Select Log</span>
              </button>
            ) : (
              <button
                type="button"
                onClick={handleCancelSelection}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-[10px] font-bold bg-rose-50 hover:bg-rose-100 dark:bg-rose-950/40 dark:hover:bg-rose-900/50 text-rose-600 dark:text-rose-300 transition-all uppercase tracking-wider border border-rose-200 dark:border-rose-800/40"
              >
                <X size={13} />
                <span>Cancel Selection</span>
              </button>
            )}
          </div>
        </div>

        {/* Action / Selection bar or Switcher Control */}
        <div className="flex items-center justify-between gap-4">
          <div className="flex items-center gap-1 bg-slate-100 dark:bg-slate-800/60 p-1 rounded-lg">
            <button
              type="button"
              onClick={() => {
                setViewMode("simple");
                setCurrentPage(1);
              }}
              className={cn(
                "px-3 py-1.5 rounded-md text-[10px] font-bold transition-all uppercase tracking-wider",
                viewMode === "simple"
                  ? "bg-white dark:bg-slate-700 text-slate-900 dark:text-white shadow-sm"
                  : "text-slate-500 dark:text-slate-400 hover:text-slate-950 dark:hover:text-white"
              )}
            >
              Simple Log
            </button>
            <button
              type="button"
              onClick={() => {
                setViewMode("complete");
                setCurrentPage(1);
              }}
              className={cn(
                "px-3 py-1.5 rounded-md text-[10px] font-bold transition-all uppercase tracking-wider",
                viewMode === "complete"
                  ? "bg-white dark:bg-slate-700 text-slate-900 dark:text-white shadow-sm"
                  : "text-slate-500 dark:text-slate-400 hover:text-slate-950 dark:hover:text-white"
              )}
            >
              Complete Log
            </button>
          </div>

          {/* Selection Controls when Selection Mode active */}
          {isSelectMode && (
            <div className="flex items-center gap-3 bg-blue-50 dark:bg-blue-950/30 px-3 py-1.5 rounded-lg border border-blue-200 dark:border-blue-800/40 text-[10px] relative">
              <label className="flex items-center gap-1.5 cursor-pointer font-bold text-slate-700 dark:text-slate-300 select-none">
                <input
                  type="checkbox"
                  checked={isCurrentPageAllSelected}
                  onChange={handleSelectAllCurrentPage}
                  className="accent-blue-600 rounded cursor-pointer"
                />
                <span>Select All (Current Page)</span>
              </label>

              <span className="text-slate-300 dark:text-slate-700">|</span>

              <span className="font-bold text-blue-600 dark:text-blue-400">
                {selectedIds.size} selected
              </span>

              <div className="relative inline-flex items-center">
                <button
                  type="button"
                  onClick={copyFormattedSelectedLogs}
                  disabled={selectedIds.size === 0}
                  className="flex items-center gap-1 px-2.5 py-1 bg-blue-600 hover:bg-blue-700 disabled:opacity-40 disabled:cursor-not-allowed text-white font-bold rounded transition-all"
                >
                  <CopyIcon size={12} />
                  <span>Copy</span>
                </button>
                {showBatchCopiedTooltip && (
                  <span className="absolute bottom-full mb-1.5 left-1/2 -translate-x-1/2 px-2 py-0.5 bg-slate-900 text-emerald-400 text-[9px] font-bold rounded shadow-lg border border-emerald-500/30 whitespace-nowrap animate-in fade-in zoom-in-95 duration-150">
                    Copied!
                  </span>
                )}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Logs Content Area */}
      <div
        ref={containerRef}
        className="flex-1 overflow-y-auto p-6 font-mono text-[10px] space-y-3 scrollbar-thin scrollbar-thumb-slate-200 dark:scrollbar-thumb-white/5 relative"
      >
        {/* Floating Pending Logs Badge */}
        {pendingLogs.length > 0 && (
          <div className="sticky top-4 z-20 flex justify-center w-full h-0 overflow-visible animate-in fade-in slide-in-from-top-2 duration-300">
            <button
              type="button"
              onMouseEnter={() => setIsHovered(true)}
              onMouseLeave={() => setIsHovered(false)}
              onClick={handleReleasePending}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white dark:bg-blue-500 dark:text-white rounded-full shadow-lg hover:shadow-xl hover:bg-blue-700 dark:hover:bg-blue-600 transition-all duration-200 cursor-pointer font-sans text-[11px] font-black uppercase tracking-wider border border-blue-500/30 shrink-0"
            >
              <ArrowUp size={12} className="animate-bounce shrink-0" />
              <span>
                {isHovered ? `show ${pendingLogs.length} new logs` : `${pendingLogs.length} new logs`}
              </span>
            </button>
          </div>
        )}

        {renderLogContent()}
      </div>

      {/* Pagination Controls */}
      {displayedLogs.length > 0 && (
        <div className="shrink-0 p-4 border-t border-slate-200 dark:border-white/5 bg-slate-50 dark:bg-slate-900/40 flex flex-wrap gap-4 items-center justify-between text-[11px] font-sans">
          <div className="text-slate-500 dark:text-slate-400 font-bold">
            Showing{" "}
            <span className="text-slate-800 dark:text-slate-200">
              {Math.min(startIndex + 1, displayedLogs.length)}
            </span>{" "}
            to{" "}
            <span className="text-slate-800 dark:text-slate-200">
              {Math.min(endIndex, displayedLogs.length)}
            </span>{" "}
            of{" "}
            <span className="text-slate-800 dark:text-slate-200">
              {displayedLogs.length}
            </span>{" "}
            entries
          </div>

          <div className="flex items-center gap-4">
            {/* Page Size Selector */}
            <div className="flex items-center gap-1.5">
              <span className="text-slate-400 font-bold">Show:</span>
              <select
                id="log-page-size"
                value={pageSize}
                onChange={(event) => {
                  setPageSize(Number(event.target.value));
                  setCurrentPage(1);
                }}
                className="bg-white dark:bg-slate-800 border border-slate-200 dark:border-white/10 rounded px-1.5 py-1 text-slate-700 dark:text-slate-300 font-bold outline-none cursor-pointer"
              >
                <option value={10}>10</option>
                <option value={20}>20</option>
                <option value={50}>50</option>
                <option value={100}>100</option>
              </select>
            </div>

            {/* Navigation Buttons */}
            <div className="flex items-center gap-1">
              <button
                type="button"
                disabled={safeCurrentPage === 1}
                onClick={() => setCurrentPage(prev => Math.max(prev - 1, 1))}
                className="px-3 py-1.5 rounded border border-slate-200 dark:border-white/10 bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white disabled:opacity-40 disabled:cursor-not-allowed font-bold"
              >
                Previous
              </button>
              {getPaginationRange(safeCurrentPage, totalPages).map((item) => {
                if (item.isDots) {
                  return (
                    <span
                      key={item.key}
                      className="px-2 py-1 text-slate-400 dark:text-slate-600 font-bold select-none"
                    >
                      ...
                    </span>
                  );
                }
                const pageNum = item.value as number;
                const isSelected = pageNum === safeCurrentPage;
                return (
                  <button
                    key={item.key}
                    type="button"
                    onClick={() => setCurrentPage(pageNum)}
                    className={cn(
                      "px-2.5 py-1.5 rounded border font-bold font-mono text-[10px] transition-all",
                      isSelected
                        ? "bg-blue-600 border-blue-600 text-white shadow-sm"
                        : "border-slate-200 dark:border-white/10 bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:border-slate-300 dark:hover:border-white/20"
                    )}
                  >
                    {pageNum}
                  </button>
                );
              })}
              <button
                type="button"
                disabled={safeCurrentPage === totalPages}
                onClick={() => setCurrentPage(prev => Math.min(prev + 1, totalPages))}
                className="px-3 py-1.5 rounded border border-slate-200 dark:border-white/10 bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white disabled:opacity-40 disabled:cursor-not-allowed font-bold"
              >
                Next
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

export default LogViewer;
