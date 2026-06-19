import React from 'react';
import { Plus } from 'lucide-react';
import { clsx } from 'clsx';
import { handleActionKey } from "../utils/a11y";
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
 return twMerge(clsx(inputs));
}

function AddPluginAction({ isAddingPlugin, setIsAddingPlugin, prerequisites, services, handleAddToHome, renderIcon }) {
 return (
 <div className="mb-3 relative">
 <button
 onKeyDown={handleActionKey(() => setIsAddingPlugin(!isAddingPlugin))} onClick={() => setIsAddingPlugin(!isAddingPlugin)}
 className="w-full bg-slate-100/50 dark:bg-white/[0.01] border border-dashed border-slate-300 dark:border-white/5 hover:border-blue-500/20 hover:bg-blue-500/5 rounded-sm p-3 transition-all flex items-center justify-center gap-3 group"
 >
 <Plus size={14} className={cn("transition-colors", isAddingPlugin ? "text-rose-500" : "text-slate-400 group-hover:text-blue-500")} />
 <span className="text-[9px] font-black uppercase tracking-[0.2em] text-slate-500 group-hover:text-blue-600">
 {isAddingPlugin ? 'Close Menu' : 'Add Plugin to Home'}
 </span>
 </button>

 {isAddingPlugin && (
 <div className="absolute left-0 right-0 top-full mt-2 p-3 bg-white dark:bg-slate-900 border border-slate-200 dark:border-white/10 rounded-sm shadow-3xl z-50 animate-in fade-in slide-in-from-top-1">
 <div className="grid grid-cols-2 gap-1.5">
 {prerequisites.filter(p => !services.find(s => s.name === p.name)).map(task => {
 return (
 <button
 key={task.name}
 onKeyDown={handleActionKey(() => handleAddToHome(task))} onClick={() => handleAddToHome(task)}
 className="flex items-center gap-2.5 p-2.5 bg-slate-50 dark:bg-white/5 hover:bg-slate-100 dark:hover:bg-white/10 rounded-sm text-left transition-all group/item"
 >
 <div className="w-7 h-7 rounded-sm bg-blue-500/10 text-blue-600 dark:text-blue-400 flex items-center justify-center group-hover/item:bg-blue-500 group-hover/item:text-white transition-all">
 {renderIcon(task.name, 14)}
 </div>
 <span className="text-xs font-bold text-slate-700 dark:text-white">{task.name}</span>
 </button>
 );
 })}
 {prerequisites.filter(p => !services.find(s => s.name === p.name)).length === 0 && (
 <div className="col-span-2 py-4 text-center text-[10px] font-bold text-slate-400 uppercase italic">All plugins are already pinned</div>
 )}
 </div>
 </div>
 )}
 </div>
 );
}

export default AddPluginAction;
