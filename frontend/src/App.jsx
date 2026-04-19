import { useState, useEffect } from 'react';
import { Play, Square, Download, Settings, Terminal as TerminalIcon, Database, Globe, FolderOpen } from 'lucide-react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { GetPrerequisites, InstallPrerequisite, StartAllServices, StopAllServices } from '../wailsjs/go/main/App';

function App() {
  const [services, setServices] = useState([
    { name: 'Apache', status: 'Stopped', version: '2.4.58' },
    { name: 'MySQL', status: 'Stopped', version: '8.0.36' },
  ]);
  const [logs, setLogs] = useState([]);
  const [installing, setInstalling] = useState(false);
  const [prerequisites, setPrerequisites] = useState([]);
  const [downloadProgress, setDownloadProgress] = useState({});

  useEffect(() => {
    if (window.go) {
      GetPrerequisites().then(setPrerequisites);
    } else {
      console.warn("Wails Go bridge not found. Using mock prerequisites.");
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

      EventsOn('service_log', (data) => {
        setLogs(prev => [...prev.slice(-100), `[${data.service}] ${data.message}`]);
      });

      EventsOn('download_progress', (data) => {
        setDownloadProgress(prev => ({
          ...prev,
          [data.name]: data
        }));
      });
    }
  }, []);

  const handleStartAll = () => {
    if (window.go) StartAllServices();
  };

  const handleStopAll = () => {
    if (window.go) StopAllServices();
  };

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
    <div className="flex flex-col h-screen bg-slate-900 text-slate-100 font-sans">
      {/* Header */}
      <header className="flex items-center justify-between px-6 py-4 bg-slate-800 border-b border-slate-700">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center font-bold text-xl">O</div>
          <h1 className="text-xl font-bold tracking-tight">Ostenia</h1>
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleStartAll}
            className="flex items-center gap-2 px-4 py-2 bg-emerald-600 hover:bg-emerald-500 rounded-md transition-colors font-medium text-sm"
          >
            <Play size={16} fill="currentColor" /> Start All
          </button>
          <button
            onClick={handleStopAll}
            className="flex items-center gap-2 px-4 py-2 bg-rose-600 hover:bg-rose-500 rounded-md transition-colors font-medium text-sm"
          >
            <Square size={16} fill="currentColor" /> Stop All
          </button>
        </div>
      </header>

      <main className="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        <aside className="w-64 bg-slate-800/50 border-r border-slate-700 p-4 flex flex-col gap-2">
          <button className="flex items-center gap-3 px-3 py-2 bg-blue-600/20 text-blue-400 rounded-md text-sm font-medium">
            <Globe size={18} /> Web Server
          </button>
          <button className="flex items-center gap-3 px-3 py-2 hover:bg-slate-700/50 rounded-md text-sm font-medium transition-colors">
            <Database size={18} /> Database
          </button>
          <button className="flex items-center gap-3 px-3 py-2 hover:bg-slate-700/50 rounded-md text-sm font-medium transition-colors">
            <FolderOpen size={18} /> Root Directory
          </button>
          <div className="mt-auto">
            <button className="w-full flex items-center gap-3 px-3 py-2 hover:bg-slate-700/50 rounded-md text-sm font-medium transition-colors">
              <Settings size={18} /> Settings
            </button>
          </div>
        </aside>

        {/* Content */}
        <section className="flex-1 flex flex-col overflow-hidden">
          {/* Dashboard Grid */}
          <div className="p-6 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {services.map(service => (
              <div key={service.name} className="bg-slate-800 border border-slate-700 rounded-xl p-4 flex flex-col gap-3 shadow-sm">
                <div className="flex justify-between items-start">
                  <div>
                    <h3 className="font-bold text-lg">{service.name}</h3>
                    <p className="text-xs text-slate-400">Version {service.version}</p>
                  </div>
                  <span className={`px-2 py-1 rounded text-[10px] font-bold uppercase tracking-wider ${
                    service.status === 'Running' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'
                  }`}>
                    {service.status}
                  </span>
                </div>
                <div className="flex gap-2 mt-2">
                  <button className="flex-1 py-1.5 bg-slate-700 hover:bg-slate-600 rounded text-xs font-medium transition-colors">Config</button>
                  <button className="flex-1 py-1.5 bg-slate-700 hover:bg-slate-600 rounded text-xs font-medium transition-colors">Logs</button>
                </div>
              </div>
            ))}

            {/* Installation Card if needed */}
            {!isAllInstalled && (
              <div className="bg-slate-800 border border-blue-500/30 rounded-xl p-4 flex flex-col gap-3 shadow-lg ring-1 ring-blue-500/20">
                <h3 className="font-bold text-lg flex items-center gap-2">
                  <Download size={18} className="text-blue-400" /> Prerequisites
                </h3>
                <p className="text-sm text-slate-400 leading-relaxed">Required binaries (Apache, PHP, MySQL) are missing.</p>
                <button
                  onClick={handleInstallAll}
                  disabled={installing}
                  className="w-full py-2 bg-blue-600 hover:bg-blue-500 disabled:bg-slate-700 rounded-md text-sm font-bold transition-all shadow-md shadow-blue-900/20"
                >
                  {installing ? 'Installing...' : 'Install Now'}
                </button>
                {installing && (
                  <div className="space-y-2 mt-1">
                    {Object.entries(downloadProgress).map(([name, prog]) => (
                      <div key={name} className="space-y-1">
                        <div className="flex justify-between text-[10px] uppercase font-bold tracking-tight text-slate-500">
                          <span>{name}</span>
                          <span>{Math.round(prog.percentage)}%</span>
                        </div>
                        <div className="w-full bg-slate-700 rounded-full h-1.5 overflow-hidden">
                          <div
                            className="bg-blue-500 h-full transition-all duration-300 ease-out"
                            style={{ width: `${prog.percentage}%` }}
                          />
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>

          {/* Terminal */}
          <div className="flex-1 m-6 mt-0 bg-black/50 rounded-xl border border-slate-700 flex flex-col overflow-hidden shadow-inner">
            <div className="px-4 py-2 border-b border-slate-700 bg-slate-800/30 flex items-center justify-between">
              <div className="flex items-center gap-2 text-xs font-bold text-slate-400 uppercase tracking-widest">
                <TerminalIcon size={14} /> Console Output
              </div>
              <button
                onClick={() => setLogs([])}
                className="text-[10px] bg-slate-700 hover:bg-slate-600 px-2 py-0.5 rounded transition-colors"
              >
                Clear
              </button>
            </div>
            <div className="flex-1 p-4 font-mono text-sm overflow-y-auto scrollbar-thin scrollbar-thumb-slate-700">
              {logs.length === 0 ? (
                <p className="text-slate-600 italic">No output yet...</p>
              ) : (
                logs.map((log, i) => (
                  <div key={i} className="mb-1">
                    <span className="text-slate-500 select-none mr-2">{i+1}</span>
                    <span className="text-slate-300">{log}</span>
                  </div>
                ))
              )}
            </div>
          </div>
        </section>
      </main>
    </div>
  );
}

export default App;
