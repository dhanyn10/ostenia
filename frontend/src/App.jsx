import { useState, useEffect } from 'react';
import { Play, Square, Download, Settings, Terminal as TerminalIcon, Database, Globe, FolderOpen, MoreVertical, ExternalLink } from 'lucide-react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { GetPrerequisites, InstallPrerequisite, StartAllServices, StopAllServices, OpenTerminal } from '../wailsjs/go/main/App';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

function App() {
  const [services, setServices] = useState([
    { name: 'Apache', status: 'Stopped', version: '2.4.58' },
    { name: 'MySQL', status: 'Stopped', version: '8.0.36' },
  ]);
  const [installing, setInstalling] = useState(false);
  const [prerequisites, setPrerequisites] = useState([]);
  const [downloadProgress, setDownloadProgress] = useState({});

  useEffect(() => {
    if (window.go) {
      GetPrerequisites().then(setPrerequisites);
    } else {
      setPrerequisites([
        { Name: 'PHP', URL: '', Version: '8.3.4', Target: 'php/php-8.3.4' },
        { Name: 'Apache', URL: '', Version: '2.4.58', Target: 'apache/httpd-2.4.58' },
        { Name: 'MySQL', URL: '', Version: '8.0.36', Target: 'mysql/mysql-8.0.36' },
        { Name: 'HeidiSQL', URL: '', Version: '12.6', Target: 'heidisql' },
      ]);
    }

    if (window.runtime) {
      EventsOn('service_status', (data) => {
        setServices(prev => prev.map(s => s.name === data.name ? { ...s, status: data.status } : s));
      });

      EventsOn('download_progress', (data) => {
        setDownloadProgress(prev => ({
          ...prev,
          [data.name]: data
        }));
      });
    }
  }, []);

  const handleStartAll = () => { if (window.go) StartAllServices(); };
  const handleStopAll = () => { if (window.go) StopAllServices(); };
  const handleTerminal = () => { if (window.go) OpenTerminal(); };

  const handleInstallAll = async () => {
    setInstalling(true);
    for (const task of prerequisites) {
      if (window.go) {
        await InstallPrerequisite(task);
      } else {
        await new Promise(r => setTimeout(r, 500));
        setDownloadProgress(prev => ({
          ...prev,
          [task.Name]: { name: task.Name, percentage: 100, status: 'Completed (Mock)' }
        }));
      }
    }
    setInstalling(false);
  };

  const isAllInstalled = Object.values(downloadProgress).every(p => p.percentage === 100);

  return (
    <div className="flex flex-col h-screen bg-[#0f172a] text-slate-200 font-sans selection:bg-blue-500/30">
      {/* Top Navbar */}
      <header className="h-14 flex items-center justify-between px-6 bg-[#1e293b] border-b border-white/5 shadow-2xl z-10">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2.5 group">
            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-lg flex items-center justify-center font-black text-white shadow-lg group-hover:scale-105 transition-transform cursor-pointer">O</div>
            <h1 className="text-lg font-bold tracking-tight text-white uppercase italic">Ostenia</h1>
          </div>
          <div className="h-4 w-px bg-white/10 mx-2" />
          <nav className="flex items-center gap-1">
            <button onClick={handleStartAll} className="flex items-center gap-2 px-3 py-1.5 hover:bg-emerald-500/10 text-emerald-400 rounded-md transition-all text-xs font-bold uppercase tracking-wider">
              <Play size={14} fill="currentColor" /> Start
            </button>
            <button onClick={handleStopAll} className="flex items-center gap-2 px-3 py-1.5 hover:bg-rose-500/10 text-rose-400 rounded-md transition-all text-xs font-bold uppercase tracking-wider">
              <Square size={14} fill="currentColor" /> Stop
            </button>
          </nav>
        </div>

        <div className="flex items-center gap-3">
          <button onClick={handleTerminal} title="Open Terminal" className="p-2 hover:bg-white/5 rounded-full transition-colors text-slate-400 hover:text-white">
            <TerminalIcon size={20} />
          </button>
          <button className="p-2 hover:bg-white/5 rounded-full transition-colors text-slate-400 hover:text-white">
            <Settings size={20} />
          </button>
        </div>
      </header>

      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar Mini */}
        <aside className="w-16 flex flex-col items-center py-6 gap-6 bg-[#1e293b]/50 border-r border-white/5">
          <button title="Web Server" className="p-3 bg-blue-600 rounded-xl text-white shadow-xl shadow-blue-900/20">
            <Globe size={22} />
          </button>
          <button title="Database" className="p-3 text-slate-500 hover:text-white transition-colors">
            <Database size={22} />
          </button>
          <button title="Project Directory" className="p-3 text-slate-500 hover:text-white transition-colors">
            <FolderOpen size={22} />
          </button>
        </aside>

        {/* Main Workspace */}
        <main className="flex-1 flex flex-col p-8 overflow-y-auto">
          <div className="max-w-4xl w-full mx-auto space-y-8">
            <div className="flex items-end justify-between">
              <div>
                <h2 className="text-3xl font-black text-white tracking-tight">Dashboard</h2>
                <p className="text-slate-500 font-medium">Manage your local development environment</p>
              </div>
              <div className="flex gap-2">
                 <button className="px-4 py-2 bg-[#1e293b] hover:bg-[#2d3a4f] rounded-lg text-sm font-bold transition-colors border border-white/5 flex items-center gap-2">
                    <ExternalLink size={14} /> Web
                 </button>
                 <button className="px-4 py-2 bg-[#1e293b] hover:bg-[#2d3a4f] rounded-lg text-sm font-bold transition-colors border border-white/5 flex items-center gap-2">
                    <Database size={14} /> DB
                 </button>
              </div>
            </div>

            {/* Service List */}
            <div className="bg-[#1e293b] rounded-2xl border border-white/5 shadow-xl overflow-hidden">
              <div className="px-6 py-4 border-b border-white/5 bg-white/[0.02] flex justify-between items-center text-[10px] font-black uppercase tracking-[0.2em] text-slate-500">
                <span>Service Name</span>
                <div className="flex gap-20 mr-24">
                   <span>Version</span>
                   <span>Status</span>
                </div>
              </div>
              <div className="divide-y divide-white/5">
                {services.map((service) => (
                  <div key={service.name} className="px-6 py-5 flex items-center justify-between hover:bg-white/[0.01] transition-colors group">
                    <div className="flex items-center gap-4">
                      <div className={cn(
                        "w-10 h-10 rounded-xl flex items-center justify-center transition-colors shadow-inner",
                        service.status === 'Running' ? "bg-emerald-500/10 text-emerald-400" : "bg-slate-700/20 text-slate-500"
                      )}>
                        {service.name === 'Apache' ? <Globe size={20} /> : <Database size={20} />}
                      </div>
                      <div>
                        <h4 className="font-bold text-white group-hover:text-blue-400 transition-colors">{service.name}</h4>
                        <button className="text-[10px] text-blue-500 hover:underline font-bold uppercase tracking-widest mt-0.5">Edit Config</button>
                      </div>
                    </div>

                    <div className="flex items-center gap-16 mr-8">
                      <span className="text-sm font-mono text-slate-500">{service.version}</span>

                      <div className="flex items-center gap-4">
                        <span className={cn(
                          "text-[10px] font-black uppercase tracking-widest px-2 py-0.5 rounded border transition-all",
                          service.status === 'Running'
                            ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20 shadow-lg shadow-emerald-900/10"
                            : "bg-slate-800 text-slate-500 border-white/5"
                        )}>
                          {service.status}
                        </span>

                        {/* Toggle Switch */}
                        <button
                          className={cn(
                            "w-11 h-6 rounded-full p-1 transition-colors duration-200 ease-in-out relative",
                            service.status === 'Running' ? "bg-emerald-600" : "bg-slate-700"
                          )}
                        >
                          <div className={cn(
                            "w-4 h-4 bg-white rounded-full transition-transform duration-200 shadow-sm",
                            service.status === 'Running' ? "translate-x-5" : "translate-x-0"
                          )} />
                        </button>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Installation Overlay */}
            {!isAllInstalled && (
              <div className="relative overflow-hidden bg-gradient-to-br from-blue-600 to-indigo-700 rounded-2xl p-8 shadow-2xl shadow-blue-900/40 group">
                <div className="absolute top-0 right-0 p-8 text-white/10 group-hover:scale-110 transition-transform duration-500">
                  <Download size={120} />
                </div>
                <div className="relative z-10 space-y-6">
                  <div>
                    <h3 className="text-2xl font-black text-white italic tracking-tight">READY TO START?</h3>
                    <p className="text-blue-100/80 font-medium">Ostenia needs to download the core server engine.</p>
                  </div>

                  <div className="flex flex-col gap-4">
                    <button
                      onClick={handleInstallAll}
                      disabled={installing}
                      className="w-fit px-8 py-3 bg-white text-blue-700 rounded-xl font-black text-sm uppercase tracking-widest hover:bg-blue-50 transition-all shadow-xl disabled:opacity-50"
                    >
                      {installing ? 'Downloading...' : 'Start Installation'}
                    </button>

                    {installing && (
                      <div className="grid grid-cols-2 gap-x-8 gap-y-4 max-w-lg mt-4">
                        {Object.entries(downloadProgress).map(([name, prog]) => (
                          <div key={name} className="space-y-1.5">
                            <div className="flex justify-between text-[9px] uppercase font-black tracking-[0.15em] text-white/60">
                              <span>{name}</span>
                              <span>{Math.round(prog.percentage)}%</span>
                            </div>
                            <div className="w-full bg-black/20 rounded-full h-1.5 overflow-hidden ring-1 ring-white/10">
                              <div
                                className="bg-white h-full transition-all duration-300 shadow-[0_0_10px_rgba(255,255,255,0.5)]"
                                style={{ width: `${prog.percentage}%` }}
                              />
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            )}
          </div>
        </main>
      </div>
    </div>
  );
}

export default App;
