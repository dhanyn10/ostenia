import React, { useState, useEffect, useRef } from 'react';
import {
  Vibrate,
  Minus,
  Square,
  X,
  Settings,
  Eye,
  FileText,
  HelpCircle,
  ChevronRight,
  Monitor,
  Terminal as TerminalIcon,
  Copy,
  ExternalLink,
  ChevronDown
} from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';
import * as AppBackend from '../../wailsjs/go/main/App';

function cn(...inputs) {
  return twMerge(clsx(inputs));
}

const MenuItem = ({ label, children, isOpen, onOpen, onClose }) => {
  const containerRef = useRef(null);

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (containerRef.current && !containerRef.current.contains(event.target)) {
        onClose();
      }
    };
    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isOpen, onClose]);

  return (
    <div className="relative" ref={containerRef}>
      <button
        onClick={onOpen}
        onMouseEnter={() => isOpen && onOpen()}
        className={cn(
          "px-3 h-full flex items-center text-[13px] hover:bg-white/10 transition-colors",
          isOpen && "bg-white/10"
        )}
      >
        {label}
      </button>
      {isOpen && (
        <div className="absolute top-full left-0 w-64 bg-[#1e1e1e] border border-white/10 shadow-2xl z-[100] py-1 animate-in fade-in zoom-in-95 duration-100">
          {children}
        </div>
      )}
    </div>
  );
};

const SubMenuItem = ({ label, icon: Icon, onClick, shortcut, hasSubmenu }) => (
  <button
    onClick={onClick}
    className="w-full flex items-center justify-between px-3 py-1.5 text-[12px] text-mui-grey-300 hover:bg-mui-blue-500 hover:text-white transition-colors group"
  >
    <div className="flex items-center gap-2.5">
      <div className="w-4 flex justify-center">
        {Icon && <Icon size={14} />}
      </div>
      <span>{label}</span>
    </div>
    <div className="flex items-center gap-2">
      {shortcut && <span className="text-[10px] text-mui-grey-500 group-hover:text-white/70">{shortcut}</span>}
      {hasSubmenu && <ChevronRight size={12} className="text-mui-grey-500 group-hover:text-white/70" />}
    </div>
  </button>
);

const MenuDivider = () => <div className="my-1 border-t border-white/5" />;

const MenuBar = ({ theme, setTheme }) => {
  const [openMenu, setOpenMenu] = useState(null);
  const [isMaximized, setIsMaximized] = useState(false);

  const handleMinimize = () => AppBackend.Minimize();
  const handleMaximize = () => {
    if (isMaximized) {
      AppBackend.Unmaximize();
    } else {
      AppBackend.Maximize();
    }
    setIsMaximized(!isMaximized);
  };
  const handleClose = () => AppBackend.Close();

  const handleExport = async (type) => {
    try {
      await AppBackend.ExportProfile(type === 'all' || type === 'config', type === 'all' || type === 'ssh');
    } catch (err) {
      console.error(err);
    }
    setOpenMenu(null);
  };

  const handleImport = async () => {
    try {
      await AppBackend.ImportProfile();
    } catch (err) {
      console.error(err);
    }
    setOpenMenu(null);
  };

  const handleToggleDevTools = () => {
    AppBackend.ToggleDevTools();
    setOpenMenu(null);
  };

  const handleSetEditor = async () => {
    try {
      await AppBackend.SelectDefaultEditor();
    } catch (err) {
      console.error(err);
    }
    setOpenMenu(null);
  };

  return (
    <div className="h-9 flex items-center justify-between bg-[#1e1e1e] text-mui-grey-300 select-none border-b border-white/5" style={{ "--wails-draggable": "drag" } }>
      <div className="flex items-center h-full no-drag" style={{ "--wails-draggable": "no-drag" } }>
        <div className="px-3 flex items-center gap-2">
          <div className="w-5 h-5 bg-mui-blue-500 rounded-sm flex items-center justify-center">
            <Vibrate size={14} className="text-white" />
          </div>
        </div>

        <MenuItem
          label="File"
          isOpen={openMenu === 'file'}
          onOpen={() => setOpenMenu('file')}
          onClose={() => setOpenMenu(null)}
        >
          <SubMenuItem label="Import Profile..." icon={FileText} onClick={handleImport} />
          <MenuDivider />
          <SubMenuItem label="Export All" icon={Copy} onClick={() => handleExport('all')} />
          <SubMenuItem label="Export Config Only" icon={Settings} onClick={() => handleExport('config')} />
          <SubMenuItem label="Export SSH Sessions Only" icon={TerminalIcon} onClick={() => handleExport('ssh')} />
          <MenuDivider />
          <SubMenuItem label="Exit" onClick={handleClose} shortcut="Alt+F4" />
        </MenuItem>

        <MenuItem
          label="View"
          isOpen={openMenu === 'view'}
          onOpen={() => setOpenMenu('view')}
          onClose={() => setOpenMenu(null)}
        >
          <SubMenuItem label="Toggle Developer Tools" icon={Monitor} onClick={handleToggleDevTools} shortcut="F12" />
        </MenuItem>

        <MenuItem
          label="Settings"
          isOpen={openMenu === 'settings'}
          onOpen={() => setOpenMenu('settings')}
          onClose={() => setOpenMenu(null)}
        >
          <SubMenuItem
            label={theme === 'dark' ? "Switch to Light Mode" : "Switch to Dark Mode"}
            icon={Eye}
            onClick={() => { setTheme(theme === 'dark' ? 'light' : 'dark'); setOpenMenu(null); }}
          />
          <SubMenuItem label="Default Editor Path..." icon={Settings} onClick={handleSetEditor} />
        </MenuItem>

        <MenuItem
          label="Help"
          isOpen={openMenu === 'help'}
          onOpen={() => setOpenMenu('help')}
          onClose={() => setOpenMenu(null)}
        >
          <SubMenuItem label="About Ostenia" icon={HelpCircle} onClick={() => { alert("Ostenia v1.0.0\nPortable Development Environment"); setOpenMenu(null); }} />
          <SubMenuItem label="Documentation" icon={ExternalLink} onClick={() => { window.open('https://github.com/dhanyn/ostenia', '_blank'); setOpenMenu(null); }} />
        </MenuItem>
      </div>

      <div className="text-[11px] font-bold tracking-widest text-white/30 uppercase italic absolute left-1/2 -translate-x-1/2 pointer-events-none">
        Ostenia
      </div>

      <div className="flex h-full no-drag" style={{ "--wails-draggable": "no-drag" } }>
        <button
          onClick={handleMinimize}
          className="w-12 h-full flex items-center justify-center hover:bg-white/10 transition-colors"
        >
          <Minus size={14} />
        </button>
        <button
          onClick={handleMaximize}
          className="w-12 h-full flex items-center justify-center hover:bg-white/10 transition-colors"
        >
          <Square size={12} />
        </button>
        <button
          onClick={handleClose}
          className="w-12 h-full flex items-center justify-center hover:bg-rose-600 hover:text-white transition-colors"
        >
          <X size={16} />
        </button>
      </div>
    </div>
  );
};

export default MenuBar;
