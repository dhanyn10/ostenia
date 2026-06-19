import React, { useState, useEffect, useRef } from 'react';
import { Circle, ChevronUp, Server } from 'lucide-react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { handleActionKey } from '../utils/a11y';

function cn(...inputs: ClassValue[]) {
 return twMerge(clsx(inputs));
}

const StatusBar = ({ services }) => {
 const [isDropdownOpen, setIsDropdownOpen] = useState(false);
 const activeServices = services.filter(s => s.status === 'Running');
 const dropdownRef = useRef(null);

 useEffect(() => {
 const handleClickOutside = (event) => {
 if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
 setIsDropdownOpen(false);
 }
 };
 if (isDropdownOpen) {
 document.addEventListener('mousedown', handleClickOutside);
 }
 return () => document.removeEventListener('mousedown', handleClickOutside);
 }, [isDropdownOpen]);

 return (
 <div className="h-6 flex items-center justify-between bg-mui-blue-600 text-white px-3 select-none text-[11px] font-medium relative z-[50]">
 <div className="flex items-center h-full gap-4">
 <div
 ref={dropdownRef}
 className="relative h-full"
 >
 <button
 role="button" tabIndex={0} onKeyDown={handleActionKey(() => setIsDropdownOpen(!isDropdownOpen))} onClick={() => setIsDropdownOpen(!isDropdownOpen)}
 className={cn(
 "flex items-center gap-1.5 px-2 h-full transition-colors",
 isDropdownOpen ? "bg-mui-blue-700" : "hover:bg-white/10"
 )}
 >
 <Circle size={10} className={cn(activeServices.length > 0 ? "fill-white" : "text-white/50")} />
 <span>{activeServices.length} Services Running</span>
 <ChevronUp size={10} className={cn("transition-transform duration-200", isDropdownOpen && "rotate-180")} />
 </button>

 {isDropdownOpen && (
 <div className="absolute bottom-full left-0 w-64 bg-white dark:bg-mui-dark-paper border border-mui-grey-200 dark:border-white/10 shadow-2xl animate-in fade-in slide-in-from-bottom-2 duration-200 rounded-t-sm overflow-hidden">
 <div className="px-3 py-2 border-b border-mui-grey-100 dark:border-white/5 bg-mui-grey-50 dark:bg-white/5 flex items-center gap-2">
 <Server size={12} className="text-mui-blue-500" />
 <span className="font-bold text-[10px] uppercase tracking-wider text-mui-grey-500 dark:text-mui-grey-400">Environment Services</span>
 </div>
 <div className="max-h-64 overflow-y-auto py-1 scrollbar-thin scrollbar-thumb-mui-grey-300 dark:scrollbar-thumb-white/10">
 {services.map(service => (
 <div
 key={service.name}
 className="flex items-center justify-between px-3 py-2 hover:bg-mui-grey-100 dark:hover:bg-white/5 transition-colors group"
 >
 <div className="flex items-center gap-2.5">
 <div className={cn(
 "w-2 h-2 rounded-full",
 service.status === 'Running' ? "bg-emerald-500 animate-pulse" : "bg-mui-grey-300 dark:bg-white/10"
 )} />
 <span className="text-mui-grey-700 dark:text-mui-grey-200">{service.name}</span>
 </div>
 <div className="flex items-center gap-2 text-[9px]">
 {service.port > 0 && (
 <span className="px-1.5 py-0.5 rounded-sm bg-mui-blue-500/10 text-mui-blue-600 dark:text-mui-blue-400 font-bold">
 :{service.port}
 </span>
 )}
 <span className={cn(
 "uppercase font-black tracking-tighter",
 service.status === 'Running' ? "text-emerald-600 dark:text-emerald-400" : "text-mui-grey-400"
 )}>
 {service.status}
 </span>
 </div>
 </div>
 ))}
 </div>
 </div>
 )}
 </div>
 </div>

 <div className="flex items-center h-full gap-3">
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
