import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useToast } from '@/hooks/useToast';

describe('useToast hook', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('starts with empty toasts array', () => {
    const { result } = renderHook(() => useToast());
    expect(result.current.toasts).toEqual([]);
  });

  it('adds and removes toasts', () => {
    const { result } = renderHook(() => useToast());

    act(() => {
      result.current.addToast('Test Title', 'Test Message', 'success');
    });

    expect(result.current.toasts).toHaveLength(1);
    expect(result.current.toasts[0].title).toBe('Test Title');
    expect(result.current.toasts[0].message).toBe('Test Message');
    expect(result.current.toasts[0].type).toBe('success');

    const toastId = result.current.toasts[0].id;

    act(() => {
      result.current.removeToast(toastId);
    });

    expect(result.current.toasts).toHaveLength(0);
  });

  it('auto-dismisses toasts after 5000ms', () => {
    const { result } = renderHook(() => useToast());

    act(() => {
      result.current.addToast('Notice', 'Auto dismiss message', 'info');
    });

    expect(result.current.toasts).toHaveLength(1);

    act(() => {
      vi.advanceTimersByTime(5000);
    });

    expect(result.current.toasts).toHaveLength(0);
  });
});
