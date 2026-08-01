import React from "react";
import { List, Copy, Check } from "lucide-react";
import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: any[]) {
  return twMerge(clsx(inputs));
}

function LogViewer({ logs }: { logs: any[] }) {
  const [logMode, setLogMode] = React.useState<"simple" | "complete">("simple");
  const [copiedId, setCopiedId] = React.useState<string | null>(null);

  const handleCopyLog = (log: any) => {
    let copyText = "";
    if (log.caller) {
      copyText = `Timestamp: [${log.time}]
Level: ${(log.type || "info").toUpperCase()}
Message: ${log.rawMsg || log.msg}
Caller Function: ${log.caller.functionName}()
File: ${log.caller.fileName}:${log.caller.line}:${log.caller.column}
Stack Trace:
${log.caller.rawStack || "N/A"}`;
    } else {
      copyText = `Timestamp: [${log.time}]
Message: ${log.msg}`;
    }

    navigator.clipboard.writeText(copyText).then(() => {
      setCopiedId(log.id);
      setTimeout(() => setCopiedId(null), 2000);
    });
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

  return (
    <div className="flex flex-col h-full animate-in fade-in slide-in-from-bottom-2 duration-500 bg-white dark:bg-[#0f172a]">
      {/* Header Area */}
      <div className="shrink-0 p-6 flex flex-col gap-4 border-b border-slate-200 dark:border-white/5 bg-white/50 dark:bg-slate-900/40 ">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="font-black text-slate-900 dark:text-white uppercase italic tracking-tighter text-sm">
              System Activity Logs
            </h3>
            <p className="text-[9px] text-slate-400 uppercase tracking-widest font-bold">
              Real-time application monitoring
            </p>
          </div>
        </div>

        {/* Log Mode Switcher (di bawah judul header) */}
        <div className="flex items-center gap-2">
          <button
            type="button"
            data-testid="simple-log-btn"
            onClick={() => setLogMode("simple")}
            className={cn(
              "px-3 py-1.5 text-[10px] font-bold uppercase tracking-wider rounded transition-all border",
              logMode === "simple"
                ? "bg-indigo-600 border-indigo-600 text-white shadow-sm shadow-indigo-600/20"
                : "bg-transparent border-slate-200 dark:border-white/10 text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-white/5"
            )}
          >
            Simple Log
          </button>
          <button
            type="button"
            data-testid="complete-log-btn"
            onClick={() => setLogMode("complete")}
            className={cn(
              "px-3 py-1.5 text-[10px] font-bold uppercase tracking-wider rounded transition-all border",
              logMode === "complete"
                ? "bg-indigo-600 border-indigo-600 text-white shadow-sm shadow-indigo-600/20"
                : "bg-transparent border-slate-200 dark:border-white/10 text-slate-600 dark:text-slate-400 hover:bg-slate-50 dark:hover:bg-white/5"
            )}
          >
            Complete Log
          </button>
        </div>
      </div>

      {/* Logs Content Area */}
      <div className="flex-1 overflow-y-auto p-6 font-mono text-[10px] space-y-3 scrollbar-thin scrollbar-thumb-slate-200 dark:scrollbar-thumb-white/5">
        {logs.length === 0 ? (
          <div className="h-full flex flex-col items-center justify-center text-slate-400 dark:text-slate-600 gap-2 opacity-50">
            <List size={32} strokeWidth={1} />
            <p className="text-[10px] font-bold uppercase tracking-widest italic">
              No activity recorded yet...
            </p>
          </div>
        ) : (
          <div className="flex flex-col-reverse justify-end min-h-full space-y-3 space-y-reverse">
            {logs.map((log) => {
              const isError =
                log.type === "error" ||
                log.msg.includes("ERR") ||
                log.msg.includes("Error") ||
                log.msg.includes("failed");
              const isWarn = log.type === "warn" || log.msg.includes("[WRN]");

              if (logMode === "simple") {
                // Simple Log View (original view with copy button on hover)
                return (
                  <div
                    key={log.id}
                    className="flex gap-4 group py-1 border-b border-transparent hover:border-slate-100 dark:hover:border-white/5 transition-all justify-between items-center"
                  >
                    <div className="flex gap-4 items-center flex-1 min-w-0">
                      <span className="text-slate-400 dark:text-slate-600 select-none shrink-0 w-20">
                        [{log.time}]
                      </span>
                      <span
                        className={cn(
                          "flex-1 break-all leading-relaxed",
                          getLogColorClass(log.msg)
                        )}
                      >
                        {log.msg}
                      </span>
                    </div>

                    <button
                      type="button"
                      onClick={() => handleCopyLog(log)}
                      title="Copy log details"
                      className="opacity-0 group-hover:opacity-100 focus:opacity-100 p-1 rounded hover:bg-slate-100 dark:hover:bg-white/5 text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 transition-all"
                    >
                      {copiedId === log.id ? (
                        <Check size={12} className="text-emerald-500" />
                      ) : (
                        <Copy size={12} />
                      )}
                    </button>
                  </div>
                );
              } else {
                // Complete Log View (Laravel Option C)
                if (log.caller) {
                  return (
                    <div
                      key={log.id}
                      className={cn(
                        "p-4 rounded-lg border flex flex-col gap-2 transition-all relative group shadow-sm",
                        isError
                          ? "bg-rose-50/10 dark:bg-rose-950/5 border-rose-100 dark:border-rose-950/20"
                          : isWarn
                          ? "bg-amber-50/10 dark:bg-amber-950/5 border-amber-100 dark:border-amber-950/20"
                          : "bg-slate-50/20 dark:bg-slate-900/10 border-slate-100 dark:border-white/5"
                      )}
                    >
                      {/* Top metadata line */}
                      <div className="flex items-center justify-between text-[10px]">
                        <div className="flex items-center gap-2">
                          <span className="text-slate-400 dark:text-slate-500 font-bold">
                            {log.time}
                          </span>
                          <span
                            className={cn(
                              "px-1.5 py-0.5 rounded text-[8px] font-black uppercase tracking-wider",
                              isError
                                ? "bg-rose-100 dark:bg-rose-950/50 text-rose-600 dark:text-rose-400"
                                : isWarn
                                ? "bg-amber-100 dark:bg-amber-950/50 text-amber-600 dark:text-amber-400"
                                : "bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400"
                            )}
                          >
                            {(log.type || "INFO").toUpperCase()}
                          </span>
                          <span className="text-slate-400 dark:text-slate-500 italic max-w-xs truncate">
                            {log.caller.fileName}:{log.caller.line}
                          </span>
                        </div>

                        {/* Copy button */}
                        <button
                          type="button"
                          onClick={() => handleCopyLog(log)}
                          title="Copy detailed log"
                          className="opacity-0 group-hover:opacity-100 focus:opacity-100 p-1 rounded hover:bg-slate-100 dark:hover:bg-white/5 text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 transition-all"
                        >
                          {copiedId === log.id ? (
                            <Check size={12} className="text-emerald-500" />
                          ) : (
                            <Copy size={12} />
                          )}
                        </button>
                      </div>

                      {/* Main Message */}
                      <div
                        className={cn(
                          "text-[11px] font-bold break-all leading-relaxed",
                          isError
                            ? "text-rose-600 dark:text-rose-400"
                            : isWarn
                            ? "text-amber-600 dark:text-amber-400"
                            : "text-slate-700 dark:text-slate-300"
                        )}
                      >
                        {log.rawMsg || log.msg}
                      </div>

                      {/* Caller details */}
                      <div className="text-[9px] text-slate-500 dark:text-slate-400 bg-slate-100/30 dark:bg-slate-900/30 p-2.5 rounded border border-slate-200/40 dark:border-white/5 flex flex-col gap-1 select-text">
                        <div>
                          <span className="font-bold text-indigo-500 dark:text-indigo-400">
                            Caller Function:
                          </span>{" "}
                          <span className="font-mono">
                            {log.caller.functionName}()
                          </span>
                        </div>
                        <div>
                          <span className="font-bold text-indigo-500 dark:text-indigo-400">
                            Location:
                          </span>{" "}
                          <span className="font-mono">
                            {log.caller.fileName}:{log.caller.line}:
                            {log.caller.column}
                          </span>
                        </div>

                        {log.caller.rawStack && (
                          <details className="mt-1 cursor-pointer">
                            <summary className="text-[8px] uppercase tracking-wider font-black text-slate-400 hover:text-slate-500 dark:hover:text-slate-300 select-none">
                              View Stack Trace
                            </summary>
                            <pre className="mt-1.5 p-2 bg-slate-950 text-slate-300 font-mono text-[8px] overflow-x-auto rounded leading-normal max-h-40 overflow-y-auto select-text whitespace-pre">
                              {log.caller.rawStack}
                            </pre>
                          </details>
                        )}
                      </div>
                    </div>
                  );
                } else {
                  // Service logs (render in unified card but with simple layout)
                  return (
                    <div
                      key={log.id}
                      className={cn(
                        "p-3 rounded-lg border flex flex-col gap-1.5 transition-all relative group shadow-sm",
                        isError
                          ? "bg-rose-50/10 dark:bg-rose-950/5 border-rose-100 dark:border-rose-950/20"
                          : isWarn
                          ? "bg-amber-50/10 dark:bg-amber-950/5 border-amber-100 dark:border-amber-950/20"
                          : "bg-slate-50/20 dark:bg-slate-900/10 border-slate-100 dark:border-white/5"
                      )}
                    >
                      <div className="flex items-center justify-between text-[10px]">
                        <div className="flex items-center gap-2">
                          <span className="text-slate-400 dark:text-slate-500 font-bold">
                            {log.time}
                          </span>
                          <span className="px-1.5 py-0.5 rounded text-[8px] font-black uppercase tracking-wider bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                            SERVICE
                          </span>
                        </div>

                        {/* Copy button */}
                        <button
                          type="button"
                          onClick={() => handleCopyLog(log)}
                          title="Copy service log"
                          className="opacity-0 group-hover:opacity-100 focus:opacity-100 p-1 rounded hover:bg-slate-100 dark:hover:bg-white/5 text-slate-400 dark:text-slate-500 hover:text-slate-600 dark:hover:text-slate-300 transition-all"
                        >
                          {copiedId === log.id ? (
                            <Check size={12} className="text-emerald-500" />
                          ) : (
                            <Copy size={12} />
                          )}
                        </button>
                      </div>

                      <div
                        className={cn(
                          "text-[10px] break-all leading-relaxed font-mono",
                          getLogColorClass(log.msg)
                        )}
                      >
                        {log.msg}
                      </div>
                    </div>
                  );
                }
              }
            })}
          </div>
        )}
      </div>
    </div>
  );
}

export default LogViewer;
