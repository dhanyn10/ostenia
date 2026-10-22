import { useState, useCallback } from 'react';
import * as AppBackend from '@/wailsjs/go/backend/App';
import { ServiceInfo, INITIAL_SERVICES } from '@/types/service';

interface UseServicesOptions {
  addToast: (title: string, message: string, type: 'success' | 'error' | 'info') => void;
  onRequireConfirmHeidiSQL?: (name: string, onConfirm: () => void) => void;
}

export async function fetchServiceStatus(service: ServiceInfo): Promise<ServiceInfo> {
  try {
    const detail = await AppBackend.GetServiceStatus(service.name);
    return detail ? { ...service, ...detail, status: detail.status ?? 'Stopped' } : service;
  } catch {
    return service;
  }
}

export function useServices({ addToast, onRequireConfirmHeidiSQL }: UseServicesOptions) {
  const [services, setServices] = useState<ServiceInfo[]>(INITIAL_SERVICES);
  const [transitioningServices, setTransitioningServices] = useState<Set<string>>(new Set());

  const refreshServicesStatus = useCallback(async () => {
    if (!AppBackend.GetServiceStatus) return;
    try {
      setServices((prev) => {
        // Will be updated via promise resolution
        return prev;
      });
      // Query status for all services in parallel
      const current = await Promise.all(INITIAL_SERVICES.map(fetchServiceStatus));
      setServices((prev) =>
        prev.map((service) => {
          const updated = current.find((currentService) => currentService.name === service.name);
          return updated ? { ...service, ...updated } : service;
        })
      );
    } catch (error) {
      console.error(error);
    }
  }, []);

  const handleServiceStatus = useCallback((data: any) => {
    setServices((prev) => {
      const next: ServiceInfo[] = [];
      for (const service of prev) {
        if (service.name === data.name) {
          next.push({ ...service, ...data });
        } else {
          next.push(service);
        }
      }
      return next;
    });
  }, []);

  const handleConfirmHeidiSQLUninstall = useCallback(async (name: string) => {
    setTransitioningServices((prev) => new Set(prev).add(name));
    try {
      await AppBackend.StopService(name);
    } catch (err: any) {
      addToast('Error', `Failed to uninstall ${name}: ${err.message || err}`, 'error');
    } finally {
      setTransitioningServices((prev) => {
        const next = new Set(prev);
        next.delete(name);
        return next;
      });
    }
  }, [addToast]);

  const handleToggleService = useCallback(async (name: string, status: string) => {
    if (status === 'Running') {
      if (name === 'HeidiSQL') {
        if (onRequireConfirmHeidiSQL) {
          onRequireConfirmHeidiSQL(name, () => handleConfirmHeidiSQLUninstall(name));
        }
        return;
      }
      setTransitioningServices((prev) => new Set(prev).add(name));
      try {
        await AppBackend.StopService(name);
      } catch (err: any) {
        addToast('Error', `Failed to stop ${name}: ${err.message || err}`, 'error');
      } finally {
        setTransitioningServices((prev) => {
          const next = new Set(prev);
          next.delete(name);
          return next;
        });
      }
    } else {
      setTransitioningServices((prev) => new Set(prev).add(name));
      try {
        await AppBackend.StartService(name);
      } catch (err: any) {
        addToast('Error', `Failed to start ${name}: ${err.message || err}`, 'error');
      } finally {
        setTransitioningServices((prev) => {
          const next = new Set(prev);
          next.delete(name);
          return next;
        });
      }
    }
  }, [addToast, onRequireConfirmHeidiSQL, handleConfirmHeidiSQLUninstall]);

  const handleStartAll = useCallback(() => AppBackend.StartAllServices(), []);
  const handleStopAll = useCallback(() => AppBackend.StopAllServices(), []);

  const handleAddToHome = useCallback((task: any) => {
    setServices((prev) => {
      let exists = false;
      for (const s of prev) {
        if (s.name === task.name) {
          exists = true;
          break;
        }
      }
      if (exists) return prev;
      return [
        ...prev,
        {
          name: task.name,
          status: 'Stopped',
          pid: 0,
          port: 0,
          ports: [],
          activeVersion: '',
          remainingDays: 0,
        },
      ];
    });
  }, []);

  const handleRemoveFromHome = useCallback((name: string) => {
    setServices((prev) => prev.filter((service) => service.name !== name));
  }, []);

  return {
    services,
    setServices,
    transitioningServices,
    refreshServicesStatus,
    handleServiceStatus,
    handleToggleService,
    handleConfirmHeidiSQLUninstall,
    handleStartAll,
    handleStopAll,
    handleAddToHome,
    handleRemoveFromHome,
  };
}
