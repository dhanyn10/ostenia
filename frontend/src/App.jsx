import { useState, useEffect } from 'react';
import { Play, Square, Download, Settings, Terminal as TerminalIcon, Database, Globe, FolderOpen, MoreVertical, ExternalLink, CheckCircle2, AlertCircle, XCircle, X, Loader2, List } from 'lucide-react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { GetPrerequisites, InstallPrerequisite, CancelDownload, StartAllServices, StopAllServices, OpenTerminal } from '../wailsjs/go/main/App';
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
      {/* Default State: Stats + Circle */}
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

      {/* Hover State: Large Cancel Button */}
      <div className="absolute inset-0 flex items-center justify-center translate-y-full group-hover/progress:translate-y-0 transition-transform duration-200 bg-rose-600 rounded-xl">
         <div className="flex items-center gap-2 px-6">
            <X size={14} className="text-white" />
            <span className="text-[10px] font-black text-white uppercase tracking-widest">Cancel</span>
         </div>
      </div>
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
  const [services, setServices] = useState([
    { name: 'Apache', status: 'Stopped', icon: Globe },
    { name: 'MySQL', status: 'Stopped', icon: Database },
  ]);
  const [installing, setInstalling] = useState(false);
  const [prerequisites, setPrerequisites] = useState([]);
  const [downloadProgress, setDownloadProgress] = useState({});
  const [toasts, setToasts] = useState([]);
  const [logs, setLogs] = useState([]);
  const [isLogOpen, setIsLogOpen] = useState(false);

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

  const refreshPrerequisites = () => {
    if (window.go) {
      GetPrerequisites().then(tasks => {
        setPrerequisites(tasks);
        const initialProgress = {};
        tasks.forEach(t => {
          if (t.isInstalled) {
            initialProgress[t.name] = { name: t.name, percentage: 100, status: 'Installed' };
          }
        });
        setDownloadProgress(prev => ({ ...initialProgress, ...prev }));
      });
    } else {
      const mockTasks = [
        { name: 'PHP', version: '8.3.6', isInstalled: false },
        { name: 'Apache', version: '2.4.59', isInstalled: false },
        { name: 'MySQL', version: '8.0.37', isInstalled: false },
        { name: 'HeidiSQL', version: '12.7', isInstalled: false },
      ];
      setPrerequisites(mockTasks);
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
        if (data.percentage === 100 && data.status === 'Completed') {
          addToast(data.name, 'Installed successfully', 'success');
          addLog(`${data.name} installation completed`);
          refreshPrerequisites();
        }
      });

      EventsOn('download_error', (data) => {
        addToast(`${data.name} Error`, data.error, 'error');
        addLog(`Error installing ${data.name}: ${data.error}`);
        setInstalling(false);
      });
      
      // Capture detailed downloader logs from backend fmt.Printf
      // (Note: This requires backend to emit them, for now we manually add from events)
    }
  }, []);

  const handleStartAll = () => { addLog('Starting all services...'); if (window.go) StartAllServices(); };
  const handleStopAll = () => { addLog('Stopping all services...'); if (window.go) StopAllServices(); };
  const handleTerminal = () => { addLog('Opening terminal...'); if (window.go) OpenTerminal(); };

  const handleCancel = (name) => {
    addLog(`Requesting cancellation for ${name}...`);
    if (window.go) {
      CancelDownload(name);
      // Immediately reset local state for this item to stop the UI from moving
      setDownloadProgress(prev => ({
        ...prev,
        [name]: { name, percentage: 0, status: 'Cancelled', speed: '', downloaded: '' }
      }));
    }
  };

  const handleInstallSingle = async (task) => {
    addLog(`Initiating installation for ${task.name}...`);
    if (window.go) {
      try {
        await InstallPrerequisite(task);
      } catch (err) {
        addLog(`Installation process for ${task.name} ended: ${err}`);
      }
    } else {
      // Mock progress
      const steps = ['Downloading...', 'Extracting 1/10...', 'Extracting 5/10...', 'Completed'];
      for (let i = 0; i <= 100; i += 25) {
        setDownloadProgress(prev => ({
          ...prev,
          [task.name]: { 
            name: task.name, 
            percentage: i, 
            status: steps[Math.floor(i / 26)],
            speed: i < 75 ? '15.4 MB/s' : '',
            downloaded: i < 75 ? `${(i * 1.5).toFixed(1)} MB` : '150 MB'
          }
        }));
        await new Promise(r => setTimeout(r, i === 0 ? 500 : 300));
      }
      addToast(task.name, 'Installed successfully (MOCK)', 'success');
    }
    refreshPrerequisites();
  };

  const someInstalled = prerequisites.some(p => p.isInstalled);
  const essentialReady = prerequisites.length > 0 && prerequisites.filter(p => !['HeidiSQL'].includes(p.name)).every(p => p.isInstalled);
  const allReady = prerequisites.length > 0 && prerequisites.every(p => p.isInstalled);

  if (!allReady) {
    return (
      <div className="min-h-screen bg-[#0f172a] text-slate-200 font-sans selection:bg-blue-500/30 overflow-hidden relative flex flex-col">
        {/* Toast system */}
        <Toast toasts={toasts} removeToast={removeToast} />
        <LogViewer logs={logs} isOpen={isLogOpen} onClose={() => setIsLogOpen(false)} />

        {/* Backdrop gradients */}
        <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-600/10 blur-[120px] rounded-full animate-pulse" />
        <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-indigo-600/10 blur-[120px] rounded-full animate-pulse" style={{ animationDelay: '1s' }} />

        <div className="flex-1 flex flex-col items-center justify-center p-8 relative z-10">
          <div className="max-w-3xl w-full space-y-10">
            <div className="flex items-center justify-between">
              <div className="space-y-1">
                <h1 className="text-4xl font-black text-white tracking-tighter italic uppercase flex items-center gap-3">
                  <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center -rotate-6">
                    <Download size={20} className="text-white" />
                  </div>
                  SetupStack
                </h1>
                <p className="text-slate-400 text-xs font-bold uppercase tracking-[0.2em] ml-11">Portable Dev Environment</p>
              </div>
              <button 
                onClick={() => setIsLogOpen(true)}
                className="p-3 bg-white/5 hover:bg-white/10 rounded-2xl border border-white/5 text-slate-400 hover:text-white transition-all flex items-center gap-2 group"
              >
                <List size={18} className="group-hover:rotate-12 transition-transform" />
                <span className="text-[10px] font-black uppercase tracking-widest px-1">View Logs</span>
              </button>
            </div>

            <div className="space-y-3">
              {prerequisites.map((task) => {
                const progress = downloadProgress[task.name];
                const isActive = progress && progress.percentage > 0 && progress.percentage < 100;
                
                return (
                  <div key={task.name} className="bg-slate-900/40 backdrop-blur-md rounded-2xl p-4 border border-white/5 hover:border-white/10 transition-all group flex items-center gap-6">
                    <div className={cn(
                      "w-12 h-12 rounded-xl flex items-center justify-center shadow-lg transition-transform group-hover:scale-105",
                      task.isInstalled ? "bg-emerald-500/10 text-emerald-400" : "bg-blue-500/10 text-blue-400"
                    )}>
                      {task.name === 'Apache' && <Globe size={22} />}
                      {task.name === 'MySQL' && <Database size={22} />}
                      {task.name === 'PHP' && <Settings size={22} />}
                      {task.name === 'HeidiSQL' && <ExternalLink size={22} />}
                    </div>

                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-3">
                        <h3 className="font-bold text-white text-sm tracking-tight uppercase italic">{task.name}</h3>
                        <span className="text-[10px] font-bold text-slate-500 uppercase tracking-widest">v{task.version}</span>
                        {task.isInstalled && (
                           <div className="flex items-center gap-1.5 px-2 py-0.5 bg-emerald-500/10 text-emerald-400 text-[8px] font-black uppercase tracking-widest rounded-md border border-emerald-500/20">
                             <CheckCircle2 size={10} />
                             Ready
                           </div>
                        )}
                      </div>
                      <p className="text-[10px] font-medium text-slate-500 mt-1">
                        {isActive ? progress.status : task.isInstalled ? 'Component verified and linked' : 'Pending installation'}
                      </p>
                    </div>

                    <div className="flex items-center gap-6">
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

                      {!task.isInstalled && !isActive && (
                        <button
                          onClick={() => handleInstallSingle(task)}
                          className="px-6 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-[10px] font-black uppercase tracking-widest shadow-lg shadow-blue-600/20 transition-all hover:scale-105"
                        >
                          Download
                        </button>
                      )}
                      
                      {task.isInstalled && (
                        <div className="w-10 h-10 bg-emerald-500/10 rounded-xl flex items-center justify-center text-emerald-400 border border-emerald-500/20 shadow-inner">
                          <CheckCircle2 size={20} />
                        </div>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>

            {essentialReady && (
              <button
                onClick={() => refreshPrerequisites()}
                className="w-full py-5 bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white rounded-[2rem] font-black text-sm uppercase tracking-[0.3em] transition-all shadow-2xl shadow-emerald-600/20 flex items-center justify-center gap-4 group"
              >
                Launch Dashboard
                <ExternalLink size={20} className="group-hover:translate-x-1 group-hover:-translate-y-1 transition-transform" />
              </button>
            )}

            <div className="flex items-center justify-center gap-8 text-slate-600">
               <div className="h-px flex-1 bg-white/5" />
               <p className="text-[10px] font-black uppercase tracking-[0.5em] whitespace-nowrap">Ostenia Core Engine 2026</p>
               <div className="h-px flex-1 bg-white/5" />
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-screen bg-[#0f172a] text-slate-200 font-sans selection:bg-blue-500/30">
      {/* Top Navbar */}
      <header className="h-16 flex items-center justify-between px-8 bg-[#1e293b] border-b border-white/5 shadow-2xl z-20">
        <div className="flex items-center gap-6">
          <div className="flex items-center gap-3 group cursor-pointer">
            <div className="w-9 h-9 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-xl flex items-center justify-center font-black text-white shadow-lg group-hover:scale-110 transition-all duration-300">O</div>
            <h1 className="text-xl font-black tracking-tighter text-white uppercase italic">Ostenia</h1>
          </div>
          <div className="h-6 w-px bg-white/10 mx-2" />
          <nav className="flex items-center gap-2">
            <button onClick={handleStartAll} className="flex items-center gap-2 px-4 py-2 hover:bg-emerald-500/10 text-emerald-400 rounded-xl transition-all text-[11px] font-black uppercase tracking-widest border border-transparent hover:border-emerald-500/20">
              <Play size={14} fill="currentColor" /> Start All
            </button>
            <button onClick={handleStopAll} className="flex items-center gap-2 px-4 py-2 hover:bg-rose-500/10 text-rose-400 rounded-xl transition-all text-[11px] font-black uppercase tracking-widest border border-transparent hover:border-rose-500/20">
              <Square size={14} fill="currentColor" /> Stop All
            </button>
          </nav>
        </div>

        <div className="flex items-center gap-4">
          <button onClick={handleTerminal} title="Open Terminal" className="p-2.5 bg-slate-800/50 border border-white/5 hover:bg-slate-700/50 rounded-xl transition-all text-slate-400 hover:text-white group">
            <TerminalIcon size={20} className="group-hover:scale-110 transition-transform" />
          </button>
          <button className="p-2.5 bg-slate-800/50 border border-white/5 hover:bg-slate-700/50 rounded-xl transition-all text-slate-400 hover:text-white group">
            <Settings size={20} className="group-hover:rotate-45 transition-transform" />
          </button>
        </div>
      </header>

      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar Mini */}
        <aside className="w-20 flex flex-col items-center py-8 gap-8 bg-[#1e293b]/50 border-r border-white/5 z-10 transition-all">
          <button title="Web Server" className="p-4 bg-blue-600 rounded-2xl text-white shadow-xl shadow-blue-900/30 ring-4 ring-blue-600/10 hover:scale-105 transition-all">
            <Globe size={24} />
          </button>
          <button title="Database" className="p-4 text-slate-400 hover:bg-white/5 hover:text-white rounded-2xl transition-all">
            <Database size={24} />
          </button>
          <button title="Project Directory" className="p-4 text-slate-400 hover:bg-white/5 hover:text-white rounded-2xl transition-all">
            <FolderOpen size={24} />
          </button>
          <div className="mt-auto">
             <div className="w-10 h-10 rounded-full bg-slate-800 border border-white/10 flex items-center justify-center text-[10px] font-black">PH</div>
          </div>
        </aside>

        {/* Main Workspace */}
        <main className="flex-1 flex flex-col p-10 overflow-y-auto bg-slate-950/20">
          <div className="max-w-5xl w-full mx-auto space-y-10">
            <div className="flex items-end justify-between">
              <div className="space-y-1">
                <h2 className="text-4xl font-black text-white tracking-tight">Dashboard</h2>
                <p className="text-slate-500 font-bold uppercase text-[11px] tracking-[0.2em] flex items-center gap-2">
                  <div className="w-2 h-2 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.8)]" />
                  Development Environment Ready
                </p>
              </div>
              <div className="flex gap-3">
                 <button className="px-5 py-2.5 bg-[#1e293b] hover:bg-[#2d3a4f] rounded-xl text-xs font-black uppercase tracking-widest transition-all border border-white/5 flex items-center gap-2.5 shadow-lg group">
                    <ExternalLink size={16} className="text-blue-400 group-hover:scale-110 transition-transform" /> Open Localhost
                 </button>
                 <button className="px-5 py-2.5 bg-[#1e293b] hover:bg-[#2d3a4f] rounded-xl text-xs font-black uppercase tracking-widest transition-all border border-white/5 flex items-center gap-2.5 shadow-lg group">
                    <Database size={16} className="text-emerald-400 group-hover:scale-110 transition-transform" /> Database
                 </button>
              </div>
            </div>

            {/* Service List */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {services.map((service) => (
                <div key={service.name} className="bg-[#1e293b]/70 backdrop-blur-md rounded-3xl border border-white/5 p-6 flex flex-col justify-between hover:bg-[#1e293b] transition-all group relative overflow-hidden">
                  <div className="absolute top-0 right-0 p-6 opacity-[0.03] group-hover:opacity-[0.08] transition-opacity">
                    <service.icon size={120} />
                  </div>
                  <div className="flex items-start justify-between relative z-10">
                    <div className="flex items-center gap-5">
                      <div className={cn(
                        "w-14 h-14 rounded-2xl flex items-center justify-center transition-all shadow-xl",
                        service.status === 'Running' 
                          ? "bg-gradient-to-br from-emerald-500 to-teal-600 text-white shadow-emerald-500/20" 
                          : "bg-slate-800 text-slate-500 border border-white/5"
                      )}>
                        <service.icon size={28} />
                      </div>
                      <div>
                        <h4 className="text-xl font-black text-white group-hover:text-blue-400 transition-colors uppercase italic">{service.name}</h4>
                        <div className="flex items-center gap-3 mt-1">
                          <span className="text-[10px] font-black text-slate-500 uppercase tracking-widest">v{prerequisites.find(p => p.name === service.name)?.version || '...'}</span>
                          <span className="w-1 h-1 rounded-full bg-slate-700" />
                          <button className="text-[10px] text-blue-500 hover:text-blue-400 font-black uppercase tracking-widest transition-colors">Config</button>
                        </div>
                      </div>
                    </div>

                    <div className={cn(
                      "text-[9px] font-black uppercase tracking-[0.2em] px-3 py-1.5 rounded-lg border flex items-center gap-2",
                      service.status === 'Running'
                        ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                        : "bg-slate-900/50 text-slate-500 border-white/5"
                    )}>
                      {service.status === 'Running' && <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />}
                      {service.status}
                    </div>
                  </div>

                  <div className="mt-8 flex items-center justify-between relative z-10">
                    <div className="flex gap-2">
                       {service.name === 'Apache' && (
                         <>
                           <button className="p-2 bg-white/5 hover:bg-white/10 rounded-lg text-slate-400 hover:text-white transition-all"><FolderOpen size={16} /></button>
                           <button className="p-2 bg-white/5 hover:bg-white/10 rounded-lg text-slate-400 hover:text-white transition-all"><Globe size={16} /></button>
                         </>
                       )}
                       {service.name === 'MySQL' && (
                         <button className="p-2 bg-white/5 hover:bg-white/10 rounded-lg text-slate-400 hover:text-white transition-all"><Database size={16} /></button>
                       )}
                    </div>
                    
                    {/* Premium Toggle Switch */}
                    <button
                      onClick={() => {/* Toggle Logic */}}
                      className={cn(
                        "w-14 h-7 rounded-full p-1.5 transition-all duration-300 ease-in-out relative ring-1 ring-inset shadow-inner",
                        service.status === 'Running' 
                          ? "bg-gradient-to-r from-emerald-500 to-teal-500 ring-emerald-400/50" 
                          : "bg-slate-800 ring-white/5"
                      )}
                    >
                      <div className={cn(
                        "w-4 h-4 bg-white rounded-full transition-all duration-300 shadow-md",
                        service.status === 'Running' ? "translate-x-7" : "translate-x-0"
                      )} />
                    </button>
                  </div>
                </div>
              ))}

              {/* Extra Tools Section */}
              <div className="bg-gradient-to-br from-indigo-900/40 to-slate-900/40 backdrop-blur-md rounded-3xl border border-white/5 p-6 flex flex-col gap-6">
                <h4 className="text-xs font-black text-slate-400 uppercase tracking-[0.3em]">Quick Access Tools</h4>
                <div className="grid grid-cols-2 gap-4">
                   <button className="flex items-center gap-3 p-4 bg-white/[0.03] hover:bg-white/[0.08] rounded-2xl border border-white/5 transition-all group">
                      <div className="w-10 h-10 bg-indigo-500/20 text-indigo-400 rounded-xl flex items-center justify-center group-hover:scale-110 transition-transform">
                        <ExternalLink size={20} />
                      </div>
                      <div className="text-left">
                        <span className="block font-bold text-white text-sm">HeidiSQL</span>
                        <span className="text-[9px] text-slate-500 font-bold uppercase">DB Browser</span>
                      </div>
                   </button>
                   <button onClick={handleTerminal} className="flex items-center gap-3 p-4 bg-white/[0.03] hover:bg-white/[0.08] rounded-2xl border border-white/5 transition-all group">
                      <div className="w-10 h-10 bg-slate-500/20 text-slate-400 rounded-xl flex items-center justify-center group-hover:scale-110 transition-transform">
                        <TerminalIcon size={20} />
                      </div>
                      <div className="text-left">
                        <span className="block font-bold text-white text-sm">Terminal</span>
                        <span className="text-[9px] text-slate-500 font-bold uppercase">CMD Shell</span>
                      </div>
                   </button>
                </div>
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}

export default App;
