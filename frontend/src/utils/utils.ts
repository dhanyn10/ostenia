import type React from "react";
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";
import type { Terminal as XTerm } from "@xterm/xterm";
import type { LogEntry, LogLevel, PaginationItem } from "@/types/logs";

// ============================================================================
// 1. STYLING (Tailwind & clsx)
// ============================================================================

export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}

// ============================================================================
// 2. ACCESSIBILITY & DOM HELPERS
// ============================================================================

export const handleActionKey =
  (callback: (event?: any) => void) => (event: React.KeyboardEvent) => {
    if (event.key === "Enter" || event.key === " ") {
      event.preventDefault();
      callback(event);
    }
  };

/**
 * Determines whether a click event target within a card or list item
 * should be ignored to avoid accidentally triggering card toggle/expand.
 */
export function shouldIgnoreCardToggle(target: HTMLElement): boolean {
  return Boolean(
    target.closest("button:not([data-card-expand])") ||
      target.closest("a") ||
      target.closest("input") ||
      target.closest("[data-no-card-toggle]"),
  );
}

// ============================================================================
// 3. ACTIVITY & PERFORMANCE LOGGING
// ============================================================================

/**
 * Measures the execution time of a synchronous or asynchronous action
 * and logs the performance and results to the console.
 */
export async function measureActivity<T>(
  funcName: string,
  action: () => Promise<T> | T,
): Promise<T> {
  const startTime = performance.now();
  try {
    const result = await action();
    const duration = (performance.now() - startTime).toFixed(1);
    console.log(
      `Activity: ${funcName} executed in ${duration}ms. Result: SUCCESS`,
    );
    return result;
  } catch (err: any) {
    const duration = (performance.now() - startTime).toFixed(1);
    const errMsg = err?.message || String(err);
    console.error(
      `Activity: ${funcName} executed in ${duration}ms. Result: ERROR - ${errMsg}`,
    );
    throw err;
  }
}

// ============================================================================
// 4. PLUGIN & PREREQUISITE HELPERS
// ============================================================================

export function resolvePrerequisiteTask(task: any, selectedVersion?: string) {
  const selectedVer = selectedVersion || task.version;
  const prefixMap: Record<string, string> = {
    PHP: "php-",
    Apache: "httpd-",
    MySQL: "mysql-",
    Nginx: "nginx-",
    OpenSSL: "openssl-",
    "Node.js": "node-v",
    Python: "python-",
  };
  const categoryMap: Record<string, string> = {
    "Node.js": "nodejs",
    HeidiSQL: "heidisql",
  };
  const category = categoryMap[task.name] || task.name.toLowerCase();
  const prefix = prefixMap[task.name] || "";

  const modifiedTask = {
    ...task,
    version: selectedVer,
    target: `${category}/${prefix}${selectedVer}`,
  };

  if (task.versionUrls?.[selectedVer]) {
    modifiedTask.url = task.versionUrls[selectedVer];
  }

  return modifiedTask;
}

export function isCustomVersionAllowed(taskName: string): boolean {
  return taskName !== "HeidiSQL" && taskName !== "OpenSSL";
}

export function isParentTaskActive(progressObj: any): boolean {
  return Boolean(
    progressObj &&
      progressObj.status !== "Ready" &&
      progressObj.status !== "Completed" &&
      !progressObj.status?.startsWith("Error"),
  );
}

export function isTaskVersionInstalled(
  task: any,
  selectedVersion?: string,
  installedVersions: string[] = [],
): boolean {
  if (task.name === "OpenSSL" || task.name === "HeidiSQL") {
    return Boolean(task.isInstalled);
  }
  return installedVersions.includes(selectedVersion || task.version);
}

// ============================================================================
// 5. SERVICE HELPERS
// ============================================================================

export interface ServiceActionFlags {
  isInstalled: boolean;
  isHttpsEnabled: boolean;
  hasOpenFolder: boolean;
  hasTerminal: boolean;
  hasHttpsToggle: boolean;
  hasPhpExtManager: boolean;
  hasHeidiOpen: boolean;
  hasExtraActions: boolean;
}

export function getServiceActionFlags(
  service: any,
  task: any,
  apacheHttps: boolean,
  nginxHttps: boolean,
  isOpenSslEnabled: boolean,
): ServiceActionFlags {
  const isInstalled =
    Boolean(task?.installedVers && task.installedVers.length > 0) ||
    service.name === "OpenSSL";
  const isWebServer = service.name === "Apache" || service.name === "Nginx";

  let isHttpsEnabled = false;
  if (service.name === "Apache") {
    isHttpsEnabled = apacheHttps;
  } else if (service.name === "Nginx") {
    isHttpsEnabled = nginxHttps;
  }

  const hasTerminalFacility =
    service.name !== "HeidiSQL" && service.name !== "OpenSSL";
  const hasOpenFolder =
    isInstalled && service.name !== "OpenSSL" && service.name !== "HeidiSQL";
  const hasTerminal = isInstalled && hasTerminalFacility;
  const hasHttpsToggle = isWebServer && isOpenSslEnabled;
  const hasPhpExtManager = service.name === "PHP";
  const hasHeidiOpen = service.name === "HeidiSQL" && isInstalled;
  const hasExtraActions =
    hasOpenFolder ||
    hasTerminal ||
    hasHttpsToggle ||
    hasPhpExtManager ||
    hasHeidiOpen;

  return {
    isInstalled,
    isHttpsEnabled,
    hasOpenFolder,
    hasTerminal,
    hasHttpsToggle,
    hasPhpExtManager,
    hasHeidiOpen,
    hasExtraActions,
  };
}

// ============================================================================
// 5. CHART & SPARKLINE UTILS
// ============================================================================

export type ResourceMetricType = "cpu" | "mem" | "disk";

export interface ResourceDataPoint {
  cpu: number | null;
  mem: number | null;
  disk: number | null;
}

export interface SparklinePaths {
  linePath: string;
  areaPath: string;
}

/**
 * Calculates continuous coordinate segments for valid numeric points,
 * breaking whenever a null/missing data point is encountered.
 */
export const buildSparklineSegments = (
  paddedData: ResourceDataPoint[],
  metric: ResourceMetricType,
  width: number,
  height: number,
  pointsCount: number,
): Array<Array<[number, number]>> => {
  const getX = (index: number) => (index / (pointsCount - 1)) * width;
  const getY = (val: number | null) =>
    val === null ? height : height - 1 - (val / 100) * (height - 2);

  let currentSegment: Array<[number, number]> = [];
  const segments: Array<Array<[number, number]>> = [];

  paddedData.forEach((d, i) => {
    const val = d[metric];
    if (val !== null) {
      currentSegment.push([getX(i), getY(val)]);
    } else if (currentSegment.length > 0) {
      segments.push(currentSegment);
      currentSegment = [];
    }
  });
  if (currentSegment.length > 0) {
    segments.push(currentSegment);
  }
  return segments;
};

/**
 * Generates SVG path commands (M, L, Z) for the stroke line and filled area.
 */
export const buildSparklinePaths = (
  segments: Array<Array<[number, number]>>,
  height: number,
): SparklinePaths => {
  let linePath = "";
  let areaPath = "";

  segments.forEach((seg) => {
    if (seg.length === 0) return;

    let segLine = `M ${seg[0][0]} ${seg[0][1]}`;
    let segArea = `M ${seg[0][0]} ${height} L ${seg[0][0]} ${seg[0][1]}`;

    for (let j = 1; j < seg.length; j++) {
      segLine += ` L ${seg[j][0]} ${seg[j][1]}`;
      segArea += ` L ${seg[j][0]} ${seg[j][1]}`;
    }

    segArea += ` L ${seg.at(-1)![0]} ${height} Z`;

    linePath += " " + segLine;
    areaPath += " " + segArea;
  });

  return { linePath, areaPath };
};

// ============================================================================
// 6. LOG FORMATTING & PAGINATION
// ============================================================================

export function getPaginationRange(
  currentPage: number,
  totalPages: number,
  siblingCount = 1,
): PaginationItem[] {
  const totalPageNumbers = siblingCount + 5;

  if (totalPageNumbers >= totalPages) {
    return Array.from({ length: totalPages }, (_, i) => ({
      key: `page-${i + 1}`,
      value: i + 1,
      isDots: false,
    }));
  }

  const leftSiblingIndex = Math.max(currentPage - siblingCount, 1);
  const rightSiblingIndex = Math.min(currentPage + siblingCount, totalPages);

  const shouldShowLeftDots = leftSiblingIndex > 2;
  const shouldShowRightDots = rightSiblingIndex < totalPages - 2;

  if (!shouldShowLeftDots && shouldShowRightDots) {
    const leftItemCount = 3 + 2 * siblingCount;
    const leftRange = Array.from({ length: leftItemCount }, (_, i) => ({
      key: `page-${i + 1}`,
      value: i + 1,
      isDots: false,
    }));
    return [
      ...leftRange,
      { key: "dots-right", value: "...", isDots: true },
      { key: `page-${totalPages}`, value: totalPages, isDots: false },
    ];
  }

  if (shouldShowLeftDots && !shouldShowRightDots) {
    const rightItemCount = 3 + 2 * siblingCount;
    const rightRange = Array.from({ length: rightItemCount }, (_, i) => {
      const pageNum = totalPages - rightItemCount + i + 1;
      return { key: `page-${pageNum}`, value: pageNum, isDots: false };
    });
    return [
      { key: "page-1", value: 1, isDots: false },
      { key: "dots-left", value: "...", isDots: true },
      ...rightRange,
    ];
  }

  if (shouldShowLeftDots && shouldShowRightDots) {
    const middleRange = Array.from(
      { length: rightSiblingIndex - leftSiblingIndex + 1 },
      (_, i) => {
        const pageNum = leftSiblingIndex + i;
        return { key: `page-${pageNum}`, value: pageNum, isDots: false };
      },
    );
    return [
      { key: "page-1", value: 1, isDots: false },
      { key: "dots-left", value: "...", isDots: true },
      ...middleRange,
      { key: "dots-right", value: "...", isDots: true },
      { key: `page-${totalPages}`, value: totalPages, isDots: false },
    ];
  }

  return [];
}

export function getLogLevel(log: LogEntry): LogLevel {
  if (log.isServiceLog) return "SERVICE";
  const type = (log.type || "").toLowerCase();
  if (type === "error" || type === "err") return "ERROR";
  if (type === "warn" || type === "wrn") return "WARN";

  const msg = (log.msg || "").toLowerCase();
  if (
    msg.includes("[err]") ||
    msg.includes("error") ||
    msg.includes("failed") ||
    msg.includes("crash") ||
    msg.includes("fatal") ||
    msg.includes("panic") ||
    msg.includes("exception")
  ) {
    return "ERROR";
  }
  if (msg.includes("[wrn]")) return "WARN";
  return "INFO";
}

export function formatCompleteLog(log: LogEntry): string {
  const level = getLogLevel(log);
  const timeStr = log.time || "";

  let message = log.cleanMsg || log.msg || "";
  if (
    message.startsWith("[SYS] ") ||
    message.startsWith("[ERR] ") ||
    message.startsWith("[WRN] ")
  ) {
    message = message.substring(6);
  }

  const callerFunc = log.caller ? `${log.caller.functionName}()` : "N/A";
  const fileInfo = log.caller
    ? `${log.caller.fileName}:${log.caller.line}:${log.caller.column}`
    : "N/A";
  const stackTraceText = log.rawStack ? log.rawStack : "N/A";

  return [
    `Timestamp: [${timeStr}]`,
    `Level: ${level}`,
    `Message: ${message}`,
    `Caller Function: ${callerFunc}`,
    `File: ${fileInfo}`,
    `Stack Trace:\n${stackTraceText}`,
  ].join("\n");
}

// ============================================================================
// 7. STACK TRACE PARSER
// ============================================================================

export interface CallerInfo {
  functionName: string;
  fileName: string;
  line: string;
  column: string;
}

export interface ParsedStack {
  caller?: CallerInfo;
  rawStack: string;
}

export function cleanFileName(path: string): string {
  if (!path) return "";
  let clean = path.split("?")[0].split("#")[0];
  const lastSlash = Math.max(clean.lastIndexOf("/"), clean.lastIndexOf("\\"));
  if (lastSlash !== -1) {
    clean = clean.substring(lastSlash + 1);
  }
  return clean;
}

export function parseStackLine(line: string): CallerInfo | null {
  const trimmed = line.trim();
  if (!trimmed) return null;

  let functionName = "";
  let location = "";

  if (trimmed.startsWith("at ")) {
    const lastParen = trimmed.lastIndexOf("(");
    if (trimmed.endsWith(")") && lastParen !== -1) {
      functionName = trimmed.substring(3, lastParen).trim();
      location = trimmed.substring(lastParen + 1, trimmed.length - 1);
    } else {
      functionName = "anonymous";
      location = trimmed.substring(3).trim();
    }
  } else if (trimmed.includes("@")) {
    const atIdx = trimmed.indexOf("@");
    functionName = trimmed.substring(0, atIdx).trim() || "anonymous";
    location = trimmed.substring(atIdx + 1);
  } else {
    return null;
  }

  const lastColon = location.lastIndexOf(":");
  const secondLastColon = location.lastIndexOf(":", lastColon - 1);
  if (lastColon === -1 || secondLastColon === -1) return null;

  const lineNum = location.substring(secondLastColon + 1, lastColon);
  const colNum = location.substring(lastColon + 1);

  if (!/^\d+$/.test(lineNum) || !/^\d+$/.test(colNum)) return null;

  const rawFilePath = location.substring(0, secondLastColon);

  return {
    functionName: functionName || "anonymous",
    fileName: cleanFileName(rawFilePath),
    line: lineNum,
    column: colNum,
  };
}

const ignorePatterns = [
  /addLog/i,
  /setupConsoleOverrides/i,
  /stackParser/i,
  /parseStackTrace/i,
  /measureActivity/i,
  /activityLogger/i,
  /console\.(log|warn|error|info)/i,
  /Object\.(log|warn|error|info)/i,
  /at\s+log\s+\(/i,
  /at\s+warn\s+\(/i,
  /at\s+error\s+\(/i,
];

function isIgnoredFrame(frame: CallerInfo): boolean {
  return ignorePatterns.some(
    (pattern) =>
      pattern.test(frame.functionName) || pattern.test(frame.fileName),
  );
}

export function parseStackTrace(stack: string = ""): ParsedStack {
  if (!stack) {
    return { rawStack: stack };
  }
  const rawStack = stack;

  const lines = rawStack.split("\n");
  const frames: CallerInfo[] = [];

  for (const line of lines) {
    const frame = parseStackLine(line);
    if (frame) {
      frames.push(frame);
    }
  }

  const caller = frames.find((frame) => !isIgnoredFrame(frame));

  return {
    caller,
    rawStack,
  };
}

// ============================================================================
// 8. SSH UTILITIES
// ============================================================================

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

export const isIgnorableSFTPError = (error: any): boolean => {
  const errorString = String(error).toLowerCase();
  return (
    errorString.includes("eof") ||
    errorString.includes("session not found") ||
    errorString.includes("session not connected") ||
    errorString.includes("sftp not connected")
  );
};

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

export const extractPathFromTitle = (title: string): string | null => {
  if (!title.includes(":")) return null;
  const parts = title.split(":");
  const potentialPath = parts.at(-1)?.trim() || "";
  return potentialPath.startsWith("/") ? potentialPath : null;
};

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
