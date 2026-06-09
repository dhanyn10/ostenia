import React, { useState, useEffect, useRef } from 'react';
import { Terminal as XTerm } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import * as AppBackend from '../../wailsjs/go/main/App';
import { X, Maximize2, Minimize2, Folder, File, ChevronLeft, ChevronRight, RefreshCw, Upload, Download, Edit3, Trash2, Home, Search, Terminal } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

const SSHSessionView = ({ session, onClose, addToast }) => {
  const terminalRef = useRef(null);
  const xterm = useRef(null);
  const fitAddon = useRef(new FitAddon());
  const [connecting, setConnecting] = useState(true);
  const [remotePath, setRemotePath] = useState('');
  const [files, setFiles] = useState([]);
  const [loadingFiles, setLoadingFiles] = useState(false);
  const [explorerVisible, setExplorerVisible] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    // Initialize XTerm
    xterm.current = new XTerm({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, "Courier New", monospace',
      theme: {
        background: '#0f172a',
        foreground: '#e2e8f0',
        cursor: '#3b82f6',
        selectionBackground: 'rgba(59, 130, 246, 0.3)',
      },
      allowProposedApi: true,
    });

    xterm.current.loadAddon(fitAddon.current);
    xterm.current.open(terminalRef.current);
    fitAddon.current.fit();

    xterm.current.onData(data => {
      AppBackend.SendSSHInput(session.id, data);
    });

    // Handle incoming data
    const handleOutput = (event) => {
      if (event.sessionId === session.id && xterm.current) {
        xterm.current.write(event.data);
      }
    };

    const handlePathChange = (event) => {
      if (event.sessionId === session.id) {
        setRemotePath(prev => {
            if (event.path !== prev) {
                loadRemoteFiles(event.path);
            }
            return event.path;
        });
      }
    };

    const handleDisconnect = (id) => {
      if (id === session.id && xterm.current) {
        xterm.current.write('\r\n\x1b[31mDisconnected from server.\x1b[0m\r\n');
        addToast('SSH', 'Disconnected from ' + session.name, 'warn');
      }
    };

    EventsOn('ssh_output', handleOutput);
    EventsOn('ssh_path_changed', handlePathChange);
    EventsOn('ssh_disconnected', handleDisconnect);

    // Connect
    connectSSH();

    const handleResize = () => {
      fitAddon.current.fit();
    };
    window.addEventListener('resize', handleResize);

    return () => {
      EventsOff('ssh_output', handleOutput);
      EventsOff('ssh_path_changed', handlePathChange);
      EventsOff('ssh_disconnected', handleDisconnect);
      window.removeEventListener('resize', handleResize);
      // Removed AppBackend.DisconnectSSH here to persist session on tab switch
      if (xterm.current) xterm.current.dispose();
    };
  }, [session.id]);

  const connectSSH = async () => {
    setConnecting(true);
    xterm.current.write(`Connecting to ${session.host}...\r\n`);
    try {
      await AppBackend.ConnectSSH(session);
      setConnecting(false);
      xterm.current.write('\x1b[32mConnected successfully.\x1b[0m\r\n\r\n');
      loadRemoteFiles('');
    } catch (err) {
      setConnecting(false);
      xterm.current.write(`\x1b[31mConnection failed: ${err}\x1b[0m\r\n`);
      addToast('Error', 'SSH connection failed: ' + err, 'error');
    }
  };

  const loadRemoteFiles = async (path) => {
    setLoadingFiles(true);
    try {
      const result = await AppBackend.GetRemoteFiles(session.id, path);
      setFiles(result || []);
      const current = await AppBackend.GetRemoteCurrentPath(session.id);
      setRemotePath(current);
    } catch (err) {
      addToast('Explorer', 'Failed to list files: ' + err, 'error');
    } finally {
      setLoadingFiles(false);
    }
  };

  const syncExplorer = async () => {
      try {
          const current = await AppBackend.GetRemoteCurrentPath(session.id);
          if (current !== remotePath) {
              loadRemoteFiles(current);
          }
      } catch(e) {}
  };

  const handleFileClick = (file) => {
    if (file.isDir) {
      const newPath = remotePath === '/' || remotePath === '' ? `/${file.name}` : `${remotePath}/${file.name}`;
      loadRemoteFiles(newPath);
      // Sync terminal
      AppBackend.SendSSHInput(session.id, `cd "${newPath}"\r`);
    }
  };

  const navigateUp = () => {
    if (remotePath === '/' || remotePath === '') return;
    const parts = remotePath.split('/');
    parts.pop();
    const newPath = parts.join('/') || '/';
    loadRemoteFiles(newPath);
    AppBackend.SendSSHInput(session.id, `cd "${newPath}"\r`);
  };

  const handleDownload = async (file) => {
      try {
          await AppBackend.DownloadRemoteFile(session.id, `${remotePath}/${file.name}`);
          addToast('Success', 'File download started');
      } catch (err) {
          addToast('Error', 'Download failed: ' + err, 'error');
      }
  };

  const handleUpload = async () => {
      try {
          await AppBackend.UploadRemoteFile(session.id, remotePath);
          addToast('Success', 'File uploaded successfully');
          loadRemoteFiles(remotePath);
      } catch (err) {
          addToast('Error', 'Upload failed: ' + err, 'error');
      }
  };

  const handleEdit = async (file) => {
      try {
          addToast('Editor', 'Opening ' + file.name + '...', 'info');
          await AppBackend.EditRemoteFile(session.id, `${remotePath}/${file.name}`);
          addToast('Success', 'File saved and uploaded');
          loadRemoteFiles(remotePath);
      } catch (err) {
          addToast('Error', 'Edit failed: ' + err, 'error');
      }
  };

  const handleDelete = async (file) => {
      if (confirm(`Are you sure you want to delete ${file.name}?`)) {
          try {
              await AppBackend.ExecuteSFTPAction(session.id, 'delete', `${remotePath}/${file.name}`, '');
              addToast('Success', 'Deleted ' + file.name);
              loadRemoteFiles(remotePath);
          } catch (err) {
              addToast('Error', 'Delete failed: ' + err, 'error');
          }
      }
  };

  const filteredFiles = files.filter(f => f.name.toLowerCase().includes(searchQuery.toLowerCase()));

  return (
    <div className="flex flex-col h-full bg-[#0f172a] rounded-xl overflow-hidden border border-slate-200 dark:border-white/5 shadow-2xl">
      {/* Toolbar */}
      <div className="h-12 flex items-center justify-between px-4 bg-slate-900 border-b border-white/5">
        <div className="flex items-center gap-3">
            <div className="flex items-center gap-2 px-2 py-1 bg-white/5 rounded text-blue-400 font-mono text-sm">
                <Terminal size={14} />
                <span>{session.name}</span>
            </div>
            <div className="h-4 w-[1px] bg-white/10 mx-1" />
            <button
                onClick={() => setExplorerVisible(!explorerVisible)}
                className={cn(
                    "p-1.5 rounded transition-colors",
                    explorerVisible ? "bg-blue-600 text-white" : "text-slate-400 hover:text-white hover:bg-white/5"
                )}
                title="Toggle File Explorer"
            >
                <Folder size={18} />
            </button>
        </div>

        <div className="flex items-center gap-2">
            <button
                onClick={() => connectSSH()}
                disabled={connecting}
                className="p-1.5 text-slate-400 hover:text-white hover:bg-white/5 rounded disabled:opacity-50"
                title="Reconnect"
            >
                <RefreshCw size={18} className={connecting ? "animate-spin" : ""} />
            </button>
            <button
                onClick={onClose}
                className="p-1.5 text-slate-400 hover:text-red-500 hover:bg-red-500/10 rounded transition-colors"
                title="Close Session"
            >
                <X size={20} />
            </button>
        </div>
      </div>

      <div className="flex-1 flex overflow-hidden">
        {/* Remote File Explorer */}
        {explorerVisible && (
            <div className="w-80 flex flex-col border-r border-white/5 bg-[#0f172a]">
                <div className="p-3 space-y-3">
                    <div className="flex items-center gap-1">
                        <button onClick={navigateUp} className="p-1.5 hover:bg-white/5 rounded text-slate-400">
                            <ChevronLeft size={18} />
                        </button>
                        <div className="flex-1 px-2 py-1.5 bg-white/5 rounded text-xs text-slate-400 truncate font-mono">
                            {remotePath || '/'}
                        </div>
                        <button onClick={() => loadRemoteFiles(remotePath)} className="p-1.5 hover:bg-white/5 rounded text-slate-400">
                            <RefreshCw size={16} />
                        </button>
                    </div>
                    <div className="relative">
                        <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-500" size={14} />
                        <input
                            type="text"
                            placeholder="Search files..."
                            className="w-full bg-white/5 border-none rounded-md py-1.5 pl-8 pr-3 text-xs text-slate-300 focus:ring-1 focus:ring-blue-500"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                        />
                    </div>
                    <div className="flex gap-2">
                        <button onClick={handleUpload} className="flex-1 flex items-center justify-center gap-1.5 py-1.5 bg-white/5 hover:bg-white/10 rounded text-xs text-slate-300 transition-colors">
                            <Upload size={14} /> Upload
                        </button>
                        <button
                            onClick={async () => {
                                const name = prompt('Enter folder name:');
                                if (name) {
                                    try {
                                        await AppBackend.ExecuteSFTPAction(session.id, 'mkdir', `${remotePath}/${name}`, '');
                                        loadRemoteFiles(remotePath);
                                    } catch (e) { addToast('Error', e.toString(), 'error'); }
                                }
                            }}
                            className="flex-1 flex items-center justify-center gap-1.5 py-1.5 bg-white/5 hover:bg-white/10 rounded text-xs text-slate-300 transition-colors"
                        >
                            <Folder size={14} /> New Folder
                        </button>
                    </div>
                </div>

                <div className="flex-1 overflow-y-auto px-2 pb-4">
                    {loadingFiles ? (
                        <div className="flex flex-col items-center justify-center h-40 space-y-2">
                            <RefreshCw className="animate-spin text-slate-600" />
                            <span className="text-xs text-slate-500">Reading directory...</span>
                        </div>
                    ) : (
                        <div className="space-y-0.5">
                            {filteredFiles.sort((a, b) => (b.isDir ? 1 : 0) - (a.isDir ? 1 : 0)).map((file) => (
                                <div
                                    key={file.name}
                                    className="group flex items-center gap-2 p-2 rounded-md hover:bg-white/5 cursor-pointer text-slate-400 hover:text-white"
                                    onClick={() => handleFileClick(file)}
                                >
                                    {file.isDir ? <Folder size={16} className="text-blue-400 shrink-0" /> : <File size={16} className="text-slate-500 shrink-0" />}
                                    <span className="flex-1 text-xs truncate">{file.name}</span>
                                    <div className="hidden group-hover:flex items-center gap-1">
                                        <button
                                            onClick={async (e) => {
                                                e.stopPropagation();
                                                const newName = prompt('Rename to:', file.name);
                                                if (newName && newName !== file.name) {
                                                    try {
                                                        await AppBackend.ExecuteSFTPAction(session.id, 'rename', `${remotePath}/${file.name}`, `${remotePath}/${newName}`);
                                                        loadRemoteFiles(remotePath);
                                                    } catch (err) { addToast('Error', err.toString(), 'error'); }
                                                }
                                            }}
                                            className="p-1 hover:text-blue-400"
                                            title="Rename"
                                        >
                                            <Edit2 size={12} />
                                        </button>
                                        {!file.isDir && (
                                            <>
                                                <button onClick={(e) => { e.stopPropagation(); handleEdit(file); }} className="p-1 hover:text-blue-400" title="Edit"><Edit3 size={12} /></button>
                                                <button onClick={(e) => { e.stopPropagation(); handleDownload(file); }} className="p-1 hover:text-green-400" title="Download"><Download size={12} /></button>
                                            </>
                                        )}
                                        <button onClick={(e) => { e.stopPropagation(); handleDelete(file); }} className="p-1 hover:text-red-400" title="Delete"><Trash2 size={12} /></button>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        )}

        {/* Terminal */}
        <div className="flex-1 flex flex-col bg-[#0f172a] relative">
          <div ref={terminalRef} className="flex-1 w-full h-full p-4" />
          {connecting && (
            <div className="absolute inset-0 bg-[#0f172a]/50 backdrop-blur-sm flex items-center justify-center">
              <div className="flex items-center gap-3 bg-slate-900 p-4 rounded-xl border border-white/10 shadow-2xl">
                <RefreshCw className="animate-spin text-blue-500" />
                <span className="text-slate-200 font-medium">Establishing connection...</span>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default SSHSessionView;
