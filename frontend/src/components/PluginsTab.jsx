import React from 'react';
import { CheckCircle2, Trash2 } from 'lucide-react';
import VersionDropdown from './VersionDropdown';
import CircularProgress from './CircularProgress';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

function PluginsTab({ 
  prerequisites, downloadProgress, openDropdown, setOpenDropdown,
  selectedVersions, setSelectedVersions, handleDeleteVersion,
  handleInstallSingle, handleCancel, ICON_MAP 
}) {
  return (
    <div className="flex flex-col h-full pt-4 animate-in fade-in slide-in-from-bottom-2 duration-300">
      <div className="flex-1 overflow-y-auto pr-3 -mr-3 scrollbar-thin scrollbar-thumb-slate-200 dark:scrollbar-thumb-white/5 space-y-2">
        {prerequisites.map((task) => {
          if (!task) return null;
          
          const progress = downloadProgress[task.name];
          const isActive = progress && progress.percentage > 0 && progress.percentage < 100;
          const isDropdownOpen = openDropdown === task.name;
          const Icon = ICON_MAP[task.name] || ICON_MAP.default;
          
          // Safety check for versions to prevent "not iterable" error
          const availableVersions = task.versions || [];
          const installedVersions = task.installedVers || [];
          const dropdownOptions = [...new Set([...availableVersions, ...installedVersions])];
          
          return (
            <div 
              key={task.name} 
              className={cn(
                "bg-white/70 dark:bg-slate-900/40 backdrop-blur-xl rounded-sm p-4 border border-slate-200 dark:border-white/5 hover:border-slate-300 dark:hover:border-white/10 transition-all group flex items-center gap-6 relative shadow-sm dark:shadow-lg",
                isDropdownOpen ? "z-[100]" : "z-0"
              )}
            >
              <div className={cn(
                "w-10 h-10 rounded-sm flex items-center justify-center shadow-lg transition-transform group-hover:scale-105",
                task.isInstalled ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400" : "bg-blue-500/10 text-blue-600 dark:text-blue-400"
              )}>
                <Icon size={18} />
              </div>

              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-3 flex-wrap">
                  <h3 className="text-base font-black text-slate-900 dark:text-white uppercase italic tracking-tighter">{task.name}</h3>
                  
                  {availableVersions.length > 0 ? (
                    <VersionDropdown 
                      current={selectedVersions[task.name] || task.version}
                      options={dropdownOptions}
                      isOpen={isDropdownOpen}
                      onToggle={() => setOpenDropdown(isDropdownOpen ? null : task.name)}
                      onChange={(v) => setSelectedVersions(prev => ({ ...prev, [task.name]: v }))}
                    />
                  ) : (
                    <span className="text-[9px] font-bold text-slate-500 uppercase tracking-widest bg-slate-100 dark:bg-white/5 px-1.5 py-0.5 rounded-sm">v{task.version}</span>
                  )}

                  {installedVersions.map(ver => (
                    <div 
                      key={ver} 
                      onClick={(e) => { e.stopPropagation(); handleDeleteVersion(task.name, ver); }}
                      className="group/tag flex items-center gap-1.5 px-2 py-0.5 bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-white/10 text-slate-500 dark:text-slate-400 text-[8px] font-bold uppercase tracking-widest rounded-sm hover:border-rose-500/30 hover:bg-rose-500/10 hover:text-rose-400 transition-all cursor-pointer shadow-sm"
                      title={`Delete v${ver}`}
                    >
                      <Trash2 size={10} className="w-0 opacity-0 group-hover/tag:w-2.5 group-hover/tag:opacity-100 transition-all text-rose-500" />
                      v{ver}
                    </div>
                  ))}
                </div>
              </div>

              <div className="flex items-center gap-6">
                {isActive && (
                  <div className="animate-in fade-in slide-in-from-right-2 duration-300">
                    <CircularProgress 
                      percentage={progress.percentage} 
                      status={progress.status} 
                      speed={progress.speed}
                      downloaded={progress.downloaded}
                      onCancel={() => handleCancel(task.name)}
                    />
                  </div>
                )}

                {(() => {
                  const selectedVer = selectedVersions[task.name] || task.version;
                  const isSelectedInstalled = installedVersions.includes(selectedVer);
                  
                  if (isActive) return null;

                  return (
                    <button
                      disabled={isSelectedInstalled}
                      onClick={() => !isSelectedInstalled && handleInstallSingle(task)}
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
                  );
                })()}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

export default PluginsTab;
