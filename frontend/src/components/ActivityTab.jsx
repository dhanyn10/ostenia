import React, { useState, useEffect } from 'react';
import { Plus, X, Activity, Globe, Trash2, FolderOpen, Clock, Lock, Unlock, Terminal, ChevronDown, Monitor, CheckCircle2, Settings2, HardDrive } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import { OpenServiceTerminal, SwitchServiceVersion, GetPHPExtensions, TogglePHPExtension } from '../../wailsjs/go/main/App';
import ExtensionModal from './ExtensionModal';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

function ActivityTab({ 
  serverRoot, appsLocation, handleBrowseAppsLocation, handleBrowseServerRoot,
  isAddingPlugin, setIsAddingPlugin, prerequisites, services, handleAddToHome,
  ICON_MAP, handleToggleService, handleRemoveFromHome, setActiveTab,
  handleOpenPluginFolder, handleOpenServerRootFolder,
  apacheHttps, nginxHttps, handleToggleHttps
}) {
  const [openTerminalDropdown, setOpenTerminalDropdown] = useState(null);
  const [activeAccordion, setActiveAccordion] = useState(null);
  const [phpExtensions, setPhpExtensions] = useState([]);
  const [isModalOpen, setIsModalOpen] = useState(false);

  const handleOpenLocalTerminal = (name, type) => {
    OpenServiceTerminal(name, type);
    setOpenTerminalDropdown(null);
  };

  const handleSwitchVersion = async (serviceName, version) => {
    try {
      await SwitchServiceVersion(serviceName, version);
    } catch (err) {
      console.error("Failed to switch version:", err);
    }
  };

  const toggleAccordion = (name, hasExtra) => {
    if (!hasExtra) return;
    setActiveAccordion(activeAccordion === name ? null : name);
    setOpenTerminalDropdown(null);
  };

  const fetchPHPExtensions = async () => {
    try {
      const exts = await GetPHPExtensions();
      setPhpExtensions(exts || []);
    } catch (err) {
      console.error("Failed to fetch PHP extensions:", err);
    }
  };

  const handleTogglePHPExtension = async (extName, enable) => {
    try {
      await TogglePHPExtension(extName, enable);
      fetchPHPExtensions(); 
    } catch (err) {
      console.error("Failed to toggle PHP extension:", err);
    }
  };

  useEffect(() => {
    if (activeAccordion === 'PHP') {
      fetchPHPExtensions();
    }
  }, [activeAccordion]);

  return (
    <div className="flex flex-col h-full animate-in fade-in slide-in-from-bottom-2 duration-500">
      <ExtensionModal 
        isOpen={isModalOpen} 
        onClose={() => setIsModalOpen(false)} 
        extensions={phpExtensions} 
        onToggle={handleTogglePHPExtension}
        serviceName="PHP"
      />

      <div className="shrink-0 pt-4 pb-3 grid grid-cols-2 gap-3">
        <div className="bg-white/50 dark:bg-slate-900/40 backdrop-blur-xl rounded-sm p-4 border border-slate-200 dark:border-white/5 shadow-sm">
          <h3 className="text-[9px] font-black text-slate-500 dark:text-slate-400 uppercase tracking-[0.2em] mb-2 flex items-center gap-2">
            <HardDrive size={10} /> Apps Location
          </h3>
          <div className="flex items-center gap-2">
            <input type="text" readOnly value={appsLocation || ''} className="flex-1 bg-slate-100 dark:bg-black/20 border border-slate-200 dark:border-white/5 rounded-sm px-3 py-1.5 text-[10px] text-slate-500 dark:text-slate-400 font-mono truncate" />
            <button onClick={handleBrowseAppsLocation} className="p-1.5 bg-blue-600/10 hover:bg-blue-600/20 text-blue-600 rounded-sm transition-all border border-blue-500/20">
              <FolderOpen size={14} />
            </button>
          </div>
        </div>

        <div className="bg-white/50 dark:bg-slate-900/40 backdrop-blur-xl rounded-sm p-4 border border-slate-200 dark:border-white/5 shadow-sm">
          <h3 className="text-[9px] font-black text-slate-500 dark:text-slate-400 uppercase tracking-[0.2em] mb-2 flex items-center gap-2">
            <Globe size={10} /> Server Root Directory
          </h3>
          <div className="flex items-center gap-2">
            <input type="text" readOnly value={serverRoot || ''} className="flex-1 bg-slate-100 dark:bg-black/20 border border-slate-200 dark:border-white/5 rounded-sm px-3 py-1.5 text-[10px] text-slate-500 dark:text-slate-400 font-mono truncate" />
            <button onClick={handleBrowseServerRoot} className="p-1.5 bg-emerald-600/10 hover:bg-emerald-600/20 text-emerald-600 rounded-sm transition-all border border-emerald-500/20">
              <FolderOpen size={14} />
            </button>
          </div>
        </div>
      </div>

      <div className="mb-3 relative">
        <button 
          onClick={() => setIsAddingPlugin(!isAddingPlugin)}
          className="w-full bg-slate-100/50 dark:bg-white/[0.01] border border-dashed border-slate-300 dark:border-white/5 hover:border-blue-500/20 hover:bg-blue-500/5 rounded-sm p-3 transition-all flex items-center justify-center gap-3 group"
        >
          <Plus size={14} className="text-slate-400 group-hover:text-blue-500 transition-colors" />
          <span className="text-[9px] font-black uppercase tracking-[0.2em] text-slate-500 group-hover:text-blue-600">Add Plugin to Home</span>
        </button>

        {isAddingPlugin && (
          <div className="absolute left-0 right-0 top-full mt-2 p-3 bg-white dark:bg-slate-900 border border-slate-200 dark:border-white/10 rounded-sm shadow-3xl z-50 animate-in fade-in slide-in-from-top-1">
              <div className="grid grid-cols-2 gap-1.5">
                {prerequisites.filter(p => !services.find(service => service.name === p.name)).map(task => (
                  <button 
                    key={task.name}
                    onClick={() => handleAddToHome(task)}
                    className="flex items-center gap-2.5 p-2.5 bg-slate-50 dark:bg-white/5 hover:bg-slate-100 dark:hover:bg-white/10 rounded-sm text-left transition-all"
                  >
                    <div className="w-7 h-7 rounded-sm bg-blue-500/10 text-blue-600 dark:text-blue-400 flex items-center justify-center">
                        {(() => { const IconComponent = ICON_MAP[task.name] || ICON_MAP.default; return <IconComponent size={14} /> })()}
                    </div>
                    <span className="text-xs font-bold text-slate-700 dark:text-white">{task.name}</span>
                  </button>
                ))}
              </div>
          </div>
        )}
      </div>

      <div className="flex-1 overflow-y-auto pr-3 -mr-3 scrollbar-thin scrollbar-thumb-slate-200 dark:scrollbar-thumb-white/5 space-y-2 pb-4">
        {services.map((service) => {
          const task = prerequisites.find(p => p.name === service.name);
          const isInstalled = (task?.installedVers && task.installedVers.length > 0) || service.name === 'OpenSSL';
          const isWebServer = service.name === 'Apache' || service.name === 'Nginx';
          const isHttpsEnabled = service.name === 'Apache' ? apacheHttps : (service.name === 'Nginx' ? nginxHttps : false);
          const hasTerminalFacility = service.name !== 'HeidiSQL' && service.name !== 'OpenSSL';
          const hasOpenFolder = isInstalled && service.name !== 'OpenSSL';
          const hasTerminal = isInstalled && hasTerminalFacility;
          const hasHttpsToggle = isWebServer; 
          const hasPhpExtManager = service.name === 'PHP'; 
          const hasExtraActions = hasOpenFolder || hasTerminal || hasHttpsToggle || hasPhpExtManager;
          const isExpanded = activeAccordion === service.name;
          const installedVersions = task?.installedVers || [];
          const ServiceIcon = ICON_MAP[service.name] || ICON_MAP.default;

          return (
            <div key={service.name} className={cn("bg-white/70 dark:bg-slate-900/40 backdrop-blur-xl rounded-sm p-4 border border-slate-200 dark:border-white/5 hover:border-slate-300 dark:hover:border-white/10 transition-all flex flex-col relative shadow-sm dark:shadow-lg", isExpanded ? "z-[100] ring-1 ring-blue-500/20" : "z-10", hasExtraActions ? "cursor-pointer" : "cursor-default")} onClick={() => toggleAccordion(service.name, hasExtraActions)}>
              <div className="flex items-center gap-5">
                <div className="flex-1 min-w-0 px-2">
                  <div className="flex items-center gap-3 flex-wrap">
                    <div className="flex items-center gap-3">
                      <ServiceIcon size={18} className="text-slate-900 dark:text-white" />
                      <h3 className="text-base font-black text-slate-900 dark:text-white uppercase italic tracking-tighter">{service.name}</h3>
                      {(service.name === 'PHP' || service.name === 'Node.js') && installedVersions.length > 0 && (
                        <div className="flex items-center gap-1 ml-1" onClick={(e) => e.stopPropagation()}>
                          {installedVersions.map(ver => {
                            const systemString = (service.activeVersion || "").toString().toLowerCase().trim();
                            const cleanVer = ver.toString().replace(/^v/, "").replace(/^[a-z. ]+-/, "").trim();
                            const isActive = systemString.includes(cleanVer.toLowerCase());
                            return (
                              <button key={ver} onClick={() => handleSwitchVersion(service.name, ver)} className={cn("px-1.5 py-0.5 rounded-sm text-[8px] font-black uppercase tracking-widest border transition-all", isActive ? "bg-blue-600 border-blue-500 text-white shadow-lg" : "bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-white/5 text-slate-400 dark:text-slate-500 hover:border-blue-500/50 hover:text-blue-500")}>
                                {isActive && <CheckCircle2 size={8} className="inline mr-1" />} {cleanVer}
                              </button>
                            );
                          })}
                        </div>
                      )}
                    </div>
                    {service.remainingDays > 0 && (
                      <div className="flex items-center gap-1 px-1.5 py-0.5 bg-indigo-500/10 border border-indigo-500/20 rounded-sm text-[8px] font-bold text-indigo-600 dark:text-indigo-400 uppercase tracking-widest">
                        <Clock size={10} /> {service.remainingDays} Days Left
                      </div>
                    )}
                    <div className={cn("text-[8px] font-black uppercase tracking-[0.2em] px-2 py-0.5 rounded-sm border flex items-center gap-1.5", service.status === 'Running' ? "bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20" : "bg-slate-100 dark:bg-slate-900/80 text-slate-400 dark:text-slate-500 border-slate-200 dark:border-white/5")}>
                      {service.status === 'Running' && <div className="w-1 h-1 rounded-full bg-emerald-500 animate-pulse" />} {service.status}
                    </div>
                    {service.status === 'Running' && service.name !== 'OpenSSL' && (
                      <div className="flex items-center gap-2 animate-in fade-in slide-in-from-left-2 duration-300">
                        {service.pid > 0 && (
                          <div className="flex items-center gap-1 px-1.5 py-0.5 bg-blue-500/10 border border-blue-500/20 rounded-sm text-[8px] font-bold text-blue-600 dark:text-blue-400 uppercase tracking-widest">
                            <Activity size={10} /> PID: {service.pid}
                          </div>
                        )}
                        {(service.ports && service.ports.length > 0) || service.port > 0 ? (
                          <div className="flex items-center gap-1 px-1.5 py-0.5 bg-amber-500/10 border border-amber-500/20 rounded-sm text-[8px] font-bold text-amber-600 dark:text-amber-400 uppercase tracking-widest">
                            <Globe size={10} /> Port: {service.ports && service.ports.length > 0 ? service.ports.join(', ') : service.port}
                          </div>
                        ) : null}
                      </div>
                    )}
                  </div>
                </div>
                <div className="flex items-center gap-3" onClick={(e) => e.stopPropagation()}>
                  {!isInstalled && service.name !== 'OpenSSL' ? (
                    <button onClick={() => setActiveTab('plugins')} className="px-4 py-1.5 bg-blue-600/10 hover:bg-blue-600/20 text-blue-600 dark:text-blue-400 rounded-sm text-[9px] font-black uppercase tracking-widest transition-all border border-blue-500/20">Install First</button>
                  ) : (
                    <button onClick={() => handleToggleService(service.name, service.status)} className={cn("w-12 h-6 rounded-sm p-0.5 transition-all duration-300 ease-in-out relative ring-1 ring-inset", service.status === 'Running' ? "bg-emerald-500 ring-emerald-400/50" : "bg-slate-200 dark:bg-slate-800 ring-slate-300 dark:ring-white/5")}>
                      <div className={cn("w-5 h-5 bg-white rounded-sm transition-all duration-300 shadow-lg", service.status === 'Running' ? "translate-x-6" : "translate-x-0")} />
                    </button>
                  )}
                  <button onClick={() => handleRemoveFromHome(service.name)} className="h-6 px-3 bg-slate-100 dark:bg-white/5 hover:bg-rose-500/10 text-slate-400 dark:text-slate-500 hover:text-rose-600 dark:hover:text-rose-400 rounded-sm text-[9px] font-black uppercase tracking-widest transition-all border border-slate-200 dark:border-white/5">
                    <Trash2 size={12} />
                  </button>
                </div>
              </div>

              {hasExtraActions && (
                <div className={cn("transition-all duration-300 ease-in-out overflow-visible", isExpanded ? "max-h-24 opacity-100 mt-4" : "max-h-0 opacity-0 mt-0 overflow-hidden")} onClick={(e) => e.stopPropagation()}>
                  <div className="flex items-center flex-wrap gap-4 px-1 pb-2">
                    {hasPhpExtManager && (
                      <button onClick={() => setIsModalOpen(true)} className="flex items-center gap-2 px-3 py-1.5 h-8 bg-slate-100 dark:bg-white/5 hover:bg-blue-600/10 text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 rounded-sm text-[9px] font-black uppercase tracking-widest transition-all border border-slate-200 dark:border-white/5">
                        <Settings2 size={14} /> Extensions
                      </button>
                    )}
                    {hasOpenFolder && (
                      <button onClick={() => handleOpenPluginFolder(service.name)} className="w-8 h-8 flex items-center justify-center bg-slate-100 dark:bg-white/5 hover:bg-blue-600/10 text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 rounded-sm border border-slate-200 dark:border-white/5 transition-all">
                        <FolderOpen size={16} />
                      </button>
                    )}
                    {hasTerminal && (
                      <div className="relative">
                        <button onClick={() => setOpenTerminalDropdown(openTerminalDropdown === service.name ? null : service.name)} className={cn("w-12 h-8 flex items-center justify-center gap-1 rounded-sm border border-slate-200 dark:border-white/5 transition-all", openTerminalDropdown === service.name ? "bg-slate-200 dark:bg-slate-700 text-slate-900 dark:text-white" : "bg-slate-100 dark:bg-white/5 text-slate-500 dark:text-slate-400")}>
                          <Terminal size={16} /> <ChevronDown size={10} />
                        </button>
                        {openTerminalDropdown === service.name && (
                          <>
                            <div className="fixed inset-0 z-[150]" onClick={() => setOpenTerminalDropdown(null)} />
                            <div className="absolute top-full left-0 mt-1 w-40 bg-white dark:bg-slate-900 border border-slate-200 dark:border-white/10 rounded-sm shadow-2xl z-[160] animate-in fade-in slide-in-from-top-1 duration-200">
                              <div className="p-1">
                                <button onClick={() => handleOpenLocalTerminal(service.name, 'cmd')} className="w-full flex items-center gap-3 px-3 py-1.5 rounded-sm text-[10px] font-bold text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white transition-all text-left"><Monitor size={12} className="text-blue-500" /> CMD</button>
                                <button onClick={() => handleOpenLocalTerminal(service.name, 'powershell')} className="w-full flex items-center gap-3 px-3 py-1.5 rounded-sm text-[10px] font-bold text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white transition-all text-left"><Monitor size={12} className="text-blue-600" /> PowerShell</button>
                              </div>
                            </div>
                          </>
                        )}
                      </div>
                    )}
                    {hasHttpsToggle && (
                      <button onClick={() => handleToggleHttps(service.name)} className={cn("w-14 h-8 rounded-sm p-1 transition-all duration-300 ease-in-out relative ring-1 ring-inset", isHttpsEnabled ? "bg-rose-500 ring-rose-400/50 shadow-[0_0_10px_rgba(244,63,94,0.3)]" : "bg-slate-200 dark:bg-slate-800 ring-slate-300 dark:ring-white/5")}>
                        <div className={cn("w-6 h-6 bg-white rounded-sm transition-all duration-300 shadow-lg flex items-center justify-center", isHttpsEnabled ? "translate-x-6" : "translate-x-0")}>
                          {isHttpsEnabled ? <Lock size={14} className="text-rose-600" /> : <Unlock size={14} className="text-slate-400" />}
                        </div>
                      </button>
                    )}
                  </div>
                </div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}

export default ActivityTab;
