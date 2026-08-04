import React from "react";
import { List, Copy, Check, ArrowUp } from "lucide-react";
import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: any[]) {
  return twMerge(clsx(inputs));
}

interface CallerInfo {
  functionName: string;
  fileName: string;
  line: string;
  column: string;
}

interface LogEntry {
  id: string | number;
  time: string;
  msg: string;
  cleanMsg?: string;
  type?: string;
  isServiceLog?: boolean;
  caller?: CallerInfo;
  rawStack?: string;
}

interface LogViewerProps {
  logs: LogEntry[];
  isActive?: boolean;
}

interface CopyLogButtonProps {
  isCopied: boolean;
  onCopy: (e: React.MouseEvent) => void;
  title: string;
}

const CopyLogButton = ({ isCopied, onCopy, title }: CopyLogButtonProps) => {
  return (
    <button
      type="button"
      onClick={onCopy}
      className={cn(
        "shrink-0 p-1 rounded border transition-all",
        isCopied
          ? "text-emerald-500 border-emerald-500/30 bg-emerald-500/10"
          : "text-slate-400 border-transparent hover:border-slate-200 hover:bg-slate-100 dark:hover:border-slate-700 dark:hover:bg-slate-800"
      )}
      title={title}
    >
      {isCopied ? <Check size={12} /> : <Copy size={12} />}
    </button>
  );
};

function getPaginationRange(currentPage: number, totalPages: number, siblingCount = 1) {
  const totalPageNumbers = siblingCount + 5; // siblingCount + firstPage + lastPage + currentPage + 2*ellipses

  if (totalPageNumbers >= totalPages) {
    return Array.from({ length: totalPages }, (_, i) => i + 1);
  }

  const leftSiblingIndex = Math.max(currentPage - siblingCount, 1);
  const rightSiblingIndex = Math.min(currentPage + siblingCount, totalPages);

  const shouldShowLeftDots = leftSiblingIndex > 2;
  const shouldShowRightDots = rightSiblingIndex < totalPages - 2;

  if (!shouldShowLeftDots && shouldShowRightDots) {
    let leftItemCount = 3 + 2 * siblingCount;
    let leftRange = Array.from({ length: leftItemCount }, (_, i) => i + 1);
    return [...leftRange, '...', totalPages];
  }

  if (shouldShowLeftDots && !shouldShowRightDots) {
    let rightItemCount = 3 + 2 * siblingCount;
    let rightRange = Array.from({ length: rightItemCount }, (_, i) => totalPages - rightItemCount + i + 1);
    return [1, '...', ...rightRange];
  }

  if (shouldShowLeftDots && shouldShowRightDots) {
    let middleRange = Array.from({ length: rightSiblingIndex - leftSiblingIndex + 1 }, (_, i) => leftSiblingIndex + i);
    return [1, '...', ...middleRange, '...', totalPages];
  }

  return [];
}

function LogViewer({ logs, isActive = false }: LogViewerProps) {
  const [viewMode, setViewMode] = React.useState<"simple" | "complete">("simple");
  const [copiedId, setCopiedId] = React.useState<string | number | null>(null);

  // Pagination states
  const [currentPage, setCurrentPage] = React.useState(1);
  const [pageSize, setPageSize] = React.useState(20);

  // Buffering states for real-time logs when active
  const [displayedLogs, setDisplayedLogs] = React.useState<LogEntry[]>([]);
  const [pendingLogs, setPendingLogs] = React.useState<LogEntry[]>([]);
  const [isHovered, setIsHovered] = React.useState(false);

  const containerRef = React.useRef<HTMLDivElement>(null);

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

    const displayedIds = new Set(displayedLogs.map(l => l.id));

    if (displayedLogs.length === 0) {
      setDisplayedLogs(logs);
      setPendingLogs([]);
      return;
    }

    // Check for a complete reload or clear (no overlap between previous displayed and new logs)
    const hasOverlap = logs.some(l => displayedIds.has(l.id));
    if (!hasOverlap) {
      setDisplayedLogs(logs);
      setPendingLogs([]);
      return;
    }

    // Find new logs that are in `logs` but not in `displayedLogs`
    const newLogs = logs.filter(l => !displayedIds.has(l.id));

    if (newLogs.length > 0) {
      const pendingIds = new Set(pendingLogs.map(l => l.id));
      const actuallyNew = newLogs.filter(l => !pendingIds.has(l.id));

      if (actuallyNew.length > 0) {
        setPendingLogs(() => {
          // Keep the exact same descending order as in the logs array
          return logs.filter(l => !displayedIds.has(l.id));
        });
      }
    } else {
      // If logs decreased or some were removed, filter them accordingly
      const logIds = new Set(logs.map(l => l.id));
      const stillValidDisplayed = displayedLogs.filter(l => logIds.has(l.id));
      if (stillValidDisplayed.length !== displayedLogs.length) {
        setDisplayedLogs(stillValidDisplayed);
      }
      const stillValidPending = pendingLogs.filter(l => logIds.has(l.id));
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

  const getLogColorClass = (msg: string) => {
    if (
      msg.includes("ERR") ||
      msg.includes("Error") ||
      msg.includes("failed")
    ) {
      return "text-rose-500 dark:text-rose-400 font-bold";
    }
    if (
      msg.includes("success") ||
      msg.includes("Ready") ||
      msg.includes("Completed")
    ) {
      return "text-emerald-500 dark:text-emerald-400 font-bold";
    }
    if (msg.includes("[WRN]")) {
      return "text-amber-500 dark:text-amber-400";
    }
    return "text-slate-600 dark:text-slate-400";
  };

  const getLogLevel = (log: LogEntry): "SERVICE" | "ERROR" | "WARN" | "INFO" => {
    if (log.isServiceLog) return "SERVICE";
    const type = log.type || "";
    if (type.toLowerCase() === "error" || type.toLowerCase() === "err") return "ERROR";
    if (type.toLowerCase() === "warn" || type.toLowerCase() === "wrn") return "WARN";

    // Fallback to checking the msg
    const msg = log.msg || "";
    if (msg.includes("[ERR]") || msg.includes("Error") || msg.includes("failed")) return "ERROR";
    if (msg.includes("[WRN]")) return "WARN";
    return "INFO";
  };

  const handleCopyLog = (log: LogEntry) => {
    const level = getLogLevel(log);
    const timeStr = log.time || "";

    let message = log.cleanMsg || log.msg || "";
    if (message.startsWith("[SYS] ") || message.startsWith("[ERR] ") || message.startsWith("[WRN] ")) {
      message = message.substring(6);
    }

    const callerFunc = log.caller ? `${log.caller.functionName}()` : "N/A";
    const fileInfo = log.caller ? `${log.caller.fileName}:${log.caller.line}:${log.caller.column}` : "N/A";
    const stackTraceText = log.rawStack ? log.rawStack : "N/A";

    const copyText = [
      `Timestamp: [${timeStr}]`,
      `Level: ${level}`,
      `Message: ${message}`,
      `Caller Function: ${callerFunc}`,
      `File: ${fileInfo}`,
      `Stack Trace:\n${stackTraceText}`
    ].join("\n");

    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(copyText).then(() => {
        setCopiedId(log.id);
        setTimeout(() => {
          setCopiedId(null);
        }, 2000);
      });
    } else {
      try {
        const textArea = document.createElement("textarea");
        textArea.value = copyText;
        document.body.appendChild(textArea);
        textArea.select();
        document.execCommand("copy");
        document.body.removeChild(textArea);
        setCopiedId(log.id);
        setTimeout(() => {
          setCopiedId(null);
        }, 2000);
      } catch (err) {
        console.error("Failed to copy using fallback:", err);
      }
    }
  };

  // Slice list for pagination. Array contains newest logs at index 0.
  const totalPages = Math.ceil(displayedLogs.length / pageSize) || 1;
  const safeCurrentPage = Math.min(currentPage, totalPages);
  const startIndex = (safeCurrentPage - 1) * pageSize;
  const endIndex = startIndex + pageSize;
  const paginatedLogs = displayedLogs.slice(startIndex, endIndex);

  return (
    <div className="flex flex-col h-full animate-in fade-in slide-in-from-bottom-2 duration-500 bg-white dark:bg-[#0f172a]">
      {/* Header Area */}
      <div className="shrink-0 p-6 border-b border-slate-200 dark:border-white/5 bg-white/50 dark:bg-slate-900/40 flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div>
              <h3 className="font-black text-slate-900 dark:text-white uppercase italic tracking-tighter text-sm">
                System Activity Logs
              </h3>
              <p className="text-[9px] text-slate-400 uppercase tracking-widest font-bold">
                Real-time application monitoring
              </p>
            </div>
          </div>
        </div>

        {/* Switcher Control */}
        <div className="flex items-center gap-1 bg-slate-100 dark:bg-slate-800/60 p-1 rounded-lg self-start">
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

        {displayedLogs.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-slate-400 dark:text-slate-600 gap-2 opacity-50">
            <List size={32} strokeWidth={1} />
            <p className="text-[10px] font-bold uppercase tracking-widest italic">
              No activity recorded yet...
            </p>
          </div>
        ) : viewMode === "simple" ? (
          /* Simple Log Mode: Newest on top (rendered using standard top-to-bottom flex-col) */
          <div className="flex flex-col min-h-full space-y-1.5">
            {paginatedLogs.map((log) => {
              const isCopied = copiedId === log.id;
              return (
                <div
                  key={log.id}
                  className="flex items-center justify-between gap-4 group py-1 px-2 border border-transparent hover:border-slate-100 dark:hover:border-white/5 hover:bg-slate-50 dark:hover:bg-slate-900/40 rounded transition-all"
                >
                  <div className="flex gap-4 items-baseline min-w-0 flex-1">
                    <span className="text-slate-400 dark:text-slate-600 select-none shrink-0 w-20">
                      [{log.time}]
                    </span>
                    <span
                      className={cn(
                        "flex-1 break-all leading-relaxed min-w-0",
                        getLogColorClass(log.msg)
                      )}
                    >
                      {log.msg}
                    </span>
                  </div>
                  <CopyLogButton
                    isCopied={isCopied}
                    onCopy={(e) => {
                      e.preventDefault();
                      handleCopyLog(log);
                    }}
                    title="Copy Log"
                  />
                </div>
              );
            })}
          </div>
        ) : (
          /* Complete Log (Laravel-Style) Mode: Newest on top (rendered using standard top-to-bottom flex-col) */
          <div className="flex flex-col min-h-full space-y-3">
            {paginatedLogs.map((log) => {
              const isCopied = copiedId === log.id;
              const level = getLogLevel(log);
              const isService = log.isServiceLog;

              const badgeClasses = {
                SERVICE: "bg-blue-100 text-blue-800 dark:bg-blue-950/40 dark:text-blue-300 border border-blue-200 dark:border-blue-800/50",
                ERROR: "bg-rose-100 text-rose-800 dark:bg-rose-950/40 dark:text-rose-300 border border-rose-200 dark:border-rose-800/50 font-bold",
                WARN: "bg-amber-100 text-amber-800 dark:bg-amber-950/40 dark:text-amber-300 border border-amber-200 dark:border-amber-800/50 font-bold",
                INFO: "bg-slate-100 text-slate-800 dark:bg-slate-800/60 dark:text-slate-300 border border-slate-200 dark:border-slate-700/50"
              };

              const currentBadgeClass = badgeClasses[level] || badgeClasses.INFO;

              return (
                <div
                  key={log.id}
                  className="p-4 rounded-lg border text-[10px] transition-all shadow-sm flex flex-col gap-3 bg-slate-50/50 dark:bg-slate-900/20 border-slate-100 dark:border-slate-800 hover:border-slate-200 dark:hover:border-slate-700"
                >
                  {/* Top line with metadata */}
                  <div className="flex items-center justify-between gap-4">
                    <div className="flex flex-wrap items-center gap-2">
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

                    <CopyLogButton
                      isCopied={isCopied}
                      onCopy={(e) => {
                        e.preventDefault();
                        handleCopyLog(log);
                      }}
                      title="Copy Laravel-Style Log"
                    />
                  </div>

                  {/* Message details */}
                  <div className={cn(
                    "font-semibold leading-relaxed break-all text-xs",
                    isService ? "text-slate-700 dark:text-slate-300 font-bold" : getLogColorClass(log.msg)
                  )}>
                    {log.cleanMsg || log.msg}
                  </div>

                  {/* Expandable Stack Trace */}
                  {!isService && log.rawStack && (
                    <details className="group/details text-[10px] text-slate-500 dark:text-slate-400">
                      <summary className="cursor-pointer font-bold select-none hover:text-slate-800 dark:hover:text-slate-200 flex items-center gap-1 outline-none">
                        <span className="inline-block transition-transform duration-100 group-open/details:rotate-90">▶</span>
                        View Stack Trace
                      </summary>
                      <pre className="bg-black text-rose-400/90 dark:text-rose-300 p-4 rounded-md overflow-x-auto text-[9px] mt-2 font-mono border border-slate-800 leading-normal whitespace-pre">
                        {log.rawStack}
                      </pre>
                    </details>
                  )}
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Pagination Controls */}
      {displayedLogs.length > 0 && (
        <div className="shrink-0 p-4 border-t border-slate-200 dark:border-white/5 bg-slate-50 dark:bg-slate-900/40 flex flex-wrap gap-4 items-center justify-between text-[11px] font-sans">
          <div className="text-slate-500 dark:text-slate-400 font-bold">
            Showing <span className="text-slate-800 dark:text-slate-200">{Math.min(startIndex + 1, displayedLogs.length)}</span> to{" "}
            <span className="text-slate-800 dark:text-slate-200">{Math.min(endIndex, displayedLogs.length)}</span> of{" "}
            <span className="text-slate-800 dark:text-slate-200">{displayedLogs.length}</span> entries
          </div>

          <div className="flex items-center gap-4">
            {/* Page Size Selector */}
            <div className="flex items-center gap-1.5">
              <span className="text-slate-400 font-bold">Show:</span>
              <select
                id="log-page-size"
                value={pageSize}
                onChange={(e) => {
                  setPageSize(Number(e.target.value));
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
              {getPaginationRange(safeCurrentPage, totalPages).map((page, idx) => {
                if (page === '...') {
                  return (
                    <span
                      key={`dots-${idx}`}
                      className="px-2 py-1 text-slate-400 dark:text-slate-600 font-bold select-none"
                    >
                      ...
                    </span>
                  );
                }
                const isSelected = page === safeCurrentPage;
                return (
                  <button
                    key={`page-${page}`}
                    type="button"
                    onClick={() => setCurrentPage(Number(page))}
                    className={cn(
                      "px-2.5 py-1.5 rounded border font-bold font-mono text-[10px] transition-all",
                      isSelected
                        ? "bg-blue-600 border-blue-600 text-white shadow-sm"
                        : "border-slate-200 dark:border-white/10 bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:border-slate-300 dark:hover:border-white/20"
                    )}
                  >
                    {page}
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
