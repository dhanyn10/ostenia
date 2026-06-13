import React, { useState, useEffect } from 'react';
import {
 X,
 User,
 Sliders,
 Terminal as TerminalIcon,
 Search,
 ChevronRight,
 FolderOpen,
 Globe,
 ShieldCheck,
 Monitor,
 Download,
 Upload,
 Plus,
 Trash2,
 Edit2,
 Settings
} from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import * as AppBackend from '../../wailsjs/go/main/App';

function cn(...inputs) {
 return twMerge(clsx(inputs));
}

const SettingsModal = ({ isOpen, onClose, initialCategory = 'profile', appConfig = {}, setConfig, theme, initApp }) => {
 const [activeCategory, setActiveCategory] = useState(initialCategory);
 const [searchQuery, setSearchQuery] = useState('');
 const [sshSessions, setSshSessions] = useState([]);

 useEffect(() => {
 if (isOpen) {
 setActiveCategory(initialCategory);
 loadSSHSessions();
 }
 }, [isOpen, initialCategory]);

 const loadSSHSessions = async () => {
 try {
 const sessions = await AppBackend.GetSSHSessions();
 setSshSessions(sessions || []);
 } catch (err) {
 console.error(err);
 }
 };

 if (!isOpen) return null;

 const categories = [
 { id: 'profile', label: 'Profile', icon: User },
 { id: 'config', label: 'Global Config', icon: Sliders },
 { id: 'ssh', label: 'SSH Management', icon: TerminalIcon },
 ];

 const handleExport = async (type) => {
 try {
 await AppBackend.ExportProfile(type === 'all' || type === 'config', type === 'all' || type === 'ssh');
 } catch (err) { console.error(err); }
 };

 const handleImport = async () => {
 try {
 await AppBackend.ImportProfile();
 if (initApp) initApp();
 } catch (err) { console.error(err); }
 };

 const renderContent = () => {
 switch (activeCategory) {
 case 'profile':
 return (
 <div className="space-y-8 animate-in fade-in slide-in-from-right-4 duration-300">
 <div>
 <h3 className="text-xl font-bold mb-1 tracking-tight text-mui-grey-900 dark:text-white">User Profile</h3>
 <p className="text-sm text-mui-grey-500 dark:text-mui-grey-400">Manage your application profile and portable data.</p>
 </div>

 <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
 <button
 onClick={handleImport}
 className="flex items-start gap-4 p-4 rounded-lg border border-mui-grey-200 dark:border-white/10 hover:bg-mui-grey-50 dark:hover:bg-white/5 transition-all group"
 >
 <div className="p-3 rounded-full bg-mui-blue-500/10 text-mui-blue-500 group-hover:scale-110 transition-transform">
 <Upload size={24} />
 </div>
 <div className="text-left">
 <div className="font-bold text-mui-grey-900 dark:text-white">Import Profile</div>
 <p className="text-[11px] text-mui-grey-500 mt-1 uppercase tracking-wider font-bold">Restore settings from JSON</p>
 </div>
 </button>

 <button
 onClick={() => handleExport('all')}
 className="flex items-start gap-4 p-4 rounded-lg border border-mui-grey-200 dark:border-white/10 hover:bg-mui-grey-50 dark:hover:bg-white/5 transition-all group"
 >
 <div className="p-3 rounded-full bg-emerald-500/10 text-emerald-500 group-hover:scale-110 transition-transform">
 <Download size={24} />
 </div>
 <div className="text-left">
 <div className="font-bold text-mui-grey-900 dark:text-white">Export All</div>
 <p className="text-[11px] text-mui-grey-500 mt-1 uppercase tracking-wider font-bold">Backup config and SSH sessions</p>
 </div>
 </button>
 </div>

 <div className="pt-4 border-t border-mui-grey-100 dark:border-white/5">
 <h4 className="text-xs font-black text-mui-grey-400 uppercase tracking-[0.2em] mb-4">Granular Export</h4>
 <div className="flex gap-3">
 <button onClick={() => handleExport('config')} className="px-4 py-2 rounded border border-mui-grey-200 dark:border-white/10 text-xs font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-50 dark:hover:bg-white/5 transition-colors">Config Only</button>
 <button onClick={() => handleExport('ssh')} className="px-4 py-2 rounded border border-mui-grey-200 dark:border-white/10 text-xs font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-50 dark:hover:bg-white/5 transition-colors">SSH Sessions Only</button>
 </div>
 </div>
 </div>
 );
 case 'config':
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

 <div className="grid grid-cols-2 gap-6 pt-4 border-t border-mui-grey-100 dark:border-white/5">
 <div className="space-y-4">
 <label className="text-[11px] font-black uppercase tracking-widest text-mui-grey-400 flex items-center gap-2">
 <ShieldCheck size={12} /> HTTPS Support
 </label>
 <div className="space-y-3">
 <label className="flex items-center justify-between group cursor-pointer">
 <span className="text-sm text-mui-grey-700 dark:text-mui-grey-200">Apache HTTPS</span>
 <input
 type="checkbox"
 checked={!!appConfig?.apacheHttps}
 onChange={async (e) => {
 const next = e.target.checked;
 await AppBackend.SetApacheHTTPS(next);
 if (initApp) initApp();
 }}
 className="w-4 h-4 accent-mui-blue-500"
 />
 </label>
 <label className="flex items-center justify-between group cursor-pointer">
 <span className="text-sm text-mui-grey-700 dark:text-mui-grey-200">Nginx HTTPS</span>
 <input
 type="checkbox"
 checked={!!appConfig?.nginxHttps}
 onChange={async (e) => {
 const next = e.target.checked;
 await AppBackend.SetNginxHTTPS(next);
 if (initApp) initApp();
 }}
 className="w-4 h-4 accent-mui-blue-500"
 />
 </label>
 </div>
 </div>

 <div className="space-y-4">
 <label className="text-[11px] font-black uppercase tracking-widest text-mui-grey-400 flex items-center gap-2">
 <Monitor size={12} /> External Editor
 </label>
 <div className="space-y-2">
 <div className="flex gap-2">
 <input
 type="text"
 readOnly
 placeholder="Using system default..."
 value={appConfig?.defaultEditor || ''}
 className="flex-1 bg-mui-grey-50 dark:bg-white/5 border border-mui-grey-200 dark:border-white/10 rounded px-3 py-2 text-sm outline-none text-mui-grey-700 dark:text-mui-grey-200"
 />
 <button
 onClick={async () => {
 await AppBackend.SelectDefaultEditor();
 if (initApp) initApp();
 }}
 className="px-4 py-2 bg-mui-blue-500 text-white rounded text-xs font-bold hover:bg-mui-blue-600 transition-colors"
 >
 Set
 </button>
 </div>
 <p className="text-[10px] text-mui-grey-400 italic">Used for "Edit Remote File" in SSH explorer.</p>
 </div>
 </div>
 </div>
 </div>
 </div>
 );
 case 'ssh':
 return (
 <div className="space-y-8 animate-in fade-in slide-in-from-right-4 duration-300 flex flex-col h-full">
 <div className="shrink-0">
 <h3 className="text-xl font-bold mb-1 tracking-tight text-mui-grey-900 dark:text-white">SSH Management</h3>
 <p className="text-sm text-mui-grey-500 dark:text-mui-grey-400">Manage saved connections and credentials.</p>
 </div>

 <div className="flex-1 min-h-0 border border-mui-grey-200 dark:border-white/10 rounded-lg overflow-hidden flex flex-col bg-mui-grey-50 dark:bg-white/5">
 <div className="px-4 py-3 border-b border-mui-grey-200 dark:border-white/10 flex justify-between items-center bg-white dark:bg-mui-dark-paper">
 <span className="text-xs font-black uppercase tracking-widest text-mui-grey-400">{sshSessions.length} Saved Connections</span>
 <button className="flex items-center gap-1.5 px-3 py-1.5 bg-mui-blue-500 text-white rounded text-[11px] font-bold hover:bg-mui-blue-600 transition-colors">
 <Plus size={14} /> New Session
 </button>
 </div>

 <div className="flex-1 overflow-y-auto p-2 space-y-1">
 {sshSessions.map(session => (
 <div
 key={session.id}
 className="flex items-center justify-between p-3 rounded hover:bg-white dark:hover:bg-white/5 group border border-transparent hover:border-mui-grey-200 dark:hover:border-white/10 transition-all"
 >
 <div className="flex items-center gap-4">
 <div className="w-8 h-8 rounded-full bg-mui-grey-200 dark:bg-white/10 flex items-center justify-center text-mui-grey-500">
 <TerminalIcon size={16} />
 </div>
 <div>
 <div className="font-bold text-sm text-mui-grey-900 dark:text-white">{session.name}</div>
 <div className="text-[11px] text-mui-grey-400">{session.user}@{session.host}:{session.port}</div>
 </div>
 </div>
 <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
 <button className="p-1.5 hover:bg-mui-grey-100 dark:hover:bg-white/10 rounded text-mui-grey-400 hover:text-mui-blue-500 transition-colors">
 <Edit2 size={14} />
 </button>
 <button
 onClick={async () => {
 if (confirm('Delete this SSH session?')) {
 await AppBackend.DeleteSSHSession(session.id);
 loadSSHSessions();
 }
 }}
 className="p-1.5 hover:bg-rose-500/10 rounded text-mui-grey-400 hover:text-rose-500 transition-colors"
 >
 <Trash2 size={14} />
 </button>
 </div>
 </div>
 ))}
 {sshSessions.length === 0 && (
 <div className="h-full flex flex-col items-center justify-center text-mui-grey-400 gap-2 py-12">
 <TerminalIcon size={32} strokeWidth={1} />
 <p className="text-xs font-bold uppercase tracking-widest">No SSH sessions saved</p>
 </div>
 )}
 </div>
 </div>
 </div>
 );
 default:
 return (
 <div className="flex items-center justify-center h-full text-mui-grey-400">
 Select a category from the sidebar
 </div>
 );
 }
 };

 return (
 <div className="fixed inset-0 z-[200] flex items-center justify-center p-8 bg-black/40 animate-in fade-in duration-300">
 <div
 className={cn(
 "w-full max-w-5xl h-[80vh] flex flex-col rounded-xl shadow-[0_32px_64px_-12px_rgba(0,0,0,0.5)] border overflow-hidden",
 "bg-white dark:bg-mui-dark-bg border-mui-grey-200 dark:border-white/10"
 )}
 >
 {/* Header */}
 <div className="h-14 shrink-0 flex items-center justify-between px-6 border-b border-mui-grey-200 dark:border-white/10 bg-mui-grey-50/50 dark:bg-white/5">
 <div className="flex items-center gap-3">
 <div className="w-8 h-8 bg-mui-blue-500 rounded flex items-center justify-center text-white shadow-lg shadow-mui-blue-500/30">
 <Settings size={18} />
 </div>
 <div>
 <h2 className="text-base font-black uppercase tracking-wider text-mui-grey-900 dark:text-white">Settings</h2>
 <div className="text-[10px] text-mui-grey-400 font-bold uppercase tracking-[0.2em] -mt-1">Ostenia Management</div>
 </div>
 </div>
 <button
 onClick={onClose}
 className="p-2 hover:bg-mui-grey-200 dark:hover:bg-white/10 rounded-full transition-colors text-mui-grey-500 dark:text-mui-grey-400"
 >
 <X size={20} />
 </button>
 </div>

 <div className="flex-1 flex min-h-0">
 {/* Sidebar */}
 <div className="w-64 shrink-0 border-r border-mui-grey-200 dark:border-white/10 flex flex-col bg-mui-grey-50 dark:bg-mui-dark-paper">
 <div className="p-4">
 <div className="relative group">
 <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-mui-grey-400 group-focus-within:text-mui-blue-500 transition-colors" />
 <input
 type="text"
 placeholder="Search settings..."
 value={searchQuery}
 onChange={(e) => setSearchQuery(e.target.value)}
 className="w-full bg-white dark:bg-white/5 border border-mui-grey-200 dark:border-white/10 rounded-lg pl-9 pr-3 py-2 text-xs outline-none focus:border-mui-blue-500 transition-all shadow-sm text-mui-grey-700 dark:text-mui-grey-200"
 />
 </div>
 </div>

 <div className="flex-1 overflow-y-auto px-2 space-y-1">
 {categories.map(cat => (
 <button
 key={cat.id}
 onClick={() => setActiveCategory(cat.id)}
 className={cn(
 "w-full flex items-center justify-between px-3 py-2.5 rounded-lg text-sm font-medium transition-all group",
 activeCategory === cat.id
 ? "bg-mui-blue-500 text-white shadow-lg shadow-mui-blue-500/30"
 : "text-mui-grey-600 dark:text-mui-grey-400 hover:bg-white dark:hover:bg-white/5"
 )}
 >
 <div className="flex items-center gap-3">
 <cat.icon size={16} />
 <span>{cat.label}</span>
 </div>
 {activeCategory === cat.id && <ChevronRight size={14} />}
 </button>
 ))}
 </div>

 <div className="p-4 border-t border-mui-grey-200 dark:border-white/5">
 <div className="p-3 rounded-lg bg-mui-blue-500/5 border border-mui-blue-500/10">
 <p className="text-[10px] text-mui-grey-500 dark:text-mui-grey-400 leading-relaxed italic">
 "Productivity is never an accident. It is always the result of a commitment to excellence."
 </p>
 </div>
 </div>
 </div>

 {/* Main Content */}
 <div className="flex-1 overflow-y-auto bg-white dark:bg-mui-dark-bg p-10">
 <div className="max-w-3xl mx-auto h-full">
 {renderContent()}
 </div>
 </div>
 </div>

 {/* Footer */}
 <div className="h-12 shrink-0 border-t border-mui-grey-200 dark:border-white/10 flex items-center justify-end px-6 gap-3 bg-mui-grey-50 dark:bg-mui-dark-paper">
 <button
 onClick={onClose}
 className="px-6 py-2 bg-mui-blue-500 hover:bg-mui-blue-600 text-white rounded text-xs font-black uppercase tracking-widest transition-all shadow-lg shadow-mui-blue-500/20"
 >
 Close
 </button>
 </div>
 </div>
 </div>
 );
};

export default SettingsModal;
