import { useState, useCallback } from 'react';
import * as AppBackend from '../../wailsjs/go/backend/App';
import { resolvePrerequisiteTask } from '../utils/pluginHelpers';

interface UsePrerequisitesOptions {
  addToast: (title: string, message: string, type: 'success' | 'error' | 'info') => void;
  addLog?: (msg: string, type?: string) => void;
}

export function usePrerequisites({ addToast, addLog }: UsePrerequisitesOptions) {
  const [prerequisites, setPrerequisites] = useState<any[]>([]);
  const [downloadProgress, setDownloadProgress] = useState<Record<string, any>>({});
  const [selectedVersions, setSelectedVersions] = useState<Record<string, string>>({});
  const [openDropdown, setOpenDropdown] = useState<string | null>(null);

  const updateSelectedVersions = useCallback((tasks: any[]) => {
    setSelectedVersions((prev) => {
      const next = { ...prev };
      for (const t of tasks) {
        if (t.name === 'OpenSSL' && t.installedVers?.length > 0) {
          next[t.name] = t.installedVers[0];
          continue;
        }
        if (t.activeVersion) {
          next[t.name] = t.activeVersion;
        } else if (!next[t.name]) {
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

  const refreshPrerequisites = useCallback(async () => {
    if (!AppBackend.GetPrerequisites) return;
    try {
      const tasks = await AppBackend.GetPrerequisites();
      const tasksArray = tasks ?? [];
      setPrerequisites(tasksArray);
      if (tasksArray.length > 0) {
        updateSelectedVersions(tasksArray);
      }
    } catch (err) {
      console.error(err);
    }
  }, [updateSelectedVersions]);

  const handleDeleteVersion = useCallback(async (name: string, ver: string) => {
    try {
      await AppBackend.DeleteVersion(name, ver);
      await refreshPrerequisites();
    } catch (err) {
      console.error(err);
    }
  }, [refreshPrerequisites]);

  const handleDownloadProgress = useCallback((data: any) => {
    setDownloadProgress((prev) => ({ ...prev, [data.name]: data }));
    if (data.status?.startsWith('Error')) {
      addToast('Installation Failed', `${data.name}: ${data.status}`, 'error');
      if (addLog) addLog(`Installation Failed: ${data.name} - ${data.status}`, 'error');
    }
    if (data.percentage === 100 && (data.status === 'Completed' || data.status === 'Ready')) {
      void refreshPrerequisites();
    }
  }, [addToast, addLog, refreshPrerequisites]);

  const handleInstallSingle = useCallback(async (task: any) => {
    const selectedVersion = selectedVersions[task.name] || task.version;
    setDownloadProgress((prev) => ({
      ...prev,
      [task.name]: { name: task.name, percentage: 0, status: 'Starting...' },
    }));

    const modifiedTask = resolvePrerequisiteTask(task, selectedVersion);
    try {
      await AppBackend.InstallPrerequisite(modifiedTask);
    } catch (error: any) {
      addToast('Error', error.toString(), 'error');
    }
  }, [selectedVersions, addToast]);

  const handleInstallModule = useCallback(async (parentName: string, moduleName: string) => {
    setDownloadProgress((prev) => ({
      ...prev,
      [moduleName]: { name: moduleName, percentage: 0, status: 'Starting...' },
    }));
    try {
      await AppBackend.InstallPluginModule(parentName, moduleName);
    } catch (error: any) {
      addToast('Error', error.toString(), 'error');
      setDownloadProgress((prev) => ({
        ...prev,
        [moduleName]: { name: moduleName, percentage: 0, status: 'Error: ' + error.toString() },
      }));
    }
  }, [addToast]);

  const handleUninstallModule = useCallback(async (parentName: string, moduleName: string) => {
    try {
      await AppBackend.UninstallPluginModule(parentName, moduleName);
      await refreshPrerequisites();
    } catch (error: any) {
      addToast('Error', error.toString(), 'error');
    }
  }, [addToast, refreshPrerequisites]);

  const handleCancelDownload = useCallback((name: string) => {
    void AppBackend.CancelDownload(name);
  }, []);

  const handleOpenPluginFolder = useCallback((name: string) => {
    void AppBackend.OpenPluginFolder(name);
  }, []);

  return {
    prerequisites,
    setPrerequisites,
    downloadProgress,
    setDownloadProgress,
    selectedVersions,
    setSelectedVersions,
    openDropdown,
    setOpenDropdown,
    updateSelectedVersions,
    refreshPrerequisites,
    handleDeleteVersion,
    handleInstallSingle,
    handleInstallModule,
    handleUninstallModule,
    handleCancelDownload,
    handleOpenPluginFolder,
    handleDownloadProgress,
  };
}
