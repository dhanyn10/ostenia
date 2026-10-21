import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useServices, fetchServiceStatus } from './useServices';
import * as AppBackend from '../../wailsjs/go/backend/App';

vi.mock('../../wailsjs/go/backend/App', () => ({
  GetServiceStatus: vi.fn(),
  StartService: vi.fn(),
  StopService: vi.fn(),
  StartAllServices: vi.fn(),
  StopAllServices: vi.fn(),
}));

describe('useServices hook', () => {
  const addToast = vi.fn();
  const onRequireConfirmHeidiSQL = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('initializes with default services list', () => {
    const { result } = renderHook(() => useServices({ addToast, onRequireConfirmHeidiSQL }));
    expect(result.current.services.length).toBeGreaterThanOrEqual(8);
    expect(result.current.services.some(s => s.name === 'Apache')).toBe(true);
  });

  it('updates service status via handleServiceStatus', () => {
    const { result } = renderHook(() => useServices({ addToast, onRequireConfirmHeidiSQL }));

    act(() => {
      result.current.handleServiceStatus({ name: 'Apache', status: 'Running', pid: 1234, port: 80 });
    });

    const apache = result.current.services.find(s => s.name === 'Apache');
    expect(apache?.status).toBe('Running');
    expect(apache?.pid).toBe(1234);
  });

  it('handles adding and removing from home', () => {
    const { result } = renderHook(() => useServices({ addToast, onRequireConfirmHeidiSQL }));

    act(() => {
      result.current.handleAddToHome({ name: 'PostgreSQL' });
    });
    expect(result.current.services.some(s => s.name === 'PostgreSQL')).toBe(true);

    act(() => {
      result.current.handleRemoveFromHome('PostgreSQL');
    });
    expect(result.current.services.some(s => s.name === 'PostgreSQL')).toBe(false);
  });

  it('triggers onRequireConfirmHeidiSQL when stopping running HeidiSQL', async () => {
    const { result } = renderHook(() => useServices({ addToast, onRequireConfirmHeidiSQL }));

    await act(async () => {
      await result.current.handleToggleService('HeidiSQL', 'Running');
    });

    expect(onRequireConfirmHeidiSQL).toHaveBeenCalledWith('HeidiSQL', expect.any(Function));
  });

  it('fetches service status safely', async () => {
    (AppBackend.GetServiceStatus as any).mockResolvedValue({ status: 'Running', pid: 555 });

    const status = await fetchServiceStatus({
      name: 'Nginx',
      status: 'Stopped',
      pid: 0,
      port: 0,
      ports: [],
      activeVersion: '',
    });

    expect(status.status).toBe('Running');
    expect(status.pid).toBe(555);
  });
});
