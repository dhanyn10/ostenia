import React from "react";
import { RefreshCw, Folder } from "lucide-react";
import PluginItem from "./PluginItem";

function PluginsTab({
  prerequisites,
  downloadProgress,
  openDropdown,
  setOpenDropdown,
  selectedVersions,
  setSelectedVersions,
  handleDeleteVersion,
  handleInstallSingle,
  handleCancel,
  renderIcon,
  handleInstallModule,
  handleUninstallModule,
  onAddCustomVersion,
  onReloadPlugins,
  onOpenPluginsFolder,
}) {
  return (
    <div className="flex flex-col h-full pt-4 animate-in fade-in slide-in-from-bottom-2 duration-300">
      <div className="flex items-center justify-between pb-3 mb-2 border-b border-slate-200 dark:border-white/5">
        <div className="text-xs font-semibold text-slate-500 dark:text-slate-400">
          JSON-based Plugin Catalog ({prerequisites?.length || 0})
        </div>
        <div className="flex items-center gap-2">
          {onReloadPlugins && (
            <button
              type="button"
              onClick={onReloadPlugins}
              className="flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 dark:bg-white/5 dark:hover:bg-white/10 text-slate-700 dark:text-slate-200 rounded text-[11px] font-medium transition-all"
              title="Reload plugin manifests from JSON"
            >
              <RefreshCw size={13} />
              Reload Plugins
            </button>
          )}
          {onOpenPluginsFolder && (
            <button
              type="button"
              onClick={onOpenPluginsFolder}
              className="flex items-center gap-1.5 px-3 py-1.5 bg-slate-100 hover:bg-slate-200 dark:bg-white/5 dark:hover:bg-white/10 text-slate-700 dark:text-slate-200 rounded text-[11px] font-medium transition-all"
              title="Open JSON plugins directory"
            >
              <Folder size={13} />
              Open Plugins Folder
            </button>
          )}
        </div>
      </div>
      <div className="flex-1 overflow-y-auto pr-3 -mr-3 scrollbar-thin scrollbar-thumb-slate-200 dark:scrollbar-thumb-white/5 space-y-2">
        {prerequisites.map((task) => {
          if (!task) return null;

          return (
            <PluginItem
              key={task.name}
              task={task}
              progress={downloadProgress}
              isDropdownOpen={openDropdown === task.name}
              onDropdownToggle={() =>
                setOpenDropdown(openDropdown === task.name ? null : task.name)
              }
              selectedVersion={selectedVersions[task.name]}
              onVersionChange={(v) =>
                setSelectedVersions((prev) => ({ ...prev, [task.name]: v }))
              }
              onDeleteVersion={handleDeleteVersion}
              onInstall={handleInstallSingle}
              onCancel={handleCancel}
              onOpenFolder={(name) => onAddCustomVersion(name)}
              renderIcon={renderIcon}
              onInstallModule={handleInstallModule}
              onUninstallModule={handleUninstallModule}
            />
          );
        })}
      </div>
    </div>
  );
}

export default PluginsTab;
