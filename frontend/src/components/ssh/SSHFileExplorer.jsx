import React, { useState, useEffect } from 'react';
import { Folder, File, ChevronLeft, RefreshCw, Upload, Search, MoreVertical, Edit2, Edit3, Download, Trash2 } from 'lucide-react';
import { twMerge } from 'tailwind-merge';
import { clsx } from 'clsx';
import * as AppBackend from '../../../wailsjs/go/backend/App';

function cn(...inputs) {
    return twMerge(clsx(inputs));
}

const SSHFileExplorer = ({
    session,
    remotePath,
    editingPath,
    setEditingPath,
    files,
    loadingFiles,
    loadRemoteFiles,
    syncExplorer,
    navigateUp,
    handleUpload,
    handleFileDoubleClick,
    handleEdit,
    handleDownload,
    handleDelete,
    addToast
}) => {
    const [searchQuery, setSearchQuery] = useState('');
    const [fileContextMenu, setFileContextMenu] = useState(null);
    const [sortConfig, setSortConfig] = useState({ key: 'name', direction: 'asc' });

    useEffect(() => {
        const handleClick = () => setFileContextMenu(null);
        window.addEventListener('click', handleClick);
        return () => window.removeEventListener('click', handleClick);
    }, []);

    const handleFileContextMenu = (e, file) => {
        e.preventDefault();
        setFileContextMenu({
            x: e.clientX,
            y: e.clientY,
            file: file
        });
    };

    const formatSize = (bytes) => {
        if (!bytes) return '-';
        const units = ['B', 'KB', 'MB', 'GB', 'TB'];
        let size = bytes;
        let unitIndex = 0;
        while (size >= 1024 && unitIndex < units.length - 1) {
            size /= 1024;
            unitIndex++;
        }
        return `${size.toFixed(1)} ${units[unitIndex]}`;
    };

    const sortedFiles = [...files]
        .filter(f => f.name.toLowerCase().includes(searchQuery.toLowerCase()))
        .sort((a, b) => {
            if (a.isDir !== b.isDir) return b.isDir ? 1 : -1;
            let comparison = 0;
            if (sortConfig.key === 'name') {
                comparison = a.name.localeCompare(b.name);
            } else if (sortConfig.key === 'size') {
                comparison = (a.size || 0) - (b.size || 0);
            }
            return sortConfig.direction === 'asc' ? comparison : -comparison;
        });

    const toggleSort = (key) => {
        setSortConfig(prev => ({
            key,
            direction: prev.key === key && prev.direction === 'asc' ? 'desc' : 'asc'
        }));
    };

    return (
        <div className="w-72 flex flex-col border-r border-mui-grey-200 dark:border-mui-grey-800 bg-white dark:bg-mui-dark-bg shrink-0">
            <div className="p-3 border-b border-mui-grey-100 dark:border-mui-grey-800 space-y-3 bg-mui-grey-50 dark:bg-mui-grey-900">
                <div className="flex items-center gap-1 bg-white dark:bg-mui-dark-bg rounded px-1 py-0.5 border border-mui-grey-200 dark:border-mui-grey-800">
                    <button onClick={navigateUp} className="p-1 text-mui-grey-500 dark:text-mui-grey-400 hover:text-mui-blue-600 dark:hover:text-white transition-colors shrink-0" title="Back">
                        <ChevronLeft size={16} />
                    </button>
                    <input
                        type="text"
                        className="flex-1 min-w-0 px-1 py-0.5 bg-transparent text-[10px] text-mui-grey-700 dark:text-mui-grey-300 font-mono outline-none focus:text-mui-blue-600 dark:focus:text-mui-blue-400 transition-colors"
                        value={editingPath}
                        onChange={(e) => setEditingPath(e.target.value)}
                        onKeyDown={(e) => {
                            if (e.key === 'Enter') {
                                const path = editingPath.trim();
                                if (path && path !== remotePath) {
                                    loadRemoteFiles(path, true).then(() => {
                                        AppBackend.SendSSHInput(session.id, `cd "${path}"\r`);
                                    });
                                }
                            } else if (e.key === 'Escape') {
                                setEditingPath(remotePath);
                            }
                        }}
                    />
                    <button onClick={() => syncExplorer()} className="p-1 text-mui-grey-500 dark:text-mui-grey-400 hover:text-mui-blue-500 transition-colors shrink-0" title="Sync with terminal">
                        <RefreshCw size={12} />
                    </button>
                </div>

                <div className="relative">
                    <Search className="absolute left-2 top-1/2 -translate-y-1/2 text-mui-grey-400 dark:text-mui-grey-500" size={12} />
                    <input
                        type="text"
                        placeholder="Search..."
                        className="w-full bg-mui-grey-50 dark:bg-mui-grey-800 border border-mui-grey-200 dark:border-mui-grey-700 rounded py-1 pl-7 pr-2 text-[11px] text-mui-grey-700 dark:text-mui-grey-300 outline-none focus:border-mui-blue-500 transition-all"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                    />
                </div>

                <div className="flex gap-1.5">
                    <button onClick={handleUpload} className="flex-1 flex items-center justify-center gap-1.5 py-1 bg-mui-blue-600 hover:bg-mui-blue-700 rounded text-[10px] font-bold text-white transition-colors shadow-sm">
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
                        className="flex-1 flex items-center justify-center gap-1.5 py-1 bg-mui-grey-100 dark:bg-mui-grey-800 hover:bg-mui-grey-200 dark:hover:bg-mui-grey-700 rounded text-[10px] font-bold text-mui-grey-700 dark:text-mui-grey-200 transition-colors"
                    >
                        <Folder size={12} /> New
                    </button>
                </div>
            </div>

            <div className="flex items-center px-3 py-2 border-b border-mui-grey-100 dark:border-white/5 bg-mui-grey-50 dark:bg-mui-grey-900 select-none">
                <button
                    onClick={() => toggleSort('name')}
                    className="flex-1 flex items-center gap-1 text-[10px] font-black uppercase tracking-tighter text-mui-grey-400 hover:text-mui-grey-900 dark:hover:text-white transition-colors"
                >
                    Name
                    {sortConfig.key === 'name' && (
                        <span className="text-mui-blue-500">{sortConfig.direction === 'asc' ? '↑' : '↓'}</span>
                    )}
                </button>
                <button
                    onClick={() => toggleSort('size')}
                    className="w-16 flex items-center justify-end gap-1 text-[10px] font-black uppercase tracking-tighter text-mui-grey-400 hover:text-mui-grey-900 dark:hover:text-white transition-colors"
                >
                    Size
                    {sortConfig.key === 'size' && (
                        <span className="text-mui-blue-500">{sortConfig.direction === 'asc' ? '↑' : '↓'}</span>
                    )}
                </button>
            </div>

            <div className="flex-1 overflow-y-auto px-1 py-1 custom-scrollbar bg-white dark:bg-mui-dark-bg">
                {loadingFiles ? (
                    <div className="flex flex-col items-center justify-center h-40 space-y-2">
                        <RefreshCw className="animate-spin text-mui-grey-400 dark:text-mui-grey-600" size={20} />
                    </div>
                ) : (
                    <div className="space-y-px">
                        {(remotePath && remotePath !== '/') && (
                            <div
                                className="group flex items-center gap-2 px-2 py-1.5 rounded hover:bg-mui-grey-100 dark:hover:bg-mui-grey-800 cursor-pointer border border-transparent transition-all select-none"
                                onDoubleClick={navigateUp}
                            >
                                <Folder size={14} className="text-mui-grey-400 dark:text-mui-grey-500 shrink-0" />
                                <span className="flex-1 text-[11px] font-bold text-mui-grey-500 dark:text-mui-grey-400 truncate">...</span>
                                <span className="w-16 text-[10px] text-right text-mui-grey-400 opacity-0 group-hover:opacity-100">UP</span>
                            </div>
                        )}

                        {sortedFiles.map((file) => (
                            <div
                                key={file.name}
                                onContextMenu={(e) => handleFileContextMenu(e, file)}
                                className="group flex items-center gap-2 px-2 py-1 rounded hover:bg-mui-grey-100 dark:hover:bg-mui-grey-800 cursor-pointer border border-transparent hover:border-mui-grey-200 dark:hover:border-mui-grey-700 transition-all select-none"
                                onDoubleClick={() => handleFileDoubleClick(file)}
                            >
                                {file.isDir ? <Folder size={14} className="text-mui-blue-500 dark:text-mui-blue-400 shrink-0" /> : <File size={14} className="text-mui-grey-400 dark:text-mui-grey-500 shrink-0" />}
                                <span className="flex-1 text-[11px] text-mui-grey-700 dark:text-mui-grey-400 group-hover:text-mui-grey-900 dark:group-hover:text-white truncate">{file.name}</span>

                                {!file.isDir && (
                                    <span className="w-16 text-[10px] text-right text-mui-grey-400 group-hover:text-mui-grey-500 transition-colors">
                                        {formatSize(file.size)}
                                    </span>
                                )}

                                <div className="flex items-center gap-1">
                                    <button
                                        onClick={(e) => { e.stopPropagation(); handleFileContextMenu(e, file); }}
                                        className="p-1 text-mui-grey-400 hover:text-mui-blue-600 opacity-0 group-hover:opacity-100 transition-opacity"
                                    >
                                        <MoreVertical size={12} />
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                )}
            </div>

            {fileContextMenu && (
                <div
                    className="fixed z-50 bg-white dark:bg-mui-grey-800 shadow-xl border border-mui-grey-200 dark:border-white/10 rounded-lg py-1 min-w-[140px] animate-in fade-in zoom-in-95 duration-100"
                    style={{ top: fileContextMenu.y, left: fileContextMenu.x }}
                    onClick={(e) => e.stopPropagation()}
                >
                    <button
                        onClick={async () => {
                            const name = prompt('Rename:', fileContextMenu.file.name);
                            if (name && name !== fileContextMenu.file.name) {
                                try {
                                    await AppBackend.ExecuteSFTPAction(session.id, 'rename', `${remotePath}/${fileContextMenu.file.name}`, `${remotePath}/${name}`);
                                    loadRemoteFiles(remotePath);
                                } catch (err) { addToast('Error', err.toString(), 'error'); }
                            }
                            setFileContextMenu(null);
                        }}
                        className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
                    >
                        <Edit2 size={14} />
                        Rename
                    </button>

                    {!fileContextMenu.file.isDir && (
                        <>
                            <button
                                onClick={() => { handleEdit(fileContextMenu.file); setFileContextMenu(null); }}
                                className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
                            >
                                <Edit3 size={14} />
                                Edit File
                            </button>
                            <button
                                onClick={() => { handleDownload(fileContextMenu.file); setFileContextMenu(null); }}
                                className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
                            >
                                <Download size={14} />
                                Download
                            </button>
                        </>
                    )}

                    <div className="h-px bg-mui-grey-100 dark:bg-white/5 my-1" />

                    <button
                        onClick={() => {
                            handleDelete(fileContextMenu.file);
                            setFileContextMenu(null);
                        }}
                        className="w-full px-4 py-2 text-left text-[11px] font-bold text-red-500 hover:bg-red-50 dark:hover:bg-red-500/10 flex items-center gap-2"
                    >
                        <Trash2 size={14} />
                        Delete
                    </button>
                </div>
            )}
        </div>
    );
};

export default SSHFileExplorer;
