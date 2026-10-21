import { useState, useCallback } from 'react';
import * as AppBackend from '../../wailsjs/go/backend/App';
import { parseStackTrace } from '../utils/stackParser';

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
  return args.map(a => typeof a === 'object' ? JSON.stringify(a) : a).join(' ');
}

function getLogPrefix(type: string): string {
  if (type === 'error') return 'ERR';
  if (type === 'warn') return 'WRN';
  return 'SYS';
}

function persistLog(logItem: LogItem, formattedMsg: string): void {
  if (AppBackend && (AppBackend as any).SaveLogEntry) {
    try {
      const p = (AppBackend as any).SaveLogEntry(logItem);
      if (p && typeof p.catch === 'function') {
        p.catch(() => {
          // Ignored: best-effort background log persistence
        });
      }
    } catch {
      // Ignored: backend call failure should not disrupt the application
    }
  } else if (AppBackend && (AppBackend as any).SaveLogToFile) {
    try {
      const p = (AppBackend as any).SaveLogToFile(formattedMsg);
      if (p && typeof p.catch === 'function') {
        p.catch(() => {
          // Ignored: best-effort background log persistence
        });
      }
    } catch {
      // Ignored: backend call failure should not disrupt the application
    }
  }
}

function createLogEntry(
  msg: string,
  type: string,
  isServiceLog: boolean
): { logItem: LogItem; formattedMsg: string } {
  const time = new Date().toLocaleTimeString();
  const timestamp = new Date().toISOString();
  const id = crypto.randomUUID();
  const prefix = getLogPrefix(type);
  const formattedMsg = `[${prefix}] ${msg}`;

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
    cleanMsg: msg,
    type,
    isServiceLog,
    caller,
    rawStack,
  };

  return { logItem, formattedMsg };
}

export function useAppLogs() {
  const [logs, setLogs] = useState<LogItem[]>([]);

  const addLog = useCallback((msg: string, type = 'info', isServiceLog = false) => {
    if (isLogging) return;
    isLogging = true;
    try {
      const { logItem, formattedMsg } = createLogEntry(msg, type, isServiceLog);
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
            const existingIds = new Set(prev.map(l => l.id));
            const newEntries = (recent as LogItem[]).filter(l => !existingIds.has(l.id));
            return [...prev, ...newEntries].slice(0, 1000);
          });
        }
      } catch (err) {
        console.error('Failed to load recent logs:', err);
      }
    }
  }, []);

  const clearAppLogs = useCallback(async () => {
    if (AppBackend && (AppBackend as any).ClearLogs) {
      try {
        await (AppBackend as any).ClearLogs();
      } catch (err) {
        console.error('Failed to clear logs:', err);
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
