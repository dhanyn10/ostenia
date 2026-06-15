import React, { useState, useEffect, useRef } from 'react';
import { Terminal as XTerm } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import * as AppBackend from '../../wailsjs/go/main/App';
import { X, Maximize2, Folder, RefreshCw } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import TerminalView from './ssh/TerminalView';
import SFTPExplorer from './ssh/SFTPExplorer';
import FileContextMenu from './ssh/FileContextMenu';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

const SSHSessionView = ({ session, onClose, addToast, isActive, theme }) => {
  const terminalRef = useRef(null);
  const xterm = useRef(null);
  const fitAddon = useRef(new FitAddon());
  const currentPathRef = useRef('');
  const [connecting, setConnecting] = useState(true);
  const [remotePath, setRemotePath] = useState('');
  const [editingPath, setEditingPath] = useState('');
  const [files, setFiles] = useState([]);
  const [loadingFiles, setLoadingFiles] = useState(false);
  const [explorerVisible, setExplorerVisible] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [fileContextMenu, setFileContextMenu] = useState(null);
  const [sortConfig, setSortConfig] = useState({ key: 'name', direction: 'asc' });

  useEffect(() => { setEditingPath(remotePath); }, [remotePath]);
  useEffect(() => {
    const handleClick = () => setFileContextMenu(null);
    window.addEventListener('click', handleClick);
    return () => window.removeEventListener('click', handleClick);
  }, []);

  const performFit = () => {
    if (!terminalRef.current || !xterm.current || !isActive) return;
    if (terminalRef.current.offsetParent === null) return;
    try {
      fitAddon.current.fit();
      const dims = fitAddon.current.proposeDimensions();
      if (dims && dims.cols >= 20 && dims.rows >= 2) {
        const safeCols = Math.max(dims.cols, 120);
        const safeRows = Math.max(dims.rows, 24);
        AppBackend.ResizeSSHTerminal(session.id, safeCols, safeRows);
        setTimeout(() => xterm.current?.focus(), 50);
      }
    } catch (e) { console.error("Fit error:", e); }
  };

  useEffect(() => {
    if (xterm.current) {
      xterm.current.options.theme = theme === 'dark' ? {
        background: '#121212', foreground: '#eeeeee', cursor: '#2196f3', selectionBackground: 'rgba(33, 150, 243, 0.3)',
      } : {
        background: '#fafafa', foreground: '#212121', cursor: '#1976d2', selectionBackground: 'rgba(25, 118, 210, 0.2)',
      };
    }
  }, [theme]);

  useEffect(() => { if (isActive) { [100, 600, 1500].forEach(t => setTimeout(performFit, t)); } }, [isActive]);

  useEffect(() => {
    xterm.current = new XTerm({
      cursorBlink: true, convertEol: true, fontSize: 14, lineHeight: 1.2,
      fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, "Courier New", monospace',
      theme: theme === 'dark' ? { background: '#121212', foreground: '#eeeeee', cursor: '#2196f3', selectionBackground: 'rgba(33, 150, 243, 0.3)' } : { background: '#fafafa', foreground: '#212121', cursor: '#1976d2', selectionBackground: 'rgba(25, 118, 210, 0.2)' },
      allowProposedApi: true, scrollback: 10000, macOptionIsMeta: true,
    });
    xterm.current.loadAddon(fitAddon.current);
    xterm.current.open(terminalRef.current);

    let resizeTimeout;
    const handleWindowResize = () => { clearTimeout(resizeTimeout); resizeTimeout = setTimeout(performFit, 200); };
    window.addEventListener('resize', handleWindowResize);
    const ro = new ResizeObserver(() => { clearTimeout(resizeTimeout); resizeTimeout = setTimeout(performFit, 250); });
    if (terminalRef.current) ro.observe(terminalRef.current);
    [500, 1500, 3000].forEach(t => setTimeout(performFit, t));

    xterm.current.onData(data => AppBackend.SendSSHInput(session.id, data));
    xterm.current.onTitleChange(title => {
      if (title.includes(':')) {
        const parts = title.split(':');
        const potentialPath = parts[parts.length - 1].trim();
        if (potentialPath.startsWith('/') || potentialPath.startsWith('~')) syncExplorer(potentialPath);
      }
    });

    const handleOutput = (event) => { if (event.sessionId === session.id && xterm.current) xterm.current.write(event.data); };
    const handlePathChange = (event) => { if (event.sessionId === session.id) setRemotePath(prev => { if (event.path !== prev) loadRemoteFiles(event.path); return event.path; }); };
    const handleDisconnect = (id) => { if (id === session.id && xterm.current) { xterm.current.write('\r\n\x1b[31mDisconnected from server.\x1b[0m\r\n'); addToast('SSH', 'Disconnected from ' + session.name, 'warn'); } };

    EventsOn('ssh_output', handleOutput); EventsOn('ssh_path_changed', handlePathChange); EventsOn('ssh_disconnected', handleDisconnect);
    connectSSH();
    return () => {
      EventsOff('ssh_output', handleOutput); EventsOff('ssh_path_changed', handlePathChange); EventsOff('ssh_disconnected', handleDisconnect);
      if (ro) ro.disconnect(); window.removeEventListener('resize', handleWindowResize); clearTimeout(resizeTimeout);
      if (xterm.current) xterm.current.dispose();
    };
  }, [session.id]);

  const connectSSH = async () => {
    setConnecting(true); xterm.current.write(`Connecting to ${session.host}...\r\n`);
    try {
      await AppBackend.ConnectSSH(session); setConnecting(false); xterm.current.write('\x1b[32mConnected successfully.\x1b[0m\r\n\r\n');
      const dims = fitAddon.current.proposeDimensions(); if (dims && dims.cols > 0 && dims.rows > 0) AppBackend.ResizeSSHTerminal(session.id, dims.cols, dims.rows);
      loadRemoteFiles('');
    } catch (err) { setConnecting(false); xterm.current.write(`\x1b[31mConnection failed: ${err}\x1b[0m\r\n`); addToast('Error', 'SSH connection failed: ' + err, 'error'); }
  };

  const loadRemoteFiles = async (path, isManualEntry = false) => {
    setLoadingFiles(true);
    try {
      const result = await AppBackend.GetRemoteFiles(session.id, path);
      if (result === null && isManualEntry) throw new Error("Directory not available");
      setFiles(result || []);
      if (path !== '') { setRemotePath(path); currentPathRef.current = path; }
      else { const current = await AppBackend.GetRemoteCurrentPath(session.id); if (current) { setRemotePath(current); currentPathRef.current = current; } }
    } catch (err) {
      if (isManualEntry) { addToast('Navigation', 'Directory not available', 'error'); setEditingPath(remotePath); }
      else addToast('Explorer', 'Failed to list files: ' + err, 'error');
    } finally { setLoadingFiles(false); }
  };

  const syncExplorer = async (forcedPath = null) => {
    try {
      let current = forcedPath || await AppBackend.GetRemoteCurrentPath(session.id);
      if (!current) return;
      let normalized = current.trim();
      if (normalized.length > 1 && normalized.endsWith('/')) normalized = normalized.substring(0, normalized.length - 1);
      if (normalized !== currentPathRef.current) loadRemoteFiles(normalized);
    } catch(e) {}
  };

  const handleFileDoubleClick = (file) => {
    if (file.isDir) {
      const current = remotePath || '/';
      const newPath = current.endsWith('/') ? `${current}${file.name}` : `${current}/${file.name}`;
      loadRemoteFiles(newPath); AppBackend.SendSSHInput(session.id, `cd "${newPath}"\r`);
    } else handleEdit(file);
  };

  const navigateUp = () => {
    if (remotePath === '/' || remotePath === '') return;
    const normalized = remotePath.endsWith('/') ? remotePath.slice(0, -1) : remotePath;
    const parts = normalized.split('/'); parts.pop();
    const newPath = parts.join('/') || '/';
    loadRemoteFiles(newPath); AppBackend.SendSSHInput(session.id, `cd "${newPath}"\r`);
  };

  const handleDownload = async (file) => { try { await AppBackend.DownloadRemoteFile(session.id, `${remotePath}/${file.name}`); addToast('Success', 'File download started'); } catch (err) { addToast('Error', 'Download failed: ' + err, 'error'); } };
  const handleUpload = async () => { try { await AppBackend.UploadRemoteFile(session.id, remotePath); addToast('Success', 'File uploaded successfully'); loadRemoteFiles(remotePath); } catch (err) { addToast('Error', 'Upload failed: ' + err, 'error'); } };
  const handleEdit = async (file) => { try { addToast('Editor', 'Opening ' + file.name + '...', 'info'); await AppBackend.EditRemoteFile(session.id, `${remotePath}/${file.name}`); addToast('Success', 'File saved and uploaded'); loadRemoteFiles(remotePath); } catch (err) { addToast('Error', 'Edit failed: ' + err, 'error'); } };
  const handleDelete = async (file) => { if (confirm(`Are you sure you want to delete ${file.name}?`)) { try { await AppBackend.ExecuteSFTPAction(session.id, 'delete', `${remotePath}/${file.name}`, ''); addToast('Success', 'Deleted ' + file.name); loadRemoteFiles(remotePath); } catch (err) { addToast('Error', 'Delete failed: ' + err, 'error'); } } };

  const sortedFiles = [...files].filter(f => f.name.toLowerCase().includes(searchQuery.toLowerCase())).sort((a, b) => {
    if (a.isDir !== b.isDir) return b.isDir ? 1 : -1;
    let comp = sortConfig.key === 'name' ? a.name.localeCompare(b.name) : (a.size || 0) - (b.size || 0);
    return sortConfig.direction === 'asc' ? comp : -comp;
  });

  return (
    <div className="flex flex-col h-full bg-white dark:bg-mui-dark-bg overflow-hidden border-t border-mui-grey-200 dark:border-white/5">
      <div className="h-10 flex items-center justify-between px-3 bg-white dark:bg-mui-dark-bg border-b border-mui-grey-100 dark:border-white/5 shrink-0">
        <div className="flex items-center gap-2">
          <button onClick={() => setExplorerVisible(!explorerVisible)} className={cn("p-1 rounded text-mui-grey-500 dark:text-mui-grey-400 hover:text-mui-blue-600 dark:hover:text-white transition-colors", explorerVisible && "text-mui-blue-600 bg-mui-blue-600/10")} title="Toggle Explorer"><Folder size={16} /></button>
        </div>
        <div className="flex items-center gap-1">
          <button onClick={() => { fitAddon.current.fit(); const dims = fitAddon.current.proposeDimensions(); if (dims) AppBackend.ResizeSSHTerminal(session.id, dims.cols, dims.rows); }} className="p-1 text-mui-grey-500 hover:text-mui-blue-600 dark:hover:text-white transition-colors" title="Fit Terminal"><Maximize2 size={14} /></button>
          <button onClick={() => connectSSH()} disabled={connecting} className="p-1 text-mui-grey-500 hover:text-mui-blue-600 dark:hover:text-white transition-colors disabled:opacity-30" title="Reconnect"><RefreshCw size={14} className={connecting ? "animate-spin" : ""} /></button>
          <button onClick={onClose} className="p-1 text-mui-grey-500 hover:text-red-500 transition-colors" title="Close"><X size={16} /></button>
        </div>
      </div>
      <div className="flex-1 flex overflow-hidden">
        {explorerVisible && <SFTPExplorer {...{remotePath, editingPath, setEditingPath, navigateUp, loadRemoteFiles: (p) => loadRemoteFiles(p, true), syncExplorer, searchQuery, setSearchQuery, handleUpload, handleNewFolder: async () => { const name = prompt('Folder Name:'); if (name) { try { await AppBackend.ExecuteSFTPAction(session.id, 'mkdir', `${remotePath}/${name}`, ''); loadRemoteFiles(remotePath); } catch (e) { addToast('Error', e.toString(), 'error'); } } }, loadingFiles, sortedFiles, handleFileDoubleClick, handleFileContextMenu: (e, file) => { e.preventDefault(); setFileContextMenu({ x: e.clientX, y: e.clientY, file }); }, formatSize: (bytes) => { if (!bytes) return '-'; const units = ['B', 'KB', 'MB', 'GB', 'TB']; let size = bytes, i = 0; while (size >= 1024 && i < units.length - 1) { size /= 1024; i++; } return `${size.toFixed(1)} ${units[i]}`; }, session}} />}
        <TerminalView terminalRef={terminalRef} connecting={connecting} />
      </div>
      {fileContextMenu && <FileContextMenu {...fileContextMenu} remotePath={remotePath} session={session} loadRemoteFiles={loadRemoteFiles} handleEdit={handleEdit} handleDownload={handleDownload} handleDelete={handleDelete} setFileContextMenu={setFileContextMenu} addToast={addToast} AppBackend={AppBackend} />}
    </div>
  );
};

export default SSHSessionView;
