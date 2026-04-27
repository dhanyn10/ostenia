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

function App() {
  const [activeTab, setActiveTab] = useState('activity');
  const [theme, setTheme] = useState('light'); 
  const [services, setServices] = useState([
    { name: 'Apache', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'Nginx', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'MySQL', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'PHP', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'Node.js', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'Python', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
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
  
  const [serverRootState, setServerRootState] = useState('');
  const [appsLocation, setAppsLocation] = useState('');
  const [apacheHttps, setApacheHttps] = useState(false);
  const [nginxHttps, setNginxHttps] = useState(false);

  const renderIcon = (name, size = 20, className = "") => {
    const task = (prerequisites || []).find(p => p.name === name);
    if (task && task.iconSvg) return <Icons.Raw svgString={task.iconSvg} size={size} className={className} />;
    return null;
  };

  const addToast = (title, message, type = 'info') => {
    const id = Math.random().toString(36).substr(2, 9);
    setToasts(prev => [...prev, { id, title, message, type }]);
    setTimeout(() => setToasts(curr => curr.filter(t => t.id !== id)), 5000);
  };

  const removeToast = (id) => setToasts(prev => prev.filter(t => t.id !== id));

  const refreshPrerequisites = async () => {
    if (!AppBackend.GetPrerequisites) return;
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
      }
    } catch (err) { console.error(err); }
  };

  const loadInitialData = async () => {
    if (!AppBackend.GetConfig) return;
    try {
      const cfg = await AppBackend.GetConfig();
      if (cfg) {
        setServerRootState(cfg.wwwRoot || '');
        setAppsLocation(cfg.baseDir || '');
        setApacheHttps(cfg.apacheHttps || false);
        setNginxHttps(cfg.nginxHttps || false);
      }
      if (AppBackend.GetServiceStatus) {
        const updatedServices = await Promise.all(
          services.map(async (service) => {
             try {
               const detail = await AppBackend.GetServiceStatus(service.name);
               if (!detail) return service;
               return { ...service, ...detail, status: detail.status || 'Stopped' };
             } catch (e) { return service; }
          })
        );
        setServices(updatedServices);
      }
    } catch (err) { console.error(err); }
  };

  const initApp = async () => {
    setLoading(true);
    try {
        await Promise.all([refreshPrerequisites(), loadInitialData()]);
    } catch (err) {
        console.error("Initialization failed:", err);
    } finally {
        setTimeout(() => setLoading(false), 600);
    }
  };

  useEffect(() => {
    initApp();

    if (window.runtime) {
      EventsOn('service_status', (data) => {
        setServices(prev => prev.map(s => s.name === data.name ? { ...s, ...data } : s));
      });

      EventsOn('download_progress', (data) => {
        setDownloadProgress(prev => ({ ...prev, [data.name]: data }));
        if (data.status?.startsWith('Error')) {
          addToast('Installation Failed', `${data.name}: ${data.status}`, 'error');
        }
        if (data.percentage === 100 && (data.status === 'Completed' || data.status === 'Ready')) {
          refreshPrerequisites();
        }
      });
      
      EventsOn('environment_changed', () => {
        initApp();
      });
    }
  }, []);

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
        renderIcon={renderIcon}
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
            {activeTab === 'activity' ? (
              <ActivityTab 
                serverRoot={serverRootState}
                appsLocation={appsLocation}
                handleBrowseAppsLocation={async () => {
                   const selected = await AppBackend.SelectServerRoot();
                   if (selected) { setAppsLocation(selected); initApp(); }
                }}
                handleBrowseServerRoot={async () => {
                   if (window.runtime) {
                     const selected = await window.runtime.OpenDirectoryDialog({ title: "Select Server Root (www)" });
                     if (selected) { await AppBackend.SetWWWRoot(selected); setServerRootState(selected); }
                   }
                }}
                isAddingPlugin={isAddingPlugin}
                setIsAddingPlugin={setIsAddingPlugin}
                prerequisites={prerequisites}
                services={services}
                handleAddToHome={(task) => {
                  setServices(prev => {
                    if (prev.find(s => s.name === task.name)) return prev;
                    return [...prev, { name: task.name, status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '', remainingDays: 0 }];
                  });
                  setIsAddingPlugin(false);
                }}
                renderIcon={renderIcon}
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
                isLoading={loading} // Pass loading state to ActivityTab
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
                    setDownloadProgress(prev => ({ ...prev, [task.name]: { name: task.name, percentage: 0, status: 'Starting...' } }));
                    const modifiedTask = { ...task, version: selectedVer };
                    const prefixMap = {
                        'PHP': 'php-', 'Apache': 'httpd-', 'MySQL': 'mysql-',
                        'Nginx': 'nginx-', 'OpenSSL': 'openssl-', 'Node.js': 'node-v', 'Python': 'python-'
                    };
                    const categoryMap = { 'Node.js': 'nodejs', 'HeidiSQL': 'heidisql' };
                    const category = categoryMap[task.name] || task.name.toLowerCase();
                    const prefix = prefixMap[task.name] || '';
                    modifiedTask.target = `${category}/${prefix}${selectedVer}`;
                    if (task.versionUrls && task.versionUrls[selectedVer]) {
                        modifiedTask.url = task.versionUrls[selectedVer];
                    }
                    try { await AppBackend.InstallPrerequisite(modifiedTask); } catch (e) { addToast('Error', e.toString(), 'error'); }
                }}
                handleCancel={(name) => AppBackend.CancelDownload(name)}
                renderIcon={renderIcon}
              />
            )}
          </div>
        </main>
      </div>
    </div>
  );
}

export default App;
