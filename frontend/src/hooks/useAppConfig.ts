import { useState, useCallback } from 'react';
import * as AppBackend from '../../wailsjs/go/backend/App';

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
      const cfg = await AppBackend.GetConfig();
      setServerRootState(cfg?.wwwRoot ?? '');
      setAppsLocation(cfg?.baseDir ?? '');
      setApacheHttps(cfg?.apacheHttps ?? false);
      setNginxHttps(cfg?.nginxHttps ?? false);
      setDefaultEditor(cfg?.defaultEditor ?? '');
    } catch (err) {
      console.error(err);
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

  const handleToggleHttps = useCallback(async (name: string) => {
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
    return AppBackend.OpenServerRootFolder();
  }, []);

  const handleOpenAppsLocationFolder = useCallback(() => {
    return AppBackend.OpenAppsLocationFolder();
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
