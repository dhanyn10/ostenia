import React from 'react';
import { Edit2, Edit3, Download, Trash2 } from 'lucide-react';

const FileContextMenu = ({ x, y, file, remotePath, session, loadRemoteFiles, handleEdit, handleDownload, handleDelete, setFileContextMenu, addToast, AppBackend }) => {
  return (
    <div
      className="fixed z-50 bg-white dark:bg-mui-grey-800 shadow-xl border border-mui-grey-200 dark:border-white/10 rounded-lg py-1 min-w-[140px] animate-in fade-in zoom-in-95 duration-100"
      style={{ top: y, left: x }}
      onClick={(e) => e.stopPropagation()}
    >
      <button
        onClick={async () => {
          const name = prompt('Rename:', file.name);
          if (name && name !== file.name) {
            try {
              await AppBackend.ExecuteSFTPAction(session.id, 'rename', `${remotePath}/${file.name}`, `${remotePath}/${name}`);
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

      {!file.isDir && (
        <>
          <button
            onClick={() => { handleEdit(file); setFileContextMenu(null); }}
            className="w-full px-4 py-2 text-left text-[11px] font-bold text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-grey-100 dark:hover:bg-white/5 flex items-center gap-2"
          >
            <Edit3 size={14} />
            Edit File
          </button>
          <button
            onClick={() => { handleDownload(file); setFileContextMenu(null); }}
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
          handleDelete(file);
          setFileContextMenu(null);
        }}
        className="w-full px-4 py-2 text-left text-[11px] font-bold text-red-500 hover:bg-red-50 dark:hover:bg-red-500/10 flex items-center gap-2"
      >
        <Trash2 size={14} />
        Delete
      </button>
    </div>
  );
};

export default FileContextMenu;
