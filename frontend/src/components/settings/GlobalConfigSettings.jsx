import React from 'react';
import PropTypes from 'prop-types';
import { FolderOpen, Globe, Monitor, Trash2 } from 'lucide-react';
import * as AppBackend from '../../../wailsjs/go/backend/App';

const GlobalConfigSettings = ({ appConfig, installedApps, initApp }) => {
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
                                        <span className="text-xs text-mui-grey-700 dark:text-mui-grey-200 truncate max-w-md">{appConfig.defaultEditor}</span>
                                    </div>
                                    <button
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
            </div>
        </div>
    );
};

GlobalConfigSettings.propTypes = {
    appConfig: PropTypes.shape({
        baseDir: PropTypes.string,
        wwwRoot: PropTypes.string,
        defaultEditor: PropTypes.string,
    }),
    installedApps: PropTypes.arrayOf(PropTypes.shape({
        name: PropTypes.string.isRequired,
        path: PropTypes.string.isRequired,
    })).isRequired,
    initApp: PropTypes.func.isRequired,
};

export default GlobalConfigSettings;
