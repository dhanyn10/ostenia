import React from 'react';
import { List, X } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

function LogViewer({ logs, isOpen, onClose }) {
  if (!isOpen) return null;
  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center p-8 bg-black/60 backdrop-blur-sm animate-in fade-in duration-300">
      <div className="bg-white dark:bg-[#1e293b] w-full max-w-4xl h-[70vh] rounded-sm border border-slate-200 dark:border-white/5 flex flex-col shadow-3xl overflow-hidden animate-in zoom-in-95 duration-300">
        <div className="p-5 border-b border-slate-200 dark:border-white/5 flex items-center justify-between bg-slate-50 dark:bg-white/[0.02]">
          <div className="flex items-center gap-3">
            <List size={18} className="text-blue-500 dark:text-blue-400" />
            <h3 className="font-black text-slate-900 dark:text-white uppercase italic tracking-tighter text-sm">System Activity Logs</h3>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-slate-200 dark:hover:bg-white/10 rounded-sm text-slate-400 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white transition-all">
            <X size={18} />
          </button>
        </div>
        <div className="flex-1 overflow-y-auto p-5 font-mono text-[10px] space-y-1 bg-slate-50 dark:bg-black/20">
           {logs.length === 0 && <p className="text-slate-600 italic">No logs recorded yet...</p>}
           {logs.map((log, i) => (
             <div key={i} className="flex gap-4 group">
               <span className="text-slate-400 dark:text-slate-700 select-none">[{log.time}]</span>
               <span className={cn(
                 "flex-1",
                 log.msg.includes('Error') || log.msg.includes('failed') ? "text-rose-600 dark:text-rose-400" :
                 log.msg.includes('success') || log.msg.includes('Ready') ? "text-emerald-600 dark:text-emerald-400" :
                 "text-slate-600 dark:text-slate-400"
               )}>{log.msg}</span>
             </div>
           ))}
        </div>
      </div>
    </div>
  );
}

export default LogViewer;
