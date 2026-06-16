import React from 'react';
import { Server, Edit2 } from 'lucide-react';

const SSHSessionCard = ({ session, handleConnect, handleContextMenu, setEditingSession, setShowForm }) => {
    return (
        <div
            onDoubleClick={() => handleConnect(session)}
            onContextMenu={(e) => handleContextMenu(e, session)}
            className="group bg-mui-grey-50 dark:bg-mui-dark-paper border border-mui-grey-200 dark:border-mui-grey-800 rounded-lg p-3 hover:border-mui-blue-500/50 hover:shadow-md transition-all relative overflow-hidden flex items-center gap-3 cursor-pointer select-none"
        >
            <div className="bg-mui-blue-600 text-white p-2 rounded-md shrink-0">
                <Server size={18} />
            </div>

            <div className="flex-1 min-w-0">
                <h4 className="font-bold text-mui-grey-900 dark:text-white truncate text-sm leading-tight">{session.host}</h4>
                <div className="flex items-center gap-2 mt-0.5">
                    <span className="text-[9px] font-black text-mui-blue-500 uppercase tracking-tighter">SSH</span>
                    <div className="w-1 h-1 bg-mui-grey-300 dark:bg-mui-grey-600 rounded-full" />
                    <span className="text-[9px] font-medium text-mui-grey-400 uppercase">{session.authMethod}</span>
                </div>
            </div>

            <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                <button
                    onClick={(e) => {
                        e.stopPropagation();
                        setEditingSession(session);
                        setShowForm(true);
                    }}
                    className="p-1.5 hover:bg-mui-grey-200 dark:hover:bg-white/10 rounded-md text-mui-grey-500"
                    title="Edit"
                >
                    <Edit2 size={12} />
                </button>
            </div>
        </div>
    );
};

export default SSHSessionCard;
