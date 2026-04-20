import { useState, useEffect } from 'react';
import { Play, Square, Download, Settings, Terminal as TerminalIcon, Database, Globe, FolderOpen, MoreVertical, ExternalLink, CheckCircle2, AlertCircle, XCircle, X, Loader2, List, Trash2, ChevronRight, Search, Home } from 'lucide-react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { GetPrerequisites, InstallPrerequisite, CancelDownload, StartAllServices, StopAllServices, OpenTerminal, DeleteVersion, StartService, StopService } from '../wailsjs/go/main/App';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

function CircularProgress({ percentage, status, speed, downloaded, onCancel }) {
  const radius = 16;
  const circumference = 2 * Math.PI * radius;
  const strokeDashoffset = circumference - (percentage / 100) * circumference;
  const isStreaming = status?.includes('Streaming');

  return (
    <div className="relative group/progress cursor-pointer overflow-hidden p-1 rounded-xl" onClick={onCancel}>
      <div className="flex items-center gap-4 group-hover/progress:opacity-0 transition-opacity duration-200">
        <div className="text-right">
          <p className="text-[10px] font-black text-white">{downloaded || '...'}</p>
          <p className="text-[9px] font-bold text-slate-500 uppercase tracking-widest leading-none">{speed || '...'}</p>
        </div>
        <div className="relative w-10 h-10">
          <svg className="w-full h-full -rotate-90">
            <circle className="text-white/5" strokeWidth="3" stroke="currentColor" fill="transparent" r={radius} cx="20" cy="20" />
            <circle
              className={cn("text-blue-500 transition-all duration-500", isStreaming && "animate-[spin_2s_linear_infinite]")}
              strokeWidth="3"
              strokeDasharray={circumference}
              strokeDashoffset={isStreaming ? circumference * 0.7 : strokeDashoffset}
              strokeLinecap="round"
              stroke="currentColor"
              fill="transparent"
              r={radius}
              cx="20"
              cy="20"
            />
          </svg>
          <div className="absolute inset-0 flex items-center justify-center text-[8px] font-black text-blue-400">
             {isStreaming ? <Loader2 size={10} className="animate-spin" /> : `${Math.round(percentage)}%`}
          </div>
        </div>
      </div>
      <div className="absolute inset-0 flex items-center justify-center translate-y-full group-hover/progress:translate-y-0 transition-transform duration-200 bg-rose-600 rounded-xl">
         <div className="flex items-center gap-2 px-6">
            <X size={14} className="text-white" />
            <span className="text-[10px] font-black text-white uppercase tracking-widest">Cancel</span>
         </div>
      </div>
    </div>
  );
}

function VersionDropdown({ current, options, onChange, isOpen, onToggle }) {
  return (
    <div className="relative">
      <button
        onClick={(e) => { e.stopPropagation(); onToggle(); }}
        className="flex items-center gap-1.5 bg-black/40 border border-white/10 rounded-md px-2 py-0.5 hover:border-blue-500/30 transition-colors group cursor-pointer"
      >
        <span className="text-[10px] font-black text-blue-400">v{current}</span>
        <MoreVertical size={10} className={cn("text-slate-500 group-hover:text-blue-400 transition-all", isOpen && "rotate-90")} />
      </button>

      {isOpen && (
        <>
          <div className="fixed inset-0 z-[60]" onClick={onToggle} />
          <div className="absolute top-full left-0 mt-2 w-32 max-h-48 overflow-y-auto bg-slate-900 shadow-2xl border border-white/10 rounded-xl backdrop-blur-xl z-[70] animate-in fade-in zoom-in-95 duration-200">
            <div className="p-1">
              {options.map((v) => (
                <div
                  key={v}
                  onClick={() => { onChange(v); onToggle(); }}
                  className={cn(
                    "px-3 py-1.5 rounded-lg text-[10px] font-bold cursor-pointer transition-all",
                    v === current ? "bg-blue-600 text-white" : "text-slate-400 hover:bg-white/10 hover:text-white"
                  )}
                >
                  v{v}
                </div>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}

function Toast({ toasts, removeToast }) {
  return (
    <div className="fixed bottom-6 right-6 z-[100] flex flex-col gap-3 w-80">
      {toasts.map((toast) => (
        <div key={toast.id} className={cn(
          "p-4 rounded-xl shadow-2xl flex items-start gap-4 transition-all duration-300 animate-in slide-in-from-right-4",
          toast.type === 'error' ? "bg-rose-950/90 border border-rose-500/30 text-rose-200" :
          toast.type === 'success' ? "bg-emerald-950/90 border border-emerald-500/30 text-emerald-200" :
          "bg-slate-900/90 border border-white/10 text-white"
        )}>
          <div className="mt-0.5">
            {toast.type === 'error' ? <XCircle size={18} /> : 
             toast.type === 'success' ? <CheckCircle2 size={18} /> : 
             <AlertCircle size={18} />}
          </div>
          <div className="flex-1 space-y-1">
            <h5 className="font-bold text-xs uppercase tracking-widest">{toast.title}</h5>
            <p className="text-xs opacity-80">{toast.message}</p>
          </div>
          <button onClick={() => removeToast(toast.id)} className="opacity-40 hover:opacity-100 transition-opacity">
            <X size={14} />
          </button>
        </div>
      ))}
    </div>
  );
}

function LogViewer({ logs, isOpen, onClose }) {
  if (!isOpen) return null;
  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center p-8 bg-black/60 backdrop-blur-sm animate-in fade-in duration-300">
      <div className="bg-[#1e293b] w-full max-w-4xl h-[70vh] rounded-[2rem] border border-white/5 flex flex-col shadow-3xl overflow-hidden animate-in zoom-in-95 duration-300">
        <div className="p-6 border-b border-white/5 flex items-center justify-between bg-white/[0.02]">
          <div className="flex items-center gap-3">
            <List size={20} className="text-blue-400" />
            <h3 className="font-black text-white uppercase italic tracking-tighter">System Activity Logs</h3>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-white/10 rounded-xl text-slate-400 hover:text-white transition-all">
            <X size={20} />
          </button>
        </div>
        <div className="flex-1 overflow-y-auto p-6 font-mono text-[11px] space-y-1 bg-black/20">
           {logs.length === 0 && <p className="text-slate-600 italic">No logs recorded yet...</p>}
           {logs.map((log, i) => (
             <div key={i} className="flex gap-4 group">
               <span className="text-slate-700 select-none">[{log.time}]</span>
               <span className={cn(
                 "flex-1",
                 log.msg.includes('Error') || log.msg.includes('failed') ? "text-rose-400" :
                 log.msg.includes('success') || log.msg.includes('Ready') ? "text-emerald-400" :
                 "text-slate-400"
               )}>{log.msg}</span>
             </div>
           ))}
        </div>
      </div>
    </div>
  );
}

function App() {
  const [activeTab, setActiveTab] = useState('activity'); // 'activity' or 'plugins'
  const [services, setServices] = useState([
    { name: 'Apache', status: 'Stopped', icon: Globe },
    { name: 'MySQL', status: 'Stopped', icon: Database },
  ]);
  const [prerequisites, setPrerequisites] = useState([]);
  const [downloadProgress, setDownloadProgress] = useState({});
  const [toasts, setToasts] = useState([]);
  const [logs, setLogs] = useState([]);
  const [isLogOpen, setIsLogOpen] = useState(false);
  const [openDropdown, setOpenDropdown] = useState(null);
  const [loading, setLoading] = useState(true);
  const [selectedVersions, setSelectedVersions] = useState({});

  const addLog = (msg) => {
    setLogs(prev => [{ time: new Date().toLocaleTimeString(), msg }, ...prev].slice(0, 500));
  };

  const addToast = (title, message, type = 'info') => {
    const id = Math.random().toString(36).substr(2, 9);
    setToasts(prev => [...prev, { id, title, message, type }]);
    setTimeout(() => removeToast(id), 5000);
  };

  const removeToast = (id) => {
    setToasts(prev => prev.filter(t => t.id !== id));
  };

  const refreshPrerequisites = async () => {
    if (window.go) {
      try {
        const tasks = await GetPrerequisites();
        setPrerequisites(tasks || []);
        
        if (tasks) {
          setSelectedVersions(prev => {
            const next = { ...prev };
            tasks.forEach(t => {
              if (!next[t.name]) {
                if (t.installedVers && t.installedVers.length > 0) {
                  next[t.name] = t.installedVers[0];
                } else if (t.version) {
                  next[t.name] = t.version;
                }
              }
            });
            return next;
          });

          const initialProgress = {};
          tasks.forEach(t => {
            if (t.isInstalled) {
              initialProgress[t.name] = { name: t.name, percentage: 100, status: 'Installed' };
            }
          });
          setDownloadProgress(prev => ({ ...initialProgress, ...prev }));
        }
      } catch (err) {
        addLog(`Error fetching prerequisites: ${err}`);
      } finally {
        setLoading(false);
      }
    }
  };

  useEffect(() => {
    refreshPrerequisites();

    if (window.runtime) {
      EventsOn('service_status', (data) => {
        setServices(prev => prev.map(s => s.name === data.name ? { ...s, status: data.status } : s));
        addLog(`Service ${data.name} status changed to ${data.status}`);
      });

      EventsOn('download_progress', (data) => {
        setDownloadProgress(prev => ({ ...prev, [data.name]: data }));
        if (data.percentage === 100 && (data.status === 'Completed' || data.status === 'Ready')) {
          if (data.status === 'Completed') addToast(data.name, 'Installed successfully', 'success');
          addLog(`${data.name} installation completed`);
          refreshPrerequisites();
        }
      });

      EventsOn('download_error', (data) => {
        addToast(`${data.name} Error`, data.error, 'error');
        addLog(`Error installing ${data.name}: ${data.error}`);
      });
    }
  }, []);

  const handleStartAll = () => { addLog('Starting all services...'); if (window.go) StartAllServices(); };
  const handleStopAll = () => { addLog('Stopping all services...'); if (window.go) StopAllServices(); };
  const handleTerminal = () => { addLog('Opening terminal...'); if (window.go) OpenTerminal(); };

  const handleToggleService = (name, currentStatus) => {
    if (!window.go) return;
    if (currentStatus === 'Running') {
      addLog(`Stopping ${name}...`);
      StopService(name);
    } else {
      addLog(`Starting ${name}...`);
      StartService(name);
    }
  };

  const handleCancel = (name) => {
    addLog(`Requesting cancellation for ${name}...`);
    if (window.go) {
      CancelDownload(name);
      setDownloadProgress(prev => ({
        ...prev,
        [name]: { name, percentage: 0, status: 'Cancelled', speed: '', downloaded: '' }
      }));
    }
  };

  const handleInstallSingle = async (task) => {
    const selectedVer = selectedVersions[task.name] || task.version;
    addLog(`Initiating installation for ${task.name} v${selectedVer}...`);
    
    const modifiedTask = { ...task };
    const arch = navigator.userAgent.includes('Win64') || navigator.userAgent.includes('x64') ? 'x64' : 'x86';

    if (task.name === 'PHP' && task.versions) {
       modifiedTask.version = selectedVer;
       modifiedTask.target = `php/php-${selectedVer}`;
       modifiedTask.url = `https://downloads.php.net/~windows/releases/php-${selectedVer}-Win32-vs16-${arch}.zip`;
    }

    if (task.name === 'Apache' && task.versions && task.versionUrls) {
       modifiedTask.version = selectedVer;
       modifiedTask.target = `apache/httpd-${selectedVer}`;
       modifiedTask.url = task.versionUrls[selectedVer];
    }

    if (task.name === 'MySQL' && task.versions && task.versionUrls) {
       modifiedTask.version = selectedVer;
       modifiedTask.target = `mysql/mysql-${selectedVer}`;
       modifiedTask.url = task.versionUrls[selectedVer];
    }

    if (window.go) {
      try {
        await InstallPrerequisite(modifiedTask);
      } catch (err) {
        addLog(`Installation process for ${task.name} ended: ${err}`);
      }
    }
    refreshPrerequisites();
  };

  const handleDeleteVersion = async (taskName, version) => {
    addLog(`Deleting ${taskName} v${version}...`);
    if (window.go) {
      try {
        await DeleteVersion(taskName, version);
        addToast("Deleted", `${taskName} v${version} has been removed.`, "success");
      } catch (err) {
        addLog(`Failed to delete ${taskName} v${version}: ${err}`);
      }
      refreshPrerequisites();
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-[#0f172a] flex items-center justify-center">
        <Loader2 className="animate-spin text-blue-500" size={40} />
      </div>
    );
  }

  return (
    <div className="flex h-screen bg-[#0f172a] text-slate-200 font-sans selection:bg-blue-500/30 overflow-hidden">
      <Toast toasts={toasts} removeToast={removeToast} />
      <LogViewer logs={logs} isOpen={isLogOpen} onClose={() => setIsLogOpen(false)} />

      {/* Vertical Navigation (VS Code Activity Bar style) */}
      <aside className="w-20 flex flex-col items-center py-8 gap-8 bg-[#1e293b] border-r border-white/5 z-20 shrink-0">
        <button 
          onClick={() => setActiveTab('activity')}
          title="Activity" 
          className={cn(
            "p-4 rounded-2xl transition-all relative group",
            activeTab === 'activity' ? "bg-blue-600 text-white shadow-xl shadow-blue-900/30 ring-4 ring-blue-600/10" : "text-slate-400 hover:bg-white/5 hover:text-white"
          )}
        >
          {activeTab === 'activity' && <div className="absolute left-[-20px] top-4 bottom-4 w-1 bg-blue-500 rounded-r-full" />}
          <Home size={24} />
        </button>
        
        <button 
          onClick={() => setActiveTab('plugins')}
          title="Plugins" 
          className={cn(
            "p-4 rounded-2xl transition-all relative group",
            activeTab === 'plugins' ? "bg-blue-600 text-white shadow-xl shadow-blue-900/30 ring-4 ring-blue-600/10" : "text-slate-400 hover:bg-white/5 hover:text-white"
          )}
        >
          {activeTab === 'plugins' && <div className="absolute left-[-20px] top-4 bottom-4 w-1 bg-blue-500 rounded-r-full" />}
          <Download size={24} />
        </button>

        <div className="mt-auto flex flex-col gap-6">
          <button onClick={() => setIsLogOpen(true)} className="p-4 text-slate-400 hover:text-white transition-colors">
            <List size={22} />
          </button>
        </div>
      </aside>

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col overflow-hidden relative">
        {/* Backdrop gradients */}
        <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-600/5 blur-[120px] rounded-full animate-pulse pointer-events-none" />
        <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-indigo-600/5 blur-[120px] rounded-full animate-pulse pointer-events-none" style={{ animationDelay: '1s' }} />

        {/* Top Header */}
        <header className="h-16 flex items-center justify-between px-10 bg-transparent shrink-0">
          <div className="space-y-0.5">
            <h2 className="text-2xl font-black text-white tracking-tight uppercase italic">{activeTab === 'activity' ? 'Activity Center' : 'Plugin Management'}</h2>
            <div className="flex items-center gap-2">
               <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.8)]" />
               <p className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">{activeTab === 'activity' ? 'Service Monitor' : 'Installed Extensions'}</p>
            </div>
          </div>

          <div className="flex items-center gap-3">
             <button onClick={handleStartAll} className="flex items-center gap-2 px-5 py-2.5 bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 rounded-xl transition-all text-[11px] font-black uppercase tracking-widest border border-emerald-500/20 group">
                <Play size={14} fill="currentColor" className="group-hover:scale-110 transition-transform" /> Start All
             </button>
             <button onClick={handleStopAll} className="flex items-center gap-2 px-5 py-2.5 bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 rounded-xl transition-all text-[11px] font-black uppercase tracking-widest border border-rose-500/20 group">
                <Square size={14} fill="currentColor" className="group-hover:scale-110 transition-transform" /> Stop All
             </button>
             <div className="w-px h-6 bg-white/10 mx-1" />
             <button onClick={handleTerminal} className="p-2.5 bg-slate-800/50 hover:bg-slate-700/50 rounded-xl text-slate-400 hover:text-white transition-all border border-white/5">
                <TerminalIcon size={20} />
             </button>
          </div>
        </header>

        {/* Dynamic Tab Content */}
        <main className="flex-1 overflow-y-auto p-10 scrollbar-thin scrollbar-thumb-white/5">
          <div className="max-w-5xl mx-auto space-y-4">
            {activeTab === 'activity' ? (
              <div className="space-y-4 animate-in fade-in slide-in-from-bottom-4 duration-500">
                {services.map((service) => {
                  const task = prerequisites.find(p => p.name === service.name);
                  const isInstalled = task?.installedVers && task.installedVers.length > 0;

                  return (
                    <div key={service.name} className="bg-slate-900/40 backdrop-blur-xl rounded-[2rem] p-6 border border-white/5 hover:border-white/10 transition-all group flex items-center gap-8 relative shadow-xl">
                      <div className={cn(
                        "w-16 h-16 rounded-[1.5rem] flex items-center justify-center shadow-2xl transition-transform group-hover:scale-105",
                        service.status === 'Running' ? "bg-emerald-500/10 text-emerald-400" : "bg-slate-800 text-slate-400"
                      )}>
                        <service.icon size={28} />
                      </div>

                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-4 flex-wrap">
                          <h3 className="text-xl font-black text-white uppercase italic tracking-tighter">{service.name}</h3>
                          <div className={cn(
                            "text-[10px] font-black uppercase tracking-[0.2em] px-3 py-1 rounded-lg border flex items-center gap-2",
                            service.status === 'Running'
                              ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                              : "bg-slate-900/80 text-slate-500 border-white/5"
                          )}>
                            {service.status === 'Running' && <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse shadow-[0_0_8px_rgba(16,185,129,0.8)]" />}
                            {service.status}
                          </div>
                        </div>
                        <p className="text-[11px] font-medium text-slate-500 mt-1.5 uppercase tracking-wider">
                          {isInstalled ? `Version ${task?.version || 'Unknown'} - Stable` : 'Component missing'}
                        </p>
                      </div>

                      <div className="flex items-center gap-8">
                        {!isInstalled ? (
                          <button onClick={() => setActiveTab('plugins')} className="px-6 py-2 bg-blue-600/10 hover:bg-blue-600/20 text-blue-400 rounded-xl text-[10px] font-black uppercase tracking-widest transition-all border border-blue-500/20">Install First</button>
                        ) : (
                          <button
                            onClick={() => handleToggleService(service.name, service.status)}
                            className={cn(
                              "w-14 h-7 rounded-full p-1 transition-all duration-300 ease-in-out relative ring-1 ring-inset shadow-2xl",
                              service.status === 'Running' 
                                ? "bg-gradient-to-r from-emerald-500 to-teal-500 ring-emerald-400/50" 
                                : "bg-slate-800 ring-white/5"
                            )}
                          >
                            <div className={cn(
                              "w-5 h-5 bg-white rounded-full transition-all duration-300 shadow-xl",
                              service.status === 'Running' ? "translate-x-7" : "translate-x-0"
                            )} />
                          </button>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="space-y-4 animate-in fade-in slide-in-from-bottom-4 duration-500">
                {prerequisites.map((task) => {
                  const progress = downloadProgress[task.name];
                  const isActive = progress && progress.percentage > 0 && progress.percentage < 100;
                  const isDropdownOpen = openDropdown === task.name;
                  
                  return (
                    <div 
                      key={task.name} 
                      className={cn(
                        "bg-slate-900/40 backdrop-blur-xl rounded-[2rem] p-6 border border-white/5 hover:border-white/10 transition-all group flex items-center gap-8 relative shadow-xl",
                        isDropdownOpen ? "z-[100]" : "z-0"
                      )}
                    >
                      <div className={cn(
                        "w-16 h-16 rounded-[1.5rem] flex items-center justify-center shadow-2xl transition-transform group-hover:scale-105",
                        task.isInstalled ? "bg-emerald-500/10 text-emerald-400" : "bg-blue-500/10 text-blue-400"
                      )}>
                        {task.name === 'Apache' && <Globe size={28} />}
                        {task.name === 'MySQL' && <Database size={28} />}
                        {task.name === 'PHP' && <Settings size={28} />}
                        {task.name === 'HeidiSQL' && <ExternalLink size={28} />}
                      </div>

                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-4 flex-wrap">
                          <h3 className="text-xl font-black text-white uppercase italic tracking-tighter">{task.name}</h3>
                          
                          {task.versions ? (
                            <VersionDropdown 
                              current={selectedVersions[task.name] || task.version}
                              options={[...new Set([...task.versions, ...(task.installedVers || [])])]}
                              isOpen={isDropdownOpen}
                              onToggle={() => setOpenDropdown(isDropdownOpen ? null : task.name)}
                              onChange={(v) => setSelectedVersions(prev => ({ ...prev, [task.name]: v }))}
                            />
                          ) : (
                            <span className="text-[11px] font-bold text-slate-500 uppercase tracking-widest bg-white/5 px-2 py-0.5 rounded-md">v{task.version}</span>
                          )}

                          {task.installedVers && task.installedVers.map(ver => (
                            <div 
                              key={ver} 
                              onClick={(e) => { e.stopPropagation(); handleDeleteVersion(task.name, ver); }}
                              className="group/tag flex items-center gap-2 px-2.5 py-1 bg-slate-800/80 border border-white/10 text-slate-400 text-[10px] font-bold uppercase tracking-widest rounded-lg hover:border-rose-500/30 hover:bg-rose-500/10 hover:text-rose-400 transition-all cursor-pointer shadow-sm"
                              title={`Delete v${ver}`}
                            >
                              <Trash2 size={12} className="w-0 opacity-0 group-hover/tag:w-3 group-hover/tag:opacity-100 transition-all text-rose-500" />
                              v{ver}
                            </div>
                          ))}
                        </div>
                        <p className="text-[11px] font-medium text-slate-500 mt-1.5 uppercase tracking-wider">
                          {isActive ? progress.status : task.isInstalled ? 'Component verified & optimized' : 'Available for installation'}
                        </p>
                      </div>

                      <div className="flex items-center gap-8">
                        {isActive && (
                          <div className="animate-in fade-in slide-in-from-right-4 duration-300">
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
                          const isSelectedInstalled = task.installedVers?.includes(selectedVer);
                          
                          if (isActive) return null;

                          return (
                            <button
                              disabled={isSelectedInstalled}
                              onClick={() => !isSelectedInstalled && handleInstallSingle(task)}
                              className={cn(
                                "px-8 py-3 rounded-2xl text-[11px] font-black uppercase tracking-widest transition-all flex items-center justify-center gap-2.5 shadow-2xl",
                                isSelectedInstalled 
                                  ? "bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 shadow-inner cursor-not-allowed"
                                  : "bg-blue-600 hover:bg-blue-500 text-white shadow-blue-600/20 hover:scale-105"
                              )}
                            >
                              {isSelectedInstalled && <CheckCircle2 size={16} />}
                              {isSelectedInstalled ? 'Ready' : 'Download'}
                            </button>
                          );
                        })()}
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </div>
        </main>

        {/* Mini Footer / Status Bar */}
        <footer className="h-8 bg-[#1e293b] border-t border-white/5 flex items-center justify-between px-6 shrink-0 z-10">
           <div className="flex items-center gap-6">
              <div className="flex items-center gap-2 text-[10px] font-bold text-slate-500 uppercase tracking-widest">
                 <Globe size={12} />
                 Ostenia Runtime 2026
              </div>
              <div className="flex items-center gap-2 text-[10px] font-bold text-emerald-500 uppercase tracking-widest">
                 <CheckCircle2 size={12} />
                 Environment Stable
              </div>
           </div>
           <div className="flex items-center gap-4 text-[10px] font-bold text-slate-600 uppercase tracking-[0.2em]">
              <span>UTF-8</span>
              <span>Go / React Stack</span>
           </div>
        </footer>
      </div>
    </div>
  );
}

export default App;
