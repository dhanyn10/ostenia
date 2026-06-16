import React from 'react';
import PropTypes from 'prop-types';
import { X } from 'lucide-react';
import { clsx } from 'clsx';

const SSHTabHeader = ({ activeSessionIds, sessions, currentSessionId, setCurrentSessionId, handleCloseSession }) => {
    return (
        <div className="flex items-center gap-[2px] overflow-x-auto no-scrollbar shrink-0 pt-2 px-6 bg-mui-grey-50 dark:bg-mui-grey-900 border-b border-mui-grey-200 dark:border-white/5">
            <div
                role="button"
                tabIndex="0"
                onClick={() => setCurrentSessionId(null)}
                onKeyDown={(e) => {
                    if (e.key === 'Enter') setCurrentSessionId(null);
                }}
                className={clsx(
                    "relative px-6 py-2 text-xs font-bold transition-all flex items-center gap-2 whitespace-nowrap cursor-pointer rounded-t-xl group min-w-[120px] max-w-[200px] outline-none",
                    currentSessionId === null
                        ? "bg-white dark:bg-mui-dark-bg text-mui-blue-600 z-10 border-t border-x border-mui-grey-200 dark:border-white/5"
                        : "text-mui-grey-500 hover:bg-mui-grey-200 dark:hover:bg-white/10 focus:bg-mui-grey-200 dark:focus:bg-white/10"
                )}
            >
                <span className="truncate">Dashboard</span>
                {currentSessionId === null && (
                    <div className="absolute -bottom-[1px] left-0 right-0 h-[1px] bg-white dark:bg-mui-dark-bg z-20" />
                )}
            </div>
            {activeSessionIds.map(id => {
                const session = sessions.find(s => s.id === id);
                if (!session) return null;
                const isActive = currentSessionId === id;
                const displayName = session.name || session.host;
                return (
                    <div
                        key={id}
                        role="button"
                        tabIndex="0"
                        onClick={() => setCurrentSessionId(id)}
                        onKeyDown={(e) => {
                            if (e.key === 'Enter') setCurrentSessionId(id);
                        }}
                        className={clsx(
                            "relative pl-6 pr-2 py-2 text-xs transition-all group cursor-pointer rounded-t-xl flex items-center justify-between min-w-[140px] max-w-[220px] outline-none",
                            isActive
                                ? "bg-white dark:bg-mui-dark-bg text-mui-blue-600 z-10 border-t border-x border-mui-grey-200 dark:border-white/5"
                                : "text-mui-grey-500 hover:bg-mui-grey-200 dark:hover:bg-white/10 focus:bg-mui-grey-200 dark:focus:bg-white/10"
                        )}
                    >
                        <span className={clsx(
                            "truncate font-bold",
                            isActive ? "text-mui-blue-600" : "text-mui-grey-400"
                        )}>
                            {displayName}
                        </span>
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                handleCloseSession(id);
                            }}
                            className={clsx(
                                "p-1 rounded-md transition-all ml-2",
                                isActive ? "hover:bg-mui-blue-500/10" : "hover:bg-mui-grey-500/10",
                                "opacity-0 group-hover:opacity-100"
                            )}
                            title="Close session"
                        >
                            <X size={12} className={isActive ? "text-mui-blue-600" : "text-mui-grey-500"} />
                        </button>
                        {isActive && (
                            <div className="absolute -bottom-[1px] left-0 right-0 h-[1px] bg-white dark:bg-mui-dark-bg z-20" />
                        )}
                    </div>
                );
            })}
        </div>
    );
};

SSHTabHeader.propTypes = {
    activeSessionIds: PropTypes.arrayOf(PropTypes.string).isRequired,
    sessions: PropTypes.arrayOf(PropTypes.shape({
        id: PropTypes.string.isRequired,
        host: PropTypes.string,
        name: PropTypes.string,
    })).isRequired,
    currentSessionId: PropTypes.string,
    setCurrentSessionId: PropTypes.func.isRequired,
    handleCloseSession: PropTypes.func.isRequired,
};

export default SSHTabHeader;
