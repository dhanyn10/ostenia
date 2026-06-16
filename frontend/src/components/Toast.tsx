import React from 'react';
import { XCircle, CheckCircle2, AlertCircle, X } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
 return twMerge(clsx(inputs));
}

function Toast({ toasts, removeToast }) {
 return (
 <div className="fixed bottom-6 right-6 z-[100] flex flex-col gap-2.5 w-72">
 {toasts.map((toast) => (
 <div key={toast.id} className={cn(
 "p-3 rounded-sm shadow-2xl flex items-start gap-3 transition-all duration-300 animate-in slide-in-from-right-4",
 toast.type === 'error' ? "bg-rose-50 dark:bg-rose-950/90 border border-rose-200 dark:border-rose-500/30 text-rose-800 dark:text-rose-200" :
 toast.type === 'success' ? "bg-emerald-50 dark:bg-emerald-950/90 border border-emerald-200 dark:border-emerald-500/30 text-emerald-800 dark:text-emerald-200" :
 "bg-white dark:bg-slate-900/90 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-white"
 )}>
 <div className="mt-0.5">
 {toast.type === 'error' ? <XCircle size={16} /> :
 toast.type === 'success' ? <CheckCircle2 size={16} /> :
 <AlertCircle size={16} />}
 </div>
 <div className="flex-1 space-y-0.5">
 <h5 className="font-bold text-[10px] uppercase tracking-widest">{toast.title}</h5>
 <p className="text-[10px] opacity-80">{toast.message}</p>
 </div>
 <button onClick={() => removeToast(toast.id)} className="opacity-40 hover:opacity-100 transition-opacity">
 <X size={12} />
 </button>
 </div>
 ))}
 </div>
 );
}

export default Toast;
