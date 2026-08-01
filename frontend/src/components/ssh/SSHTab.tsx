import React, { useState, useEffect } from "react";
import * as AppBackend from "../../../wailsjs/go/backend/App";
import {
  Plus,
  Trash2,
  Edit2,
  X,
  Server,
  ChevronRight,
  RefreshCw,
} from "lucide-react";
import SSHSessionView from "../SSHSessionView";
import SSHSessionForm from "./SSHSessionForm";
import { clsx } from "clsx";
import ConfirmationModal from "../ConfirmationModal";
import { handleActionKey } from "../../utils/a11y";

interface SSHTabProps {
  /** Callback to trigger user-visible toast notifications */
  addToast: (
    title: string,
    message: string,
    type?: "info" | "success" | "warn" | "error",
  ) => void;
  /** Active UI color theme ('light' or 'dark') */
  theme?: string;
  /** Action handler to trigger global modal settings window category */
  onOpenSettings: (category: string) => void;
}

/**
 * SSHTab Component
 *
 * Manages the main SSH dashboard interface, including:
 * 1. Creating, editing, and deleting saved SSH and WSL sessions.
 * 2. Connecting to and disconnecting from SSH/WSL terminals.
 * 3. A multi-tab interface where active connection sessions are toggled via CSS hidden styles
 *    instead of unmounting, thereby preserving terminal history, metrics charts, and processes.
 */
const SSHTab: React.FC<SSHTabProps> = ({ addToast, theme, onOpenSettings }) => {
  // --- Persistent & Local States ---
  const [sessions, setSessions] = useState<any[]>([]);
  const [activeSessionIds, setActiveSessionIds] = useState<string[]>([]);
  const [currentSessionId, setCurrentSessionId] = useState<string | null>(null);
  const [showForm, setShowForm] = useState(false);
  const [editingSession, setEditingSession] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [contextMenu, setContextMenu] = useState<any>(null);
  const [hostSearchQuery, setHostSearchQuery] = useState("");
  const [searchWidth, setSearchWidth] = useState<number>(180);
  const [isResizing, setIsResizing] = useState<boolean>(false);
  const [deleteSessionId, setDeleteSessionId] = useState<string | null>(null);

  const contextMenuRef = React.useRef<HTMLDivElement>(null);

  // Handle horizontal resizing of the Search input
  useEffect(() => {
    if (!isResizing) return;

    const handleMouseMove = (e: MouseEvent) => {
      // Find search input wrapper or use relative change
      setSearchWidth((prev) => Math.max(100, Math.min(600, prev + e.movementX)));
    };

    const handleMouseUp = () => {
      setIsResizing(false);
    };

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseup", handleMouseUp);
    return () => {
      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("mouseup", handleMouseUp);
    };
  }, [isResizing]);

  // Load saved configurations upon mounting
  useEffect(() => {
    loadSessions();
  }, []);

  // Dismiss context menu when clicking anywhere else
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (contextMenuRef.current?.contains(e.target as Node)) {
        return;
      }
      setContextMenu(null);
    };
    window.addEventListener("click", handleClick);
    return () => window.removeEventListener("click", handleClick);
  }, []);

  /**
   * Fetches saved connection profiles from persistent backend storage.
   */
  const loadSessions = async () => {
    setLoading(true);
    try {
      const data = await AppBackend.GetSSHSessions();
      setSessions(data || []);
    } catch (err) {
      console.error("Failed to load SSH sessions:", err);
      addToast("Error", "Failed to load SSH sessions", "error");
    } finally {
      setLoading(false);
    }
  };

  /**
   * Registers a target connection session into the multi-tab layout and focuses on it.
   */
  const handleConnect = (session: any) => {
    if (!activeSessionIds.includes(session.id)) {
      setActiveSessionIds([...activeSessionIds, session.id]);
    }
    setCurrentSessionId(session.id);
  };

  /**
   * Requests backend disconnection and removes the target tab from the layout.
   */
  const handleCloseSession = (id: string) => {
    AppBackend.DisconnectSSH(id);
    const nextActive = activeSessionIds.filter((sid) => sid !== id);
    setActiveSessionIds(nextActive);

    // Autofocus another active tab or default to main Dashboard
    if (currentSessionId === id) {
      setCurrentSessionId(
        nextActive.length > 0 ? nextActive[nextActive.length - 1] : null,
      );
    }
  };

  /**
   * Confirms and deletes a saved connection profile from persistent storage.
   */
  const handleDelete = (id: string) => {
    setDeleteSessionId(id);
  };

  const handleConfirmDelete = async () => {
    if (!deleteSessionId) return;
    const id = deleteSessionId;
    setDeleteSessionId(null);
    try {
      await AppBackend.DeleteSSHSession(id);
      loadSessions();
      addToast("Success", "Session deleted successfully", "success");
    } catch (err) {
      addToast("Error", "Failed to delete session", "error");
    }
  };

  const handleCancelDelete = () => {
    setDeleteSessionId(null);
  };

  /**
   * Spawns a custom contextual action menu on right-click.
   */
  const handleContextMenu = (e: React.MouseEvent, session: any) => {
    e.preventDefault();
    setContextMenu({
      x: e.clientX,
      y: e.clientY,
      sessionId: session.id,
    });
  };

  /**
   * Adds a new empty/virtual tab.
   */
  const handleAddNewTab = () => {
    const newTabId = `new-tab-${Date.now()}`;
    setActiveSessionIds([...activeSessionIds, newTabId]);
    setCurrentSessionId(newTabId);
  };

  /**
   * Binds a host session to a virtual tab.
   */
  const handleSelectHostForTab = (virtualTabId: string, session: any) => {
    if (activeSessionIds.includes(session.id)) {
      addToast("Info", `${session.name || session.host} is already open in another tab.`, "info");
      setCurrentSessionId(session.id);
      setActiveSessionIds(activeSessionIds.filter((id) => id !== virtualTabId));
      return;
    }

    const nextActive = activeSessionIds.map((id) => (id === virtualTabId ? session.id : id));
    setActiveSessionIds(nextActive);
    setCurrentSessionId(session.id);
  };

  /**
   * Renders host selection cards inside a virtual tab.
   */
  const renderHostSelectionScreen = (virtualTabId: string) => {
    if (sessions.length === 0) {
      return (
        <div className="flex-1 flex flex-col items-center justify-center border-2 border-dashed border-mui-grey-200 dark:border-white/5 rounded-xl p-12 text-center h-full">
          <div className="bg-mui-grey-100 dark:bg-white/5 p-4 rounded-full mb-4">
            <Server size={32} className="text-mui-grey-400" />
          </div>
          <h3 className="text-lg font-semibold text-mui-grey-900 dark:text-white">
            No sessions found
          </h3>
          <p className="text-mui-grey-600 dark:text-mui-grey-400 mt-1 max-w-sm">
            Add your first SSH connection to start managing remote servers and files.
          </p>
          <button
            type="button"
            onClick={() => {
              setEditingSession(null);
              setShowForm(true);
            }}
            className="mt-6 text-mui-blue-500 hover:text-mui-blue-600 font-medium flex items-center gap-1 bg-transparent border-none"
          >
            Create session now <ChevronRight size={18} />
          </button>
        </div>
      );
    }

    const filteredSessions = sessions.filter((session) => {
      const nameMatch = (session.name || "").toLowerCase().includes(hostSearchQuery.toLowerCase());
      const hostMatch = (session.host || "").toLowerCase().includes(hostSearchQuery.toLowerCase());
      return nameMatch || hostMatch;
    });

    return (
      <div className="flex flex-col h-full overflow-y-auto pr-2 pb-4 custom-scrollbar">
        <div className="mb-6">
          <h3 className="text-lg font-bold text-mui-grey-900 dark:text-white">
            Connect to a Host
          </h3>
          <p className="text-xs text-mui-grey-500 dark:text-mui-grey-400 mt-1">
            Select a saved host below to establish a connection in this tab.
          </p>
        </div>

        {filteredSessions.length === 0 ? (
          <div className="flex-1 flex flex-col items-center justify-center p-8 text-center">
            <p className="text-sm text-mui-grey-500 dark:text-mui-grey-400">
              No hosts match your search query.
            </p>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-3">
            {filteredSessions.map((session) => (
              <div
                key={session.id}
                className="group bg-mui-grey-50 dark:bg-mui-dark-paper border border-mui-grey-200 dark:border-white/10 rounded-lg p-3 hover:border-mui-blue-500/50 hover:shadow-md transition-all relative overflow-hidden flex items-center gap-3 select-none"
              >
                <button
                  type="button"
                  onClick={() => handleSelectHostForTab(virtualTabId, session)}
                  className="flex-1 flex items-center gap-3 outline-none text-left min-w-0 bg-transparent border-none p-0 cursor-pointer"
                >
                  <div className="bg-mui-blue-600 text-white p-2 rounded-md shrink-0">
                    <Server size={18} />
                  </div>

                  <div className="flex-1 min-w-0">
                    <h4 className="font-bold text-mui-grey-900 dark:text-white truncate text-sm leading-tight">
                      {session.name || session.host}
                    </h4>
                    <div className="flex items-center gap-2 mt-0.5">
                      <span className="text-[9px] font-black text-mui-blue-500 uppercase tracking-tighter">
                        SSH
                      </span>
                      <div className="w-1 h-1 bg-mui-grey-300 dark:bg-mui-grey-600 rounded-full" />
                      <span className="text-[9px] font-medium text-mui-grey-400 uppercase">
                        {session.authMethod}
                      </span>
                    </div>
                  </div>
                </button>
              </div>
            ))}
          </div>
        )}
      </div>
    );
  };

  /**
   * Renders the loading animation, empty dashboard state, or list of connection cards.
   */
  const renderDashboardContent = () => {
    if (loading) {
      return (
        <div className="flex-1 flex items-center justify-center">
          <RefreshCw className="animate-spin text-mui-grey-400" size={32} />
        </div>
      );
    }

    if (sessions.length === 0) {
      return (
        <div className="flex-1 flex flex-col items-center justify-center border-2 border-dashed border-mui-grey-200 dark:border-white/5 rounded-xl p-12 text-center">
          <div className="bg-mui-grey-100 dark:bg-white/5 p-4 rounded-full mb-4">
            <Server size={32} className="text-mui-grey-400" />
          </div>
          <h3 className="text-lg font-semibold text-mui-grey-900 dark:text-white">
            No sessions found
          </h3>
          <p className="text-mui-grey-600 dark:text-mui-grey-400 mt-1 max-w-sm">
            Add your first SSH connection to start managing remote servers and
            files.
          </p>
          <button
            type="button"
            onClick={() => setShowForm(true)}
            className="mt-6 text-mui-blue-500 hover:text-mui-blue-600 font-medium flex items-center gap-1"
          >
            Create session now <ChevronRight size={18} />
          </button>
        </div>
      );
    }

    const filteredSessions = sessions.filter((session) => {
      const nameMatch = (session.name || "").toLowerCase().includes(hostSearchQuery.toLowerCase());
      const hostMatch = (session.host || "").toLowerCase().includes(hostSearchQuery.toLowerCase());
      return nameMatch || hostMatch;
    });

    if (filteredSessions.length === 0 && hostSearchQuery) {
      return (
        <div className="flex-1 flex flex-col items-center justify-center p-8 text-center">
          <p className="text-sm text-mui-grey-500 dark:text-mui-grey-400">
            No hosts match your search query.
          </p>
        </div>
      );
    }

    return (
      <div className="flex-1 overflow-y-auto pr-2 pb-4 custom-scrollbar">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-3">
          {filteredSessions.map((session) => (
            <div
              key={session.id}
              className="group bg-mui-grey-50 dark:bg-mui-dark-paper border border-mui-grey-200 dark:border-white/10 rounded-lg p-3 hover:border-mui-blue-500/50 hover:shadow-md transition-all relative overflow-hidden flex items-center gap-3 select-none"
            >
              <button
                type="button"
                onDoubleClick={() => handleConnect(session)}
                onKeyDown={handleActionKey(() => handleConnect(session))}
                onContextMenu={(e) => handleContextMenu(e, session)}
                className="flex-1 flex items-center gap-3 outline-none text-left min-w-0 bg-transparent border-none p-0 cursor-pointer"
              >
                <div className="bg-mui-blue-600 text-white p-2 rounded-md shrink-0">
                  <Server size={18} />
                </div>

                <div className="flex-1 min-w-0">
                  <h4 className="font-bold text-mui-grey-900 dark:text-white truncate text-sm leading-tight">
                    {session.name || session.host}
                  </h4>
                  <div className="flex items-center gap-2 mt-0.5">
                    <span className="text-[9px] font-black text-mui-blue-500 uppercase tracking-tighter">
                      SSH
                    </span>
                    <div className="w-1 h-1 bg-mui-grey-300 dark:bg-mui-grey-600 rounded-full" />
                    <span className="text-[9px] font-medium text-mui-grey-400 uppercase">
                      {session.authMethod}
                    </span>
                  </div>
                </div>
              </button>

              <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 focus-within:opacity-100 transition-opacity">
                <button
                  type="button"
                  onClick={(e) => {
                    e.stopPropagation();
                    setEditingSession(session);
                    setShowForm(true);
                  }}
                  className="p-1.5 hover:bg-mui-grey-200 dark:hover:bg-white/10 rounded-md text-mui-grey-500 border-none bg-transparent cursor-pointer"
                  title="Edit"
                >
                  <Edit2 size={12} />
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  };

  return (
    <div className="relative flex h-full overflow-hidden bg-white dark:bg-mui-dark-bg transition-colors duration-300">
      <div className="flex-1 flex flex-col min-w-0 h-full overflow-hidden">
        {/* --- Top Tabbed Bar --- */}
        <div className="flex items-center gap-[2px] overflow-x-auto no-scrollbar shrink-0 pt-2 px-6 bg-mui-grey-50 dark:bg-mui-grey-900 border-b border-mui-grey-200 dark:border-white/5">
          {/* Main Dashboard Tab-Button */}
          <button
              type="button"
              onClick={() => setCurrentSessionId(null)}
              onKeyDown={handleActionKey(() => setCurrentSessionId(null))}
              className={clsx(
                "relative px-6 py-2 text-xs font-bold transition-all flex items-center gap-2 whitespace-nowrap cursor-pointer rounded-t-xl group min-w-[120px] max-w-[200px] outline-none border-none p-0",
                currentSessionId === null
                  ? "bg-white dark:bg-mui-dark-bg text-mui-blue-600 z-10 border-t border-x border-mui-grey-200 dark:border-white/5"
                  : "text-mui-grey-500 hover:bg-mui-grey-200 dark:hover:bg-white/10 focus:bg-mui-grey-100 dark:focus:bg-white/5",
              )}
            >
              <span className="truncate">Dashboard</span>
              {currentSessionId === null && (
                <div className="absolute -bottom-[1px] left-0 right-0 h-[1px] bg-white dark:bg-mui-dark-bg z-20" />
              )}
            </button>

            {/* Render dynamically created connection tabs */}
            {activeSessionIds.map((id) => {
              const session = sessions.find((s) => s.id === id);
              const isVirtualTab = id.startsWith("new-tab-");
              const isActive = currentSessionId === id;
              const displayName = isVirtualTab ? "New Tab" : (session?.name || session?.host || id);
              return (
                <div
                  key={id}
                  className={clsx(
                    "relative py-2 text-xs transition-all rounded-t-xl flex items-center min-w-[140px] max-w-[220px] overflow-hidden pr-2 group",
                    isActive
                      ? "bg-white dark:bg-mui-dark-bg z-10 border-t border-x border-mui-grey-200 dark:border-white/80"
                      : "text-mui-grey-500 hover:bg-mui-grey-200 dark:hover:bg-white/10",
                  )}
                >
                  <button
                    type="button"
                    onClick={() => setCurrentSessionId(id)}
                    onKeyDown={handleActionKey(() => setCurrentSessionId(id))}
                    className="flex-1 min-w-0 pl-4 pr-1 h-full text-left outline-none bg-transparent border-none p-0 cursor-pointer"
                  >
                    <span
                      className={clsx(
                        "truncate font-bold block",
                        isActive ? "text-mui-blue-600" : "text-mui-grey-400",
                      )}
                    >
                      {displayName}
                    </span>
                  </button>
                  <button
                    type="button"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleCloseSession(id);
                    }}
                    className="p-0.5 rounded-full transition-all ml-2 opacity-50 group-hover:opacity-100 hover:bg-mui-grey-200 dark:hover:bg-white/10 flex items-center justify-center cursor-pointer select-none border-none bg-transparent"
                    title="Close Connection"
                  >
                    <X
                      size={12}
                      className={
                        isActive ? "text-mui-blue-600" : "text-mui-grey-500"
                      }
                    />
                  </button>

                  {isActive && (
                    <div className="absolute -bottom-[1px] left-0 right-0 h-[1px] bg-white dark:bg-mui-dark-bg z-20" />
                  )}
                </div>
              );
            })}

            {/* "+" tab-button to add an empty/virtual new tab */}
            <button
              type="button"
              onClick={handleAddNewTab}
              className="p-1 rounded-full text-mui-grey-500 hover:text-mui-blue-600 dark:hover:text-white hover:bg-mui-grey-200 dark:hover:bg-white/10 transition-all ml-2 flex items-center justify-center cursor-pointer shrink-0 border-none bg-transparent"
              title="New Connection"
            >
              <Plus size={14} />
            </button>
          </div>

          {/* --- "New Host" & Search bookmark-like bar directly below the tabbed bar (only visible on Dashboard or New Tab selection screens) --- */}
          {(currentSessionId === null || currentSessionId.startsWith("new-tab-")) && (
            <div className="flex items-center justify-start gap-4 px-6 py-2 bg-mui-grey-100 dark:bg-mui-grey-950 border-b border-mui-grey-200 dark:border-white/5 shrink-0 select-none">
              {/* New Host button */}
              <button
                type="button"
                onClick={() => {
                  setEditingSession(null);
                  setShowForm(true);
                }}
                className="flex items-center gap-1.5 px-3 py-1 bg-mui-blue-600 hover:bg-mui-blue-700 text-white text-xs font-bold rounded shadow transition-all duration-200 border-none cursor-pointer shrink-0"
                title="Add New Host"
              >
                <Plus size={14} />
                <span>New Host</span>
              </button>

              {/* Resizable Search input */}
              <div
                className="flex items-center gap-1 bg-white dark:bg-mui-grey-900 px-1 py-0.5 rounded border border-mui-grey-200 dark:border-white/10 shrink-0"
                style={{ width: searchWidth + 24 }}
              >
                <div className="relative flex-1 flex items-center min-w-0" style={{ width: searchWidth }}>
                  <input
                    type="text"
                    placeholder="Search host..."
                    value={hostSearchQuery}
                    onChange={(e) => setHostSearchQuery(e.target.value)}
                    className="w-full px-3 py-1 text-xs rounded border border-transparent bg-white dark:bg-mui-grey-900 text-mui-grey-900 dark:text-white focus:outline-none"
                  />
                  {hostSearchQuery && (
                    <button
                      type="button"
                      onClick={() => setHostSearchQuery("")}
                      className="absolute right-2.5 top-1/2 -translate-y-1/2 text-mui-grey-400 hover:text-mui-grey-600 dark:hover:text-white border-none bg-transparent cursor-pointer flex items-center justify-center"
                    >
                      <X size={12} />
                    </button>
                  )}
                </div>

                {/* Resizer handle */}
                <div
                  onMouseDown={(e) => {
                    e.preventDefault();
                    setIsResizing(true);
                  }}
                  className="w-1.5 h-6 bg-mui-grey-300 dark:bg-white/10 hover:bg-mui-blue-500 active:bg-mui-blue-600 cursor-col-resize rounded shrink-0 transition-colors ml-1"
                  title="Resize Search horizontally"
                />
              </div>
            </div>
          )}

        {/* --- Main Content Panel --- */}
        <div
          className={clsx(
            "flex-1 min-h-0 relative",
            currentSessionId === null && "p-6",
          )}
        >
          {/* Dashboard container - kept alive but hidden if not selected to retain state */}
          <div className={clsx("h-full flex flex-col", currentSessionId !== null && "hidden")}>
            {renderDashboardContent()}
          </div>

          {/* Active SSH sessions container - kept alive but hidden if Dashboard is selected.
              Toggling via CSS 'hidden' preserves terminal interactions, history buffer,
              and active resource history charts. */}
          <div className={clsx("h-full", currentSessionId === null && "hidden")}>
            {activeSessionIds.map((id) => {
              const isVirtualTab = id.startsWith("new-tab-");
              return (
                <div
                  key={id}
                  className={clsx("h-full", currentSessionId === id ? "block" : "hidden")}
                >
                  {isVirtualTab ? (
                    <div className="p-6 h-full flex flex-col">
                      {renderHostSelectionScreen(id)}
                    </div>
                  ) : (
                    <SSHSessionView
                      session={sessions.find((s) => s.id === id)}
                      onClose={() => handleCloseSession(id)}
                      addToast={addToast}
                      isActive={currentSessionId === id}
                      theme={theme}
                      onOpenSettings={onOpenSettings}
                    />
                  )}
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* --- Overlay Form Modal --- */}
      {showForm && (
        <SSHSessionForm
          session={editingSession}
          onClose={() => setShowForm(false)}
          onSave={() => {
            setShowForm(false);
            loadSessions();
          }}
          addToast={addToast}
        />
      )}

      {/* --- Left Panel Context Action Menu --- */}
      {contextMenu && (
        <div
          ref={contextMenuRef}
          className="fixed z-50 bg-white dark:bg-mui-grey-800 shadow-xl border border-mui-grey-200 dark:border-white/10 rounded-lg py-1 min-w-[140px] animate-in fade-in zoom-in-95 duration-100 cursor-default p-0"
          style={{ top: contextMenu.y, left: contextMenu.x }}
        >
          <button
            type="button"
            onClick={() => {
              handleDelete(contextMenu.sessionId);
              setContextMenu(null);
            }}
            className="w-full px-4 py-2 text-left text-[11px] font-bold text-red-500 hover:bg-red-50 dark:hover:bg-red-500/10 flex items-center gap-2 transition-all"
          >
            <Trash2 size={14} />
            Delete Session
          </button>
        </div>
      )}

      <ConfirmationModal
        isOpen={deleteSessionId !== null}
        title="Delete Session Profile"
        message="Are you sure you want to delete this session?"
        type="danger"
        confirmText="Delete"
        cancelText="Cancel"
        onConfirm={handleConfirmDelete}
        onCancel={handleCancelDelete}
      />
    </div>
  );
};

export default SSHTab;
