import React from 'react';
import PropTypes from 'prop-types';
import { List } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
 return twMerge(clsx(inputs));
}

function LogViewer({ logs }) {
 return (
 <div className="flex flex-col h-full animate-in fade-in slide-in-from-bottom-2 duration-500 bg-white dark:bg-[#0f172a]">
 {/* Header Area */}
 <div className="shrink-0 p-6 flex items-center justify-between border-b border-slate-200 dark:border-white/5 bg-white/50 dark:bg-slate-900/40 ">
 <div className="flex items-center gap-3">
 {/* Icon container removed */}
 <div>
 <h3 className="font-black text-slate-900 dark:text-white uppercase italic tracking-tighter text-sm">System Activity Logs</h3>
 <p className="text-[9px] text-slate-400 uppercase tracking-widest font-bold">Real-time application monitoring</p>
 </div>
 </div>
 </div>

 {/* Logs Content Area */}
 <div className="flex-1 overflow-y-auto p-6 font-mono text-[10px] space-y-1.5 scrollbar-thin scrollbar-thumb-slate-200 dark:scrollbar-thumb-white/5">
 {logs.length === 0 ? (
 <div className="h-full flex flex-col items-center justify-center text-slate-400 dark:text-slate-600 gap-2 opacity-50">
 <List size={32} strokeWidth={1} />
 <p className="text-[10px] font-bold uppercase tracking-widest italic">No activity recorded yet...</p>
 </div>
 ) : (
 <div className="flex flex-col-reverse justify-end min-h-full">
 {logs.map((log) => (
 <div key={`${log.time}-${log.msg}`} className="flex gap-4 group py-0.5 border-b border-transparent hover:border-slate-100 dark:hover:border-white/5 transition-all">
 <span className="text-slate-400 dark:text-slate-600 select-none shrink-0 w-20">[{log.time}]</span>
 <span className={cn(
 "flex-1 break-all leading-relaxed",
 log.msg.includes('ERR') || log.msg.includes('Error') || log.msg.includes('failed') ? "text-rose-500 dark:text-rose-400 font-bold" :
 log.msg.includes('success') || log.msg.includes('Ready') || log.msg.includes('Completed') ? "text-emerald-500 dark:text-emerald-400 font-bold" :
 log.msg.includes('[WRN]') ? "text-amber-500 dark:text-amber-400" :
 "text-slate-600 dark:text-slate-400"
 )}>{log.msg}</span>
 </div>
 ))}
 </div>
 )}
 </div>
 </div>
 );
}

LogViewer.propTypes = {
    logs: PropTypes.arrayOf(PropTypes.shape({
        time: PropTypes.string.isRequired,
        msg: PropTypes.string.isRequired,
    })).isRequired,
};

export default LogViewer;
