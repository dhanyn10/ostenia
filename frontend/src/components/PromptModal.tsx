import React, { useState, useEffect, useRef } from "react";
import { Edit3, X } from "lucide-react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import { handleActionKey } from "../utils/a11y";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

interface PromptModalProps {
  isOpen: boolean;
  title: string;
  message: string;
  defaultValue?: string;
  placeholder?: string;
  onConfirm: (value: string) => void;
  onCancel: () => void;
  confirmText?: string;
  cancelText?: string;
}

const PromptModal: React.FC<PromptModalProps> = ({
  isOpen,
  title,
  message,
  defaultValue = "",
  placeholder = "",
  onConfirm,
  onCancel,
  confirmText = "Save",
  cancelText = "Cancel",
}) => {
  const [value, setValue] = useState(defaultValue);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isOpen) {
      setValue(defaultValue);
      setTimeout(() => {
        inputRef.current?.focus();
        inputRef.current?.select();
      }, 50);
    }
  }, [isOpen, defaultValue]);

  if (!isOpen) return null;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (value.trim()) {
      onConfirm(value);
    }
  };

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center p-4 sm:p-6">
      {/* Backdrop */}
      <button
        type="button"
        className="absolute inset-0 bg-slate-900/60 animate-in fade-in duration-300 w-full h-full border-none p-0 cursor-default focus:outline-none"
        onKeyDown={handleActionKey(onCancel)}
        onClick={onCancel}
      />

      {/* Modal */}
      <div
        className={cn(
          "relative w-full max-w-sm bg-white dark:bg-slate-900 rounded-sm shadow-2xl border border-slate-200 dark:border-white/10 overflow-hidden animate-in zoom-in-95 fade-in duration-200",
        )}
      >
        <form onSubmit={handleSubmit}>
          {/* Header */}
          <div className="px-6 py-4 border-b border-slate-100 dark:border-white/5 flex items-center justify-between">
            <h3 className="text-sm font-black text-slate-900 dark:text-white uppercase italic tracking-tighter flex items-center gap-2">
              <Edit3 size={16} className="text-blue-500" />
              {title}
            </h3>
            <button
              type="button"
              onClick={onCancel}
              className="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors"
            >
              <X size={18} />
            </button>
          </div>

          {/* Content */}
          <div className="px-6 py-6 flex flex-col gap-3">
            <label className="text-xs font-bold text-slate-500 dark:text-slate-400 leading-relaxed uppercase tracking-widest">
              {message}
            </label>
            <input
              ref={inputRef}
              type="text"
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder={placeholder}
              className={cn(
                "w-full px-3 py-2 text-xs border rounded-sm outline-none transition-all",
                "bg-slate-50 dark:bg-white/5 border-slate-200 dark:border-white/10 text-slate-900 dark:text-white",
                "focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500",
              )}
            />
          </div>

          {/* Footer */}
          <div className="px-6 py-4 bg-slate-50 dark:bg-white/5 flex items-center justify-end gap-3">
            <button
              type="button"
              onClick={onCancel}
              className="px-4 py-2 rounded-sm text-[10px] font-black uppercase tracking-widest text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white transition-all"
            >
              {cancelText}
            </button>
            <button
              type="submit"
              disabled={!value.trim()}
              className={cn(
                "px-5 py-2 rounded-sm text-[10px] font-black uppercase tracking-widest text-white shadow-lg transition-all hover:scale-105 active:scale-95",
                "bg-blue-600 hover:bg-blue-500 shadow-blue-500/20 disabled:opacity-50 disabled:pointer-events-none",
              )}
            >
              {confirmText}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default PromptModal;
