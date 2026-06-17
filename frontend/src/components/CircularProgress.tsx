import React from 'react';
import { X } from 'lucide-react';
import { handleActionKey } from '../utils/a11y';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

interface CircularProgressProps {
  percentage: number;
  status?: string;
  speed?: string;
  downloaded?: string;
  onCancel?: () => void;
  size?: number;
}

const CircularProgress: React.FC<CircularProgressProps> = ({ percentage, status, speed, downloaded, onCancel, size = 36 }) => {
  const radius = (size / 2) - 2;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference - (percentage / 100) * circumference;
  const isStreaming = percentage > 0 && percentage < 100;

  return (
    <div
      role="button"
      tabIndex={0}
      className="relative group/progress cursor-pointer overflow-hidden p-0.5 rounded-sm outline-none focus:ring-1 focus:ring-blue-500/40"
      onClick={onCancel}
      onKeyDown={onCancel ? handleActionKey(onCancel) : undefined}
    >
      <svg width={size} height={size} className={cn("transform -rotate-90", isStreaming && "animate-[spin_3s_linear_infinite]")}>
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          stroke="currentColor"
          strokeWidth="2.5"
          fill="transparent"
          className="text-slate-200 dark:text-white/10"
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          stroke="currentColor"
          strokeWidth="2.5"
          fill="transparent"
          strokeDasharray={circumference}
          style={{ strokeDashoffset: offset, transition: 'stroke-dashoffset 0.5s ease-in-out' }}
          className="text-blue-600 dark:text-blue-500"
        />
      </svg>

      <div className="absolute inset-0 flex items-center justify-center">
        <div className="group-hover/progress:hidden flex flex-col items-center">
           <span className="text-[9px] font-black text-slate-900 dark:text-white">{Math.round(percentage)}%</span>
        </div>
        <div className="hidden group-hover/progress:flex items-center justify-center bg-white dark:bg-slate-900 w-full h-full">
           <X size={14} className="text-rose-500" />
        </div>
      </div>
    </div>
  );
};

export default CircularProgress;
