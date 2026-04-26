import { useState, useEffect } from 'react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import { GetPrerequisites, InstallPrerequisite, CancelDownload, StartAllServices, StopAllServices, OpenTerminal, DeleteVersion, StartService, StopService, GetServerRoot, SetServerRoot, GetServiceStatus, SelectServerRoot, OpenPluginFolder, OpenServerRootFolder, UpdateActiveTab, SetApacheHTTPS, SetNginxHTTPS, GetConfig } from '../wailsjs/go/main/App';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

// Components
import VerticalNav from './components/VerticalNav';
import AppHeader from './components/AppHeader';
import Toast from './components/Toast';
import LogViewer from './components/LogViewer';
import ActivityTab from './components/ActivityTab';
import PluginsTab from './components/PluginsTab';

// Icons
import { Globe, Database, Settings, ExternalLink, Server, Loader2, Shield } from 'lucide-react';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

const ICON_MAP = {
  'Apache': Globe,
  'Nginx': Server,
  'MySQL': Database,
  'PHP': Settings,
  'HeidiSQL': ExternalLink,
  'OpenSSL': Shield,
  'Node.js': Settings,
  'default': Database
};

function App() {
  const [activeTab, setActiveTab] = useState('activity');
  const [theme, setTheme] = useState('dark');
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
  const [apacheHttps, setApacheHttps] = useState(false);
  const [nginxHttps, setNginxHttps] = useState(false);

  const toggleTheme = () => {
    const newTheme = theme === 'dark' ? 'light' : 'dark';
    setTheme(newTheme);
    localStorage.setItem('ostenia-theme', newTheme);
  };

  useEffect(() => {
    const savedTheme = localStorage.getItem('ostenia-theme') || 'dark';
    setTheme(savedTheme);
  }, []);

  useEffect(() => {
    if (theme === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }, [theme]);

  // Sync active tab to backend
  useEffect(() => {
    if (window.go) {
      UpdateActiveTab(activeTab);
    }
  }, [activeTab]);

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
    const loadInitialData = async () => {
      if (window.go) {
        try {
          const cfg = await GetConfig();
          setServerRoot(cfg.wwwRoot);
          setApacheHttps(cfg.apacheHttps);
          setNginxHttps(cfg.nginxHttps);

          const updatedServices = await Promise.all(
            services.map(async (service) => {
               const detail = await GetServiceStatus(service.name);
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
        }
      }
    };

    refreshPrerequisites();
    loadInitialData();

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
        
        if (data.name === 'OpenSSL' && data.status === 'Stopped') {
          setApacheHttps(false);
          setNginxHttps(false);
        }
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
  
  const handleTerminal = (type) => { 
    addLog(`Opening ${type || 'default'} terminal...`); 
    if (window.go) OpenTerminal(type || 'cmd'); 
    setIsTerminalOpen(false);
  };

  const handleServerRootChange = async (e) => {
    const newRoot = e.target.value;
    setServerRoot(newRoot);
    if (window.go) {
      try {
        await SetServerRoot(newRoot);
        addToast('Server Root', 'Server root updated successfully', 'success');
        addLog(`Server root updated to: ${newRoot}`);
      } catch (err) {
        addToast('Server Root Error', `Failed to update server root: ${err}`, 'error');
        addLog(`Error setting server root: ${err}`);
      }
    }
  };

  const handleBrowseServerRoot = async () => {
    if (window.go) {
      try {
        const selectedDirectory = await SelectServerRoot();
        if (selectedDirectory) {
          setServerRoot(selectedDirectory);
          addToast('Server Root', 'Server root updated successfully', 'success');
          addLog(`Server root updated to: ${selectedDirectory}`);
        }
      } catch (err) {
        addToast('Server Root Error', `Failed to select directory: ${err}`, 'error');
        addLog(`Error selecting directory: ${err}`);
      }
    }
  };

  const handleOpenServerRootFolder = async () => {
    if (window.go) {
      try {
        await OpenServerRootFolder();
        addLog(`Opened Server Root folder`);
      } catch (err) {
        addToast('Error', `Failed to open folder: ${err}`, 'error');
        addLog(`Error opening server root folder: ${err}`);
      }
    }
  };

  const handleOpenPluginFolder = async (serviceName) => {
    if (window.go) {
      try {
        await OpenPluginFolder(serviceName);
        addLog(`Opened folder for ${serviceName}`);
      } catch (err) {
        addToast('Error', `Failed to open folder: ${err}`, 'error');
        addLog(`Error opening folder for ${serviceName}: ${err}`);
      }
    }
  };

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

  const handleRemoveFromHome = (name) => {
    setServices(prev => prev.filter(service => service.name !== name));
    addLog(`Removed ${name} from home screen.`);
  };

  const handleAddToHome = (task) => {
    if (!services.find(service => service.name === task.name)) {
      setServices(prev => [...prev, { name: task.name, status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '', remainingDays: 0 }]);
      addLog(`Added ${task.name} to home screen.`);
    }
    setIsAddingPlugin(false);
  };

  const handleToggleHttps = async (name) => {
    if (!window.go) return;
    try {
      if (name === 'Apache') {
        const newValue = !apacheHttps;
        setApacheHttps(newValue);
        await SetApacheHTTPS(newValue);
        addLog(`Apache HTTPS ${newValue ? 'enabled' : 'disabled'}`);
      } else if (name === 'Nginx') {
        const newValue = !nginxHttps;
        setNginxHttps(newValue);
        await SetNginxHTTPS(newValue);
        addLog(`Nginx HTTPS ${newValue ? 'enabled' : 'disabled'}`);
      }
    } catch (err) {
      addToast('HTTPS Error', `Failed to toggle HTTPS: ${err}`, 'error');
      if (name === 'Apache') setApacheHttps(apacheHttps);
      else if (name === 'Nginx') setNginxHttps(nginxHttps);
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

    if (task.name === 'Node.js' && task.versions) {
       modifiedTask.version = selectedVer;
       modifiedTask.target = `nodejs/node-${selectedVer}`;
       modifiedTask.url = `https://nodejs.org/dist/v${selectedVer}/node-v${selectedVer}-win-${arch}.zip`;
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
      <div className="min-h-screen bg-slate-50 dark:bg-[#0f172a] flex items-center justify-center">
        <Loader2 className="animate-spin text-blue-500" size={32} />
      </div>
    );
  }

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
        toggleTheme={toggleTheme} 
        theme={theme} 
        setIsLogOpen={setIsLogOpen} 
      />

      <div className="flex-1 flex flex-col min-h-0 overflow-hidden relative">
        <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-600/5 blur-[120px] rounded-full animate-pulse pointer-events-none" />
        <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-indigo-600/5 blur-[120px] rounded-full animate-pulse pointer-events-none" style={{ animationDelay: '1s' }} />

        <AppHeader 
          activeTab={activeTab} 
          handleStartAll={handleStartAll} 
          handleStopAll={handleStopAll} 
          handleTerminal={handleTerminal} 
          isTerminalOpen={isTerminalOpen} 
          setIsTerminalOpen={setIsTerminalOpen} 
        />

        <main className="flex-1 flex flex-col min-h-0 overflow-hidden">
          <div className="max-w-4xl w-full mx-auto flex flex-col h-full px-8 pb-8">
            {activeTab === 'activity' ? (
              <ActivityTab 
                serverRoot={serverRoot}
                handleServerRootChange={handleServerRootChange}
                handleBrowseServerRoot={handleBrowseServerRoot}
                isAddingPlugin={isAddingPlugin}
                setIsAddingPlugin={setIsAddingPlugin}
                prerequisites={prerequisites}
                services={services}
                handleAddToHome={handleAddToHome}
                ICON_MAP={ICON_MAP}
                handleToggleService={handleToggleService}
                handleRemoveFromHome={handleRemoveFromHome}
                setActiveTab={setActiveTab}
                handleOpenPluginFolder={handleOpenPluginFolder}
                handleOpenServerRootFolder={handleOpenServerRootFolder}
                apacheHttps={apacheHttps}
                nginxHttps={nginxHttps}
                handleToggleHttps={handleToggleHttps}
              />
            ) : (
              <PluginsTab 
                prerequisites={prerequisites}
                downloadProgress={downloadProgress}
                openDropdown={openDropdown}
                setOpenDropdown={setOpenDropdown}
                selectedVersions={selectedVersions}
                setSelectedVersions={setSelectedVersions}
                handleDeleteVersion={handleDeleteVersion}
                handleInstallSingle={handleInstallSingle}
                handleCancel={handleCancel}
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
