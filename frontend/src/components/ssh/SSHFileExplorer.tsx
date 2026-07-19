import React from "react";
import {
  ChevronLeft,
  RefreshCw,
  Search,
  Upload,
  Folder,
  File,
  MoreVertical,
} from "lucide-react";
import { handleActionKey } from "../../utils/a11y";

interface SSHFileExplorerProps {
  remotePath: string;
  editingPath: string;
  setEditingPath: (path: string) => void;
  onNavigateUp: () => void;
  onSync: () => void;
  searchQuery: string;
  setSearchQuery: (query: string) => void;
  onUpload: () => void;
  onNewFolder: () => void;
  loadingFiles: boolean;
  sortedFiles: any[];
  onFileDoubleClick: (file: any) => void;
  onFileContextMenu: (e: React.MouseEvent, file: any) => void;
  onExplorerContextMenu: (e: React.MouseEvent) => void;
  formatSize: (bytes: number) => string;
  toggleSort: (key: string) => void;
  sortConfig: { key: string; direction: "asc" | "desc" };
  onManualNavigation: (path: string) => void;
}

const SSHFileExplorer: React.FC<SSHFileExplorerProps> = ({
  remotePath,
  editingPath,
  setEditingPath,
  onNavigateUp,
  onSync,
  searchQuery,
  setSearchQuery,
  onUpload,
  onNewFolder,
  loadingFiles,
  sortedFiles,
  onFileDoubleClick,
  onFileContextMenu,
  onExplorerContextMenu,
  formatSize,
  toggleSort,
  sortConfig,
  onManualNavigation,
}) => {
  return (
    <div
      onContextMenu={onExplorerContextMenu}
      className="w-72 flex flex-col border-r border-mui-grey-200 dark:border-mui-grey-800 bg-white dark:bg-mui-dark-bg shrink-0"
    >
      <div className="p-3 border-b border-mui-grey-100 dark:border-mui-grey-800 space-y-3 bg-mui-grey-50 dark:bg-mui-grey-900">
        <div className="flex items-center gap-1 bg-white dark:bg-mui-dark-bg rounded px-1 py-0.5 border border-mui-grey-200 dark:border-mui-grey-800">
          <button
            type="button"
            onClick={onNavigateUp}
            className="p-1 text-mui-grey-500 dark:text-mui-grey-400 hover:text-mui-blue-600 dark:hover:text-white transition-colors shrink-0"
            title="Back"
          >
            <ChevronLeft size={16} />
          </button>
          <input
            type="text"
            className="flex-1 min-w-0 px-1 py-0.5 bg-transparent text-[10px] text-mui-grey-700 dark:text-mui-grey-300 font-mono outline-none focus:text-mui-blue-600 dark:focus:text-mui-blue-400 transition-colors"
            value={editingPath}
            onChange={(e) => setEditingPath(e.target.value)}
            onContextMenu={(e) => e.stopPropagation()}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                onManualNavigation(editingPath);
              } else if (e.key === "Escape") {
                setEditingPath(remotePath);
              }
            }}
          />
          <button
            type="button"
            onClick={onSync}
            className="p-1 text-mui-grey-500 dark:text-mui-grey-400 hover:text-mui-blue-500 transition-colors shrink-0"
            title="Sync with terminal"
          >
            <RefreshCw size={12} />
          </button>
        </div>

        <div className="relative">
          <Search
            className="absolute left-2 top-1/2 -translate-y-1/2 text-mui-grey-400 dark:text-mui-grey-500"
            size={12}
          />
          <input
            type="text"
            placeholder="Search..."
            className="w-full bg-mui-grey-50 dark:bg-mui-grey-800 border border-mui-grey-200 dark:border-mui-grey-700 rounded py-1 pl-7 pr-2 text-[11px] text-mui-grey-700 dark:text-mui-grey-300 outline-none focus:border-mui-blue-500 transition-all"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onContextMenu={(e) => e.stopPropagation()}
          />
        </div>

        <div className="flex gap-1.5">
          <button
            type="button"
            onClick={onUpload}
            className="flex-1 flex items-center justify-center gap-1.5 py-1 bg-mui-blue-600 hover:bg-mui-blue-700 rounded text-[10px] font-bold text-white transition-colors shadow-sm"
          >
            <Upload size={12} /> Upload
          </button>
          <button
            type="button"
            onClick={onNewFolder}
            className="flex-1 flex items-center justify-center gap-1.5 py-1 bg-mui-grey-100 dark:bg-mui-grey-800 hover:bg-mui-grey-200 dark:hover:bg-mui-grey-700 rounded text-[10px] font-bold text-mui-grey-700 dark:text-mui-grey-200 transition-colors"
          >
            <Folder size={12} /> New
          </button>
        </div>
      </div>

      {/* Column Headers */}
      <div className="flex items-center px-3 py-2 border-b border-mui-grey-100 dark:border-white/5 bg-mui-grey-50 dark:bg-mui-grey-900 select-none">
        <button
          type="button"
          onClick={() => toggleSort("name")}
          className="flex-1 flex items-center gap-1 text-[10px] font-black uppercase tracking-tighter text-mui-grey-400 hover:text-mui-grey-900 dark:hover:text-white transition-colors"
        >
          Name
          {sortConfig.key === "name" && (
            <span className="text-mui-blue-500">
              {sortConfig.direction === "asc" ? "↑" : "↓"}
            </span>
          )}
        </button>
        <button
          type="button"
          onClick={() => toggleSort("size")}
          className="w-16 flex items-center justify-end gap-1 text-[10px] font-black uppercase tracking-tighter text-mui-grey-400 hover:text-mui-grey-900 dark:hover:text-white transition-colors"
        >
          Size
          {sortConfig.key === "size" && (
            <span className="text-mui-blue-500">
              {sortConfig.direction === "asc" ? "↑" : "↓"}
            </span>
          )}
        </button>
      </div>

      <div className="flex-1 overflow-y-auto px-1 py-1 custom-scrollbar bg-white dark:bg-mui-dark-bg">
        {loadingFiles ? (
          <div className="flex flex-col items-center justify-center h-40 space-y-2">
            <RefreshCw
              className="animate-spin text-mui-grey-400 dark:text-mui-grey-600"
              size={20}
            />
          </div>
        ) : (
          <div className="space-y-px">
            {/* Go Up Directory */}
            {remotePath && remotePath !== "/" && (
              <button
                type="button"
                className="w-full group flex items-center gap-2 px-2 py-1.5 rounded hover:bg-mui-grey-100 dark:hover:bg-mui-grey-800 cursor-pointer border border-transparent transition-all select-none outline-none focus:bg-mui-grey-100 dark:focus:bg-mui-grey-800 bg-transparent p-0 text-left"
                onDoubleClick={onNavigateUp}
                onKeyDown={handleActionKey(onNavigateUp)}
              >
                <Folder
                  size={14}
                  className="text-mui-grey-400 dark:text-mui-grey-500 shrink-0"
                />
                <span className="flex-1 text-[11px] font-bold text-mui-grey-500 dark:text-mui-grey-400 truncate">
                  ...
                </span>
                <span className="w-16 text-[10px] text-right text-mui-grey-400 opacity-0 group-hover:opacity-100">
                  UP
                </span>
              </button>
            )}

            {sortedFiles.map((file) => (
              <div
                key={file.name}
                className="w-full group flex items-center gap-2 px-2 py-1 rounded hover:bg-mui-grey-100 dark:hover:bg-mui-grey-800 border border-transparent hover:border-mui-grey-200 dark:hover:border-mui-grey-700 transition-all select-none relative"
              >
                <button
                  type="button"
                  className="flex-1 flex items-center gap-2 outline-none text-left min-w-0 bg-transparent border-none p-0"
                  onDoubleClick={() => onFileDoubleClick(file)}
                  onKeyDown={handleActionKey(() => onFileDoubleClick(file))}
                  onContextMenu={(e) => onFileContextMenu(e, file)}
                >
                  {file.isDir ? (
                    <Folder
                      size={14}
                      className="text-mui-blue-500 dark:text-mui-blue-400 shrink-0"
                    />
                  ) : (
                    <File
                      size={14}
                      className="text-mui-grey-400 dark:text-mui-grey-500 shrink-0"
                    />
                  )}
                  <span className="flex-1 text-[11px] text-mui-grey-700 dark:text-mui-grey-400 group-hover:text-mui-grey-900 dark:group-hover:text-white truncate">
                    {file.name}
                  </span>

                  {!file.isDir && (
                    <span className="w-16 text-[10px] text-right text-mui-grey-400 group-hover:text-mui-grey-500 transition-colors">
                      {formatSize(file.size)}
                    </span>
                  )}
                </button>

                <div className="flex items-center gap-1">
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      onFileContextMenu(e, file);
                    }}
                    className="p-1 text-mui-grey-400 hover:text-mui-blue-600 opacity-0 group-hover:opacity-100 transition-opacity focus:opacity-100"
                  >
                    <MoreVertical size={12} />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default SSHFileExplorer;
