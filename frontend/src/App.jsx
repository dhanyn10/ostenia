import { useState, useEffect } from 'react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import * as AppBackend from '../wailsjs/go/main/App';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

// Components
import VerticalNav from './components/VerticalNav';
import AppHeader from './components/AppHeader';
import Toast from './components/Toast';
import LogViewer from './components/LogViewer';
import ActivityTab from './components/ActivityTab';
import PluginsTab from './components/PluginsTab';
import Icons from './components/Icons';

// Icons
import { Loader2 } from 'lucide-react';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

const ICON_MAP = {
  'Apache': Icons.Apache,
  'Nginx': Icons.Nginx,
  'MySQL': Icons.MySQL,
  'PHP': Icons.PHP,
  'HeidiSQL': Icons.HeidiSQL,
  'OpenSSL': Icons.OpenSSL,
  'Node.js': Icons.Node,
  'default': Icons.MySQL
};

function App() {
  const [activeTab, setActiveTab] = useState('activity');
  const [theme, setTheme] = useState('light'); 
  const [services, setServices] = useState([
    { name: 'Apache', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'Nginx', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'MySQL', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'PHP', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'Node.js', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'HeidiSQL', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'OpenSSL', status: 'Stopped', pid: 0, port: 0, remainingDays: 0, ports: [], activeVersion: '' },
  ]);
  const [prerequisites, setPrerequisites] = useState([]);
  const [downloadProgress, setDownloadProgress] = useState({});
  const [toasts, setToasts] = useState([]);
  const [logs, setLogs] = useState([]);
  const [isLogOpen, setIsLogOpen] = useState(false);
  const [openDropdown, setOpenDropdown] = useState(null);
  const [isTerminalOpen, setIsTerminalOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [selectedVersions, setSelectedVersions] = useState({});
  const [isAddingPlugin, setIsAddingPlugin] = useState(false);
  
  const [serverRoot, setServerRoot] = useState('');
  const [appsLocation, setAppsLocation] = useState('');
  const [apacheHttps, setApacheHttps] = useState(false);
  const [nginxHttps, setNginxHttps] = useState(false);

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
    try {
      const tasks = await AppBackend.GetPrerequisites();
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
    }
  };

  const loadInitialData = async () => {
    try {
      const cfg = await AppBackend.GetConfig();
      setServerRoot(cfg.wwwRoot || '');
      setAppsLocation(cfg.baseDir || '');
      setApacheHttps(cfg.apacheHttps || false);
      setNginxHttps(cfg.nginxHttps || false);

      const updatedServices = await Promise.all(
        services.map(async (service) => {
           const detail = await AppBackend.GetServiceStatus(service.name);
           return { 
             ...service, 
             status: detail.status, 
             pid: detail.pid, 
             port: detail.port, 
             ports: detail.ports || [], 
             remainingDays: detail.remainingDays || 0,
             activeVersion: detail.activeVersion || ''
           };
        })
      );
      setServices(updatedServices);
    } catch (err) {
      addLog(`Error loading initial data: ${err}`);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    const init = async () => {
      await refreshPrerequisites();
      await loadInitialData();
    };
    init();

    if (window.runtime) {
      EventsOn('service_status', (data) => {
        setServices(prev => prev.map(service => service.name === data.name ? { 
          ...service, 
          status: data.status, 
          pid: data.pid, 
          port: data.port, 
          ports: data.ports || [], 
          remainingDays: data.remainingDays || 0,
          activeVersion: data.activeVersion || ''
        } : service));
      });

      EventsOn('environment_changed', () => {
        loadInitialData();
        refreshPrerequisites();
      });

      EventsOn('download_progress', (data) => {
        setDownloadProgress(prev => ({ ...prev, [data.name]: data }));
        if (data.percentage === 100 && (data.status === 'Completed' || data.status === 'Ready')) {
          refreshPrerequisites();
        }
      });
    }
  }, []);

  const handleBrowseAppsLocation = async () => {
    try {
      const selected = await AppBackend.SelectServerRoot();
      if (selected) {
        setAppsLocation(selected);
        addToast('Apps Location', 'Apps location updated', 'success');
      }
    } catch (err) {
      addLog(`Error selecting apps location: ${err}`);
    }
  };

  const handleBrowseServerRoot = async () => {
    try {
      // Use Wails runtime via window.runtime if available
      if (window.runtime) {
        const selected = await window.runtime.OpenDirectoryDialog({ title: "Select Server Root (www)" });
        if (selected) {
          await AppBackend.SetWWWRoot(selected);
          setServerRoot(selected);
          addToast('Server Root', 'Server root updated', 'success');
        }
      }
    } catch (err) {
      addLog(`Error selecting server root: ${err}`);
    }
  };

  return (
    <div className={cn(
      "flex h-screen font-sans selection:bg-blue-500/30 overflow-hidden transition-colors duration-300",
      theme === 'dark' ? "bg-[#0f172a] text-slate-200" : "bg-slate-50 text-slate-900"
    )}>
      <Toast toasts={toasts} removeToast={removeToast} />
      <LogViewer logs={logs} isOpen={isLogOpen} onClose={() => setIsLogOpen(false)} />

      <VerticalNav 
        activeTab={activeTab} 
        setActiveTab={setActiveTab} 
        toggleTheme={() => setTheme(t => t === 'dark' ? 'light' : 'dark')} 
        theme={theme} 
        setIsLogOpen={setIsLogOpen} 
      />

      <div className="flex-1 flex flex-col min-h-0 overflow-hidden relative">
        <AppHeader 
          activeTab={activeTab} 
          handleStartAll={() => AppBackend.StartAllServices()} 
          handleStopAll={() => AppBackend.StopAllServices()} 
          handleTerminal={(type) => { AppBackend.OpenTerminal(type); setIsTerminalOpen(false); }} 
          isTerminalOpen={isTerminalOpen} 
          setIsTerminalOpen={setIsTerminalOpen} 
        />

        <main className="flex-1 flex flex-col min-h-0 overflow-hidden">
          <div className="max-w-4xl w-full mx-auto flex flex-col h-full px-8 pb-8">
            {loading ? (
              <div className="flex-1 flex items-center justify-center">
                <Loader2 className="animate-spin text-blue-500" size={32} />
              </div>
            ) : activeTab === 'activity' ? (
              <ActivityTab 
                serverRoot={serverRoot}
                appsLocation={appsLocation}
                handleBrowseAppsLocation={handleBrowseAppsLocation}
                handleBrowseServerRoot={handleBrowseServerRoot}
                isAddingPlugin={isAddingPlugin}
                setIsAddingPlugin={setIsAddingPlugin}
                prerequisites={prerequisites}
                services={services}
                handleAddToHome={(task) => {
                  if (!services.find(s => s.name === task.name)) {
                    setServices(prev => [...prev, { name: task.name, status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '', remainingDays: 0 }]);
                  }
                  setIsAddingPlugin(false);
                }}
                ICON_MAP={ICON_MAP}
                handleToggleService={(name, status) => status === 'Running' ? AppBackend.StopService(name) : AppBackend.StartService(name)}
                handleRemoveFromHome={(name) => setServices(prev => prev.filter(s => s.name !== name))}
                setActiveTab={setActiveTab}
                handleOpenPluginFolder={(name) => AppBackend.OpenPluginFolder(name)}
                handleOpenServerRootFolder={() => AppBackend.OpenServerRootFolder()}
                apacheHttps={apacheHttps}
                nginxHttps={nginxHttps}
                handleToggleHttps={async (name) => {
                  if (name === 'Apache') {
                    const next = !apacheHttps; setApacheHttps(next); await AppBackend.SetApacheHTTPS(next);
                  } else {
                    const next = !nginxHttps; setNginxHttps(next); await AppBackend.SetNginxHTTPS(next);
                  }
                }}
              />
            ) : (
              <PluginsTab 
                prerequisites={prerequisites}
                downloadProgress={downloadProgress}
                openDropdown={openDropdown}
                setOpenDropdown={setOpenDropdown}
                selectedVersions={selectedVersions}
                setSelectedVersions={setSelectedVersions}
                handleDeleteVersion={(name, ver) => AppBackend.DeleteVersion(name, ver).then(refreshPrerequisites)}
                handleInstallSingle={async (task) => {
                    const selectedVer = selectedVersions[task.name] || task.version;
                    const arch = navigator.userAgent.includes('Win64') || navigator.userAgent.includes('x64') ? 'x64' : 'x86';
                    const modifiedTask = { ...task, version: selectedVer };
                    if (task.name === 'Node.js') {
                        modifiedTask.target = `nodejs/node-v${selectedVer}`;
                        modifiedTask.url = `https://nodejs.org/dist/v${selectedVer}/node-v${selectedVer}-win-${arch}.zip`;
                    }
                    await AppBackend.InstallPrerequisite(modifiedTask);
                    refreshPrerequisites();
                }}
                handleCancel={(name) => AppBackend.CancelDownload(name)}
                ICON_MAP={ICON_MAP}
              />
            )}
          </div>
        </main>
      </div>
    </div>
  );
}

export default App;
