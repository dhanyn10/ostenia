import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useAppLogs, formatLogArgs } from './useAppLogs';
import * as AppBackend from '../../wailsjs/go/backend/App';

vi.mock('../../wailsjs/go/backend/App', () => ({
  SaveLogToFile: vi.fn().mockResolvedValue(undefined),
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

  it('adds logs with appropriate prefixes', () => {
    const { result } = renderHook(() => useAppLogs());

    act(() => {
      result.current.addLog('General message', 'info');
    });

    expect(result.current.logs).toHaveLength(1);
    expect(result.current.logs[0].msg).toContain('[SYS] General message');
    expect(result.current.logs[0].type).toBe('info');

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
});
