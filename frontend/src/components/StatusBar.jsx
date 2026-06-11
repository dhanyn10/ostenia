import React from 'react';
import { Play, Square, Circle } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

const StatusBar = ({ services }) => {
  const activeServices = services.filter(s => s.status === 'Running');

  return (
    <div className="h-6 flex items-center justify-between bg-mui-blue-600 text-white px-3 select-none text-[11px] font-medium">
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-1.5 hover:bg-white/10 px-1.5 h-full cursor-default transition-colors">
          <Circle size={10} className={cn(activeServices.length > 0 ? "fill-white" : "text-white/50")} />
          <span>{activeServices.length} Services Running</span>
        </div>

        <div className="flex items-center gap-3">
          {services.map(service => (
            <div
              key={service.name}
              className={cn(
                "flex items-center gap-1.5 px-1.5 h-full cursor-default transition-colors",
                service.status === 'Running' ? "text-white" : "text-white/40"
              )}
            >
              <div className={cn(
                "w-1.5 h-1.5 rounded-full",
                service.status === 'Running' ? "bg-emerald-400 animate-pulse" : "bg-white/20"
              )} />
              <span>{service.name}</span>
              {service.port > 0 && <span className="text-[9px] opacity-70">:{service.port}</span>}
            </div>
          ))}
        </div>
      </div>

      <div className="flex items-center gap-3">
        <div className="hover:bg-white/10 px-2 h-full flex items-center transition-colors cursor-default">
          UTF-8
        </div>
        <div className="hover:bg-white/10 px-2 h-full flex items-center transition-colors cursor-default">
          Ostenia Environment
        </div>
      </div>
    </div>
  );
};

export default StatusBar;
