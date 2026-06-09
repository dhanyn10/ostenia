import React, { useState, useEffect } from 'react';
import * as AppBackend from '../../wailsjs/go/main/App';
import { Plus, Terminal, Trash2, Edit2, Play, AlertCircle, X, Server, Key, Lock, ChevronRight, Folder, File, Download, Upload, RefreshCw, MoreVertical } from 'lucide-react';
import SSHSessionView from './SSHSessionView';
import SSHSessionForm from './SSHSessionForm';
import { clsx } from 'clsx';

const SSHTab = ({ addToast }) => {
  const [sessions, setSessions] = useState([]);
  const [activeSessionIds, setActiveSessionIds] = useState([]);
  const [currentSessionId, setCurrentSessionId] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [editingSession, setEditingSession] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadSessions();
  }, []);

  const loadSessions = async () => {
    setLoading(true);
    try {
      const data = await AppBackend.GetSSHSessions();
      setSessions(data || []);
    } catch (err) {
      console.error('Failed to load SSH sessions:', err);
      addToast('Error', 'Failed to load SSH sessions', 'error');
    } finally {
      setLoading(false);
    }
  };

  const handleConnect = (session) => {
    if (!activeSessionIds.includes(session.id)) {
      setActiveSessionIds([...activeSessionIds, session.id]);
    }
    setCurrentSessionId(session.id);
  };

  const handleCloseSession = (id) => {
      const nextActive = activeSessionIds.filter(sid => sid !== id);
      setActiveSessionIds(nextActive);
      if (currentSessionId === id) {
          setCurrentSessionId(nextActive.length > 0 ? nextActive[nextActive.length - 1] : null);
      }
  };

  const handleDelete = async (id) => {
    if (confirm('Are you sure you want to delete this session?')) {
      try {
        await AppBackend.DeleteSSHSession(id);
        loadSessions();
        addToast('Success', 'Session deleted successfully');
      } catch (err) {
        addToast('Error', 'Failed to delete session', 'error');
      }
    }
  };

  return (
    <div className="flex flex-col h-full space-y-4">
      {/* Tab Header for Active Sessions */}
      {activeSessionIds.length > 0 && (
          <div className="flex items-center gap-1 overflow-x-auto pb-1 no-scrollbar border-b border-slate-200 dark:border-white/5">
              <button
                onClick={() => setCurrentSessionId(null)}
                className={clsx(
                    "px-4 py-2 text-sm font-medium rounded-t-lg transition-all flex items-center gap-2 whitespace-nowrap",
                    currentSessionId === null ? "bg-white dark:bg-[#1e293b] text-blue-500 border-x border-t border-slate-200 dark:border-white/5" : "text-slate-500 hover:text-slate-700 dark:hover:text-slate-300"
                )}
              >
                  <Server size={14} />
                  Dashboard
              </button>
              {activeSessionIds.map(id => {
                  const session = sessions.find(s => s.id === id);
                  if (!session) return null;
                  return (
                      <div
                        key={id}
                        className={clsx(
                            "flex items-center rounded-t-lg transition-all border-t border-x overflow-hidden",
                            currentSessionId === id ? "bg-white dark:bg-[#0f172a] text-blue-500 border-slate-200 dark:border-white/5" : "text-slate-500 hover:text-slate-700 dark:hover:text-slate-300 border-transparent"
                        )}
                      >
                          <button
                            onClick={() => setCurrentSessionId(id)}
                            className="px-4 py-2 text-sm font-medium flex items-center gap-2 whitespace-nowrap"
                          >
                              <Terminal size={14} />
                              {session.name}
                          </button>
                          <button
                            onClick={() => handleCloseSession(id)}
                            className="pr-2 pl-1 hover:text-red-500 transition-colors"
                          >
                              <X size={14} />
                          </button>
                      </div>
                  );
              })}
          </div>
      )}

      {currentSessionId ? (
          <div className="flex-1 min-h-0">
              {activeSessionIds.map(id => (
                  <div key={id} className={currentSessionId === id ? "h-full" : "hidden"}>
                      <SSHSessionView
                        session={sessions.find(s => s.id === id)}
                        onClose={() => handleCloseSession(id)}
                        addToast={addToast}
                      />
                  </div>
              ))}
          </div>
      ) : (
          <div className="flex flex-col h-full space-y-6">
            <div className="flex justify-between items-center">
                <div>
                <h2 className="text-2xl font-bold text-slate-900 dark:text-white flex items-center gap-2">
                    <Terminal className="text-blue-500" />
                    SSH Connections
                </h2>
                <p className="text-slate-500 dark:text-slate-400">Manage and connect to your remote servers.</p>
                </div>
                <button
                onClick={() => {
                    setEditingSession(null);
                    setShowForm(true);
                }}
                className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors font-medium shadow-sm"
                >
                <Plus size={18} />
                New Connection
                </button>
            </div>

            {loading ? (
                <div className="flex-1 flex items-center justify-center">
                <RefreshCw className="animate-spin text-slate-400" size={32} />
                </div>
            ) : sessions.length === 0 ? (
                <div className="flex-1 flex flex-col items-center justify-center border-2 border-dashed border-slate-200 dark:border-white/5 rounded-xl p-12 text-center">
                <div className="bg-slate-100 dark:bg-white/5 p-4 rounded-full mb-4">
                    <Server size={32} className="text-slate-400" />
                </div>
                <h3 className="text-lg font-semibold text-slate-900 dark:text-white">No sessions found</h3>
                <p className="text-slate-500 dark:text-slate-400 mt-1 max-w-sm">
                    Add your first SSH connection to start managing remote servers and files.
                </p>
                <button
                    onClick={() => setShowForm(true)}
                    className="mt-6 text-blue-500 hover:text-blue-600 font-medium flex items-center gap-1"
                >
                    Create session now <ChevronRight size={18} />
                </button>
                </div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 overflow-y-auto pr-2 pb-4">
                {sessions.map((session) => (
                    <div
                    key={session.id}
                    className="group bg-white dark:bg-[#1e293b] border border-slate-200 dark:border-white/5 rounded-xl p-5 hover:border-blue-500/50 hover:shadow-md transition-all relative overflow-hidden"
                    >
                    <div className="absolute top-0 right-0 p-2 opacity-0 group-hover:opacity-100 transition-opacity flex gap-1">
                        <button
                        onClick={() => {
                            setEditingSession(session);
                            setShowForm(true);
                        }}
                        className="p-1.5 hover:bg-slate-100 dark:hover:bg-white/10 rounded-md text-slate-500"
                        title="Edit"
                        >
                        <Edit2 size={14} />
                        </button>
                        <button
                        onClick={() => handleDelete(session.id)}
                        className="p-1.5 hover:bg-red-50 dark:hover:bg-red-500/10 rounded-md text-slate-500 hover:text-red-500"
                        title="Delete"
                        >
                        <Trash2 size={14} />
                        </button>
                    </div>

                    <div className="flex items-start gap-4">
                        <div className="bg-blue-50 dark:bg-blue-500/10 p-3 rounded-lg">
                        <Server className="text-blue-500" size={24} />
                        </div>
                        <div className="flex-1 min-w-0">
                        <h4 className="font-bold text-slate-900 dark:text-white truncate">{session.name}</h4>
                        <p className="text-xs text-slate-500 dark:text-slate-400 truncate mt-0.5">
                            {session.user}@{session.host}:{session.port}
                        </p>
                        </div>
                    </div>

                    <div className="mt-6 flex items-center justify-between">
                        <div className="flex items-center gap-2 text-xs font-medium text-slate-400 uppercase tracking-wider">
                        {session.authMethod === 'password' ? <Lock size={12} /> : <Key size={12} />}
                        {session.authMethod}
                        </div>
                        <button
                        onClick={() => handleConnect(session)}
                        className="px-4 py-1.5 bg-slate-900 dark:bg-blue-600 hover:bg-black dark:hover:bg-blue-700 text-white rounded-md text-sm font-medium transition-colors flex items-center gap-1.5"
                        >
                        <Play size={14} />
                        Connect
                        </button>
                    </div>
                    </div>
                ))}
                </div>
            )}
          </div>
      )}

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
    </div>
  );
};

export default SSHTab;
