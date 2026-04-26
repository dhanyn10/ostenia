import React from 'react';
import { X, Settings2, Search } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

function ExtensionModal({ isOpen, onClose, extensions, onToggle, serviceName }) {
  const [searchTerm, setSearchString] = React.useState('');

  if (!isOpen) return null;

  const filteredExtensions = extensions.filter(ext => 
    ext.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="fixed inset-0 z-[200] flex items-center justify-center px-4">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-slate-950/40 backdrop-blur-sm animate-in fade-in duration-300" 
        onClick={onClose} 
      />
      
      {/* Modal Content */}
      <div className="bg-white dark:bg-slate-900 w-full max-w-2xl max-h-[80vh] rounded-sm shadow-3xl border border-slate-200 dark:border-white/10 flex flex-col relative animate-in zoom-in-95 fade-in duration-200 overflow-hidden">
        
        {/* Header */}
        <div className="p-4 border-b border-slate-200 dark:border-white/5 flex items-center justify-between shrink-0">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-sm bg-blue-500/10 text-blue-600 dark:text-blue-400 flex items-center justify-center">
              <Settings2 size={18} />
            </div>
            <div>
              <h2 className="text-sm font-black text-slate-900 dark:text-white uppercase tracking-widest">{serviceName} Extensions</h2>
              <p className="text-[10px] text-slate-500 dark:text-slate-400 font-bold uppercase italic opacity-70 mt-0.5 tracking-tight">Managed via php.ini</p>
            </div>
          </div>
          <button 
            onClick={onClose}
            className="p-2 hover:bg-slate-100 dark:hover:bg-white/5 text-slate-400 hover:text-slate-900 dark:hover:text-white rounded-sm transition-all"
          >
            <X size={18} />
          </button>
        </div>

        {/* Search Bar */}
        <div className="p-3 border-b border-slate-200 dark:border-white/5 shrink-0 bg-slate-50/50 dark:bg-black/10">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" size={14} />
            <input 
              type="text"
              placeholder="Search extensions..."
              value={searchTerm}
              onChange={(e) => setSearchString(e.target.value)}
              className="w-full bg-white dark:bg-slate-800 border border-slate-200 dark:border-white/5 rounded-sm pl-9 pr-4 py-2 text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-blue-500/50 transition-all"
            />
          </div>
        </div>

        {/* List Content */}
        <div className="flex-1 overflow-y-auto p-4 custom-scrollbar">
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
            {filteredExtensions.map(ext => (
              <label 
                key={ext.name} 
                className={cn(
                  "flex items-center gap-3 p-3 rounded-sm border transition-all cursor-pointer group/ext relative",
                  ext.enabled 
                    ? "bg-blue-500/5 border-blue-500/20 text-blue-600 dark:text-blue-400" 
                    : "bg-white dark:bg-white/[0.02] border-slate-200 dark:border-white/5 text-slate-500 dark:text-slate-400 hover:border-slate-300 dark:hover:border-white/10"
                )}
              >
                <input 
                  type="checkbox" 
                  checked={ext.enabled} 
                  onChange={() => onToggle(ext.name, !ext.enabled)}
                  className="form-checkbox h-4 w-4 rounded-sm border-slate-300 dark:border-white/10 text-blue-600 focus:ring-0 focus:ring-offset-0 transition-all"
                />
                <span className="text-[11px] font-bold truncate group-hover/ext:translate-x-0.5 transition-transform">{ext.name}</span>
                {ext.enabled && (
                  <div className="absolute top-1.5 right-1.5 w-1 h-1 rounded-full bg-blue-500 animate-pulse" />
                )}
              </label>
            ))}
            {filteredExtensions.length === 0 && (
              <div className="col-span-full py-12 text-center">
                <p className="text-xs font-bold text-slate-500 uppercase tracking-widest opacity-50">No extensions found matching your search</p>
              </div>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="p-4 border-t border-slate-200 dark:border-white/5 shrink-0 bg-slate-50/50 dark:bg-black/10 flex justify-end">
           <button 
             onClick={onClose}
             className="px-6 py-2 bg-slate-900 dark:bg-white text-white dark:text-slate-900 rounded-sm text-[10px] font-black uppercase tracking-[0.2em] hover:opacity-90 transition-all shadow-lg"
           >
             Close
           </button>
        </div>
      </div>
    </div>
  );
}

export default ExtensionModal;
