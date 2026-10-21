import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useAppLogs, formatLogArgs } from './useAppLogs';
import * as AppBackend from '../../wailsjs/go/backend/App';

vi.mock('../../wailsjs/go/backend/App', () => ({
  SaveLogToFile: vi.fn().mockResolvedValue(undefined),
  SaveLogEntry: vi.fn().mockResolvedValue(undefined),
  GetRecentLogs: vi.fn().mockResolvedValue([]),
  ClearLogs: vi.fn().mockResolvedValue(undefined),
  OpenLogsFolder: vi.fn().mockResolvedValue(undefined),
}));

describe('useAppLogs hook', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('formatLogArgs formats objects and primitives correctly', () => {
    expect(formatLogArgs(['hello', 123, true])).toBe('hello 123 true');
    expect(formatLogArgs([{ key: 'val' }])).toBe('{"key":"val"}');
  });

  it('starts with empty logs', () => {
    const { result } = renderHook(() => useAppLogs());
    expect(result.current.logs).toEqual([]);
  });

  it('adds logs with appropriate prefixes and calls SaveLogEntry', () => {
    const { result } = renderHook(() => useAppLogs());

    act(() => {
      result.current.addLog('General message', 'info');
    });

    expect(result.current.logs).toHaveLength(1);
    expect(result.current.logs[0].msg).toContain('[SYS] General message');
    expect(result.current.logs[0].type).toBe('info');
    expect(AppBackend.SaveLogEntry).toHaveBeenCalledTimes(1);

    act(() => {
      result.current.addLog('Warning message', 'warn');
    });

    expect(result.current.logs).toHaveLength(2);
    expect(result.current.logs[0].msg).toContain('[WRN] Warning message');

    act(() => {
      result.current.addLog('Error message', 'error');
    });

    expect(result.current.logs).toHaveLength(3);
    expect(result.current.logs[0].msg).toContain('[ERR] Error message');
  });

  it('captures console overrides', () => {
    const { result } = renderHook(() => useAppLogs());

    const origLog = vi.fn();
    const origWarn = vi.fn();
    const origError = vi.fn();

    act(() => {
      result.current.setupConsoleOverrides(origLog, origWarn, origError);
    });

    act(() => {
      console.log('Console test');
    });

    expect(origLog).toHaveBeenCalledWith('Console test');
    expect(result.current.logs.some(l => l.cleanMsg === 'Console test')).toBe(true);
  });

  it('loads recent logs via loadRecentLogs', async () => {
    const mockRecentLogs = [
      {
        id: 'rec-1',
        time: '10:00:00',
        msg: '[SYS] Restored log 1',
        cleanMsg: 'Restored log 1',
        type: 'info',
        isServiceLog: false,
      },
      {
        id: 'rec-2',
        time: '10:01:00',
        msg: '[ERR] Restored log 2',
        cleanMsg: 'Restored log 2',
        type: 'error',
        isServiceLog: false,
      },
    ];
    vi.mocked(AppBackend.GetRecentLogs).mockResolvedValueOnce(mockRecentLogs as any);

    const { result } = renderHook(() => useAppLogs());

    await act(async () => {
      await result.current.loadRecentLogs(50);
    });

    expect(AppBackend.GetRecentLogs).toHaveBeenCalledWith(50);
    expect(result.current.logs).toHaveLength(2);
    expect(result.current.logs[0].id).toBe('rec-1');
  });

  it('clears logs and calls ClearLogs on backend', async () => {
    const { result } = renderHook(() => useAppLogs());

    act(() => {
      result.current.addLog('Some log', 'info');
    });
    expect(result.current.logs).toHaveLength(1);

    await act(async () => {
      await result.current.clearAppLogs();
    });

    expect(AppBackend.ClearLogs).toHaveBeenCalledTimes(1);
    expect(result.current.logs).toHaveLength(0);
  });

  it('handles backend error in clearAppLogs', async () => {
    vi.mocked(AppBackend.ClearLogs).mockRejectedValueOnce(new Error('Clear error'));
    const { result } = renderHook(() => useAppLogs());

    await act(async () => {
      await result.current.clearAppLogs();
    });
    expect(result.current.logs).toHaveLength(0);
  });

  it('handles backend error in loadRecentLogs', async () => {
    vi.mocked(AppBackend.GetRecentLogs).mockRejectedValueOnce(new Error('Load error'));
    const { result } = renderHook(() => useAppLogs());

    await act(async () => {
      await result.current.loadRecentLogs(10);
    });
    expect(result.current.logs).toEqual([]);
  });

  it('handles console.warn and console.error overrides', () => {
    const { result } = renderHook(() => useAppLogs());
    const origLog = vi.fn();
    const origWarn = vi.fn();
    const origError = vi.fn();

    act(() => {
      result.current.setupConsoleOverrides(origLog, origWarn, origError);
    });

    act(() => {
      console.warn('Warning test');
      console.error('Error test');
    });

    expect(origWarn).toHaveBeenCalledWith('Warning test');
    expect(origError).toHaveBeenCalledWith('Error test');
    expect(result.current.logs.some(l => l.type === 'warn')).toBe(true);
    expect(result.current.logs.some(l => l.type === 'error')).toBe(true);
  });
});
