import React, { useState, useEffect, useRef } from 'react';
import { Terminal as XTerm } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { WebglAddon } from '@xterm/addon-webgl';
import '@xterm/xterm/css/xterm.css';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import * as AppBackend from '../../wailsjs/go/main/App';
import { X, Maximize2, Minimize2, Folder, File, ChevronLeft, RefreshCw, Upload, Download, Edit2, Edit3, Trash2, Home, Search, Terminal } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

const SSHSessionView = ({ session, onClose, addToast, isActive }) => {
  const terminalRef = useRef(null);
  const xterm = useRef(null);
  const fitAddon = useRef(new FitAddon());
  const currentPathRef = useRef('');
  const [connecting, setConnecting] = useState(true);
  const [remotePath, setRemotePath] = useState('');
  const [files, setFiles] = useState([]);
  const [loadingFiles, setLoadingFiles] = useState(false);
  const [explorerVisible, setExplorerVisible] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');

  // Manual Fit function defined outside useEffect so it can be called by multiple effects
  const performFit = () => {
    if (!terminalRef.current || !xterm.current || !isActive) return;

    // Check visibility
    if (terminalRef.current.offsetParent === null) return;

    try {
        fitAddon.current.fit();
        const dims = fitAddon.current.proposeDimensions();

        // Ensure dimensions are sane before notifying backend.
        // We force a minimum width of 100 columns to prevent the "5 character wrap" issue.
        if (dims && dims.cols >= 20 && dims.rows >= 2) {
            const safeCols = Math.max(dims.cols, 120);
            const safeRows = Math.max(dims.rows, 24);

            console.log(`[SSH] Resizing ${session.name} to ${safeCols}x${safeRows}`);
            AppBackend.ResizeSSHTerminal(session.id, safeCols, safeRows);
            setTimeout(() => xterm.current?.focus(), 50);
        }
    } catch (e) {
        console.error("Fit error:", e);
    }
  };

  // Re-fit when tab becomes active (unhidden) or isActive changes
  useEffect(() => {
    if (isActive) {
        // Multiple attempts to ensure the layout has settled
        setTimeout(performFit, 100);
        setTimeout(performFit, 600);
        setTimeout(performFit, 1500);
    }
  }, [isActive]);

  useEffect(() => {
    // Initialize XTerm
    xterm.current = new XTerm({
      cursorBlink: true,
      convertEol: true,
      fontSize: 14,
      lineHeight: 1.2,
      fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, "Courier New", monospace',
      theme: {
        background: '#121212', // mui-dark-bg
        foreground: '#eeeeee', // mui-grey-200
        cursor: '#2196f3',     // mui-blue-500
        selectionBackground: 'rgba(33, 150, 243, 0.3)',
      },
      allowProposedApi: true,
      scrollback: 10000,
      macOptionIsMeta: true,
    });

    xterm.current.loadAddon(fitAddon.current);
    xterm.current.open(terminalRef.current);

    let resizeTimeout;
    const handleWindowResize = () => {
        clearTimeout(resizeTimeout);
        resizeTimeout = setTimeout(performFit, 200);
    };

    window.addEventListener('resize', handleWindowResize);

    // Use ResizeObserver with debounce
    const ro = new ResizeObserver(() => {
        clearTimeout(resizeTimeout);
        resizeTimeout = setTimeout(performFit, 250);
    });

    if (terminalRef.current) {
        ro.observe(terminalRef.current);
    }

    // Multiple fit attempts during initialization
    setTimeout(performFit, 500);
    setTimeout(performFit, 1500);
    setTimeout(performFit, 3000);

    xterm.current.onData(data => {
      AppBackend.SendSSHInput(session.id, data);
    });
    xterm.current.onTitleChange(title => {
        // Extract path from title if it follows common patterns like "user@host: /path"
        if (title.includes(':')) {
            const parts = title.split(':');
            const potentialPath = parts[parts.length - 1].trim();
            if (potentialPath.startsWith('/') || potentialPath.startsWith('~')) {
                // If it's a home relative path like ~ or ~/projects, the backend GetRemoteCurrentPath
                // will return the absolute path. For title sync, we might need more logic but
                // let's try to sync based on raw potentialPath if it differs.
                syncExplorer(potentialPath);
            }
        }
    });

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

    connectSSH();

    return () => {
      EventsOff('ssh_output', handleOutput);
      EventsOff('ssh_path_changed', handlePathChange);
      EventsOff('ssh_disconnected', handleDisconnect);
      if (ro) ro.disconnect();
      window.removeEventListener('resize', handleWindowResize);
      clearTimeout(resizeTimeout);
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

      // Send initial resize again after connection
      const dims = fitAddon.current.proposeDimensions();
      if (dims && dims.cols > 0 && dims.rows > 0) {
          AppBackend.ResizeSSHTerminal(session.id, dims.cols, dims.rows);
      }

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
      if (current) {
          setRemotePath(current);
          currentPathRef.current = current;
      }
    } catch (err) {
      addToast('Explorer', 'Failed to list files: ' + err, 'error');
    } finally {
      setLoadingFiles(false);
    }
  };

  const syncExplorer = async (forcedPath = null) => {
      try {
          let current = forcedPath;

          if (!current) {
              current = await AppBackend.GetRemoteCurrentPath(session.id);
          }

          if (!current) return;

          // Normalize: remove trailing slash and resolve ~ if possible
          let normalized = current.trim();
          if (normalized.length > 1 && normalized.endsWith('/')) {
              normalized = normalized.substring(0, normalized.length - 1);
          }

          // Use the ref for comparison to avoid stale state in callbacks
          if (normalized !== currentPathRef.current) {
              loadRemoteFiles(normalized);
          }
      } catch(e) {}
  };

  const handleFileClick = (file) => {
    if (file.isDir) {
      const newPath = remotePath === '/' || remotePath === '' ? `/${file.name}` : `${remotePath}/${file.name}`;
      loadRemoteFiles(newPath);
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
    <div className="flex flex-col h-full bg-mui-dark-bg overflow-hidden border-t border-mui-grey-200 dark:border-white/5">
      {/* Header / Toolbar */}
      <div className="h-10 flex items-center justify-between px-3 bg-white dark:bg-mui-dark-bg border-b border-mui-grey-100 dark:border-white/5 shrink-0">
        <div className="flex items-center gap-2">
            <button
                onClick={() => setExplorerVisible(!explorerVisible)}
                className={cn(
                    "p-1 rounded text-mui-grey-400 hover:text-white transition-colors",
                    explorerVisible && "text-mui-blue-500 bg-mui-blue-500/10"
                )}
                title="Toggle Explorer"
            >
                <Folder size={16} />
            </button>
        </div>

        <div className="flex items-center gap-1">
            <button
                onClick={() => {
                    fitAddon.current.fit();
                    const dims = fitAddon.current.proposeDimensions();
                    if (dims) AppBackend.ResizeSSHTerminal(session.id, dims.cols, dims.rows);
                }}
                className="p-1 text-mui-grey-500 hover:text-white transition-colors"
                title="Fit Terminal"
            >
                <Maximize2 size={14} />
            </button>
            <button
                onClick={() => connectSSH()}
                disabled={connecting}
                className="p-1 text-mui-grey-500 hover:text-white transition-colors disabled:opacity-30"
                title="Reconnect"
            >
                <RefreshCw size={14} className={connecting ? "animate-spin" : ""} />
            </button>
            <button
                onClick={onClose}
                className="p-1 text-mui-grey-500 hover:text-red-500 transition-colors"
                title="Close"
            >
                <X size={16} />
            </button>
        </div>
      </div>

      <div className="flex-1 flex overflow-hidden">
        {/* SFTP Explorer Sidebar */}
        {explorerVisible && (
            <div className="w-72 flex flex-col border-r border-white/5 bg-mui-dark-bg shrink-0">
                <div className="p-3 border-b border-white/5 space-y-3 bg-mui-grey-900/30">
                    <div className="flex items-center gap-1 bg-white/5 rounded px-1 py-0.5 border border-white/5">
                        <button onClick={navigateUp} className="p-1 text-mui-grey-400 hover:text-white transition-colors" title="Back">
                            <ChevronLeft size={16} />
                        </button>
                        <div className="flex-1 px-1 text-[10px] text-mui-grey-400 truncate font-mono">
                            {remotePath || '/'}
                        </div>
                        <button onClick={() => syncExplorer()} className="p-1 text-mui-grey-400 hover:text-mui-blue-400 transition-colors" title="Sync with terminal">
                            <RefreshCw size={12} />
                        </button>
                    </div>

                    <div className="relative">
                        <Search className="absolute left-2 top-1/2 -translate-y-1/2 text-mui-grey-500" size={12} />
                        <input
                            type="text"
                            placeholder="Search..."
                            className="w-full bg-white/5 border border-white/5 rounded py-1 pl-7 pr-2 text-[11px] text-mui-grey-300 outline-none focus:border-mui-blue-500 transition-all"
                            value={searchQuery}
                            onChange={(e) => setSearchQuery(e.target.value)}
                        />
                    </div>

                    <div className="flex gap-1.5">
                        <button onClick={handleUpload} className="flex-1 flex items-center justify-center gap-1.5 py-1 bg-mui-blue-600 hover:bg-mui-blue-700 rounded text-[10px] font-bold text-white transition-colors">
                            <Upload size={12} /> Upload
                        </button>
                        <button
                            onClick={async () => {
                                const name = prompt('Folder Name:');
                                if (name) {
                                    try {
                                        await AppBackend.ExecuteSFTPAction(session.id, 'mkdir', `${remotePath}/${name}`, '');
                                        loadRemoteFiles(remotePath);
                                    } catch (e) { addToast('Error', e.toString(), 'error'); }
                                }
                            }}
                            className="flex-1 flex items-center justify-center gap-1.5 py-1 bg-white/10 hover:bg-white/20 rounded text-[10px] font-bold text-mui-grey-200 transition-colors"
                        >
                            <Folder size={12} /> New
                        </button>
                    </div>
                </div>

                <div className="flex-1 overflow-y-auto px-1 py-2 custom-scrollbar">
                    {loadingFiles ? (
                        <div className="flex flex-col items-center justify-center h-40 space-y-2">
                            <RefreshCw className="animate-spin text-mui-grey-600" size={20} />
                        </div>
                    ) : (
                        <div className="space-y-px">
                            {filteredFiles.sort((a, b) => (b.isDir ? 1 : 0) - (a.isDir ? 1 : 0)).map((file) => (
                                <div
                                    key={file.name}
                                    className="group flex items-center gap-2 px-2 py-1.5 rounded hover:bg-white/5 cursor-pointer border border-transparent hover:border-white/5 transition-all"
                                    onClick={() => handleFileClick(file)}
                                >
                                    {file.isDir ? <Folder size={14} className="text-mui-blue-400 shrink-0" /> : <File size={14} className="text-mui-grey-500 shrink-0" />}
                                    <span className="flex-1 text-[11px] text-mui-grey-400 group-hover:text-white truncate">{file.name}</span>
                                    <div className="opacity-0 group-hover:opacity-100 flex items-center gap-1 transition-opacity">
                                        <button
                                            onClick={async (e) => {
                                                e.stopPropagation();
                                                const newName = prompt('Rename:', file.name);
                                                if (newName && newName !== file.name) {
                                                    try {
                                                        await AppBackend.ExecuteSFTPAction(session.id, 'rename', `${remotePath}/${file.name}`, `${remotePath}/${newName}`);
                                                        loadRemoteFiles(remotePath);
                                                    } catch (err) { addToast('Error', err.toString(), 'error'); }
                                                }
                                            }}
                                            className="p-1 text-mui-grey-500 hover:text-mui-blue-400"
                                        >
                                            <Edit2 size={10} />
                                        </button>
                                        {!file.isDir && (
                                            <>
                                                <button onClick={(e) => { e.stopPropagation(); handleEdit(file); }} className="p-1 text-mui-grey-500 hover:text-mui-blue-400" title="Edit"><Edit3 size={10} /></button>
                                                <button onClick={(e) => { e.stopPropagation(); handleDownload(file); }} className="p-1 text-mui-grey-500 hover:text-green-400" title="Download"><Download size={10} /></button>
                                            </>
                                        )}
                                        <button onClick={(e) => { e.stopPropagation(); handleDelete(file); }} className="p-1 text-mui-grey-500 hover:text-red-500" title="Delete"><Trash2 size={10} /></button>
                                    </div>
                                </div>
                            ))}
                        </div>
                    )}
                </div>
            </div>
        )}

        {/* Integrated Terminal */}
        <div className="flex-1 bg-mui-dark-bg relative overflow-hidden">
          <div ref={terminalRef} className="absolute inset-0" />
          {connecting && (
            <div className="absolute inset-0 bg-mui-dark-bg flex items-center justify-center">
              <div className="flex items-center gap-3">
                <RefreshCw className="animate-spin text-mui-blue-500" size={18} />
                <span className="text-mui-grey-400 text-xs font-bold uppercase tracking-widest">Connecting...</span>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default SSHSessionView;
