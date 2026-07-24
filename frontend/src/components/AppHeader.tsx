import React from "react";
import {
  Play,
  Square,
  Terminal as TerminalIcon,
  ChevronDown,
  Monitor,
  ExternalLink,
} from "lucide-react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { handleActionKey } from "../utils/a11y";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

function AppHeader({
  activeTab,
  handleStartAll,
  handleStopAll,
  handleTerminal,
  isTerminalOpen,
  setIsTerminalOpen,
}) {
  if (activeTab === "logs" || activeTab === "ssh") return null;

  const title =
    {
      activity: "Activity Center",
      plugins: "Plugin Management",
      proxy: "Proxy Management",
      ssh: "",
      logs: "System Activity Logs",
    }[activeTab] || "";

  return (
    <header className="h-14 shrink-0">
      <div className="max-w-4xl mx-auto h-full px-8 flex items-center justify-between">
        <div className="space-y-0.5">
          <h2 className="text-lg font-black text-slate-900 dark:text-white tracking-tight uppercase italic">
            {title}
          </h2>
        </div>

        {activeTab === "activity" && (
          <div className="flex items-center gap-2.5">
            <button
              type="button"
              onClick={handleStartAll}
              className="flex items-center gap-1.5 px-4 py-2 bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 rounded-sm transition-all text-[9px] font-black uppercase tracking-widest border border-emerald-500/20 group"
            >
              <Play
                size={12}
                fill="currentColor"
                className="group-hover:scale-110 transition-transform"
              />{" "}
              Start All
            </button>
            <button
              type="button"
              onClick={handleStopAll}
              className="flex items-center gap-1.5 px-4 py-2 bg-rose-500/10 hover:bg-rose-500/20 text-rose-600 dark:text-rose-400 rounded-sm transition-all text-[9px] font-black uppercase tracking-widest border border-rose-500/20 group"
            >
              <Square
                size={12}
                fill="currentColor"
                className="group-hover:scale-110 transition-transform"
              />{" "}
              Stop All
            </button>
            <div className="w-px h-5 bg-slate-200 dark:bg-white/10 mx-1" />

            {/* Integrated Terminal Dropdown */}
            <div className="relative">
              <button
                type="button"
                onClick={() => setIsTerminalOpen(!isTerminalOpen)}
                className={cn(
                  "flex items-center gap-1.5 p-2 rounded-sm transition-all border border-slate-200 dark:border-white/5",
                  isTerminalOpen
                    ? "bg-slate-200 dark:bg-slate-700 text-slate-900 dark:text-white"
                    : "bg-white dark:bg-slate-800/50 text-slate-400 hover:text-slate-900 dark:hover:text-white",
                )}
              >
                <TerminalIcon size={16} />
                <ChevronDown
                  size={10}
                  className={cn(
                    "transition-transform",
                    isTerminalOpen && "rotate-180",
                  )}
                />
              </button>

              {isTerminalOpen && (
                <>
                  <button
                    type="button"
                    className="fixed inset-0 z-[60] w-full h-full cursor-default bg-transparent border-none p-0 focus:outline-none"
                    onKeyDown={handleActionKey(() => setIsTerminalOpen(false))}
                    onClick={() => setIsTerminalOpen(false)}
                  />
                  <div className="absolute top-full right-0 mt-2 w-48 bg-white dark:bg-slate-900 border border-slate-200 dark:border-white/10 rounded-sm shadow-2xl z-[70] animate-in fade-in zoom-in-95 duration-200">
                    <div className="p-1">
                      <button
                        type="button"
                        onClick={() => handleTerminal("cmd")}
                        className="w-full flex items-center gap-3 px-3 py-2 rounded-sm text-[11px] font-bold text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white transition-all text-left"
                      >
                        <Monitor size={14} className="text-blue-500" /> Command
                        Prompt (CMD)
                      </button>
                      <button
                        type="button"
                        onClick={() => handleTerminal("powershell")}
                        className="w-full flex items-center gap-3 px-3 py-2 rounded-sm text-[11px] font-bold text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white transition-all text-left"
                      >
                        <Monitor size={14} className="text-blue-600" />{" "}
                        PowerShell
                      </button>
                      <button
                        type="button"
                        onClick={() => handleTerminal("gitbash")}
                        className="w-full flex items-center gap-3 px-3 py-2 rounded-sm text-[11px] font-bold text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white transition-all text-left"
                      >
                        <ExternalLink size={14} className="text-orange-500" />{" "}
                        Git Bash
                      </button>
                    </div>
                  </div>
                </>
              )}
            </div>
          </div>
        )}
      </div>
    </header>
  );
}

export default AppHeader;
