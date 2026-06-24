import { useState, useEffect, useCallback } from 'react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import * as AppBackend from '../wailsjs/go/backend/App';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

// Components
import MenuBar from './components/MenuBar';
import StatusBar from './components/StatusBar';
import VerticalNav from './components/VerticalNav';
import AppHeader from './components/AppHeader';
import SettingsModal from './components/SettingsModal';
import Toast from './components/Toast';
import LogViewer from './components/LogViewer';
import ActivityTab from './components/ActivityTab';
import PluginsTab from './components/PluginsTab';
import ProxyTab from './components/ProxyTab';
import SSHTab from './components/ssh/SSHTab';
import Icons from './components/Icons';
import ConfirmationModal from './components/ConfirmationModal';

function cn(...inputs: ClassValue[]) {
 return twMerge(clsx(inputs));
}

interface ServiceInfo {
  name: string;
  status: string;
  pid: number;
  port: number;
  ports: number[];
  activeVersion: string;
  remainingDays?: number;
}

function App() {
 const [activeTab, setActiveTab] = useState('activity');
 const [theme, setTheme] = useState(() => {
 return localStorage.getItem('theme') || (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light');
 });
 const [services, setServices] = useState<ServiceInfo[]>([
 { name: 'Apache', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
 { name: 'Nginx', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
 { name: 'MySQL', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
 { name: 'PHP', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
 { name: 'Node.js', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
 { name: 'Python', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
 { name: 'HeidiSQL', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
 { name: 'OpenSSL', status: 'Stopped', pid: 0, port: 0, remainingDays: 0, ports: [], activeVersion: '' },
 ]);
 const [prerequisites, setPrerequisites] = useState<any[]>([]);
 const [downloadProgress, setDownloadProgress] = useState<Record<string, any>>({});
 const [toasts, setToasts] = useState<any[]>([]);
 const [logs, setLogs] = useState<any[]>([]);
 const [openDropdown, setOpenDropdown] = useState<string | null>(null);
 const [isTerminalOpen, setIsTerminalOpen] = useState(false);
 const [loading, setLoading] = useState(true);
 const [selectedVersions, setSelectedVersions] = useState<Record<string, string>>({});
 const [isAddingPlugin, setIsAddingPlugin] = useState(false);
 const [settingsModal, setSettingsModal] = useState({ isOpen: false, category: 'profile' });

 const [serverRootState, setServerRootState] = useState('');
 const [appsLocation, setAppsLocation] = useState('');
 const [apacheHttps, setApacheHttps] = useState(false);
 const [nginxHttps, setNginxHttps] = useState(false);
 const [defaultEditor, setDefaultEditor] = useState('');

 const [confirmModal, setConfirmModal] = useState<any>({
 isOpen: false,
 title: '',
 message: '',
 onConfirm: () => {},
 type: 'danger'
 });

 useEffect(() => {
 if (theme === 'dark') {
 document.documentElement.classList.add('dark');
 document.documentElement.classList.remove('light');
 } else {
 document.documentElement.classList.add('light');
 document.documentElement.classList.remove('dark');
 }
 localStorage.setItem('theme', theme);
 }, [theme]);

 const toggleTheme = () => setTheme(prev => prev === 'dark' ? 'light' : 'dark');

 const addLog = useCallback((msg: string, type = 'info') => {
 const time = new Date().toLocaleTimeString();
 const id = crypto.randomUUID();
 const prefix = type === 'error' ? 'ERR' : type === 'warn' ? 'WRN' : 'SYS';
 setLogs(prev => [{ id, time, msg: `[${prefix}] ${msg}` }, ...prev].slice(0, 1000));
 }, []);

 const renderIcon = (name: string, size = 20, className = "") => {
   const task = prerequisites?.find(p => p.name === name);
   return task?.iconSvg ? <Icons.Raw svgString={task.iconSvg} size={size} className={className} /> : null;
 };

 const addToast = (title: string, message: string, type: 'info' | 'success' | 'warn' | 'error' = 'info') => {
 const id = crypto.randomUUID();
 setToasts(prev => [...prev, { id, title, message, type }]);
 setTimeout(() => setToasts(curr => curr.filter(t => t.id !== id)), 5000);
 };

 const removeToast = (id: string) => setToasts(prev => prev.filter(t => t.id !== id));

 const updateSelectedVersions = useCallback((tasks: any[]) => {
   setSelectedVersions(prev => {
     const next = { ...prev };
     for (const t of tasks) {
       if (t.name === 'OpenSSL' && t.installedVers?.length > 0) {
         next[t.name] = t.installedVers[0];
         continue;
       }
       if (!next[t.name]) {
         if (t.installedVers?.length > 0) {
           next[t.name] = t.installedVers[0];
         } else if (t.version) {
           next[t.name] = t.version;
         }
       }
     }
     return next;
   });
 }, []);

 const refreshPrerequisites = async () => {
   if (!(AppBackend as any).GetPrerequisites) return;
   try {
     const tasks = await AppBackend.GetPrerequisites();
     const tasksArray = tasks ?? [];
     setPrerequisites(tasksArray);
     if (tasksArray.length > 0) {
       updateSelectedVersions(tasksArray);
     }
   } catch (err) { console.error(err); }
 };

 const loadInitialData = async () => {
   if (!(AppBackend as any).GetConfig) return;
   try {
     const cfg = await AppBackend.GetConfig();
     setServerRootState(cfg?.wwwRoot ?? '');
     setAppsLocation(cfg?.baseDir ?? '');
     setApacheHttps(cfg?.apacheHttps ?? false);
     setNginxHttps(cfg?.nginxHttps ?? false);
     setDefaultEditor(cfg?.defaultEditor ?? '');

     if (AppBackend.GetServiceStatus) {
       const updatedServices = await Promise.all(
         services.map(async (service) => {
           try {
             const detail = await AppBackend.GetServiceStatus(service.name);
             return detail ? { ...service, ...detail, status: detail.status ?? 'Stopped' } : service;
           } catch (e) { return service; }
         })
       );
       setServices(updatedServices as ServiceInfo[]);
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

 // Extracted event handlers to reduce nesting
 const handleServiceLog = (data: any) => {
   addLog(`[${data.service}] ${data.message}`, 'info');
 };

 const handleServiceStatus = (data: any) => {
   setServices(prev => prev.map(s => s.name === data.name ? { ...s, ...data } : s));
 };

 const handleDownloadProgress = (data: any) => {
   setDownloadProgress(prev => ({ ...prev, [data.name]: data }));
   if (data.status?.startsWith('Error')) {
     addToast('Installation Failed', `${data.name}: ${data.status}`, 'error');
     addLog(`Installation Failed: ${data.name} - ${data.status}`, 'error');
   }
   if (data.percentage === 100 && (data.status === 'Completed' || data.status === 'Ready')) {
     refreshPrerequisites();
   }
 };

 const setupConsoleOverrides = useCallback((originalLog: any, originalWarn: any, originalError: any) => {
   console.log = (...args: any[]) => {
     addLog(args.map(a => typeof a === 'object' ? JSON.stringify(a) : a).join(' '), 'info');
     originalLog.apply(console, args);
   };
   console.warn = (...args: any[]) => {
     addLog(args.map(a => typeof a === 'object' ? JSON.stringify(a) : a).join(' '), 'warn');
     originalWarn.apply(console, args);
   };
   console.error = (...args: any[]) => {
     addLog(args.map(a => typeof a === 'object' ? JSON.stringify(a) : a).join(' '), 'error');
     originalError.apply(console, args);
   };
 }, [addLog]);

 useEffect(() => {
   const originalLog = console.log;
   const originalWarn = console.warn;
   const originalError = console.error;

   setupConsoleOverrides(originalLog, originalWarn, originalError);
   initApp();

   if (window.runtime) {
     EventsOn('service_log', handleServiceLog);
     EventsOn('service_status', handleServiceStatus);
     EventsOn('download_progress', handleDownloadProgress);
     EventsOn('environment_changed', initApp);
   }

   return () => {
     console.log = originalLog;
     console.warn = originalWarn;
     console.error = originalError;
   };
 }, [addLog, setupConsoleOverrides]);

 // Main action handlers
 const handleBrowseAppsLocation = async () => {
   const selected = await AppBackend.SelectServerRoot();
   if (selected) { setAppsLocation(selected); initApp(); }
 };

 const handleBrowseServerRoot = async () => {
   const selected = await AppBackend.SelectWWWRoot();
   if (selected) { setServerRootState(selected); initApp(); }
 };

 const handleAddToHome = (task: any) => {
   setServices(prev => {
     if (prev.find(s => s.name === task.name)) return prev;
     return [...prev, { name: task.name, status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '', remainingDays: 0 }];
   });
   setIsAddingPlugin(false);
 };

 const handleConfirmHeidiSQLUninstall = useCallback((name: string) => {
   AppBackend.StopService(name);
   handleCloseConfirmModal();
 }, [handleCloseConfirmModal]);

 const handleToggleService = (name: string, status: string) => {
   if (status === 'Running') {
     if (name === 'HeidiSQL') {
       setConfirmModal({
         isOpen: true,
         title: 'Uninstall HeidiSQL',
         message: 'Are you sure you want to uninstall HeidiSQL from your system? This will remove the application but your database data should remain intact.',
         type: 'danger',
         onConfirm: () => handleConfirmHeidiSQLUninstall(name)
       });
       return;
     }
     AppBackend.StopService(name);
   } else {
     AppBackend.StartService(name);
   }
 };

 const handleToggleHttps = async (name: string) => {
   if (name === 'Apache') {
     const next = !apacheHttps; setApacheHttps(next); await AppBackend.SetApacheHTTPS(next);
   } else {
     const next = !nginxHttps; setNginxHttps(next); await AppBackend.SetNginxHTTPS(next);
   }
 };

 const handleInstallSingle = async (task: any) => {
   const selectedVer = selectedVersions[task.name] || task.version;
   setDownloadProgress(prev => ({ ...prev, [task.name]: { name: task.name, percentage: 0, status: 'Starting...' } }));
   const modifiedTask = { ...task, version: selectedVer };
   const prefixMap: Record<string, string> = {
     'PHP': 'php-', 'Apache': 'httpd-', 'MySQL': 'mysql-',
     'Nginx': 'nginx-', 'OpenSSL': 'openssl-', 'Node.js': 'node-v', 'Python': 'python-'
   };
   const categoryMap: Record<string, string> = { 'Node.js': 'nodejs', 'HeidiSQL': 'heidisql' };
   const category = categoryMap[task.name] || task.name.toLowerCase();
   const prefix = prefixMap[task.name] || '';
   modifiedTask.target = `${category}/${prefix}${selectedVer}`;
   if (task.versionUrls?.[selectedVer]) {
     modifiedTask.url = task.versionUrls[selectedVer];
   }
   try { await AppBackend.InstallPrerequisite(modifiedTask); } catch (e: any) { addToast('Error', e.toString(), 'error'); }
 };

 const handleInstallModule = async (parentName: string, modName: string) => {
   setDownloadProgress(prev => ({ ...prev, [modName]: { name: modName, percentage: 0, status: 'Starting...' } }));
   try {
     await AppBackend.InstallPluginModule(parentName, modName);
   } catch (e: any) {
     addToast('Error', e.toString(), 'error');
     setDownloadProgress(prev => ({ ...prev, [modName]: { name: modName, percentage: 0, status: 'Error: ' + e.toString() } }));
   }
 };

 const handleUninstallModule = async (parentName: string, modName: string) => {
   try {
     await AppBackend.UninstallPluginModule(parentName, modName);
     refreshPrerequisites();
   } catch (e: any) {
     addToast('Error', e.toString(), 'error');
   }
 };

 const handleOpenSettings = (category: string) => setSettingsModal({ isOpen: true, category });
 const handleRemoveFromHome = (name: string) => setServices(prev => prev.filter(s => s.name !== name));
 const handleDeleteVersion = (name: string, ver: string) => AppBackend.DeleteVersion(name, ver).then(refreshPrerequisites);
 const handleOpenPluginFolder = (name: string) => AppBackend.OpenPluginFolder(name);
 const handleCloseSettings = () => setSettingsModal(prev => ({ ...prev, isOpen: false }));
 const handleCloseConfirmModal = () => setConfirmModal(prev => ({ ...prev, isOpen: false }));
 const handleStartAll = () => AppBackend.StartAllServices();
 const handleStopAll = () => AppBackend.StopAllServices();
 const handleTerminal = (type: string) => { AppBackend.OpenTerminal(type); setIsTerminalOpen(false); };
 const handleOpenServerRootFolder = () => AppBackend.OpenServerRootFolder();
 const handleOpenAppsLocationFolder = () => AppBackend.OpenAppsLocationFolder();
 const handleCancelDownload = (name: string) => AppBackend.CancelDownload(name);

 return (
 <div className={cn(
 "flex flex-col h-screen font-sans selection:bg-mui-blue-500/30 overflow-hidden transition-colors duration-300 fixed inset-0",
 theme === 'dark' ? "bg-mui-dark-bg text-mui-grey-200" : "bg-mui-grey-50 text-mui-grey-900"
 )}>
 <MenuBar
 theme={theme}
 setTheme={setTheme}
 onOpenSettings={handleOpenSettings}
 />

 <div className="flex-1 flex min-h-0 overflow-hidden relative">
 <Toast toasts={toasts} removeToast={removeToast} />

 <VerticalNav
 activeTab={activeTab}
 setActiveTab={setActiveTab}
 toggleTheme={toggleTheme}
 theme={theme}
 renderIcon={renderIcon}
 />

 <div className="flex-1 flex flex-col min-h-0 overflow-hidden relative">
 <AppHeader
 activeTab={activeTab}
 handleStartAll={handleStartAll}
 handleStopAll={handleStopAll}
 handleTerminal={handleTerminal}
 isTerminalOpen={isTerminalOpen}
 setIsTerminalOpen={setIsTerminalOpen}
 />

 <main className="flex-1 flex flex-col min-h-0 overflow-hidden relative">
 <div className={cn(
 "w-full mx-auto flex flex-col h-full",
 (activeTab === 'logs' || activeTab === 'ssh') ? "max-w-none" : "max-w-4xl px-8 pb-8"
 )}>
 <div className={cn("h-full flex flex-col", activeTab !== 'activity' && "hidden")}>
 <ActivityTab
 serverRoot={serverRootState}
 appsLocation={appsLocation}
 handleBrowseAppsLocation={handleBrowseAppsLocation}
 handleBrowseServerRoot={handleBrowseServerRoot}
 isAddingPlugin={isAddingPlugin}
 setIsAddingPlugin={setIsAddingPlugin}
 prerequisites={prerequisites}
 services={services}
 handleAddToHome={handleAddToHome}
 renderIcon={renderIcon}
 handleToggleService={handleToggleService}
 handleRemoveFromHome={handleRemoveFromHome}
 setActiveTab={setActiveTab}
 handleOpenPluginFolder={handleOpenPluginFolder}
 handleOpenServerRootFolder={handleOpenServerRootFolder}
 handleOpenAppsLocationFolder={handleOpenAppsLocationFolder}
 apacheHttps={apacheHttps}
 nginxHttps={nginxHttps}
 handleToggleHttps={handleToggleHttps}
 isLoading={loading}
 />
 </div>

 <div className={cn("h-full flex flex-col", activeTab !== 'plugins' && "hidden")}>
 <PluginsTab
 prerequisites={prerequisites}
 downloadProgress={downloadProgress}
 openDropdown={openDropdown}
 setOpenDropdown={setOpenDropdown}
 selectedVersions={selectedVersions}
 setSelectedVersions={setSelectedVersions}
 handleDeleteVersion={handleDeleteVersion}
 handleInstallSingle={handleInstallSingle}
 handleCancel={handleCancelDownload}
 renderIcon={renderIcon}
 handleInstallModule={handleInstallModule}
 handleUninstallModule={handleUninstallModule}
 />
 </div>

 <div className={cn("h-full flex flex-col", activeTab !== 'proxy' && "hidden")}>
 <ProxyTab addToast={addToast} />
 </div>

 <div className={cn("h-full flex flex-col", activeTab !== 'ssh' && "hidden")}>
 <SSHTab addToast={addToast} theme={theme} onOpenSettings={handleOpenSettings} />
 </div>

 <div className={cn("h-full flex flex-col", activeTab !== 'logs' && "hidden")}>
 <LogViewer logs={logs} />
 </div>
 </div>
 </main>
 </div>
 </div>

 <StatusBar services={services} />

 <SettingsModal
 isOpen={settingsModal.isOpen}
 onClose={handleCloseSettings}
 initialCategory={settingsModal.category}
 appConfig={{
 baseDir: appsLocation,
 wwwRoot: serverRootState,
 apacheHttps,
 nginxHttps,
 defaultEditor: defaultEditor
 }}
 initApp={initApp}
 theme={theme}
 />

 <ConfirmationModal
 isOpen={confirmModal.isOpen}
 title={confirmModal.title}
 message={confirmModal.message}
 type={confirmModal.type}
 onConfirm={confirmModal.onConfirm}
 onCancel={handleCloseConfirmModal}
 />
 </div>
 );
}

export default App;
