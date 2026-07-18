import React, { useState, useEffect } from 'react';
import { FolderOpen, Globe, HardDrive, Loader2 } from 'lucide-react';
import { OpenServiceTerminal, SwitchServiceVersion, GetPHPExtensions, TogglePHPExtension } from '../../wailsjs/go/backend/App';
import ExtensionModal from './ExtensionModal';
import ServiceItem from './ServiceItem';
import AddPluginAction from './AddPluginAction';

function ActivityTab({ 
 serverRoot, appsLocation, handleBrowseAppsLocation, handleBrowseServerRoot,
 isAddingPlugin, setIsAddingPlugin, prerequisites, services, handleAddToHome,
 renderIcon, handleToggleService, handleRemoveFromHome, setActiveTab,
 handleOpenPluginFolder, handleOpenServerRootFolder, handleOpenAppsLocationFolder,
 apacheHttps, nginxHttps, handleToggleHttps,
 isLoading, transitioningServices
}) {
 const [openTerminalDropdown, setOpenTerminalDropdown] = useState(null);
 const [activeAccordion, setActiveAccordion] = useState(null);
 const [phpExtensions, setPhpExtensions] = useState([]);
 const [isModalOpen, setIsModalOpen] = useState(false);

 const openSslService = services?.find(s => s.name === 'OpenSSL');
 const isOpenSslEnabled = openSslService?.status === 'Running';

 const toggleAccordion = (name, hasExtra) => {
 if (!hasExtra) return;
 setActiveAccordion(activeAccordion === name ? null : name);
 setOpenTerminalDropdown(null);
 };

 const fetchPHPExtensions = async () => {
 try {
 const exts = await GetPHPExtensions();
 setPhpExtensions(exts || []);
 } catch (err) {
 console.error("Failed to fetch PHP extensions:", err);
 }
 };

 useEffect(() => {
 if (activeAccordion === 'PHP') {
 fetchPHPExtensions();
 }
 }, [activeAccordion]);

 return (
 <div className="flex flex-col h-full animate-in fade-in slide-in-from-bottom-2 duration-500">
 <ExtensionModal
 isOpen={isModalOpen}
 onClose={() => setIsModalOpen(false)}
 extensions={phpExtensions}
 onToggle={async (name, enable) => {
 await TogglePHPExtension(name, enable);
 fetchPHPExtensions();
 }}
 serviceName="PHP"
 />

 {/* Path Configuration */}
 <div className="shrink-0 pt-4 pb-3 grid grid-cols-2 gap-3">
 <div className="bg-white/50 dark:bg-slate-900/40 rounded-sm p-4 border border-slate-200 dark:border-white/5 shadow-sm">
 <h3 className="text-[9px] font-black text-slate-500 dark:text-slate-400 uppercase tracking-[0.2em] mb-2 flex items-center gap-2">
 <HardDrive size={10} /> Apps Location
 </h3>
 <div className="flex items-center gap-2">
 <input type="text" readOnly value={appsLocation || ''} className="flex-1 bg-slate-100 dark:bg-black/20 border border-slate-200 dark:border-white/5 rounded-sm px-3 py-1.5 text-[10px] text-slate-500 dark:text-slate-400 font-mono truncate" />
 <button type="button" onClick={handleBrowseAppsLocation} title="Browse Directory" className="px-2 py-1.5 bg-slate-200 dark:bg-white/10 hover:bg-slate-300 dark:hover:bg-white/20 text-slate-600 dark:text-slate-300 rounded-sm text-[10px] font-bold uppercase tracking-wider transition-all">Browse</button>
 <button type="button" onClick={handleOpenAppsLocationFolder} title="Open in Explorer" className="p-1.5 bg-blue-600/10 hover:bg-blue-600/20 text-blue-600 rounded-sm transition-all border border-blue-500/20"><FolderOpen size={14} /></button>
 </div>
 </div>

 <div className="bg-white/50 dark:bg-slate-900/40 rounded-sm p-4 border border-slate-200 dark:border-white/5 shadow-sm">
 <h3 className="text-[9px] font-black text-slate-500 dark:text-slate-400 uppercase tracking-[0.2em] mb-2 flex items-center gap-2">
 <Globe size={10} /> Server Root Directory
 </h3>
 <div className="flex items-center gap-2">
 <input type="text" readOnly value={serverRoot || ''} className="flex-1 bg-slate-100 dark:bg-black/20 border border-slate-200 dark:border-white/5 rounded-sm px-3 py-1.5 text-[10px] text-slate-500 dark:text-slate-400 font-mono truncate" />
 <button type="button" onClick={handleBrowseServerRoot} title="Browse Directory" className="px-2 py-1.5 bg-slate-200 dark:bg-white/10 hover:bg-slate-300 dark:hover:bg-white/20 text-slate-600 dark:text-slate-300 rounded-sm text-[10px] font-bold uppercase tracking-wider transition-all">Browse</button>
 <button type="button" onClick={handleOpenServerRootFolder} title="Open in Explorer" className="p-1.5 bg-emerald-600/10 hover:bg-emerald-600/20 text-emerald-600 rounded-sm transition-all border border-emerald-500/20"><FolderOpen size={14} /></button>
 </div>
 </div>
 </div>

 <AddPluginAction
 isAddingPlugin={isAddingPlugin}
 setIsAddingPlugin={setIsAddingPlugin}
 prerequisites={prerequisites || []}
 services={services || []}
 handleAddToHome={handleAddToHome}
 renderIcon={renderIcon}
 />

 <div className="flex-1 overflow-y-auto pr-3 -mr-3 scrollbar-thin scrollbar-thumb-slate-200 dark:scrollbar-thumb-white/5 space-y-2 pb-4 relative">
 {isLoading ? (
 <div className="absolute inset-0 flex flex-col items-center justify-center py-20 z-50 rounded-sm">
 <Loader2 className="animate-spin text-blue-500 mb-2" size={24} />
 <span className="text-[9px] font-bold text-slate-400 uppercase tracking-widest animate-pulse">Scanning Plugins...</span>
 </div>
 ) : (
 <>
 {(services || []).map((service) => (
 <ServiceItem
 key={service.name}
 service={service}
 task={prerequisites?.find(p => p.name === service.name)}
 isExpanded={activeAccordion === service.name}
 onToggleAccordion={toggleAccordion}
 renderIcon={renderIcon}
 handleToggleService={handleToggleService}
 handleRemoveFromHome={handleRemoveFromHome}
 handleSwitchVersion={async (name, ver) => await SwitchServiceVersion(name, ver)}
 handleOpenLocalTerminal={(name, type) => OpenServiceTerminal(name, type)}
 handleToggleHttps={handleToggleHttps}
 openTerminalDropdown={openTerminalDropdown}
 setOpenTerminalDropdown={setOpenTerminalDropdown}
 setIsModalOpen={setIsModalOpen}
 apacheHttps={apacheHttps}
 nginxHttps={nginxHttps}
 isOpenSslEnabled={isOpenSslEnabled}
 setActiveTab={setActiveTab}
 handleOpenPluginFolder={handleOpenPluginFolder}
 isTransitioning={transitioningServices?.has(service.name)}
 />
 ))}

 {services?.length === 0 && (
 <div className="flex flex-col items-center justify-center py-12 text-slate-400 dark:text-slate-500 border border-dashed border-slate-200 dark:border-white/5 rounded-sm">
 <p className="text-[10px] font-bold uppercase tracking-widest">No services active</p>
 <p className="text-[9px] opacity-60">Add some plugins to get started</p>
 </div>
 )}
 </>
 )}
 </div>
 </div>
 );
}

export default ActivityTab;
