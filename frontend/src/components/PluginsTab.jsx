import React from 'react';
import PluginItem from './PluginItem';
import { OpenPluginFolder } from '../../wailsjs/go/main/App';

function PluginsTab({ 
 prerequisites, downloadProgress, openDropdown, setOpenDropdown,
 selectedVersions, setSelectedVersions, handleDeleteVersion,
 handleInstallSingle, handleCancel, renderIcon, handleInstallModule, handleUninstallModule
}) {
 return (
 <div className="flex flex-col h-full pt-4 animate-in fade-in slide-in-from-bottom-2 duration-300">
 <div className="flex-1 overflow-y-auto pr-3 -mr-3 scrollbar-thin scrollbar-thumb-slate-200 dark:scrollbar-thumb-white/5 space-y-2">
 {prerequisites.map((task) => {
 if (!task) return null;

 return (
 <PluginItem
 key={task.name}
 task={task}
 progress={downloadProgress}
 isDropdownOpen={openDropdown === task.name}
 onDropdownToggle={() => setOpenDropdown(openDropdown === task.name ? null : task.name)}
 selectedVersion={selectedVersions[task.name]}
 onVersionChange={(v) => setSelectedVersions(prev => ({ ...prev, [task.name]: v }))}
 onDeleteVersion={handleDeleteVersion}
 onInstall={handleInstallSingle}
 onCancel={handleCancel}
 onOpenFolder={(name) => OpenPluginFolder(name)}
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
