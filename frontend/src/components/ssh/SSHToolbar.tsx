import React from "react";
import { Maximize2, RefreshCw, X, Folder } from "lucide-react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

interface SSHToolbarProps {
  explorerVisible: boolean;
  setExplorerVisible: (visible: boolean) => void;
  onReconnect: () => void;
  connecting: boolean;
}

const SSHToolbar: React.FC<SSHToolbarProps> = ({
  explorerVisible,
  setExplorerVisible,
  onReconnect,
  connecting,
}) => {
  return (
    <div className="h-10 flex items-center justify-between px-3 bg-white dark:bg-mui-dark-bg border-b border-mui-grey-100 dark:border-white/5 shrink-0">
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={() => setExplorerVisible(!explorerVisible)}
          className={cn(
            "p-1 rounded text-mui-grey-500 dark:text-mui-grey-400 hover:text-mui-blue-600 dark:hover:text-white transition-colors",
            explorerVisible && "text-mui-blue-600 bg-mui-blue-600/10",
          )}
          title="Toggle Explorer"
        >
          <Folder size={16} />
        </button>
      </div>

      <div className="flex items-center gap-1">
        <button
          type="button"
          onClick={onReconnect}
          disabled={connecting}
          className="p-1 text-mui-grey-500 hover:text-mui-blue-600 dark:hover:text-white transition-colors disabled:opacity-30"
          title="Reconnect"
        >
          <RefreshCw size={14} className={connecting ? "animate-spin" : ""} />
        </button>
      </div>
    </div>
  );
};

export default SSHToolbar;
