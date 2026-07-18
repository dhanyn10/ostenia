import React from 'react';
import { ChevronDown, FolderPlus, MoreVertical } from 'lucide-react';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { handleActionKey } from '../utils/a11y';

function cn(...inputs: ClassValue[]) {
 return twMerge(clsx(inputs));
}

interface VersionDropdownProps {
  current: string;
  options: string[];
  onChange: (v: string) => void;
  isOpen: boolean;
  onToggle: () => void;
  allowCustom?: boolean;
  onCustomClick?: () => void;
}

const VersionDropdown: React.FC<VersionDropdownProps> = ({ current, options, onChange, isOpen, onToggle, allowCustom, onCustomClick }) => {
 return (
 <div className="relative">
 <button
   type="button"
 onClick={(e) => { e.stopPropagation(); onToggle(); }}
 className="flex items-center gap-1.5 bg-slate-100 dark:bg-black/40 border border-slate-200 dark:border-white/10 rounded-sm px-1.5 py-0.5 hover:border-blue-500/30 transition-colors group cursor-pointer"
 >
 <span className="text-[9px] font-black text-blue-600 dark:text-blue-400 uppercase tracking-tighter">v{current}</span>
 <MoreVertical size={9} className={cn("text-slate-400 dark:text-slate-500 group-hover:text-blue-500 transition-all", isOpen && "rotate-90")} />
 </button>

 {isOpen && (
 <>
 <button type="button" className="fixed inset-0 z-[60] w-full h-full bg-transparent border-none p-0 cursor-default focus:outline-none" onKeyDown={handleActionKey(onToggle)} onClick={onToggle} />
 <div className="absolute top-full left-0 mt-1 w-32 max-h-56 overflow-y-auto bg-white dark:bg-slate-900 shadow-2xl border border-slate-200 dark:border-white/10 rounded-sm z-[70] animate-in fade-in zoom-in-95 duration-200">
 <div className="p-1 space-y-0.5">
 {options.map((v) => (
 <button
   type="button"
 key={v}
 onClick={() => { onChange(v); onToggle(); }}
 onKeyDown={handleActionKey(() => { onChange(v); onToggle(); })}
 className={cn(
 "w-full text-left px-2 py-1 rounded-sm text-[9px] font-bold cursor-pointer transition-all outline-none bg-transparent border-none",
 v === current ? "bg-blue-600 text-white" : "text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-white/10 hover:text-slate-900 dark:hover:text-white focus:bg-slate-100 dark:focus:bg-white/10"
 )}
 >
 v{v}
 </button>
 ))}

 {allowCustom && (
 <div className="border-t border-slate-100 dark:border-white/5 mt-1 pt-1">
 <button
   type="button"
 onClick={(e) => { e.stopPropagation(); onCustomClick?.(); onToggle(); }}
 className="w-full flex items-center gap-2 px-2 py-1.5 rounded-sm text-[9px] font-bold text-slate-500 dark:text-slate-400 hover:bg-blue-500/10 hover:text-blue-600 transition-all text-left"
 >
 <FolderPlus size={10} /> Add Custom...
 </button>
 </div>
 )}
 </div>
 </div>
 </>
 )}
 </div>
 );
}

export default VersionDropdown;
