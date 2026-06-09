import React from 'react';
import { Home, List, Sun, Moon, Globe, Server } from 'lucide-react';
import Icons from './Icons';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

function VerticalNav({ activeTab, setActiveTab, toggleTheme, theme }) {
  return (
    <aside className="w-16 flex flex-col items-center py-6 gap-5 bg-white dark:bg-[#1e293b] border-r border-slate-200 dark:border-white/5 z-20 shrink-0 shadow-sm">
      {/* Activity Center */}
      <button 
        onClick={() => setActiveTab('activity')}
        title="Activity Center" 
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
      
      {/* Proxy Management */}
      <button
        onClick={() => setActiveTab('proxy')}
        title="Proxy Management"
        className={cn(
          "p-3 rounded-sm transition-all relative group",
          activeTab === 'proxy'
            ? "bg-blue-600 text-white shadow-lg shadow-blue-900/30"
            : "text-slate-400 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white"
        )}
      >
        {activeTab === 'proxy' && <div className="absolute left-[-16px] top-3 bottom-3 w-1 bg-blue-500 rounded-r-sm" />}
        <Globe size={20} />
      </button>

      {/* SSH Management */}
      <button
        onClick={() => setActiveTab('ssh')}
        title="SSH & Remote Files"
        className={cn(
          "p-3 rounded-sm transition-all relative group",
          activeTab === 'ssh'
            ? "bg-blue-600 text-white shadow-lg shadow-blue-900/30"
            : "text-slate-400 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white"
        )}
      >
        {activeTab === 'ssh' && <div className="absolute left-[-16px] top-3 bottom-3 w-1 bg-blue-500 rounded-r-sm" />}
        <Server size={20} />
      </button>

      {/* Plugin Management */}
      <button 
        onClick={() => setActiveTab('plugins')}
        title="Plugin Management" 
        className={cn(
          "p-3 rounded-sm transition-all relative group",
          activeTab === 'plugins' 
            ? "bg-blue-600 text-white shadow-lg shadow-blue-900/30" 
            : "text-slate-400 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white"
        )}
      >
        {activeTab === 'plugins' && <div className="absolute left-[-16px] top-3 bottom-3 w-1 bg-blue-500 rounded-r-sm" />}
        <Icons.Plugins size={20} />
      </button>

      {/* System Logs Tab */}
      <button 
        onClick={() => setActiveTab('logs')} 
        title="System Activity Logs"
        className={cn(
          "p-3 rounded-sm transition-all relative group",
          activeTab === 'logs' 
            ? "bg-blue-600 text-white shadow-lg shadow-blue-900/30" 
            : "text-slate-400 hover:bg-slate-100 dark:hover:bg-white/5 hover:text-slate-900 dark:hover:text-white"
        )}
      >
        {activeTab === 'logs' && <div className="absolute left-[-16px] top-3 bottom-3 w-1 bg-blue-500 rounded-r-sm" />}
        <List size={20} />
      </button>

      {/* Theme Toggle */}
      <button 
        onClick={toggleTheme}
        title={theme === 'dark' ? 'Switch to Light Mode' : 'Switch to Dark Mode'}
        className="p-3 text-slate-400 hover:text-slate-900 dark:hover:text-white transition-all rounded-sm"
      >
        {theme === 'dark' ? <Sun size={20} /> : <Moon size={20} />}
      </button>
    </aside>
  );
}

export default VerticalNav;
