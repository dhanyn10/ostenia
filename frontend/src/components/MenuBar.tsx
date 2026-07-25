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

  const snapWindow = async (
    layout:
      | "left"
      | "right"
      | "left-third"
      | "middle-third"
      | "right-third"
      | "left-twothirds"
      | "top-left"
      | "top-right"
      | "bottom-left"
      | "bottom-right"
      | "full",
  ) => {
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

      if (layout === "left") {
        runtime.WindowSetPosition(0, 0);
        runtime.WindowSetSize(Math.floor(screenWidth / 2), screenHeight);
      } else if (layout === "right") {
        runtime.WindowSetPosition(Math.floor(screenWidth / 2), 0);
        runtime.WindowSetSize(Math.floor(screenWidth / 2), screenHeight);
      } else if (layout === "left-third") {
        runtime.WindowSetPosition(0, 0);
        runtime.WindowSetSize(Math.floor(screenWidth / 3), screenHeight);
      } else if (layout === "middle-third") {
        runtime.WindowSetPosition(Math.floor(screenWidth / 3), 0);
        runtime.WindowSetSize(Math.floor(screenWidth / 3), screenHeight);
      } else if (layout === "right-third") {
        runtime.WindowSetPosition(Math.floor(screenWidth * 2 / 3), 0);
        runtime.WindowSetSize(Math.floor(screenWidth / 3), screenHeight);
      } else if (layout === "left-twothirds") {
        runtime.WindowSetPosition(0, 0);
        runtime.WindowSetSize(Math.floor((screenWidth * 2) / 3), screenHeight);
      } else if (layout === "top-left") {
        runtime.WindowSetPosition(0, 0);
        runtime.WindowSetSize(Math.floor(screenWidth / 2), Math.floor(screenHeight / 2));
      } else if (layout === "top-right") {
        runtime.WindowSetPosition(Math.floor(screenWidth / 2), 0);
        runtime.WindowSetSize(Math.floor(screenWidth / 2), Math.floor(screenHeight / 2));
      } else if (layout === "bottom-left") {
        runtime.WindowSetPosition(0, Math.floor(screenHeight / 2));
        runtime.WindowSetSize(Math.floor(screenWidth / 2), Math.floor(screenHeight / 2));
      } else if (layout === "bottom-right") {
        runtime.WindowSetPosition(Math.floor(screenWidth / 2), Math.floor(screenHeight / 2));
        runtime.WindowSetSize(Math.floor(screenWidth / 2), Math.floor(screenHeight / 2));
      } else if (layout === "full") {
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
                "absolute top-full right-0 mt-2 p-3 border shadow-2xl z-[150] rounded-lg bg-white dark:bg-mui-dark-paper border-mui-grey-200 dark:border-white/10 flex flex-col gap-4 animate-in fade-in zoom-in-95 duration-100 w-[240px]",
              )}
            >
              <div className="text-[11px] font-bold text-mui-grey-500 dark:text-mui-grey-400 uppercase tracking-wider">
                Snap Assist
              </div>

              <div className="grid grid-cols-2 gap-3">
                {/* 50/50 Layout */}
                <div className="flex flex-col gap-1.5">
                  <span className="text-[10px] text-mui-grey-400">50 / 50</span>
                  <div className="w-[100px] h-[64px] border border-mui-grey-300 dark:border-white/10 rounded-md p-1 flex gap-1 bg-mui-grey-50 dark:bg-white/5">
                    <button
                      type="button"
                      onClick={() => snapWindow("left")}
                      className="w-1/2 h-full border border-mui-grey-200 dark:border-white/10 rounded bg-mui-grey-100/50 dark:bg-white/5 hover:bg-mui-blue-500 hover:border-mui-blue-500 transition-all duration-100"
                      title="Snap Left Half"
                    />
                    <button
                      type="button"
                      onClick={() => snapWindow("right")}
                      className="w-1/2 h-full border border-mui-grey-200 dark:border-white/10 rounded bg-mui-grey-100/50 dark:bg-white/5 hover:bg-mui-blue-500 hover:border-mui-blue-500 transition-all duration-100"
                      title="Snap Right Half"
                    />
                  </div>
                </div>

                {/* 2/3 & 1/3 Layout */}
                <div className="flex flex-col gap-1.5">
                  <span className="text-[10px] text-mui-grey-400">2/3 & 1/3</span>
                  <div className="w-[100px] h-[64px] border border-mui-grey-300 dark:border-white/10 rounded-md p-1 flex gap-1 bg-mui-grey-50 dark:bg-white/5">
                    <button
                      type="button"
                      onClick={() => snapWindow("left-twothirds")}
                      className="w-2/3 h-full border border-mui-grey-200 dark:border-white/10 rounded bg-mui-grey-100/50 dark:bg-white/5 hover:bg-mui-blue-500 hover:border-mui-blue-500 transition-all duration-100"
                      title="Snap Left 2/3"
                    />
                    <button
                      type="button"
                      onClick={() => snapWindow("right-third")}
                      className="w-1/3 h-full border border-mui-grey-200 dark:border-white/10 rounded bg-mui-grey-100/50 dark:bg-white/5 hover:bg-mui-blue-500 hover:border-mui-blue-500 transition-all duration-100"
                      title="Snap Right 1/3"
                    />
                  </div>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                {/* 3 Columns Layout */}
                <div className="flex flex-col gap-1.5">
                  <span className="text-[10px] text-mui-grey-400">3 Columns</span>
                  <div className="w-[100px] h-[64px] border border-mui-grey-300 dark:border-white/10 rounded-md p-1 flex gap-1 bg-mui-grey-50 dark:bg-white/5">
                    <button
                      type="button"
                      onClick={() => snapWindow("left-third")}
                      className="w-1/3 h-full border border-mui-grey-200 dark:border-white/10 rounded bg-mui-grey-100/50 dark:bg-white/5 hover:bg-mui-blue-500 hover:border-mui-blue-500 transition-all duration-100"
                      title="Snap Left 1/3"
                    />
                    <button
                      type="button"
                      onClick={() => snapWindow("middle-third")}
                      className="w-1/3 h-full border border-mui-grey-200 dark:border-white/10 rounded bg-mui-grey-100/50 dark:bg-white/5 hover:bg-mui-blue-500 hover:border-mui-blue-500 transition-all duration-100"
                      title="Snap Middle 1/3"
                    />
                    <button
                      type="button"
                      onClick={() => snapWindow("right-third")}
                      className="w-1/3 h-full border border-mui-grey-200 dark:border-white/10 rounded bg-mui-grey-100/50 dark:bg-white/5 hover:bg-mui-blue-500 hover:border-mui-blue-500 transition-all duration-100"
                      title="Snap Right 1/3"
                    />
                  </div>
                </div>

                {/* 2x2 Grid Layout */}
                <div className="flex flex-col gap-1.5">
                  <span className="text-[10px] text-mui-grey-400">2 x 2 Grid</span>
                  <div className="w-[100px] h-[64px] border border-mui-grey-300 dark:border-white/10 rounded-md p-1 flex flex-col gap-1 bg-mui-grey-50 dark:bg-white/5">
                    <div className="flex gap-1 h-1/2">
                      <button
                        type="button"
                        onClick={() => snapWindow("top-left")}
                        className="w-1/2 h-full border border-mui-grey-200 dark:border-white/10 rounded bg-mui-grey-100/50 dark:bg-white/5 hover:bg-mui-blue-500 hover:border-mui-blue-500 transition-all duration-100"
                        title="Snap Top Left"
                      />
                      <button
                        type="button"
                        onClick={() => snapWindow("top-right")}
                        className="w-1/2 h-full border border-mui-grey-200 dark:border-white/10 rounded bg-mui-grey-100/50 dark:bg-white/5 hover:bg-mui-blue-500 hover:border-mui-blue-500 transition-all duration-100"
                        title="Snap Top Right"
                      />
                    </div>
                    <div className="flex gap-1 h-1/2">
                      <button
                        type="button"
                        onClick={() => snapWindow("bottom-left")}
                        className="w-1/2 h-full border border-mui-grey-200 dark:border-white/10 rounded bg-mui-grey-100/50 dark:bg-white/5 hover:bg-mui-blue-500 hover:border-mui-blue-500 transition-all duration-100"
                        title="Snap Bottom Left"
                      />
                      <button
                        type="button"
                        onClick={() => snapWindow("bottom-right")}
                        className="w-1/2 h-full border border-mui-grey-200 dark:border-white/10 rounded bg-mui-grey-100/50 dark:bg-white/5 hover:bg-mui-blue-500 hover:border-mui-blue-500 transition-all duration-100"
                        title="Snap Bottom Right"
                      />
                    </div>
                  </div>
                </div>
              </div>

              <button
                type="button"
                onClick={() => snapWindow("full")}
                className="group w-full flex items-center justify-between p-2 rounded-md hover:bg-mui-grey-100 dark:hover:bg-white/5 border border-mui-grey-200 dark:border-white/10 transition-colors text-left"
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
