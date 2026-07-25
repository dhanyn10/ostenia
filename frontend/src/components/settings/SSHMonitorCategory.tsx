import React, { useState } from 'react';
import { Sliders, Monitor } from 'lucide-react';

const SSHMonitorCategory: React.FC = () => {
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

  // Strict sanitization before writing to browser storage
  const handleToggleEnabled = (enabled: boolean) => {
    const sanitized = enabled ? 'true' : 'false';
    setMonitorEnabled(enabled);
    localStorage.setItem('ostenia_ssh_monitor_enabled', sanitized);
    window.dispatchEvent(new Event('ostenia_ssh_monitor_settings_changed'));
  };

  const handleIntervalChange = (rawValue: string) => {
    const parsed = Number.parseInt(rawValue, 10);
    if (!Number.isNaN(parsed)) {
      // Clamp values between 1 and 60 to prevent taint/overflow/underflow
      const clamped = Math.min(Math.max(parsed, 1), 60);
      setMonitorInterval(clamped);
      localStorage.setItem('ostenia_ssh_monitor_interval', String(clamped));
      window.dispatchEvent(new Event('ostenia_ssh_monitor_settings_changed'));
    }
  };

  const handleDisplayModeChange = (rawMode: string) => {
    // Only allow expected enum values to sanitize tainted input
    const sanitizedMode = ['tooltip', 'hover-inline', 'always'].includes(rawMode) ? rawMode : 'tooltip';
    setDisplayMode(sanitizedMode);
    localStorage.setItem('ostenia_ssh_monitor_display_mode', sanitizedMode);
    window.dispatchEvent(new Event('ostenia_ssh_monitor_settings_changed'));
  };

  return (
    <div className="space-y-8 animate-in fade-in slide-in-from-right-4 duration-300">
      <div>
        <h3 className="text-xl font-bold mb-1 tracking-tight text-mui-grey-900 dark:text-white">SSH Resource Monitoring</h3>
        <p className="text-sm text-mui-grey-500 dark:text-mui-grey-400">Configure real-time background resource tracking for active SSH and WSL sessions.</p>
      </div>

      <div className="space-y-6">
        <div className="space-y-4 bg-mui-grey-50/50 dark:bg-white/5 p-5 rounded-lg border border-mui-grey-200 dark:border-white/10">
          <label className="flex items-center gap-2.5 cursor-pointer text-mui-grey-700 dark:text-mui-grey-300">
            <input
              type="checkbox"
              checked={monitorEnabled}
              onChange={(e) => handleToggleEnabled(e.target.checked)}
              className="rounded border-mui-grey-300 dark:border-white/10 text-mui-blue-600 focus:ring-mui-blue-500"
            />
            <span className="font-bold text-xs">Enable Real-Time Resource Monitoring</span>
          </label>

          <div className="space-y-2 max-w-xs pt-2">
            <label htmlFor="monitor-interval-input" className="block text-[10px] font-bold text-mui-grey-500 uppercase tracking-wider">
              Refresh Interval (seconds)
            </label>
            <input
              id="monitor-interval-input"
              type="number"
              min={1}
              max={60}
              value={monitorInterval}
              onChange={(e) => handleIntervalChange(e.target.value)}
              className="w-full px-3 py-1.5 bg-white dark:bg-mui-grey-900 border border-mui-grey-200 dark:border-white/10 rounded outline-none text-sm text-mui-grey-900 dark:text-white focus:border-mui-blue-500 font-bold"
            />
          </div>

          <div className="space-y-2 max-w-xs pt-2">
            <label htmlFor="monitor-display-select" className="block text-[10px] font-bold text-mui-grey-500 uppercase tracking-wider">
              Chart Display Style
            </label>
            <select
              id="monitor-display-select"
              value={displayMode}
              onChange={(e) => handleDisplayModeChange(e.target.value)}
              className="w-full px-3 py-1.5 bg-white dark:bg-mui-grey-900 border border-mui-grey-200 dark:border-white/10 rounded outline-none text-sm text-mui-grey-700 dark:text-mui-grey-200 focus:border-mui-blue-500 font-bold cursor-pointer"
            >
              <option value="tooltip">Tooltip on Hover</option>
              <option value="hover-inline">Show Inline on Hover</option>
              <option value="always">Always Show Inline</option>
            </select>
          </div>
        </div>

        <div className="p-4 rounded-lg bg-mui-blue-500/5 border border-mui-blue-500/10 space-y-1">
          <div className="flex items-center gap-2 text-mui-blue-600 dark:text-mui-blue-400 font-bold text-xs uppercase tracking-wider">
            <Monitor size={14} /> Information
          </div>
          <p className="text-xs text-mui-grey-600 dark:text-mui-grey-400 leading-relaxed">
            Real-time tracking displays total used, capacity, percentage usage, and visual sliding line charts of CPU, RAM, and Disk metrics directly on the active SSH/WSL connection status bar. Disabling this option prevents background terminal polling commands.
          </p>
        </div>
      </div>
    </div>
  );
};

export default SSHMonitorCategory;
