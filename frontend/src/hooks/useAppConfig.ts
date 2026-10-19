import { useState, useEffect, useCallback } from 'react';
import { AppService } from '../services/appService';

export function useAppConfig() {
  const [config, setConfig] = useState<any>(null);
  const [loading, setLoading] = useState<boolean>(true);

  const fetchConfig = useCallback(async () => {
    try {
      setLoading(true);
      const res = await AppService.getConfig();
      setConfig(res);
    } catch (err) {
      console.error('Failed to load config:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  const saveConfig = useCallback(async (newCfg: any) => {
    try {
      await AppService.saveConfig(newCfg);
      setConfig(newCfg);
      return true;
    } catch (err) {
      console.error('Failed to save config:', err);
      return false;
    }
  }, []);

  useEffect(() => {
    fetchConfig();
  }, [fetchConfig]);

  return { config, loading, fetchConfig, saveConfig };
}
