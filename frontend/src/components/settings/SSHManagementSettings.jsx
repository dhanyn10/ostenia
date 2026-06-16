import React from 'react';

const SSHManagementSettings = ({ sshSessions }) => {
    return (
        <div className="space-y-8 animate-in fade-in slide-in-from-right-4 duration-300 flex flex-col h-full">
            <div className="shrink-0">
                <h3 className="text-xl font-bold mb-1 tracking-tight text-mui-grey-900 dark:text-white">SSH Management</h3>
                <p className="text-sm text-mui-grey-500 dark:text-mui-grey-400">Current active session data in read-only JSON format.</p>
            </div>

            <div className="flex-1 min-h-0 border border-mui-grey-200 dark:border-white/10 rounded-lg overflow-hidden flex flex-col bg-mui-grey-50 dark:bg-white/5">
                <div className="px-4 py-3 border-b border-mui-grey-200 dark:border-white/10 flex justify-between items-center bg-white dark:bg-mui-dark-paper">
                    <span className="text-xs font-black uppercase tracking-widest text-mui-grey-400">ssh_sessions.json</span>
                    <div className="px-2 py-1 rounded bg-mui-grey-100 dark:bg-white/5 text-[10px] font-bold text-mui-grey-500 dark:text-mui-grey-400 uppercase tracking-tighter">Read Only</div>
                </div>

                <div className="flex-1 overflow-y-auto p-4 font-mono text-[12px] leading-relaxed">
                    <pre className="text-mui-grey-700 dark:text-mui-blue-200">
                        {JSON.stringify(sshSessions.map(({ password, passphrase, ...s }) => ({ ...s, password: "***", passphrase: "***" })), null, 2)}
                    </pre>
                </div>
            </div>
            <p className="text-[10px] text-mui-grey-400 italic">Sensitive fields like password and passphrase are masked for security. Manage sessions via the main SSH Tab.</p>
        </div>
    );
};

export default SSHManagementSettings;
