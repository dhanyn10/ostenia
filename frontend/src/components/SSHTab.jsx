import React, { useState, useEffect } from 'react';
import PropTypes from 'prop-types';
import * as AppBackend from '../../wailsjs/go/backend/App';
import { Plus, Trash2, Server, ChevronRight, RefreshCw } from 'lucide-react';
import SSHSessionView from './SSHSessionView';
import SSHSessionForm from './SSHSessionForm';
import SSHTabHeader from './ssh/SSHTabHeader';
import SSHSessionCard from './ssh/SSHSessionCard';
import { clsx } from 'clsx';

const SSHTab = ({ addToast, theme }) => {
    const [sessions, setSessions] = useState([]);
    const [activeSessionIds, setActiveSessionIds] = useState([]);
    const [currentSessionId, setCurrentSessionId] = useState(null);
    const [showForm, setShowForm] = useState(false);
    const [editingSession, setEditingSession] = useState(null);
    const [loading, setLoading] = useState(true);
    const [contextMenu, setContextMenu] = useState(null);

    useEffect(() => {
        loadSessions();
        const handleClick = () => setContextMenu(null);
        window.addEventListener('click', handleClick);
        return () => window.removeEventListener('click', handleClick);
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
        AppBackend.DisconnectSSH(id);
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

    const handleContextMenu = (e, session) => {
        e.preventDefault();
        setContextMenu({
            x: e.clientX,
            y: e.clientY,
            sessionId: session.id
        });
    };

    return (
        <div className="flex h-full overflow-hidden bg-white dark:bg-mui-dark-bg transition-colors duration-300">
            <div className="flex-1 flex flex-col min-w-0 h-full overflow-hidden">
                <div className="px-6 pt-2 pb-4 flex justify-between items-center shrink-0 border-b border-mui-grey-100 dark:border-white/5">
                    <div>
                        <h2 className="text-2xl font-bold text-mui-grey-900 dark:text-white">SSH Connections</h2>
                        <p className="text-mui-grey-600 dark:text-mui-grey-400 text-sm">Manage and connect to your remote servers.</p>
                    </div>
                    {!showForm && (
                        <button
                            onClick={() => { setEditingSession(null); setShowForm(true); }}
                            className="flex items-center gap-2 px-4 py-2 bg-mui-blue-600 hover:bg-mui-blue-700 text-white rounded-lg transition-colors font-medium shadow-sm"
                        >
                            <Plus size={18} /> New Connection
                        </button>
                    )}
                </div>

                {activeSessionIds.length > 0 && (
                    <SSHTabHeader
                        activeSessionIds={activeSessionIds}
                        sessions={sessions}
                        currentSessionId={currentSessionId}
                        setCurrentSessionId={setCurrentSessionId}
                        handleCloseSession={handleCloseSession}
                    />
                )}

                <div className={clsx("flex-1 min-h-0 relative", currentSessionId === null && "p-6")}>
                    {currentSessionId ? (
                        <div className="h-full">
                            {activeSessionIds.map(id => (
                                <div key={id} className={currentSessionId === id ? "h-full" : "hidden"}>
                                    <SSHSessionView
                                        session={sessions.find(s => s.id === id)}
                                        onClose={() => handleCloseSession(id)}
                                        addToast={addToast}
                                        isActive={currentSessionId === id}
                                        theme={theme}
                                    />
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="flex flex-col h-full">
                            {loading ? (
                                <div className="flex-1 flex items-center justify-center">
                                    <RefreshCw className="animate-spin text-mui-grey-400" size={32} />
                                </div>
                            ) : sessions.length === 0 ? (
                                <div className="flex-1 flex flex-col items-center justify-center border-2 border-dashed border-mui-grey-200 dark:border-white/5 rounded-xl p-12 text-center">
                                    <div className="bg-mui-grey-100 dark:bg-white/5 p-4 rounded-full mb-4">
                                        <Server size={32} className="text-mui-grey-400" />
                                    </div>
                                    <h3 className="text-lg font-semibold text-mui-grey-900 dark:text-white">No sessions found</h3>
                                    <p className="text-mui-grey-600 dark:text-mui-grey-400 mt-1 max-w-sm">Add your first SSH connection to start managing remote servers and files.</p>
                                    <button onClick={() => setShowForm(true)} className="mt-6 text-mui-blue-500 hover:text-mui-blue-600 font-medium flex items-center gap-1">
                                        Create session now <ChevronRight size={18} />
                                    </button>
                                </div>
                            ) : (
                                <div className="flex-1 overflow-y-auto pr-2 pb-4 custom-scrollbar">
                                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-5 gap-3">
                                        {sessions.map((session) => (
                                            <SSHSessionCard
                                                key={session.id}
                                                session={session}
                                                handleConnect={handleConnect}
                                                handleContextMenu={handleContextMenu}
                                                setEditingSession={setEditingSession}
                                                setShowForm={setShowForm}
                                            />
                                        ))}
                                    </div>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>

            {showForm && (
                <SSHSessionForm
                    session={editingSession}
                    onClose={() => setShowForm(false)}
                    onSave={() => { setShowForm(false); loadSessions(); }}
                    addToast={addToast}
                />
            )}

            {contextMenu && (
                <div
                    className="fixed z-50 bg-white dark:bg-mui-grey-800 shadow-xl border border-mui-grey-200 dark:border-white/10 rounded-lg py-1 min-w-[140px] animate-in fade-in zoom-in-95 duration-100"
                    style={{ top: contextMenu.y, left: contextMenu.x }}
                    onClick={(e) => e.stopPropagation()}
                >
                    <button
                        onClick={() => { handleDelete(contextMenu.sessionId); setContextMenu(null); }}
                        className="w-full px-4 py-2 text-left text-[11px] font-bold text-red-500 hover:bg-red-50 dark:hover:bg-red-500/10 flex items-center gap-2 transition-all"
                    >
                        <Trash2 size={14} /> Delete Session
                    </button>
                </div>
            )}
        </div>
    );
};

SSHTab.propTypes = {
    addToast: PropTypes.func.isRequired,
    theme: PropTypes.string.isRequired,
};

export default SSHTab;
