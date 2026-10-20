import { useState, useCallback } from 'react';
import * as AppBackend from '../../wailsjs/go/backend/App';
import { parseStackTrace } from '../utils/stackParser';

export interface LogItem {
  id: string;
  time: string;
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

export function useAppLogs() {
  const [logs, setLogs] = useState<LogItem[]>([]);

  const addLog = useCallback((msg: string, type = 'info', isServiceLog = false) => {
    if (isLogging) return;
    isLogging = true;
    try {
      const time = new Date().toLocaleTimeString();
      const id = crypto.randomUUID();
      let prefix = 'SYS';
      if (type === 'error') {
        prefix = 'ERR';
      } else if (type === 'warn') {
        prefix = 'WRN';
      }

      const stackTrace = isServiceLog ? '' : new Error().stack;
      const { caller, rawStack } = isServiceLog ? { caller: undefined, rawStack: '' } : parseStackTrace(stackTrace);

      const formattedMsg = `[${prefix}] ${msg}`;
      if (AppBackend && (AppBackend as any).SaveLogToFile) {
        (AppBackend as any).SaveLogToFile(formattedMsg).catch(() => {});
      }

      setLogs(prev => [
        {
          id,
          time,
          msg: formattedMsg,
          cleanMsg: msg,
          type,
          isServiceLog,
          caller,
          rawStack,
        },
        ...prev
      ].slice(0, 1000));
    } finally {
      isLogging = false;
    }
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

  return { logs, setLogs, addLog, setupConsoleOverrides };
}
