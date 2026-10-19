import { useState, useCallback } from 'react';

export interface LogEntry {
  id: string;
  timestamp: string;
  level: string;
  message: string;
  callerFunc?: string;
  file?: string;
  stackTrace?: string;
}

export function useLogViewerLogic(initialLogs: LogEntry[] = []) {
  const [logs, setLogs] = useState<LogEntry[]>(initialLogs);
  const [viewMode, setViewMode] = useState<'simple' | 'complete'>('simple');
  const [pageSize, setPageSize] = useState<number>(20);
  const [currentPage, setCurrentPage] = useState<number>(1);
  const [selectedLogIds, setSelectedLogIds] = useState<Set<string>>(new Set());

  const toggleLogSelection = useCallback((id: string) => {
    setSelectedLogIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  const selectAllOnPage = useCallback((pageLogIds: string[]) => {
    setSelectedLogIds(new Set(pageLogIds));
  }, []);

  const clearSelection = useCallback(() => {
    setSelectedLogIds(new Set());
  }, []);

  const addLogs = useCallback((newLogs: LogEntry[]) => {
    setLogs((prev) => [...newLogs, ...prev]);
  }, []);

  const clearLogs = useCallback(() => {
    setLogs([]);
    setSelectedLogIds(new Set());
  }, []);

  return {
    logs,
    viewMode,
    setViewMode,
    pageSize,
    setPageSize,
    currentPage,
    setCurrentPage,
    selectedLogIds,
    toggleLogSelection,
    selectAllOnPage,
    clearSelection,
    addLogs,
    clearLogs,
  };
}
