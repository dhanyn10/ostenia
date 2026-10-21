import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { useAppConfig } from './useAppConfig';
import * as AppBackend from '../../wailsjs/go/backend/App';

vi.mock('../../wailsjs/go/backend/App', () => ({
  GetConfig: vi.fn(),
  SelectServerRoot: vi.fn(),
  SelectWWWRoot: vi.fn(),
  SetApacheHTTPS: vi.fn().mockResolvedValue(undefined),
  SetNginxHTTPS: vi.fn().mockResolvedValue(undefined),
  OpenServerRootFolder: vi.fn().mockResolvedValue(undefined),
  OpenAppsLocationFolder: vi.fn().mockResolvedValue(undefined),
}));

describe('useAppConfig hook', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('loads config and updates state', async () => {
    (AppBackend.GetConfig as any).mockResolvedValue({
      wwwRoot: '/test/www',
      baseDir: '/test/apps',
      apacheHttps: true,
      nginxHttps: false,
      defaultEditor: 'vscode',
    });

    const { result } = renderHook(() => useAppConfig());

    await act(async () => {
      await result.current.loadConfig();
    });

    expect(result.current.serverRootState).toBe('/test/www');
    expect(result.current.appsLocation).toBe('/test/apps');
    expect(result.current.apacheHttps).toBe(true);
    expect(result.current.nginxHttps).toBe(false);
    expect(result.current.defaultEditor).toBe('vscode');
  });

  it('handles browsing server and apps root and triggers onConfigChanged', async () => {
    const onConfigChanged = vi.fn();
    (AppBackend.SelectServerRoot as any).mockResolvedValue('/selected/apps');
    (AppBackend.SelectWWWRoot as any).mockResolvedValue('/selected/www');

    const { result } = renderHook(() => useAppConfig({ onConfigChanged }));

    await act(async () => {
      await result.current.handleBrowseAppsLocation();
    });
    expect(result.current.appsLocation).toBe('/selected/apps');
    expect(onConfigChanged).toHaveBeenCalledTimes(1);

    await act(async () => {
      await result.current.handleBrowseServerRoot();
    });
    expect(result.current.serverRootState).toBe('/selected/www');
    expect(onConfigChanged).toHaveBeenCalledTimes(2);
  });

  it('toggles HTTPS for Apache and Nginx', async () => {
    const { result } = renderHook(() => useAppConfig());

    await act(async () => {
      await result.current.handleToggleHttps('Apache');
    });
    expect(result.current.apacheHttps).toBe(true);
    expect(AppBackend.SetApacheHTTPS).toHaveBeenCalledWith(true);

    await act(async () => {
      await result.current.handleToggleHttps('Nginx');
    });
    expect(result.current.nginxHttps).toBe(true);
    expect(AppBackend.SetNginxHTTPS).toHaveBeenCalledWith(true);
  });
});
