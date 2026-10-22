import type React from "react";
import type { Terminal as XTerm } from "@xterm/xterm";

export interface ZoomHandlers {
  handleZoomIn: () => void;
  handleZoomOut: () => void;
  handleZoomReset: () => void;
}

export interface SSHStorageConfig {
  maxTimeout: number;
  maxRetries: number;
}

export interface SortConfigFile {
  key: string;
  direction: "asc" | "desc";
}

/**
 * Checks whether a given filename represents a known archive file format.
 */
export const isArchiveFile = (fileName?: string): boolean => {
  if (!fileName) return false;
  const name = fileName.toLowerCase();
  return (
    name.endsWith(".zip") ||
    name.endsWith(".tar") ||
    name.endsWith(".gz") ||
    name.endsWith(".7z") ||
    name.endsWith(".rar") ||
    name.endsWith(".bz2") ||
    name.endsWith(".xz")
  );
};

/**
 * Calculates Y position for custom context menus to avoid viewport clipping.
 */
export const calculateContextMenuY = (
  clientY: number,
  menuHeight: number,
): number => {
  if (
    typeof window !== "undefined" &&
    clientY + menuHeight > window.innerHeight
  ) {
    return Math.max(10, clientY - menuHeight);
  }
  return clientY;
};

/**
 * Checks if an SFTP listing error is expected/ignorable during background sync.
 */
export const isIgnorableSFTPError = (error: any): boolean => {
  const errorString = String(error).toLowerCase();
  return (
    errorString.includes("eof") ||
    errorString.includes("session not found") ||
    errorString.includes("session not connected") ||
    errorString.includes("sftp not connected")
  );
};

/**
 * Extracts timeout and retry options from localStorage.
 */
export const getSSHConfigFromStorage = (): SSHStorageConfig => {
  if (typeof localStorage === "undefined") {
    return { maxTimeout: 10, maxRetries: 3 };
  }
  const timeout = Number.parseInt(
    localStorage.getItem("ostenia_ssh_max_timeout") || "10",
    10,
  );
  const retries = Number.parseInt(
    localStorage.getItem("ostenia_ssh_max_retries") || "3",
    10,
  );
  return {
    maxTimeout: Number.isNaN(timeout) || timeout < 1 ? 10 : timeout,
    maxRetries: Number.isNaN(retries) || retries < 1 ? 3 : retries,
  };
};

/**
 * Extracts target path from xterm title string if valid.
 */
export const extractPathFromTitle = (title: string): string | null => {
  if (!title.includes(":")) return null;
  const parts = title.split(":");
  const potentialPath = parts.at(-1)?.trim() || "";
  return potentialPath.startsWith("/") ? potentialPath : null;
};

/**
 * Attaches custom keyboard handlers for zoom shortcuts in xterm.
 */
export const setupZoomKeyHandler = (
  term: XTerm,
  handlersRef: React.RefObject<ZoomHandlers>,
): void => {
  term.attachCustomKeyEventHandler((event: KeyboardEvent) => {
    if (event.type !== "keydown") return true;
    const isCtrlOrCmd = event.ctrlKey || event.metaKey;
    if (!isCtrlOrCmd) return true;
    const enabled =
      typeof localStorage !== "undefined" &&
      localStorage.getItem("ostenia_ssh_zoom_enabled") !== "false";
    if (!enabled) return true;

    if (event.key === "=" || event.key === "+") {
      handlersRef.current?.handleZoomIn();
      return false;
    }
    if (event.key === "-") {
      handlersRef.current?.handleZoomOut();
      return false;
    }
    if (event.key === "0") {
      handlersRef.current?.handleZoomReset();
      return false;
    }
    return true;
  });
};

/**
 * Helper utility to convert raw file sizes into readable units.
 */
export const formatSize = (bytes: number): string => {
  if (!bytes) return "-";
  const units = ["B", "KB", "MB", "GB", "TB"];
  let size = bytes;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }
  return `${size.toFixed(1)} ${units[unitIndex]}`;
};

/**
 * Filters and sorts remote file list based on search query, hidden file toggle, and sort column.
 */
export const filterAndSortFiles = (
  files: any[],
  searchQuery: string,
  showHiddenFiles: boolean,
  sortConfig: SortConfigFile,
): any[] => {
  return [...files]
    .filter((file) => {
      const matchesSearch = file.name
        .toLowerCase()
        .includes(searchQuery.toLowerCase());
      if (!showHiddenFiles) {
        return matchesSearch && !file.name.startsWith(".");
      }
      return matchesSearch;
    })
    .sort((a, b) => {
      if (a.isDir !== b.isDir) return b.isDir ? 1 : -1;
      const comparison =
        sortConfig.key === "name"
          ? a.name.localeCompare(b.name)
          : (a.size || 0) - (b.size || 0);
      return sortConfig.direction === "asc" ? comparison : -comparison;
    });
};
