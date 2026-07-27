import React, { useState, useEffect } from 'react';
import { Eye, EyeOff, ZoomIn, Activity } from 'lucide-react';
import { handleActionKey } from "../../utils/a11y";
import * as AppBackend from '../../../wailsjs/go/backend/App';

const SSHManagementCategory: React.FC = () => {
  const [sshSessions, setSshSessions] = useState<any[]>([]);
  const [showPasswords, setShowPasswords] = useState(false);
  const [zoomEnabled, setZoomEnabled] = useState<boolean>(() => {
    return localStorage.getItem('ostenia_ssh_zoom_enabled') !== 'false';
  });

  const [monitorEnabled, setMonitorEnabled] = useState<boolean>(() => {
    return localStorage.getItem('ostenia_ssh_monitor_enabled') !== 'false';
  });
  const [monitorInterval, setMonitorInterval] = useState<number>(() => {
    const val = Number.parseInt(localStorage.getItem('ostenia_ssh_monitor_interval') || '3', 10);
    return Number.isNaN(val) || val < 1 ? 3 : Math.min(Math.max(val, 1), 60);
  });
  const [displayMode, setDisplayMode] = useState<string>(() => {
    const mode = localStorage.getItem('ostenia_ssh_monitor_display_mode') || 'tooltip';
    return ['tooltip', 'hover-inline', 'always'].includes(mode) ? mode : 'tooltip';
  });

  useEffect(() => {
    loadSSHSessions();
  }, []);

  const loadSSHSessions = async () => {
    try {
      const sessions = await AppBackend.GetSSHSessions();
      setSshSessions(sessions || []);
    } catch (err) {
      console.error(err);
    }
  };

  const handleToggleZoom = (enabled: boolean) => {
    setZoomEnabled(enabled);
    localStorage.setItem('ostenia_ssh_zoom_enabled', enabled ? 'true' : 'false');
    window.dispatchEvent(new Event('ostenia_ssh_zoom_settings_changed'));
  };

  const handleToggleMonitorEnabled = (enabled: boolean) => {
    setMonitorEnabled(enabled);
    localStorage.setItem('ostenia_ssh_monitor_enabled', enabled ? 'true' : 'false');
    window.dispatchEvent(new Event('ostenia_ssh_monitor_settings_changed'));
  };

  const handleIntervalChange = (rawValue: string) => {
    const parsed = Number.parseInt(rawValue, 10);
    if (!Number.isNaN(parsed)) {
      const clamped = Math.min(Math.max(parsed, 1), 60);
      setMonitorInterval(clamped);
      localStorage.setItem('ostenia_ssh_monitor_interval', String(clamped));
      window.dispatchEvent(new Event('ostenia_ssh_monitor_settings_changed'));
    }
  };

  const handleDisplayModeChange = (rawMode: string) => {
    const sanitizedMode = ['tooltip', 'hover-inline', 'always'].includes(rawMode) ? rawMode : 'tooltip';
    setDisplayMode(sanitizedMode);
    localStorage.setItem('ostenia_ssh_monitor_display_mode', sanitizedMode);
    window.dispatchEvent(new Event('ostenia_ssh_monitor_settings_changed'));
  };

  return (
    <div className="space-y-6 animate-in fade-in slide-in-from-right-4 duration-300 flex flex-col h-full">
      <div className="shrink-0">
        <h3 className="text-xl font-bold mb-1 tracking-tight text-mui-grey-900 dark:text-white">SSH Settings</h3>
        <p className="text-sm text-mui-grey-500 dark:text-mui-grey-400">Configure connection settings, resource monitoring, and view active sessions.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 shrink-0">
        {/* Card 1: Zoom Settings */}
        <div className="p-5 rounded-lg border border-mui-grey-200 dark:border-white/10 bg-mui-grey-50/50 dark:bg-white/5 space-y-4">
          <div className="flex items-center gap-2.5">
            <ZoomIn className="text-mui-blue-500" size={16} />
            <h4 className="text-xs font-black uppercase tracking-widest text-mui-grey-400">Zoom Settings</h4>
          </div>
          <label className="flex items-center gap-2.5 cursor-pointer text-mui-grey-700 dark:text-mui-grey-300">
            <input
              type="checkbox"
              checked={zoomEnabled}
              onChange={(e) => handleToggleZoom(e.target.checked)}
              className="rounded border-mui-grey-300 dark:border-white/10 text-mui-blue-600 focus:ring-mui-blue-500"
            />
            <span className="font-bold text-xs">Enable Terminal Zoom (Ctrl +/-)</span>
          </label>
          <p className="text-[10px] text-mui-grey-400 leading-normal italic">
            Adds vertical Google Maps-style zoom controls to the active terminal panel, and enables standard keyboard shortcuts (Ctrl + Plus, Ctrl + Minus, Ctrl + 0).
          </p>
        </div>

        {/* Card 2: Monitoring Settings */}
        <div className="p-5 rounded-lg border border-mui-grey-200 dark:border-white/10 bg-mui-grey-50/50 dark:bg-white/5 space-y-4">
          <div className="flex items-center gap-2.5">
            <Activity className="text-mui-blue-500" size={16} />
            <h4 className="text-xs font-black uppercase tracking-widest text-mui-grey-400">Resource Monitoring</h4>
          </div>

          <label className="flex items-center gap-2.5 cursor-pointer text-mui-grey-700 dark:text-mui-grey-300">
            <input
              type="checkbox"
              checked={monitorEnabled}
              onChange={(e) => handleToggleMonitorEnabled(e.target.checked)}
              className="rounded border-mui-grey-300 dark:border-white/10 text-mui-blue-600 focus:ring-mui-blue-500"
            />
            <span className="font-bold text-xs">Enable Resource Tracking</span>
          </label>

          <div className="grid grid-cols-2 gap-2 pt-1">
            <div className="space-y-1">
              <label htmlFor="monitor-interval-input" className="block text-[9px] font-black text-mui-grey-400 uppercase tracking-wider">
                Interval (sec)
              </label>
              <input
                id="monitor-interval-input"
                type="number"
                min={1}
                max={60}
                disabled={!monitorEnabled}
                value={monitorInterval}
                onChange={(e) => handleIntervalChange(e.target.value)}
                className="w-full px-2 py-1 bg-white dark:bg-mui-grey-900 border border-mui-grey-200 dark:border-white/10 rounded outline-none text-xs text-mui-grey-900 dark:text-white focus:border-mui-blue-500 font-bold disabled:opacity-50"
              />
            </div>

            <div className="space-y-1">
              <label htmlFor="monitor-display-select" className="block text-[9px] font-black text-mui-grey-400 uppercase tracking-wider">
                Display Style
              </label>
              <select
                id="monitor-display-select"
                disabled={!monitorEnabled}
                value={displayMode}
                onChange={(e) => handleDisplayModeChange(e.target.value)}
                className="w-full px-2 py-1 bg-white dark:bg-mui-grey-900 border border-mui-grey-200 dark:border-white/10 rounded outline-none text-xs text-mui-grey-700 dark:text-mui-grey-200 focus:border-mui-blue-500 font-bold cursor-pointer disabled:opacity-50"
              >
                <option value="tooltip">Tooltip</option>
                <option value="hover-inline">Hover Inline</option>
                <option value="always">Always Inline</option>
              </select>
            </div>
          </div>
        </div>
      </div>

      {/* JSON Viewer */}
      <div className="flex-1 min-h-0 border border-mui-grey-200 dark:border-white/10 rounded-lg overflow-hidden flex flex-col bg-mui-grey-50 dark:bg-white/5">
        <div className="px-4 py-3 border-b border-mui-grey-200 dark:border-white/10 flex justify-between items-center bg-white dark:bg-mui-dark-paper">
          <span className="text-xs font-black uppercase tracking-widest text-mui-grey-400">ssh_sessions.json</span>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onKeyDown={handleActionKey(() => setShowPasswords(!showPasswords))} onClick={() => setShowPasswords(!showPasswords)}
              className="flex items-center gap-1.5 px-2 py-1 rounded bg-mui-blue-500/10 text-mui-blue-500 hover:bg-mui-blue-500/20 transition-colors text-[10px] font-bold uppercase tracking-tight border-none"
            >
              {showPasswords ? <EyeOff size={12} /> : <Eye size={12} />}
              {showPasswords ? 'Mask Passwords' : 'Show Passwords'}
            </button>
            <div className="px-2 py-1 rounded bg-mui-grey-100 dark:bg-white/5 text-[10px] font-bold text-mui-grey-500 dark:text-mui-grey-400 uppercase tracking-tighter">Read Only</div>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-4 font-mono text-[12px] leading-relaxed">
          <pre className="text-mui-grey-700 dark:text-mui-blue-200">
            {JSON.stringify(sshSessions.map(({ password, passphrase, ...s }: any) => ({
              ...s,
              password: showPasswords ? password : "***",
              passphrase: showPasswords ? passphrase : "***"
            })), null, 2)}
          </pre>
        </div>
      </div>
      <p className="text-[10px] text-mui-grey-400 italic">Sensitive fields like password and passphrase are masked for security. Manage sessions via the main SSH Tab.</p>
    </div>
  );
};

export default SSHManagementCategory;
