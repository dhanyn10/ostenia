import React, { useState, useEffect, useRef } from 'react';
import PropTypes from 'prop-types';
import { Terminal as XTerm } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import * as AppBackend from '../../wailsjs/go/backend/App';

import SSHToolbar from './ssh/SSHToolbar';
import SSHFileExplorer from './ssh/SSHFileExplorer';
import SSHTerminal from './ssh/SSHTerminal';

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

    useEffect(() => {
        setEditingPath(remotePath);
    }, [remotePath]);

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
        } catch (e) {
            console.error("Fit error:", e);
        }
    };

    useEffect(() => {
        if (xterm.current) {
            const newTheme = theme === 'dark' ? {
                background: '#121212',
                foreground: '#eeeeee',
                cursor: '#2196f3',
                selectionBackground: 'rgba(33, 150, 243, 0.3)',
            } : {
                background: '#fafafa',
                foreground: '#212121',
                cursor: '#1976d2',
                selectionBackground: 'rgba(25, 118, 210, 0.2)',
            };
            xterm.current.options.theme = newTheme;
        }
    }, [theme]);

    useEffect(() => {
        if (isActive) {
            setTimeout(performFit, 100);
            setTimeout(performFit, 600);
            setTimeout(performFit, 1500);
        }
    }, [isActive]);

    useEffect(() => {
        xterm.current = new XTerm({
            cursorBlink: true,
            convertEol: true,
            fontSize: 14,
            lineHeight: 1.2,
            fontFamily: 'JetBrains Mono, Menlo, Monaco, Consolas, "Courier New", monospace',
            theme: theme === 'dark' ? {
                background: '#121212',
                foreground: '#eeeeee',
                cursor: '#2196f3',
                selectionBackground: 'rgba(33, 150, 243, 0.3)',
            } : {
                background: '#fafafa',
                foreground: '#212121',
                cursor: '#1976d2',
                selectionBackground: 'rgba(25, 118, 210, 0.2)',
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
        const ro = new ResizeObserver(() => {
            clearTimeout(resizeTimeout);
            resizeTimeout = setTimeout(performFit, 250);
        });

        if (terminalRef.current) ro.observe(terminalRef.current);

        setTimeout(performFit, 500);
        setTimeout(performFit, 1500);

        xterm.current.onData(data => {
            AppBackend.SendSSHInput(session.id, data);
        });
        xterm.current.onTitleChange(title => {
            if (title.includes(':')) {
                const parts = title.split(':');
                const potentialPath = parts[parts.length - 1].trim();
                if (potentialPath.startsWith('/') || potentialPath.startsWith('~')) {
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

    const loadRemoteFiles = async (path, isManualEntry = false) => {
        setLoadingFiles(true);
        try {
            const result = await AppBackend.GetRemoteFiles(session.id, path);
            if (result === null && isManualEntry) {
                throw new Error("Directory not available");
            }
            setFiles(result || []);
            if (path !== '') {
                setRemotePath(path);
                currentPathRef.current = path;
            } else {
                const current = await AppBackend.GetRemoteCurrentPath(session.id);
                if (current) {
                    setRemotePath(current);
                    currentPathRef.current = current;
                }
            }
        } catch (err) {
            if (isManualEntry) {
                addToast('Navigation', 'Directory not available', 'error');
                setEditingPath(remotePath);
            } else {
                addToast('Explorer', 'Failed to list files: ' + err, 'error');
            }
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
            let normalized = current.trim();
            if (normalized.length > 1 && normalized.endsWith('/')) {
                normalized = normalized.substring(0, normalized.length - 1);
            }
            if (normalized !== currentPathRef.current) {
                loadRemoteFiles(normalized);
            }
        } catch(e) {}
    };

    const handleFileDoubleClick = (file) => {
        if (file.isDir) {
            const current = remotePath || '/';
            const newPath = current.endsWith('/') ? `${current}${file.name}` : `${current}/${file.name}`;
            loadRemoteFiles(newPath);
            AppBackend.SendSSHInput(session.id, `cd "${newPath}"\r`);
        } else {
            handleEdit(file);
        }
    };

    const navigateUp = () => {
        if (remotePath === '/' || remotePath === '') return;
        const normalized = remotePath.endsWith('/') ? remotePath.slice(0, -1) : remotePath;
        const parts = normalized.split('/');
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

    return (
        <div className="flex flex-col h-full bg-white dark:bg-mui-dark-bg overflow-hidden border-t border-mui-grey-200 dark:border-white/5">
            <SSHToolbar
                explorerVisible={explorerVisible}
                setExplorerVisible={setExplorerVisible}
                onFit={performFit}
                onReconnect={connectSSH}
                onClose={onClose}
                connecting={connecting}
            />

            <div className="flex-1 flex overflow-hidden">
                {explorerVisible && (
                    <SSHFileExplorer
                        session={session}
                        remotePath={remotePath}
                        editingPath={editingPath}
                        setEditingPath={setEditingPath}
                        files={files}
                        loadingFiles={loadingFiles}
                        loadRemoteFiles={loadRemoteFiles}
                        syncExplorer={syncExplorer}
                        navigateUp={navigateUp}
                        handleUpload={handleUpload}
                        handleFileDoubleClick={handleFileDoubleClick}
                        handleEdit={handleEdit}
                        handleDownload={handleDownload}
                        handleDelete={handleDelete}
                        addToast={addToast}
                    />
                )}
                <SSHTerminal ref={terminalRef} connecting={connecting} />
            </div>
        </div>
    );
};

SSHSessionView.propTypes = {
    session: PropTypes.shape({
        id: PropTypes.string.isRequired,
        host: PropTypes.string.isRequired,
        name: PropTypes.string,
    }).isRequired,
    onClose: PropTypes.func.isRequired,
    addToast: PropTypes.func.isRequired,
    isActive: PropTypes.bool.isRequired,
    theme: PropTypes.string.isRequired,
};

export default SSHSessionView;
