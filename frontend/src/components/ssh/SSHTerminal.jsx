import React from 'react';
import PropTypes from 'prop-types';
import { RefreshCw } from 'lucide-react';

const SSHTerminal = React.forwardRef(({ connecting }, ref) => {
    return (
        <div className="flex-1 bg-white dark:bg-mui-dark-bg relative overflow-hidden">
            <div ref={ref} className="absolute inset-0 px-2 pt-2" />
            {connecting && (
                <div className="absolute inset-0 bg-white dark:bg-mui-dark-bg flex items-center justify-center">
                    <div className="flex items-center gap-3">
                        <RefreshCw className="animate-spin text-mui-blue-600 dark:text-mui-blue-500" size={18} />
                        <span className="text-mui-grey-600 dark:text-mui-grey-400 text-xs font-bold uppercase tracking-widest">Connecting...</span>
                    </div>
                </div>
            )}
        </div>
    );
});

SSHTerminal.displayName = 'SSHTerminal';

SSHTerminal.propTypes = {
    connecting: PropTypes.bool.isRequired,
};

export default SSHTerminal;
