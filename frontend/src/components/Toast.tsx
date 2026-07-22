import React, { useState } from "react";
import { XCircle, CheckCircle2, AlertCircle, X, Copy, Check } from "lucide-react";
import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

function getToastClasses(type: string) {
  if (type === "error") {
    return "bg-rose-50 dark:bg-rose-950/90 border border-rose-200 dark:border-rose-500/30 text-rose-800 dark:text-rose-200";
  }
  if (type === "success") {
    return "bg-emerald-50 dark:bg-emerald-950/90 border border-emerald-200 dark:border-emerald-500/30 text-emerald-800 dark:text-emerald-200";
  }
  return "bg-white dark:bg-slate-900/90 border border-slate-200 dark:border-white/10 text-slate-900 dark:text-white";
}

function getToastIcon(type: string) {
  if (type === "error") {
    return <XCircle size={16} />;
  }
  if (type === "success") {
    return <CheckCircle2 size={16} />;
  }
  return <AlertCircle size={16} />;
}

function Toast({ toasts, removeToast }) {
  const [copiedId, setCopiedId] = useState<string | null>(null);

  const handleCopy = (id: string, text: string) => {
    navigator.clipboard.writeText(text);
    setCopiedId(id);
    setTimeout(() => {
      setCopiedId(null);
    }, 2000);
  };

  return (
    <div className="fixed bottom-6 right-6 z-[100] flex flex-col gap-2.5 w-72">
      {toasts.map((toast) => {
        const toastClasses = getToastClasses(toast.type);
        const toastIcon = getToastIcon(toast.type);
        return (
          <div
            key={toast.id}
            className={cn(
              "p-3 rounded-sm shadow-2xl flex items-start gap-3 transition-all duration-300 animate-in slide-in-from-right-4",
              toastClasses,
            )}
          >
            <div className="mt-0.5">{toastIcon}</div>
            <div className="flex-1 space-y-0.5 select-text cursor-text">
              <h5 className="font-bold text-[10px] uppercase tracking-widest select-text cursor-text">
                {toast.title}
              </h5>
              <p className="text-[10px] opacity-80 select-text cursor-text break-words">
                {toast.message}
              </p>
            </div>
            <div className="flex flex-col gap-1.5 self-stretch justify-between items-end">
              <button
                type="button"
                onClick={() => removeToast(toast.id)}
                className="opacity-40 hover:opacity-100 transition-opacity"
              >
                <X size={12} />
              </button>
              <button
                type="button"
                onClick={() => handleCopy(toast.id, toast.message)}
                title="Copy message"
                className="opacity-40 hover:opacity-100 transition-opacity"
              >
                {copiedId === toast.id ? (
                  <Check size={12} className="text-emerald-500" />
                ) : (
                  <Copy size={12} />
                )}
              </button>
            </div>
          </div>
        );
      })}
    </div>
  );
}

export default Toast;
