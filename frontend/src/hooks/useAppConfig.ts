import { useState, useCallback } from 'react';
import * as AppBackend from '@/wailsjs/go/backend/App';

interface UseAppConfigOptions {
  onConfigChanged?: () => void;
}

export function useAppConfig({ onConfigChanged }: UseAppConfigOptions = {}) {
  const [serverRootState, setServerRootState] = useState('');
  const [appsLocation, setAppsLocation] = useState('');
  const [apacheHttps, setApacheHttps] = useState(false);
  const [nginxHttps, setNginxHttps] = useState(false);
  const [defaultEditor, setDefaultEditor] = useState('');

  const loadConfig = useCallback(async () => {
    if (!AppBackend.GetConfig) return;
    try {
      const config = await AppBackend.GetConfig();
      setServerRootState(config?.wwwRoot ?? '');
      setAppsLocation(config?.baseDir ?? '');
      setApacheHttps(config?.apacheHttps ?? false);
      setNginxHttps(config?.nginxHttps ?? false);
      setDefaultEditor(config?.defaultEditor ?? '');
    } catch (error) {
      console.error(error);
    }
  }, []);

  const handleBrowseAppsLocation = useCallback(async () => {
    const selected = await AppBackend.SelectServerRoot();
    if (selected) {
      setAppsLocation(selected);
      onConfigChanged?.();
    }
  }, [onConfigChanged]);

  const handleBrowseServerRoot = useCallback(async () => {
    const selected = await AppBackend.SelectWWWRoot();
    if (selected) {
      setServerRootState(selected);
      onConfigChanged?.();
    }
  }, [onConfigChanged]);

  const handleToggleHttps = useCallback((name: string) => {
    if (name === 'Apache') {
      setApacheHttps((prev) => {
        const next = !prev;
        void AppBackend.SetApacheHTTPS(next);
        return next;
      });
    } else {
      setNginxHttps((prev) => {
        const next = !prev;
        void AppBackend.SetNginxHTTPS(next);
        return next;
      });
    }
  }, []);

  const handleOpenServerRootFolder = useCallback(() => {
    void AppBackend.OpenServerRootFolder();
  }, []);

  const handleOpenAppsLocationFolder = useCallback(() => {
    void AppBackend.OpenAppsLocationFolder();
  }, []);

  return {
    serverRootState,
    setServerRootState,
    appsLocation,
    setAppsLocation,
    apacheHttps,
    setApacheHttps,
    nginxHttps,
    setNginxHttps,
    defaultEditor,
    setDefaultEditor,
    loadConfig,
    handleBrowseAppsLocation,
    handleBrowseServerRoot,
    handleToggleHttps,
    handleOpenServerRootFolder,
    handleOpenAppsLocationFolder,
  };
}
