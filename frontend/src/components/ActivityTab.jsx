import React from 'react';
import { Plus, X, Activity, Globe, Trash2, FolderOpen } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

function ActivityTab({ 
  serverRoot, handleServerRootChange, handleBrowseServerRoot,
  isAddingPlugin, setIsAddingPlugin, prerequisites, services, handleAddToHome,
  ICON_MAP, handleToggleService, handleRemoveFromHome, setActiveTab,
  handleOpenPluginFolder, handleOpenServerRootFolder // Added handleOpenServerRootFolder
}) {
  return (
    <div className="flex flex-col h-full animate-in fade-in slide-in-from-bottom-2 duration-500">
      {/* Dashboard Controls */}
      <div className="shrink-0 pt-4 pb-3 space-y-3">
        {/* Server Root Configuration */}
        <div className="bg-white/50 dark:bg-slate-900/40 backdrop-blur-xl rounded-sm p-4 border border-slate-200 dark:border-white/5 hover:border-slate-300 dark:hover:border-white/10 transition-all group flex items-center gap-4 shadow-sm dark:shadow-lg">
          <div className="flex-1 min-w-0">
            <h3 className="text-[9px] font-black text-slate-500 dark:text-slate-400 uppercase tracking-[0.2em] mb-1">Server Root Directory</h3>
            <div className="flex items-center gap-2">
              <input
                type="text"
                value={serverRoot}
                onChange={handleServerRootChange}
                placeholder="C:/ostenia/www"
                className="flex-1 bg-slate-100 dark:bg-black/20 border border-slate-200 dark:border-white/5 rounded-sm px-3 py-1.5 text-[11px] text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-600 focus:outline-none focus:ring-1 focus:ring-blue-500/50 transition-all font-mono"
              />
              <div className="flex items-center gap-1.5">
                <button 
                  onClick={handleBrowseServerRoot}
                  className="px-3 py-1.5 bg-blue-600/10 hover:bg-blue-600/20 text-blue-600 dark:text-blue-400 rounded-sm text-[9px] font-black uppercase tracking-widest transition-all border border-blue-500/20"
                >
                  Browse
                </button>
                <button 
                  onClick={handleOpenServerRootFolder}
                  title="Open Server Root in Explorer"
                  className="p-1.5 bg-slate-100 dark:bg-white/5 hover:bg-blue-600/10 text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 rounded-sm transition-all border border-slate-200 dark:border-white/5"
                >
                  <FolderOpen size={16} />
                </button>
              </div>
            </div>
          </div>
        </div>

        <div className="relative">
          <button 
            onClick={() => setIsAddingPlugin(!isAddingPlugin)}
            className="w-full bg-slate-100/50 dark:bg-white/[0.01] border border-dashed border-slate-300 dark:border-white/5 hover:border-blue-500/20 hover:bg-blue-500/5 rounded-sm p-4 transition-all flex items-center justify-center gap-3 group"
          >
            <div className="w-8 h-8 rounded-sm bg-slate-200 dark:bg-slate-800 flex items-center justify-center text-slate-500 group-hover:bg-blue-500 group-hover:text-white transition-all shrink-0">
                <Plus size={16} />
            </div>
            <span className="text-[10px] font-black uppercase tracking-[0.2em] text-slate-500 group-hover:text-blue-600 dark:group-hover:text-blue-400">Add Plugin to Home</span>
          </button>

          {isAddingPlugin && (
            <div className="absolute top-full left-0 right-0 mt-3 p-3 bg-white dark:bg-slate-900 border border-white/10 rounded-sm shadow-3xl z-50 animate-in fade-in slide-in-from-top-1">
                <div className="flex items-center justify-between mb-3 px-1.5">
                  <span className="text-[9px] font-black text-slate-500 dark:text-slate-400 uppercase tracking-widest">Available to Pin</span>
                  <button onClick={() => setIsAddingPlugin(false)} className="text-slate-400 hover:text-slate-900 dark:hover:text-white"><X size={12} /></button>
                </div>
                <div className="grid grid-cols-2 gap-1.5">
                  {prerequisites.filter(p => !services.find(service => service.name === p.name)).map(task => (
                    <button 
                      key={task.name}
                      onClick={() => handleAddToHome(task)}
                      className="flex items-center gap-2.5 p-2.5 bg-slate-50 dark:bg-white/5 hover:bg-slate-100 dark:hover:bg-white/10 rounded-sm text-left transition-all group/item"
                    >
                      <div className="w-7 h-7 rounded-sm bg-blue-500/10 text-blue-600 dark:text-blue-400 flex items-center justify-center group-hover/item:bg-blue-500 group-hover/item:text-white transition-all">
                          {(() => { const Icon = ICON_MAP[task.name] || ICON_MAP.default; return <Icon size={14} /> })()}
                      </div>
                      <span className="text-xs font-bold text-slate-700 dark:text-white">{task.name}</span>
                    </button>
                  ))}
                </div>
            </div>
          )}
        </div>
      </div>

      {/* Scrollable Services List */}
      <div className="flex-1 overflow-y-auto pr-3 -mr-3 scrollbar-thin scrollbar-thumb-slate-200 dark:scrollbar-thumb-white/5 space-y-2">
        {services.map((service) => {
          const task = prerequisites.find(p => p.name === service.name);
          const isInstalled = task?.installedVers && task.installedVers.length > 0;

          return (
            <div key={service.name} className="bg-white/70 dark:bg-slate-900/40 backdrop-blur-xl rounded-sm p-4 border border-slate-200 dark:border-white/5 hover:border-slate-300 dark:hover:border-white/10 transition-all group flex items-center gap-5 relative shadow-sm dark:shadow-lg">
              <div className="flex-1 min-w-0 px-2">
                <div className="flex items-center gap-3 flex-wrap">
                  <h3 className="text-base font-black text-slate-900 dark:text-white uppercase italic tracking-tighter">{service.name}</h3>
                  <div className={cn(
                    "text-[8px] font-black uppercase tracking-[0.2em] px-2 py-0.5 rounded-sm border flex items-center gap-1.5",
                    service.status === 'Running'
                      ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20"
                      : "bg-slate-100 dark:bg-slate-900/80 text-slate-400 dark:text-slate-500 border-slate-200 dark:border-white/5"
                  )}>
                    {service.status === 'Running' && <div className="w-1 h-1 rounded-full bg-emerald-500 animate-pulse" />}
                    {service.status}
                  </div>
                  
                  {/* Runtime Stats (PID & Port) */}
                  {service.status === 'Running' && (
                    <div className="flex items-center gap-2 animate-in fade-in slide-in-from-left-2 duration-300">
                      {service.pid > 0 && (
                        <div className="flex items-center gap-1 px-1.5 py-0.5 bg-blue-500/10 border border-blue-500/20 rounded-sm text-[8px] font-bold text-blue-600 dark:text-blue-400 uppercase tracking-widest">
                          <Activity size={10} />
                          PID: {service.pid}
                        </div>
                      )}
                      {service.port > 0 && (
                        <div className="flex items-center gap-1 px-1.5 py-0.5 bg-amber-500/10 border border-amber-500/20 rounded-sm text-[8px] font-bold text-amber-600 dark:text-amber-400 uppercase tracking-widest">
                          <Globe size={10} />
                          Port: {service.port}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              </div>

              <div className="flex items-center gap-3">
                {isInstalled && (
                  <button 
                    onClick={() => handleOpenPluginFolder(service.name)}
                    className="h-6 px-3 bg-slate-100 dark:bg-white/5 hover:bg-blue-600/10 text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 rounded-sm text-[9px] font-black uppercase tracking-widest transition-all border border-slate-200 dark:border-white/5"
                    title={`Open ${service.name} Folder`}
                  >
                    <FolderOpen size={12} />
                  </button>
                )}
                {!isInstalled ? (
                  <button onClick={() => setActiveTab('plugins')} className="px-4 py-1.5 bg-blue-600/10 hover:bg-blue-600/20 text-blue-600 dark:text-blue-400 rounded-sm text-[9px] font-black uppercase tracking-widest transition-all border border-blue-500/20">Install First</button>
                ) : (
                  <button
                    onClick={() => handleToggleService(service.name, service.status)}
                    className={cn(
                      "w-12 h-6 rounded-sm p-0.5 transition-all duration-300 ease-in-out relative ring-1 ring-inset",
                      service.status === 'Running' 
                        ? "bg-emerald-500 ring-emerald-400/50" 
                        : "bg-slate-200 dark:bg-slate-800 ring-slate-300 dark:ring-white/5"
                    )}
                  >
                    <div className={cn(
                      "w-5 h-5 bg-white rounded-sm transition-all duration-300 shadow-lg",
                      service.status === 'Running' ? "translate-x-6" : "translate-x-0"
                    )} />
                  </button>
                )}

                <button 
                  onClick={() => handleRemoveFromHome(service.name)}
                  className="h-6 px-3 bg-slate-100 dark:bg-white/5 hover:bg-rose-500/10 text-slate-400 dark:text-slate-500 hover:text-rose-600 dark:hover:text-rose-400 rounded-sm text-[9px] font-black uppercase tracking-widest transition-all border border-slate-200 dark:border-white/5"
                >
                  <Trash2 size={12} />
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

export default ActivityTab;
