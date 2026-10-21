import React, { useState, useEffect } from 'react';
import { FolderOpen, Globe, Monitor, Trash2 } from 'lucide-react';
import * as AppBackend from '../../../wailsjs/go/backend/App';

interface GlobalConfigCategoryProps {
  appConfig: any;
  initApp?: () => void;
}

const GlobalConfigCategory: React.FC<GlobalConfigCategoryProps> = ({ appConfig, initApp }) => {
  const [installedApps, setInstalledApps] = useState<any[]>([]);

  useEffect(() => {
    void loadInstalledApps();
  }, []);

  const loadInstalledApps = async () => {
    try {
      const apps = await AppBackend.GetInstalledApps();
      setInstalledApps(apps || []);
    } catch (err) { console.error(err); }
  };

  const handleSelectServerRoot = async () => {
    const selected = await AppBackend.SelectServerRoot();
    if (selected && initApp) initApp();
  };

  const handleSelectWWWRoot = async () => {
    const selected = await AppBackend.SelectWWWRoot();
    if (selected && initApp) initApp();
  };

  const handleEditorChange = async (event: React.ChangeEvent<HTMLSelectElement>) => {
    if (event.target.value) {
      await AppBackend.SetDefaultEditor(event.target.value);
      if (initApp) initApp();
    }
  };

  const handleSelectDefaultEditor = async () => {
    await AppBackend.SelectDefaultEditor();
    if (initApp) initApp();
  };

  const handleClearDefaultEditor = async () => {
    await AppBackend.SetDefaultEditor("");
    if (initApp) initApp();
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
              onClick={() => {
                void handleSelectServerRoot();
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
              onClick={() => {
                void handleSelectWWWRoot();
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
                  value={installedApps.find(app => app.path === appConfig?.defaultEditor)?.path || ""}
                  onChange={(event) => {
                    void handleEditorChange(event);
                  }}
                  className="flex-1 bg-mui-grey-50 dark:bg-white/5 border border-mui-grey-200 dark:border-white/10 rounded px-3 py-2 text-sm outline-none text-mui-grey-700 dark:text-mui-grey-200 appearance-none cursor-pointer"
                >
                  <option value="">Select from installed apps...</option>
                  {installedApps.toSorted((appA, appB) => appA.name.localeCompare(appB.name)).map((app) => (
                    <option key={app.path || app.name} value={app.path}>{app.name}</option>
                  ))}
                </select>
                <div className="flex items-center px-3 border border-mui-grey-200 dark:border-white/10 rounded bg-mui-grey-50 dark:bg-white/5 text-xs font-bold text-mui-grey-400">
                  OR
                </div>
                <button
                  type="button"
                  onClick={() => {
                    void handleSelectDefaultEditor();
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
                    onClick={() => {
                      void handleClearDefaultEditor();
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
      </div>
    </div>
  );
};

export default GlobalConfigCategory;
