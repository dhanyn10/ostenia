import { useState, useEffect, useCallback } from 'react';
import { EventsOn } from '../wailsjs/runtime/runtime';
import * as AppBackend from '../wailsjs/go/backend/App';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

// Hooks
import { useTheme } from './hooks/useTheme';
import { useToast } from './hooks/useToast';
import { useAppLogs } from './hooks/useAppLogs';
import { useServices } from './hooks/useServices';
import { usePrerequisites } from './hooks/usePrerequisites';
import { useAppConfig } from './hooks/useAppConfig';

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
import AddCustomVersionModal from './components/AddCustomVersionModal';

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

function App() {
  const [activeTab, setActiveTab] = useState('activity');
  const [loading, setLoading] = useState(true);
  const [isTerminalOpen, setIsTerminalOpen] = useState(false);
  const [isAddingPlugin, setIsAddingPlugin] = useState(false);
  const [customVersionModal, setCustomVersionModal] = useState({ isOpen: false, serviceName: '' });
  const [settingsModal, setSettingsModal] = useState({ isOpen: false, category: 'profile' });
  const [confirmModal, setConfirmModal] = useState<any>({
    isOpen: false,
    title: '',
    message: '',
    type: 'danger',
    onConfirm: () => {},
  });

  const { theme, setTheme, toggleTheme } = useTheme();
  const { toasts, addToast, removeToast } = useToast();
  const { logs, addLog, loadRecentLogs, clearAppLogs, setupConsoleOverrides } = useAppLogs();

  const handleCloseConfirmModal = useCallback(() => {
    setConfirmModal((prev: any) => ({ ...prev, isOpen: false }));
  }, []);

  const handleRequireConfirmHeidiSQL = useCallback((name: string, onConfirm: () => void) => {
    setConfirmModal({
      isOpen: true,
      title: 'Uninstall HeidiSQL',
      message:
        'Are you sure you want to uninstall HeidiSQL from your system? This will remove the application but your database data should remain intact.',
      type: 'danger',
      onConfirm: async () => {
        await onConfirm();
        handleCloseConfirmModal();
      },
    });
  }, [handleCloseConfirmModal]);

  const {
    services,
    transitioningServices,
    refreshServicesStatus,
    handleServiceStatus,
    handleToggleService,
    handleStartAll,
    handleStopAll,
    handleAddToHome,
    handleRemoveFromHome,
  } = useServices({
    addToast,
    onRequireConfirmHeidiSQL: handleRequireConfirmHeidiSQL,
  });

  const {
    prerequisites,
    downloadProgress,
    selectedVersions,
    setSelectedVersions,
    openDropdown,
    setOpenDropdown,
    refreshPrerequisites,
    handleDeleteVersion,
    handleInstallSingle,
    handleInstallModule,
    handleUninstallModule,
    handleCancelDownload,
    handleOpenPluginFolder,
    handleDownloadProgress,
  } = usePrerequisites({ addToast, addLog });

  const {
    serverRootState,
    appsLocation,
    apacheHttps,
    nginxHttps,
    defaultEditor,
    loadConfig,
    handleBrowseAppsLocation,
    handleBrowseServerRoot,
    handleToggleHttps,
    handleOpenServerRootFolder,
    handleOpenAppsLocationFolder,
  } = useAppConfig({
    onConfigChanged: () => {
      void initApp();
    },
  });

  const handleOpenSettings = useCallback((category: string) => {
    setSettingsModal({ isOpen: true, category });
  }, []);

  const handleCloseSettings = useCallback(() => {
    setSettingsModal((prev) => ({ ...prev, isOpen: false }));
  }, []);

  const handleTerminal = useCallback((type: string) => {
    AppBackend.OpenTerminal(type);
    setIsTerminalOpen(false);
  }, []);

  const handleServiceLog = useCallback(
    (data: any) => {
      addLog(`[${data.service}] ${data.message}`, 'info', true);
    },
    [addLog]
  );

  const initApp = useCallback(async () => {
    setLoading(true);
    try {
      await Promise.all([
        refreshPrerequisites(),
        loadConfig(),
        refreshServicesStatus(),
        loadRecentLogs(),
      ]);
    } catch (err) {
      console.error('Initialization failed:', err);
    } finally {
      setTimeout(() => setLoading(false), 600);
    }
  }, [refreshPrerequisites, loadConfig, refreshServicesStatus, loadRecentLogs]);

  useEffect(() => {
    const originalLog = console.log;
    const originalWarn = console.warn;
    const originalError = console.error;

    setupConsoleOverrides(originalLog, originalWarn, originalError);
    void initApp();

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
  }, [setupConsoleOverrides, initApp, handleServiceLog, handleServiceStatus, handleDownloadProgress]);

  const renderIcon = useCallback(
    (name: string, size = 20, className = '') => {
      const task = prerequisites?.find((p) => p.name === name);
      return task?.iconSvg ? (
        <Icons.Raw svgString={task.iconSvg} size={size} className={className} />
      ) : null;
    },
    [prerequisites]
  );

  return (
    <div
      className={cn(
        'flex flex-col h-screen font-sans selection:bg-mui-blue-500/30 overflow-hidden transition-colors duration-300 fixed inset-0',
        theme === 'dark' ? 'bg-mui-dark-bg text-mui-grey-200' : 'bg-mui-grey-50 text-mui-grey-900'
      )}
    >
      <MenuBar theme={theme} setTheme={setTheme} onOpenSettings={handleOpenSettings} />

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
            <div
              className={cn(
                'w-full mx-auto flex flex-col h-full',
                activeTab === 'logs' || activeTab === 'ssh' ? 'max-w-none' : 'max-w-4xl px-8 pb-8'
              )}
            >
              <div className={cn('h-full flex flex-col', activeTab !== 'activity' && 'hidden')}>
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
                  transitioningServices={transitioningServices}
                />
              </div>

              <div className={cn('h-full flex flex-col', activeTab !== 'plugins' && 'hidden')}>
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
                  onAddCustomVersion={(name) => setCustomVersionModal({ isOpen: true, serviceName: name })}
                />
              </div>

              <div className={cn('h-full flex flex-col', activeTab !== 'proxy' && 'hidden')}>
                <ProxyTab addToast={addToast} />
              </div>

              <div className={cn('h-full flex flex-col', activeTab !== 'ssh' && 'hidden')}>
                <SSHTab addToast={addToast} theme={theme} onOpenSettings={handleOpenSettings} />
              </div>

              <div className={cn('h-full flex flex-col', activeTab !== 'logs' && 'hidden')}>
                <LogViewer logs={logs} isActive={activeTab === 'logs'} onClearLogs={clearAppLogs} />
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
          defaultEditor,
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

      <AddCustomVersionModal
        isOpen={customVersionModal.isOpen}
        onClose={() => setCustomVersionModal((prev) => ({ ...prev, isOpen: false }))}
        serviceName={customVersionModal.serviceName}
        onSuccess={refreshPrerequisites}
        appsLocation={appsLocation}
      />
    </div>
  );
}

export default App;
