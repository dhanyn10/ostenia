import { useState, useCallback } from 'react';
import * as AppBackend from '@/wailsjs/go/backend/App';
import { parseStackTrace } from '@/utils/stackParser';

export interface LogItem {
  id: string;
  time: string;
  timestamp?: string;
  msg: string;
  cleanMsg: string;
  type: string;
  isServiceLog: boolean;
  caller?: any;
  rawStack?: string;
}

let isLogging = false;

export function formatLogArgs(args: any[]): string {
  return args.map(arg => typeof arg === 'object' ? JSON.stringify(arg) : arg).join(' ');
}

function getLogPrefix(type: string): string {
  if (type === 'error') return 'ERR';
  if (type === 'warn') return 'WRN';
  return 'SYS';
}

function persistLog(logItem: LogItem, formattedMsg: string): void {
  if (AppBackend && (AppBackend as any).SaveLogEntry) {
    try {
      const promise = (AppBackend as any).SaveLogEntry(logItem);
      if (promise && typeof promise.catch === 'function') {
        promise.catch(() => {
          // Ignored: best-effort background log persistence
        });
      }
    } catch {
      // Ignored: backend call failure should not disrupt the application
    }
  } else if (AppBackend && (AppBackend as any).SaveLogToFile) {
    try {
      const promise = (AppBackend as any).SaveLogToFile(formattedMsg);
      if (promise && typeof promise.catch === 'function') {
        promise.catch(() => {
          // Ignored: best-effort background log persistence
        });
      }
    } catch {
      // Ignored: backend call failure should not disrupt the application
    }
  }
}

function createLogEntry(
  message: string,
  type: string,
  isServiceLog: boolean
): { logItem: LogItem; formattedMsg: string } {
  const time = new Date().toLocaleTimeString();
  const timestamp = new Date().toISOString();
  const id = crypto.randomUUID();
  const prefix = getLogPrefix(type);
  const formattedMsg = `[${prefix}] ${message}`;

  let caller;
  let rawStack = '';
  if (!isServiceLog) {
    const trace = parseStackTrace(new Error('Trace capture').stack);
    caller = trace.caller;
    rawStack = trace.rawStack;
  }

  const logItem: LogItem = {
    id,
    time,
    timestamp,
    msg: formattedMsg,
    cleanMsg: message,
    type,
    isServiceLog,
    caller,
    rawStack,
  };

  return { logItem, formattedMsg };
}

export function useAppLogs() {
  const [logs, setLogs] = useState<LogItem[]>([]);

  const addLog = useCallback((message: string, type = 'info', isServiceLog = false) => {
    if (isLogging) return;
    isLogging = true;
    try {
      const { logItem, formattedMsg } = createLogEntry(message, type, isServiceLog);
      persistLog(logItem, formattedMsg);
      setLogs(prev => [logItem, ...prev].slice(0, 1000));
    } finally {
      isLogging = false;
    }
  }, []);

  const loadRecentLogs = useCallback(async (limit = 500) => {
    if (AppBackend && (AppBackend as any).GetRecentLogs) {
      try {
        const recent = await (AppBackend as any).GetRecentLogs(limit);
        if (Array.isArray(recent) && recent.length > 0) {
          setLogs(prev => {
            const existingIds = new Set(prev.map(log => log.id));
            const newEntries = (recent as LogItem[]).filter(log => !existingIds.has(log.id));
            return [...prev, ...newEntries].slice(0, 1000);
          });
        }
      } catch (error) {
        console.error('Failed to load recent logs:', error);
      }
    }
  }, []);

  const clearAppLogs = useCallback(async () => {
    if (AppBackend && (AppBackend as any).ClearLogs) {
      try {
        await (AppBackend as any).ClearLogs();
      } catch (error) {
        console.error('Failed to clear logs:', error);
      }
    }
    setLogs([]);
  }, []);

  const setupConsoleOverrides = useCallback((originalLog: any, originalWarn: any, originalError: any) => {
    console.log = (...args: any[]) => {
      addLog(formatLogArgs(args), 'info');
      originalLog.apply(console, args);
    };
    console.warn = (...args: any[]) => {
      addLog(formatLogArgs(args), 'warn');
      originalWarn.apply(console, args);
    };
    console.error = (...args: any[]) => {
      addLog(formatLogArgs(args), 'error');
      originalError.apply(console, args);
    };
  }, [addLog]);

  return { logs, setLogs, addLog, loadRecentLogs, clearAppLogs, setupConsoleOverrides };
}
