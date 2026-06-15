import React, { useState, useEffect } from 'react';
import { X, User, Sliders, Terminal as TerminalIcon, Search, ChevronRight, Settings } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import * as AppBackend from '../../wailsjs/go/main/App';
import ProfileSettings from './settings/ProfileSettings';
import GlobalConfigSettings from './settings/GlobalConfigSettings';
import SSHManagementSettings from './settings/SSHManagementSettings';

function cn(...inputs) { return twMerge(clsx(inputs)); }

const SettingsModal = ({ isOpen, onClose, initialCategory = 'profile', appConfig = {}, setConfig, theme, initApp }) => {
  const [activeCategory, setActiveCategory] = useState(initialCategory);
  const [searchQuery, setSearchQuery] = useState('');
  const [sshSessions, setSshSessions] = useState([]);
  const [installedApps, setInstalledApps] = useState([]);

  useEffect(() => { if (isOpen) { setActiveCategory(initialCategory); loadSSHSessions(); loadInstalledApps(); } }, [isOpen, initialCategory]);

  const loadInstalledApps = async () => { try { const apps = await AppBackend.GetInstalledApps(); setInstalledApps(apps || []); } catch (err) { console.error(err); } };
  const loadSSHSessions = async () => { try { const sessions = await AppBackend.GetSSHSessions(); setSshSessions(sessions || []); } catch (err) { console.error(err); } };

  if (!isOpen) return null;

  const categories = [
    { id: 'profile', label: 'Profile', icon: User },
    { id: 'config', label: 'Global Config', icon: Sliders },
    { id: 'ssh', label: 'SSH Management', icon: TerminalIcon },
  ];

  const handleExport = async (type) => { try { await AppBackend.ExportProfile(type === 'all' || type === 'config', type === 'all' || type === 'ssh'); } catch (err) { console.error(err); } };
  const handleImport = async () => { try { await AppBackend.ImportProfile(); if (initApp) initApp(); } catch (err) { console.error(err); } };

  const renderContent = () => {
    switch (activeCategory) {
      case 'profile': return <ProfileSettings handleImport={handleImport} handleExport={handleExport} />;
      case 'config': return <GlobalConfigSettings {...{appConfig, installedApps, AppBackend, initApp}} />;
      case 'ssh': return <SSHManagementSettings sshSessions={sshSessions} />;
      default: return <div className="flex items-center justify-center h-full text-mui-grey-400">Select a category from the sidebar</div>;
    }
  };

  return (
    <div
      onClick={onClose}
      onKeyDown={(e) => e.key === 'Escape' && onClose()}
      tabIndex={-1}
      role="button"
      aria-label="Close Settings"
      className="fixed inset-0 z-[200] flex items-center justify-center p-8 bg-transparent animate-in fade-in duration-300"
    >
      <div
        onClick={(e) => e.stopPropagation()}
        onKeyDown={(e) => e.stopPropagation()}
        role="presentation"
        className={cn("w-full max-w-5xl h-[80vh] flex flex-col rounded-xl shadow-[0_32px_64px_-12px_rgba(0,0,0,0.5)] border overflow-hidden bg-white dark:bg-mui-dark-bg border-mui-grey-200 dark:border-white/10")}
      >
        <div className="h-14 shrink-0 flex items-center justify-between px-6 border-b border-mui-grey-200 dark:border-white/10 bg-mui-grey-50/50 dark:bg-white/5">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-mui-blue-500 rounded flex items-center justify-center text-white shadow-lg shadow-mui-blue-500/30"><Settings size={18} /></div>
            <div><h2 className="text-base font-black uppercase tracking-wider text-mui-grey-900 dark:text-white">Settings</h2><div className="text-[10px] text-mui-grey-400 font-bold uppercase tracking-[0.2em] -mt-1">Ostenia Management</div></div>
          </div>
          <button onClick={onClose} className="p-2 hover:bg-mui-grey-200 dark:hover:bg-white/10 rounded-full transition-colors text-mui-grey-500 dark:text-mui-grey-400"><X size={20} /></button>
        </div>
        <div className="flex-1 flex min-h-0">
          <div className="w-64 shrink-0 border-r border-mui-grey-200 dark:border-white/10 flex flex-col bg-mui-grey-50 dark:bg-mui-dark-paper">
            <div className="p-4"><div className="relative group"><Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-mui-grey-400 group-focus-within:text-mui-blue-500 transition-colors" /><input type="text" placeholder="Search settings..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} className="w-full bg-white dark:bg-white/5 border border-mui-grey-200 dark:border-white/10 rounded-lg pl-9 pr-3 py-2 text-xs outline-none focus:border-mui-blue-500 transition-all shadow-sm text-mui-grey-700 dark:text-mui-grey-200" /></div></div>
            <div className="flex-1 overflow-y-auto px-2 space-y-1">
              {categories.map(cat => (
                <button key={cat.id} onClick={() => setActiveCategory(cat.id)} className={cn("w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-sm font-medium transition-all group", activeCategory === cat.id ? "bg-mui-blue-500 text-white shadow-lg shadow-mui-blue-500/30" : "text-mui-grey-600 dark:text-mui-grey-400 hover:bg-white dark:hover:bg-white/5")}>
                  <div className="flex items-center gap-3"><cat.icon size={16} /><span>{cat.label}</span></div>
                  {activeCategory === cat.id && <ChevronRight size={14} />}
                </button>
              ))}
            </div>
            <div className="p-4 border-t border-mui-grey-200 dark:border-white/5"><div className="p-3 rounded-lg bg-mui-blue-500/5 border border-mui-blue-500/10"><p className="text-[10px] text-mui-grey-500 dark:text-mui-grey-400 leading-relaxed italic">"Productivity is never an accident. It is always the result of a commitment to excellence."</p></div></div>
          </div>
          <div className="flex-1 overflow-y-auto bg-white dark:bg-mui-dark-bg p-10"><div className="max-w-3xl mx-auto h-full">{renderContent()}</div></div>
        </div>
        <div className="h-12 shrink-0 border-t border-mui-grey-200 dark:border-white/10 flex items-center justify-end px-6 gap-3 bg-mui-grey-50 dark:bg-mui-dark-paper"><button onClick={onClose} className="px-6 py-2 bg-mui-blue-500 hover:bg-mui-blue-600 text-white rounded text-xs font-black uppercase tracking-widest transition-all shadow-lg shadow-mui-blue-500/20">Close</button></div>
      </div>
    </div>
  );
};

export default SettingsModal;
