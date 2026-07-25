import React, { useState, useEffect, useRef } from "react";
import {
  Vibrate,
  Minus,
  Square,
  X,
  Settings,
  Eye,
  HelpCircle,
  ChevronRight,
  Monitor,
  Terminal as TerminalIcon,
  ExternalLink,
  User,
  Sliders,
  type LucideIcon,
} from "lucide-react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import * as AppBackend from "../../wailsjs/go/backend/App";
import { handleActionKey } from "../utils/a11y";

function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

interface MenuItemProps {
  label: string;
  children: React.ReactNode;
  isOpen: boolean;
  onOpen: () => void;
  onHover: () => void;
  onClose: () => void;
}

const MenuItem: React.FC<MenuItemProps> = ({
  label,
  children,
  isOpen,
  onOpen,
  onHover,
  onClose,
}) => {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        onClose();
      }
    };
    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isOpen, onClose]);

  return (
    <div className="relative h-full" ref={containerRef}>
      <button
        type="button"
        onClick={onOpen}
        onMouseEnter={onHover}
        className={cn(
          "px-3 h-full flex items-center text-[13px] transition-colors",
          isOpen
            ? "bg-mui-blue-500 text-white"
            : "hover:bg-mui-grey-200 dark:hover:bg-white/10 text-mui-grey-700 dark:text-mui-grey-300",
        )}
      >
        {label}
      </button>
      {isOpen && (
        <div
          className={cn(
            "absolute top-full left-0 w-64 border shadow-2xl z-[100] py-1 animate-in fade-in zoom-in-95 duration-100",
            "bg-white dark:bg-mui-dark-paper border-mui-grey-200 dark:border-white/10 rounded-b-sm",
          )}
        >
          {children}
        </div>
      )}
    </div>
  );
};

interface SubMenuItemProps {
  label: string;
  icon?: LucideIcon;
  onClick: () => void;
  shortcut?: string;
  hasSubmenu?: boolean;
}

const SubMenuItem: React.FC<SubMenuItemProps> = ({
  label,
  icon: Icon,
  onClick,
  shortcut,
  hasSubmenu,
}) => (
  <button
    type="button"
    onClick={onClick}
    onKeyDown={handleActionKey(onClick)}
    className={cn(
      "w-full flex items-center justify-between px-3 py-1.5 text-[12px] transition-colors group",
      "text-mui-grey-700 dark:text-mui-grey-300 hover:bg-mui-blue-500 hover:text-white",
    )}
  >
    <div className="flex items-center gap-2.5">
      <div className="w-4 flex justify-center">
        {Icon && <Icon size={14} />}
      </div>
      <span>{label}</span>
    </div>
    <div className="flex items-center gap-2">
      {shortcut && (
        <span className="text-[10px] text-mui-grey-500 group-hover:text-white/70">
          {shortcut}
        </span>
      )}
      {hasSubmenu && (
        <ChevronRight
          size={12}
          className="text-mui-grey-500 group-hover:text-white/70"
        />
      )}
    </div>
  </button>
);

const MenuDivider = () => (
  <div className="my-1 border-t border-mui-grey-200 dark:border-white/5" />
);

interface MenuBarProps {
  theme: string;
  setTheme: (theme: string) => void;
  onOpenSettings: (category: string) => void;
}

const MenuBar: React.FC<MenuBarProps> = ({
  theme,
  setTheme,
  onOpenSettings,
}) => {
  const [openMenu, setOpenMenu] = useState<string | null>(null);
  const [isMaximized, setIsMaximized] = useState(false);
  const [showSnapMenu, setShowSnapMenu] = useState(false);
  const hoverTimeoutRef = useRef<any>(null);

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

  const handleMouseEnterMaximize = () => {
    if (hoverTimeoutRef.current) clearTimeout(hoverTimeoutRef.current);
    hoverTimeoutRef.current = setTimeout(() => {
      setShowSnapMenu(true);
    }, 400);
  };

  const handleMouseLeaveMaximize = () => {
    if (hoverTimeoutRef.current) clearTimeout(hoverTimeoutRef.current);
    hoverTimeoutRef.current = setTimeout(() => {
      setShowSnapMenu(false);
    }, 300);
  };

  const handleMouseEnterSnapMenu = () => {
    if (hoverTimeoutRef.current) clearTimeout(hoverTimeoutRef.current);
  };

  const handleMouseLeaveSnapMenu = () => {
    if (hoverTimeoutRef.current) clearTimeout(hoverTimeoutRef.current);
    hoverTimeoutRef.current = setTimeout(() => {
      setShowSnapMenu(false);
    }, 200);
  };

  const snapWindow = async (layout: 'left' | 'right' | 'left-third' | 'right-third' | 'full') => {
    if (isMaximized) {
      AppBackend.Unmaximize();
      setIsMaximized(false);
    }
    try {
      const runtime = await import("../../wailsjs/runtime/runtime");
      const screens = await runtime.ScreenGetAll();
      const currentScreen = screens.find((s: any) => s.isCurrent) || screens[0] || { width: 1920, height: 1080 };
      const screenWidth = currentScreen.width;
      const screenHeight = currentScreen.height - 48; // Reserve pixels for system taskbar

      if (layout === 'left') {
        runtime.WindowSetPosition(0, 0);
        runtime.WindowSetSize(Math.floor(screenWidth / 2), screenHeight);
      } else if (layout === 'right') {
        runtime.WindowSetPosition(Math.floor(screenWidth / 2), 0);
        runtime.WindowSetSize(Math.floor(screenWidth / 2), screenHeight);
      } else if (layout === 'left-third') {
        runtime.WindowSetPosition(0, 0);
        runtime.WindowSetSize(Math.floor(screenWidth / 3), screenHeight);
      } else if (layout === 'right-third') {
        runtime.WindowSetPosition(Math.floor(screenWidth * 2 / 3), 0);
        runtime.WindowSetSize(Math.floor(screenWidth / 3), screenHeight);
      } else if (layout === 'full') {
        AppBackend.Maximize();
        setIsMaximized(true);
      }
    } catch (e) {
      console.error("Failed to snap window:", e);
    }
    setShowSnapMenu(false);
  };

  const handleOpenCategory = (category: string) => {
    onOpenSettings(category);
    setOpenMenu(null);
  };

  const menuItems = [
    {
      id: "file",
      label: "File",
      content: (
        <SubMenuItem
          label="Exit"
          onClick={handleClose}
          onKeyDown={handleActionKey(handleClose)}
        />
      ),
    },
    {
      id: "view",
      label: "View",
      content: (
        <SubMenuItem
          label="Toggle Developer Tools"
          icon={Monitor}
          onClick={() => {
            AppBackend.ToggleDevTools();
            setOpenMenu(null);
          }}
          onKeyDown={handleActionKey(() => {
            AppBackend.ToggleDevTools();
            setOpenMenu(null);
          })}
          shortcut="F12"
        />
      ),
    },
    {
      id: "settings",
      label: "Settings",
      content: (
        <>
          <SubMenuItem
            label="Profile"
            icon={User}
            onClick={() => handleOpenCategory("profile")}
            onKeyDown={handleActionKey(() => handleOpenCategory("profile"))}
          />
          <SubMenuItem
            label="Config"
            icon={Sliders}
            onClick={() => handleOpenCategory("config")}
            onKeyDown={handleActionKey(() => handleOpenCategory("config"))}
          />
          <SubMenuItem
            label="SSH"
            icon={TerminalIcon}
            onClick={() => handleOpenCategory("ssh")}
            onKeyDown={handleActionKey(() => handleOpenCategory("ssh"))}
          />
          <MenuDivider />
          <SubMenuItem
            label={
              theme === "dark" ? "Switch to Light Mode" : "Switch to Dark Mode"
            }
            onKeyDown={handleActionKey(() => {
              setTheme(theme === "dark" ? "light" : "dark");
              setOpenMenu(null);
            })}
            icon={Eye}
            onClick={() => {
              setTheme(theme === "dark" ? "light" : "dark");
              setOpenMenu(null);
            }}
          />
        </>
      ),
    },
    {
      id: "help",
      label: "Help",
      content: (
        <>
          <SubMenuItem
            label="About Ostenia"
            icon={HelpCircle}
            onClick={() => {
              alert("Ostenia v1.0.0\nPortable Development Environment");
              setOpenMenu(null);
            }}
            onKeyDown={handleActionKey(() => {
              alert("Ostenia v1.0.0\nPortable Development Environment");
              setOpenMenu(null);
            })}
          />
          <SubMenuItem
            label="Documentation"
            icon={ExternalLink}
            onClick={() => {
              window.open(
                "https://github.com/dhanyn/ostenia",
                "_blank",
                "noopener,noreferrer",
              );
              setOpenMenu(null);
            }}
            onKeyDown={handleActionKey(() => {
              window.open(
                "https://github.com/dhanyn/ostenia",
                "_blank",
                "noopener,noreferrer",
              );
              setOpenMenu(null);
            })}
          />
        </>
      ),
    },
  ];

  return (
    <div
      className={cn(
        "h-9 flex items-center justify-between select-none border-b transition-colors duration-300",
        "bg-white dark:bg-mui-dark-paper border-mui-grey-200 dark:border-white/5 text-mui-grey-700 dark:text-mui-grey-300",
      )}
      style={{ "--wails-draggable": "drag" } as React.CSSProperties}
    >
      <div
        className="flex items-center h-full no-drag"
        style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}
      >
        <div className="px-3 flex items-center gap-2">
          <div className="w-5 h-5 bg-mui-blue-500 rounded-sm flex items-center justify-center">
            <Vibrate size={14} className="text-white" />
          </div>
        </div>

        {menuItems.map((item) => (
          <MenuItem
            key={item.id}
            label={item.label}
            isOpen={openMenu === item.id}
            onOpen={() => setOpenMenu(openMenu === item.id ? null : item.id)}
            onHover={() => openMenu && setOpenMenu(item.id)}
            onClose={() => setOpenMenu(null)}
          >
            {item.content}
          </MenuItem>
        ))}
      </div>

      <div className="text-[11px] font-bold tracking-widest opacity-30 uppercase italic absolute left-1/2 -translate-x-1/2 pointer-events-none">
        Ostenia
      </div>

      <div
        className="flex h-full no-drag"
        style={{ "--wails-draggable": "no-drag" } as React.CSSProperties}
      >
        <button
          type="button"
          onClick={handleMinimize}
          className="w-12 h-full flex items-center justify-center hover:bg-mui-grey-200 dark:hover:bg-white/10 transition-colors"
        >
          <Minus size={14} />
        </button>

        <div
          className="relative h-full flex items-center"
          onMouseEnter={handleMouseEnterMaximize}
          onMouseLeave={handleMouseLeaveMaximize}
          data-testid="maximize-container"
        >
          <button
            type="button"
            onClick={handleMaximize}
            className="w-12 h-full flex items-center justify-center hover:bg-mui-grey-200 dark:hover:bg-white/10 transition-colors"
          >
            <Square size={12} />
          </button>

          {showSnapMenu && (
            <div
              onMouseEnter={handleMouseEnterSnapMenu}
              onMouseLeave={handleMouseLeaveSnapMenu}
              className={cn(
                "absolute top-full right-0 mt-1 w-48 p-3 border shadow-2xl z-[150] rounded bg-white dark:bg-mui-dark-paper border-mui-grey-200 dark:border-white/10 flex flex-col gap-3 animate-in fade-in zoom-in-95 duration-100",
              )}
            >
              <div className="text-[10px] font-semibold text-mui-grey-500 dark:text-mui-grey-400 uppercase tracking-wider mb-1">
                Snap Window
              </div>

              <div className="grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onClick={() => snapWindow('left')}
                  className="group flex flex-col items-center gap-1.5 p-2 rounded hover:bg-mui-grey-100 dark:hover:bg-white/5 border border-mui-grey-200 dark:border-white/5 transition-colors"
                >
                  <div className="w-12 h-8 border border-mui-grey-300 dark:border-white/10 rounded flex overflow-hidden">
                    <div className="w-1/2 bg-mui-blue-500/20 dark:bg-mui-blue-500/30 border-r border-mui-grey-300 dark:border-white/10 group-hover:bg-mui-blue-500/40" />
                    <div className="w-1/2 bg-transparent" />
                  </div>
                  <span className="text-[11px] text-mui-grey-600 dark:text-mui-grey-300">Left Half</span>
                </button>

                <button
                  type="button"
                  onClick={() => snapWindow('right')}
                  className="group flex flex-col items-center gap-1.5 p-2 rounded hover:bg-mui-grey-100 dark:hover:bg-white/5 border border-mui-grey-200 dark:border-white/5 transition-colors"
                >
                  <div className="w-12 h-8 border border-mui-grey-300 dark:border-white/10 rounded flex overflow-hidden">
                    <div className="w-1/2 bg-transparent" />
                    <div className="w-1/2 bg-mui-blue-500/20 dark:bg-mui-blue-500/30 border-l border-mui-grey-300 dark:border-white/10 group-hover:bg-mui-blue-500/40" />
                  </div>
                  <span className="text-[11px] text-mui-grey-600 dark:text-mui-grey-300">Right Half</span>
                </button>
              </div>

              <div className="grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onClick={() => snapWindow('left-third')}
                  className="group flex flex-col items-center gap-1.5 p-2 rounded hover:bg-mui-grey-100 dark:hover:bg-white/5 border border-mui-grey-200 dark:border-white/5 transition-colors"
                >
                  <div className="w-12 h-8 border border-mui-grey-300 dark:border-white/10 rounded flex overflow-hidden">
                    <div className="w-1/3 bg-mui-blue-500/20 dark:bg-mui-blue-500/30 border-r border-mui-grey-300 dark:border-white/10 group-hover:bg-mui-blue-500/40" />
                    <div className="w-2/3 bg-transparent" />
                  </div>
                  <span className="text-[11px] text-mui-grey-600 dark:text-mui-grey-300">Left 1/3</span>
                </button>

                <button
                  type="button"
                  onClick={() => snapWindow('right-third')}
                  className="group flex flex-col items-center gap-1.5 p-2 rounded hover:bg-mui-grey-100 dark:hover:bg-white/5 border border-mui-grey-200 dark:border-white/5 transition-colors"
                >
                  <div className="w-12 h-8 border border-mui-grey-300 dark:border-white/10 rounded flex overflow-hidden">
                    <div className="w-2/3 bg-transparent" />
                    <div className="w-1/3 bg-mui-blue-500/20 dark:bg-mui-blue-500/30 border-l border-mui-grey-300 dark:border-white/10 group-hover:bg-mui-blue-500/40" />
                  </div>
                  <span className="text-[11px] text-mui-grey-600 dark:text-mui-grey-300">Right 1/3</span>
                </button>
              </div>

              <button
                type="button"
                onClick={() => snapWindow('full')}
                className="group w-full flex items-center justify-between p-1.5 px-2.5 rounded hover:bg-mui-grey-100 dark:hover:bg-white/5 border border-mui-grey-200 dark:border-white/5 transition-colors text-left"
              >
                <span className="text-[11px] text-mui-grey-600 dark:text-mui-grey-300">Full Maximize</span>
                <div className="w-4 h-3 border border-mui-grey-400 dark:border-white/20 rounded-sm bg-mui-blue-500/10 group-hover:bg-mui-blue-500/30" />
              </button>
            </div>
          )}
        </div>

        <button
          type="button"
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
