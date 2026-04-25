import React from 'react';
import { X, Loader2 } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

function CircularProgress({ percentage, status, speed, downloaded, onCancel }) {
  const radius = 12;
  const circumference = 2 * Math.PI * radius;
  const strokeDashoffset = circumference - (percentage / 100) * circumference;
  const isStreaming = status?.includes('Streaming');

  return (
    <div className="relative group/progress cursor-pointer overflow-hidden p-0.5 rounded-sm" onClick={onCancel}>
      <div className="flex items-center gap-3 group-hover/progress:opacity-0 transition-opacity duration-200">
        <div className="text-right">
          <p className="text-[9px] font-black text-slate-800 dark:text-white">{downloaded || '...'}</p>
          <p className="text-[8px] font-bold text-slate-500 uppercase tracking-widest leading-none">{speed || '...'}</p>
        </div>
        <div className="relative w-8 h-8">
          <svg className="w-full h-full -rotate-90">
            <circle className="text-slate-200 dark:text-white/5" strokeWidth="2.5" stroke="currentColor" fill="transparent" r={radius} cx="16" cy="16" />
            <circle
              className={cn("text-blue-500 transition-all duration-500", isStreaming && "animate-[spin_2s_linear_infinite]")}
              strokeWidth="2.5"
              strokeDasharray={circumference}
              strokeDashoffset={isStreaming ? circumference * 0.7 : strokeDashoffset}
              strokeLinecap="round"
              stroke="currentColor"
              fill="transparent"
              r={radius}
              cx="16"
              cy="16"
            />
          </svg>
          <div className="absolute inset-0 flex items-center justify-center text-[7px] font-black text-blue-400">
             {isStreaming ? <Loader2 size={8} className="animate-spin" /> : `${Math.round(percentage)}%`}
          </div>
        </div>
      </div>
      <div className="absolute inset-0 flex items-center justify-center translate-y-full group-hover/progress:translate-y-0 transition-transform duration-200 bg-rose-600 rounded-sm">
         <div className="flex items-center gap-1.5 px-4">
            <X size={12} className="text-white" />
            <span className="text-[9px] font-black text-white uppercase tracking-widest">Cancel</span>
         </div>
      </div>
    </div>
  );
}

export default CircularProgress;
