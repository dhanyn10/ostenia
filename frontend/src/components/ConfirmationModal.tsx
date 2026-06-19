import React from 'react';
import { AlertTriangle, X } from 'lucide-react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { handleActionKey } from '../utils/a11y';

function cn(...inputs: ClassValue[]) {
 return twMerge(clsx(inputs));
}

function ConfirmationModal({ isOpen, title, message, onConfirm, onCancel, confirmText = "Confirm", cancelText = "Cancel", type = "danger" }) {
 if (!isOpen) return null;

 return (
 <div className="fixed inset-0 z-[1000] flex items-center justify-center p-4 sm:p-6">
 {/* Backdrop */}
 <button
 className="absolute inset-0 bg-slate-900/60 animate-in fade-in duration-300 w-full h-full border-none p-0 cursor-default focus:outline-none"
 onKeyDown={handleActionKey(onCancel)} onClick={onCancel}
 />

 {/* Modal */}
 <div className={cn(
 "relative w-full max-w-md bg-white dark:bg-slate-900 rounded-sm shadow-2xl border border-slate-200 dark:border-white/10 overflow-hidden animate-in zoom-in-95 fade-in duration-200",
 )}>
 {/* Header */}
 <div className="px-6 py-4 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
 <h3 className="text-sm font-black text-slate-900 dark:text-white uppercase italic tracking-tighter flex items-center gap-2">
 <AlertTriangle size={16} className={type === 'danger' ? 'text-rose-500' : 'text-blue-500'} />
 {title}
 </h3>
 <button
 onClick={onCancel}
 className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors"
 >
 <X size={18} />
 </button>
 </div>

 {/* Content */}
 <div className="px-6 py-6">
 <p className="text-xs font-bold text-slate-500 dark:text-slate-400 leading-relaxed uppercase tracking-widest">
 {message}
 </p>
 </div>

 {/* Footer */}
 <div className="px-6 py-4 bg-slate-50 dark:bg-white/5 flex items-center justify-end gap-3">
 <button
 onClick={onCancel}
 className="px-4 py-2 rounded-sm text-[10px] font-black uppercase tracking-widest text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white transition-all"
 >
 {cancelText}
 </button>
 <button
 onClick={onConfirm}
 className={cn(
 "px-5 py-2 rounded-sm text-[10px] font-black uppercase tracking-widest text-white shadow-lg transition-all hover:scale-105 active:scale-95",
 type === 'danger' ? "bg-rose-600 hover:bg-rose-500 shadow-rose-500/20" : "bg-blue-600 hover:bg-blue-500 shadow-blue-500/20"
 )}
 >
 {confirmText}
 </button>
 </div>
 </div>
 </div>
 );
}

export default ConfirmationModal;
