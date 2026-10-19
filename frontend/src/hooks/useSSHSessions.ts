import { useState, useEffect, useCallback } from 'react';
import { AppService } from '../services/appService';

export function useSSHSessions() {
  const [sessions, setSessions] = useState<any[]>([]);
  const [loading, setLoading] = useState<boolean>(true);

  const fetchSessions = useCallback(async () => {
    try {
      setLoading(true);
      const res = await AppService.getSSHSessions();
      setSessions(res || []);
    } catch (err) {
      console.error('Failed to load SSH sessions:', err);
    } finally {
      setLoading(false);
    }
  }, []);

  const saveProfile = useCallback(async (profile: any) => {
    try {
      await AppService.saveSSHProfile(profile);
      await fetchSessions();
      return true;
    } catch (err) {
      console.error('Failed to save profile:', err);
      return false;
    }
  }, [fetchSessions]);

  const deleteProfile = useCallback(async (id: string) => {
    try {
      await AppService.deleteSSHProfile(id);
      await fetchSessions();
      return true;
    } catch (err) {
      console.error('Failed to delete profile:', err);
      return false;
    }
  }, [fetchSessions]);

  useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  return { sessions, loading, fetchSessions, saveProfile, deleteProfile };
}
