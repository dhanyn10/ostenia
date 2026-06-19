import React, { useState } from 'react';
import { CheckCircle2, Trash2, ChevronDown, ChevronUp, Download } from 'lucide-react';
import VersionDropdown from './VersionDropdown';
import CircularProgress from './CircularProgress';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { handleActionKey } from '../utils/a11y';

function cn(...inputs: ClassValue[]) {
 return twMerge(clsx(inputs));
}

function PluginItem({ 
 task,
 progress,
 isDropdownOpen,
 onDropdownToggle,
 selectedVersion,
 onVersionChange,
 onDeleteVersion,
 onInstall,
 onCancel,
 onOpenFolder,
 renderIcon,
 onInstallModule,
 onUninstallModule
}) {
 const [isModulesExpanded, setIsModulesExpanded] = useState(false);
 const availableVersions = task.versions || [];
 const installedVersions = task.installedVers || [];
 const dropdownOptions = [...new Set([...availableVersions, ...installedVersions])];
 const isCustomAllowed = task.name !== 'HeidiSQL' && task.name !== 'OpenSSL';

 const parentProgress = progress[task.name];
 const isActive = parentProgress &&
 parentProgress.status !== 'Ready' &&
 parentProgress.status !== 'Completed' &&
 !parentProgress.status?.startsWith('Error');

 const isSelectedInstalled = (task.name === 'OpenSSL' || task.name === 'HeidiSQL')
 ? task.isInstalled
 : installedVersions.includes(selectedVersion || task.version);

 return (
 <div className={cn(
 "bg-white/70 dark:bg-slate-900/40 rounded-sm p-4 border border-slate-200 dark:border-white/5 hover:border-slate-300 dark:hover:border-white/10 transition-all group flex flex-col gap-0 relative shadow-sm dark:shadow-lg",
 isDropdownOpen ? "z-[100]" : "z-0"
 )}>
 <div className="flex items-center gap-6 w-full">
 <div className={cn(
 "w-10 h-10 rounded-sm flex items-center justify-center shadow-lg transition-transform group-hover:scale-105 shrink-0",
 task.isInstalled ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400" : "bg-blue-500/10 text-blue-600 dark:text-blue-400"
 )}>
 {renderIcon(task.name, 18, "text-slate-900 dark:text-white")}
 </div>

 <div className="flex-1 min-w-0">
 <div className="flex items-center gap-3 flex-wrap">
 <h3 className="text-base font-black text-slate-900 dark:text-white uppercase italic tracking-tighter">{task.name}</h3>

 {(availableVersions.length > 0 || isCustomAllowed) ? (
 <VersionDropdown
 current={selectedVersion || task.version}
 options={dropdownOptions}
 isOpen={isDropdownOpen}
 onToggle={onDropdownToggle}
 onChange={onVersionChange}
 allowCustom={isCustomAllowed}
 onCustomClick={() => onOpenFolder(task.name)}
 />
 ) : (
 <span className="text-[9px] font-bold text-slate-500 uppercase tracking-widest bg-slate-100 dark:bg-white/5 px-1.5 py-0.5 rounded-sm">v{task.version}</span>
 )}

 {task.info && (
 <span className="text-[9px] font-bold text-blue-500 dark:text-blue-400 uppercase tracking-widest bg-blue-500/10 px-1.5 py-0.5 rounded-sm border border-blue-500/20">
 {task.info}
 </span>
 )}

 {installedVersions.map(ver => (
 <div
 key={ver}
 role="button"
 tabIndex={0}
 onClick={(e) => { e.stopPropagation(); onDeleteVersion(task.name, ver); }}
 onKeyDown={handleActionKey(() => onDeleteVersion(task.name, ver))}
 className="group/tag flex items-center gap-1.5 px-2 py-0.5 bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-white/10 text-slate-500 dark:text-slate-400 text-[8px] font-bold uppercase tracking-widest rounded-sm hover:border-rose-500/30 hover:bg-rose-500/10 hover:text-rose-400 transition-all cursor-pointer shadow-sm outline-none focus:ring-1 focus:ring-rose-500/40"
 title={`Delete v${ver}`}
 >
 <Trash2 size={10} className="w-0 opacity-0 group-hover/tag:w-2.5 group-hover/tag:opacity-100 transition-all text-rose-500" />
 {ver}
 </div>
 ))}
 </div>
 </div>

 {task.modules && task.modules.length > 0 && (
 <button
 onClick={() => setIsModulesExpanded(!isModulesExpanded)}
 className="p-2 hover:bg-slate-100 dark:hover:bg-white/5 rounded-sm transition-all text-slate-400 hover:text-slate-900 dark:hover:text-white"
 >
 {isModulesExpanded ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
 </button>
 )}

 <div className="flex items-center gap-6 shrink-0">
 {isActive && (
 <div className="animate-in fade-in slide-in-from-right-2 duration-300">
 <CircularProgress
 percentage={parentProgress.percentage}
 status={parentProgress.status}
 speed={parentProgress.speed}
 downloaded={parentProgress.downloaded}
 onCancel={() => onCancel(task.name)}
 />
 </div>
 )}

 {!isActive && (
 <button
 disabled={isSelectedInstalled}
 onClick={() => !isSelectedInstalled && onInstall(task)}
 className={cn(
 "px-5 py-2 rounded-sm text-[9px] font-black uppercase tracking-widest transition-all flex items-center justify-center gap-2 shadow-lg",
 isSelectedInstalled
 ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-500 border border-emerald-500/20 cursor-not-allowed"
 : "bg-blue-600 hover:bg-blue-500 text-white hover:scale-105"
 )}
 >
 {isSelectedInstalled && <CheckCircle2 size={14} />}
 {isSelectedInstalled ? 'Ready' : 'Download'}
 </button>
 )}
 </div>
 </div>

 {isModulesExpanded && task.modules && task.modules.length > 0 && (
 <div className="mt-4 pt-4 border-t border-slate-200 dark:border-white/5 space-y-2 animate-in slide-in-from-top-2 duration-200">
 <p className="text-[9px] font-black text-slate-400 uppercase tracking-widest mb-3">Available Modules</p>
 <div className="grid grid-cols-1 gap-2">
 {task.modules.map(mod => {
 const modProgress = progress[mod.name];
 const isModActive = modProgress && modProgress.status !== 'Ready' && modProgress.status !== 'Completed' && !modProgress.status?.startsWith('Error');
 const isModInstalled = mod.isInstalled || mod.status === 'Ready' || mod.status === 'Completed';

 return (
 <div key={mod.name} className="flex items-center justify-between p-3 bg-slate-50 dark:bg-white/5 rounded-sm border border-slate-100 dark:border-white/5 group/mod">
 <div className="flex items-center gap-3">
 <div className={cn(
 "w-6 h-6 rounded-sm flex items-center justify-center text-[10px] font-black",
 isModInstalled ? "bg-emerald-500/10 text-emerald-600" : "bg-blue-500/10 text-blue-600"
 )}>
 {mod.name[0]}
 </div>
 <div className="flex items-center gap-3">
 <div className="flex flex-col">
 <span className="text-xs font-bold text-slate-700 dark:text-slate-300">{mod.name}</span>
 <p className="text-[8px] text-slate-400 font-bold uppercase tracking-tighter mt-0.5">{mod.status}</p>
 </div>

 {mod.version && (
 <span className="text-[8px] font-bold text-slate-500 uppercase tracking-widest bg-slate-100 dark:bg-white/5 px-1.5 py-0.5 rounded-sm h-fit">v{mod.version}</span>
 )}
 </div>
 </div>

 <div className="flex items-center gap-4">
 {isModActive && (
 <div className="flex items-center gap-3">
 <div className="flex flex-col items-end">
 <span className="text-[8px] font-black text-blue-500 uppercase">{modProgress.status}</span>
 <span className="text-[7px] text-slate-400 font-bold">{modProgress.percentage.toFixed(0)}%</span>
 </div>
 <div className="w-8 h-8">
 <CircularProgress percentage={modProgress.percentage}  size={32} />
 </div>
 </div>
 )}

 {!isModActive && (
 <div className="flex items-center gap-2">
 {isModInstalled && (
 <button
 onClick={() => onUninstallModule(task.name, mod.name)}
 className="p-1.5 bg-slate-100 dark:bg-white/5 hover:bg-rose-500/10 text-slate-400 dark:text-slate-500 hover:text-rose-600 dark:hover:text-rose-400 rounded-sm transition-all border border-slate-200 dark:border-white/5"
 title={`Uninstall ${mod.name}`}
 >
 <Trash2 size={12} />
 </button>
 )}

 <button
 disabled={isModInstalled || !task.isInstalled}
 onClick={() => onInstallModule(task.name, mod.name)}
 className={cn(
 "p-2 rounded-sm transition-all",
 isModInstalled
 ? "text-emerald-500 cursor-not-allowed"
 : !task.isInstalled
 ? "text-slate-300 cursor-not-allowed"
 : "text-blue-500 hover:bg-blue-500/10 hover:scale-110"
 )}
 title={!task.isInstalled ? `Install ${task.name} first` : isModInstalled ? "Installed" : `Install ${mod.name}`}
 >
 {isModInstalled ? <CheckCircle2 size={16} /> : <Download size={16} />}
 </button>
 </div>
 )}
 </div>
 </div>
 );
 })}
 </div>
 </div>
 )}
 </div>
 );
}

export default PluginItem;
