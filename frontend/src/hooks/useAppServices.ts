import { useState, useEffect, useCallback } from 'react';
import { AppService } from '../services/appService';

export function useAppServices() {
  const [services, setServices] = useState<any[]>([
    { name: 'Apache', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'Nginx', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'MySQL', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'PHP', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'Node.js', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'Python', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'HeidiSQL', status: 'Stopped', pid: 0, port: 0, ports: [], activeVersion: '' },
    { name: 'OpenSSL', status: 'Stopped', pid: 0, port: 0, remainingDays: 0, ports: [], activeVersion: '' },
  ]);

  const fetchServices = useCallback(async () => {
    try {
      const list = await AppService.getServices();
      if (list && Array.isArray(list)) {
        setServices(list);
      }
    } catch (err) {
      console.error('Failed to fetch services:', err);
    }
  }, []);

  const startService = useCallback(async (name: string) => {
    await AppService.startService(name);
    await fetchServices();
  }, [fetchServices]);

  const stopService = useCallback(async (name: string) => {
    await AppService.stopService(name);
    await fetchServices();
  }, [fetchServices]);

  useEffect(() => {
    fetchServices();
  }, [fetchServices]);

  return { services, setServices, fetchServices, startService, stopService };
}
