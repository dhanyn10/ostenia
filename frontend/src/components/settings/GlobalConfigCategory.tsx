import React, { useState, useEffect } from 'react';
import { FolderOpen, Globe, Monitor, Trash2, Sliders } from 'lucide-react';
import * as AppBackend from '../../../wailsjs/go/backend/App';

interface GlobalConfigCategoryProps {
  appConfig: any;
  initApp?: () => void;
}

const GlobalConfigCategory: React.FC<GlobalConfigCategoryProps> = ({ appConfig, initApp }) => {
  const [installedApps, setInstalledApps] = useState<any[]>([]);
  const [monitorEnabled, setMonitorEnabled] = useState<boolean>(() => {
    return localStorage.getItem('ostenia_ssh_monitor_enabled') !== 'false';
  });
  const [monitorInterval, setMonitorInterval] = useState<number>(() => {
    const val = Number.parseInt(localStorage.getItem('ostenia_ssh_monitor_interval') || '3', 10);
    return Number.isNaN(val) || val < 1 ? 3 : val;
  });

  useEffect(() => {
    loadInstalledApps();
  }, []);

  const loadInstalledApps = async () => {
    try {
      const apps = await AppBackend.GetInstalledApps();
      setInstalledApps(apps || []);
    } catch (err) { console.error(err); }
  };

  return (
    <div className="space-y-8 animate-in fade-in slide-in-from-right-4 duration-300">
      <div>
        <h3 className="text-xl font-bold mb-1 tracking-tight text-mui-grey-900 dark:text-white">Global Configuration</h3>
        <p className="text-sm text-mui-grey-500 dark:text-mui-grey-400">Environment paths and core application behavior.</p>
      </div>

      <div className="space-y-6">
        <div className="space-y-2">
          <label className="text-[11px] font-black uppercase tracking-widest text-mui-grey-400 flex items-center gap-2">
            <FolderOpen size={12} /> Apps Location (Ostenia Home)
          </label>
          <div className="flex gap-2">
            <input
              type="text"
              readOnly
              value={appConfig?.baseDir || ''}
              className="flex-1 bg-mui-grey-50 dark:bg-white/5 border border-mui-grey-200 dark:border-white/10 rounded px-3 py-2 text-sm outline-none text-mui-grey-700 dark:text-mui-grey-200"
            />
            <button
              type="button"
              onClick={async () => {
                const selected = await AppBackend.SelectServerRoot();
                if (selected && initApp) initApp();
              }}
              className="px-4 py-2 bg-mui-blue-500 text-white rounded text-xs font-bold hover:bg-mui-blue-600 transition-colors"
            >
              Browse
            </button>
          </div>
        </div>

        <div className="space-y-2">
          <label className="text-[11px] font-black uppercase tracking-widest text-mui-grey-400 flex items-center gap-2">
            <Globe size={12} /> Server Root (www)
          </label>
          <div className="flex gap-2">
            <input
              type="text"
              readOnly
              value={appConfig?.wwwRoot || ''}
              className="flex-1 bg-mui-grey-50 dark:bg-white/5 border border-mui-grey-200 dark:border-white/10 rounded px-3 py-2 text-sm outline-none text-mui-grey-700 dark:text-mui-grey-200"
            />
            <button
              type="button"
              onClick={async () => {
                const selected = await AppBackend.SelectWWWRoot();
                if (selected && initApp) initApp();
              }}
              className="px-4 py-2 bg-mui-blue-500 text-white rounded text-xs font-bold hover:bg-mui-blue-600 transition-colors"
            >
              Browse
            </button>
          </div>
        </div>

        <div className="pt-4 border-t border-mui-grey-100 dark:border-white/5">
          <div className="space-y-4">
            <label className="text-[11px] font-black uppercase tracking-widest text-mui-grey-400 flex items-center gap-2">
              <Monitor size={12} /> External Editor
            </label>
            <div className="space-y-3">
              <div className="flex gap-2">
                <select
                  value={installedApps.find(a => a.path === appConfig?.defaultEditor)?.path || ""}
                  onChange={async (e) => {
                    if (e.target.value) {
                      await AppBackend.SetDefaultEditor(e.target.value);
                      if (initApp) initApp();
                    }
                  }}
                  className="flex-1 bg-mui-grey-50 dark:bg-white/5 border border-mui-grey-200 dark:border-white/10 rounded px-3 py-2 text-sm outline-none text-mui-grey-700 dark:text-mui-grey-200 appearance-none cursor-pointer"
                >
                  <option value="">Select from installed apps...</option>
                  {installedApps.sort((a,b) => a.name.localeCompare(b.name)).map((app, idx) => (
                    <option key={idx} value={app.path}>{app.name}</option>
                  ))}
                </select>
                <div className="flex items-center px-3 border border-mui-grey-200 dark:border-white/10 rounded bg-mui-grey-50 dark:bg-white/5 text-xs font-bold text-mui-grey-400">
                  OR
                </div>
                <button
                  type="button"
                  onClick={async () => {
                    await AppBackend.SelectDefaultEditor();
                    if (initApp) initApp();
                  }}
                  className="px-4 py-2 bg-mui-blue-500 text-white rounded text-xs font-bold hover:bg-mui-blue-600 transition-colors whitespace-nowrap"
                >
                  Custom Browse
                </button>
              </div>

              {appConfig?.defaultEditor && (
                <div className="p-3 rounded border border-mui-blue-500/20 bg-mui-blue-500/5 flex items-center justify-between group">
                  <div className="flex flex-col">
                    <span className="text-[10px] font-black uppercase tracking-widest text-mui-blue-500">Current Editor</span>
                    <span className="text-xs text-mui-grey-700 dark:text-mui-grey-200 truncate max-md:max-w-xs">{appConfig.defaultEditor}</span>
                  </div>
                  <button
                    type="button"
                    onClick={async () => {
                      await AppBackend.SetDefaultEditor("");
                      if (initApp) initApp();
                    }}
                    className="p-1.5 hover:bg-rose-500/10 rounded text-mui-grey-400 hover:text-rose-500 transition-colors"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              )}

              <p className="text-[10px] text-mui-grey-400 italic">Used for "Edit Remote File" in SSH explorer. If unset, Ostenia uses system default.</p>
            </div>
          </div>
        </div>

        <div className="pt-6 border-t border-mui-grey-100 dark:border-white/5 space-y-4">
          <div>
            <label className="text-[11px] font-black uppercase tracking-widest text-mui-grey-400 flex items-center gap-2">
              <Sliders size={12} /> SSH Connection Resource Monitoring
            </label>
            <p className="text-[10px] text-mui-grey-400 mt-1">Configure real-time background resource tracking for SSH and WSL sessions.</p>
          </div>

          <div className="space-y-4 bg-mui-grey-50/50 dark:bg-white/5 p-4 rounded border border-mui-grey-100 dark:border-white/5">
            <label className="flex items-center gap-2.5 cursor-pointer text-mui-grey-700 dark:text-mui-grey-300">
              <input
                type="checkbox"
                checked={monitorEnabled}
                onChange={(e) => {
                  setMonitorEnabled(e.target.checked);
                  localStorage.setItem('ostenia_ssh_monitor_enabled', e.target.checked ? 'true' : 'false');
                  window.dispatchEvent(new Event('ostenia_ssh_monitor_settings_changed'));
                }}
                className="rounded border-mui-grey-300 dark:border-white/10 text-mui-blue-600 focus:ring-mui-blue-500"
              />
              <span className="font-bold text-xs">Enable Real-Time Resource Monitoring</span>
            </label>

            <div className="space-y-2 max-w-xs">
              <label htmlFor="monitor-interval-input" className="block text-[10px] font-bold text-mui-grey-500 uppercase tracking-wider">
                Refresh Interval (seconds)
              </label>
              <input
                id="monitor-interval-input"
                type="number"
                min={1}
                max={60}
                value={monitorInterval}
                onChange={(e) => {
                  const val = Number.parseInt(e.target.value, 10);
                  if (!Number.isNaN(val) && val >= 1) {
                    setMonitorInterval(val);
                    localStorage.setItem('ostenia_ssh_monitor_interval', String(val));
                    window.dispatchEvent(new Event('ostenia_ssh_monitor_settings_changed'));
                  }
                }}
                className="w-full px-3 py-1.5 bg-white dark:bg-mui-grey-900 border border-mui-grey-200 dark:border-white/10 rounded outline-none text-sm text-mui-grey-900 dark:text-white focus:border-mui-blue-500 font-bold"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default GlobalConfigCategory;
