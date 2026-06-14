import React, { useState, useEffect } from 'react';
import { Save, ExternalLink, Search, Folder, Terminal, ChevronDown, Monitor, Activity } from 'lucide-react';
import { EventsOn } from '../../wailsjs/runtime/runtime';
import * as AppBackend from '../../wailsjs/go/main/App';

function ProxyTab({ addToast }) {
 const [apps, setApps] = useState([]);
 const [loading, setLoading] = useState(true);
 const [searchTerm, setSearchTerm] = useState('');
 const [savingMap, setSavingMap] = useState({});
 const [openTerminalDropdown, setOpenTerminalDropdown] = useState(null);
 const [proxyStatuses, setProxyStatuses] = useState({});

 const fetchApps = async () => {
 try {
 const data = await AppBackend.GetProxyApps();
 setApps(data || []);
 } catch (err) {
 console.error(err);
 } finally {
 setLoading(false);
 }
 };

 useEffect(() => {
 fetchApps();

 if (window.runtime) {
 return EventsOn('proxy_status', (data) => {
 const statusMap = {};
 data.forEach(item => {
 statusMap[item.name] = item.isUp;
 });
 setProxyStatuses(statusMap);
 });
 }
 }, []);

 const handlePortChange = (name, value) => {
 setApps(prev => prev.map(app => app.name === name ? { ...app, port: parseInt(value) || 0 } : app));
 };

 const handleSave = async (name, port) => {
 setSavingMap(prev => ({ ...prev, [name]: true }));
 try {
 await AppBackend.SaveProxyPort(name, port);
 addToast('Success', `Proxy for ${name} updated to port ${port}`, 'info');
 } catch (err) {
 addToast('Error', err.toString(), 'error');
 } finally {
 setSavingMap(prev => ({ ...prev, [name]: false }));
 }
 };

 const filteredApps = apps.filter(app => app.name.toLowerCase().includes(searchTerm.toLowerCase()));

 return (
 <div className="flex flex-col h-full animate-in fade-in slide-in-from-bottom-4 duration-500">
 <div className="mb-6 flex justify-between items-center">
 <div>
 <p className="text-slate-500 dark:text-slate-400 text-sm">Configure local proxy pass for folders in your WWW directory.</p>
 </div>
 <div className="relative">
 <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" size={18} />
 <input
 type="text"
 placeholder="Search folders..."
 className="pl-10 pr-4 py-2 bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all w-64"
 value={searchTerm}
 onChange={(e) => setSearchTerm(e.target.value)}
 />
 </div>
 </div>

 {loading ? (
 <div className="flex-1 flex items-center justify-center">
 <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
 </div>
 ) : filteredApps.length === 0 ? (
 <div className="flex-1 flex flex-col items-center justify-center text-slate-400 border-2 border-dashed border-slate-200 dark:border-white/5 rounded-xl">
 <Folder size={48} className="mb-4 opacity-20" />
 <p>No folders found in WWW directory.</p>
 </div>
 ) : (
 <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pb-8 overflow-y-auto">
 {filteredApps.map((app) => (
 <div
 key={app.name}
 className="bg-white dark:bg-white/5 border border-slate-200 dark:border-white/10 p-5 rounded-xl hover:border-blue-500/50 transition-all group"
 >
 <div className="flex items-center justify-between mb-4">
 <div className="flex items-center gap-3">
 <div className="p-2 bg-blue-500/10 text-blue-500 rounded-lg group-hover:bg-blue-500 group-hover:text-white transition-colors">
 <Folder size={20} />
 </div>
 <div>
 <div className="flex items-center gap-2">
 <h3 className="font-semibold text-lg">{app.name}</h3>
 <div className={`w-2 h-2 rounded-full ${proxyStatuses[app.name] ? 'bg-emerald-500 animate-pulse' : 'bg-slate-300 dark:bg-slate-700'}`} title={proxyStatuses[app.name] ? 'Online' : 'Offline'}></div>
 </div>
 <div className="flex items-center gap-1 text-slate-500 text-xs">
 <span className="truncate max-w-[150px]">{app.name}.test</span>
 <a
 href={`http://${app.name}.test`}
 target="_blank"
 rel="noreferrer"
 className="hover:text-blue-500 transition-colors"
 >
 <ExternalLink size={12} />
 </a>
 </div>
 </div>
 </div>

 <div className="relative">
 <button
 onClick={() => setOpenTerminalDropdown(openTerminalDropdown === app.name ? null : app.name)}
 className={`w-10 h-8 flex items-center justify-center gap-1 rounded-lg border border-slate-200 dark:border-white/10 transition-all ${
 openTerminalDropdown === app.name ? "bg-slate-200 dark:bg-slate-700 text-slate-900 dark:text-white" : "bg-slate-50 dark:bg-white/5 text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-white/10"
 }`}
 title="Terminal"
 >
 <Terminal size={16} /> <ChevronDown size={10} />
 </button>

 {openTerminalDropdown === app.name && (
 <>
 <div className="fixed inset-0 z-10" onClick={() => setOpenTerminalDropdown(null)} />
 <div className="absolute top-full right-0 mt-1 w-40 bg-white dark:bg-slate-900 border border-slate-200 dark:border-white/10 rounded-lg shadow-xl z-20 animate-in fade-in slide-in-from-top-1 duration-200 overflow-hidden">
 <div className="p-1">
 <button
 onClick={() => { AppBackend.OpenProxyTerminal(app.name, 'cmd'); setOpenTerminalDropdown(null); }}
 className="w-full flex items-center gap-3 px-3 py-2 rounded-md text-xs font-medium text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white transition-all text-left"
 >
 <Monitor size={14} className="text-blue-500" /> CMD
 </button>
 <button
 onClick={() => { AppBackend.OpenProxyTerminal(app.name, 'powershell'); setOpenTerminalDropdown(null); }}
 className="w-full flex items-center gap-3 px-3 py-2 rounded-md text-xs font-medium text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white transition-all text-left"
 >
 <Monitor size={14} className="text-blue-600" /> PowerShell
 </button>
 </div>
 </div>
 </>
 )}
 </div>
 </div>

 <div className="space-y-4">
 <div>
 <label className="text-xs font-medium text-slate-500 uppercase tracking-wider mb-1.5 block">Target Localhost Port</label>
 <div className="flex gap-2">
 <input
 type="number"
 placeholder="e.g. 3000"
 className="flex-1 px-4 py-2 bg-slate-50 dark:bg-black/20 border border-slate-200 dark:border-white/10 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all font-mono"
 value={app.port || ''}
 onChange={(e) => handlePortChange(app.name, e.target.value)}
 />
 <button
 onClick={() => handleSave(app.name, app.port)}
 disabled={savingMap[app.name]}
 className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-all flex items-center gap-2 shadow-lg shadow-blue-900/20 disabled:opacity-50"
 >
 {savingMap[app.name] ? (
 <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
 ) : (
 <Save size={18} />
 )}
 <span>Save</span>
 </button>
 </div>
 <p className="mt-2 text-[10px] text-slate-400 italic">
 Points {app.name}.test to http://127.0.0.1:{app.port || '____'}
 </p>
 </div>
 </div>
 </div>
 ))}
 </div>
 )}
 </div>
 );
}

export default ProxyTab;
