import React from 'react';
import { Home, Download, List, Sun, Moon } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

function VerticalNav({ activeTab, setActiveTab, toggleTheme, theme, setIsLogOpen }) {
  return (
    <aside className="w-16 flex flex-col items-center py-6 gap-6 bg-white dark:bg-[#1e293b] border-r border-slate-200 dark:border-white/5 z-20 shrink-0 shadow-sm">
      <button 
        onClick={() => setActiveTab('activity')}
        title="Home" 
        className={cn(
          "p-3 rounded-sm transition-all relative group",
          activeTab === 'activity' 
            ? "bg-blue-600 text-white shadow-lg shadow-blue-900/30" 
            : "text-slate-400 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white"
        )}
      >
        {activeTab === 'activity' && <div className="absolute left-[-16px] top-3 bottom-3 w-1 bg-blue-500 rounded-r-sm" />}
        <Home size={20} />
      </button>
      
      <button 
        onClick={() => setActiveTab('plugins')}
        title="Plugins" 
        className={cn(
          "p-3 rounded-sm transition-all relative group",
          activeTab === 'plugins' 
            ? "bg-blue-600 text-white shadow-lg shadow-blue-900/30" 
            : "text-slate-400 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white"
        )}
      >
        {activeTab === 'plugins' && <div className="absolute left-[-16px] top-3 bottom-3 w-1 bg-blue-500 rounded-r-sm" />}
        <Download size={20} />
      </button>

      <div className="mt-auto flex flex-col gap-4">
        <button 
          onClick={toggleTheme}
          title={theme === 'dark' ? 'Switch to Light Mode' : 'Switch to Dark Mode'}
          className="p-3 text-slate-400 hover:text-slate-900 dark:hover:text-white transition-colors"
        >
          {theme === 'dark' ? <Sun size={20} /> : <Moon size={20} />}
        </button>
        <button onClick={() => setIsLogOpen(true)} className="p-3 text-slate-400 hover:text-slate-900 dark:hover:text-white transition-colors">
          <List size={18} />
        </button>
      </div>
    </aside>
  );
}

export default VerticalNav;
