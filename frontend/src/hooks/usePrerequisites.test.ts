import { renderHook, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { usePrerequisites } from './usePrerequisites';
import * as AppBackend from '../../wailsjs/go/backend/App';

vi.mock('../../wailsjs/go/backend/App', () => ({
  GetPrerequisites: vi.fn(),
  DeleteVersion: vi.fn().mockResolvedValue(undefined),
  InstallPrerequisite: vi.fn().mockResolvedValue(undefined),
  InstallPluginModule: vi.fn().mockResolvedValue(undefined),
  UninstallPluginModule: vi.fn().mockResolvedValue(undefined),
  CancelDownload: vi.fn().mockResolvedValue(undefined),
  OpenPluginFolder: vi.fn().mockResolvedValue(undefined),
}));

describe('usePrerequisites hook', () => {
  const addToast = vi.fn();
  const addLog = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('initializes with empty prerequisites and progress', () => {
    const { result } = renderHook(() => usePrerequisites({ addToast, addLog }));
    expect(result.current.prerequisites).toEqual([]);
    expect(result.current.downloadProgress).toEqual({});
  });

  it('refreshes prerequisites and updates selected versions', async () => {
    (AppBackend.GetPrerequisites as any).mockResolvedValue([
      { name: 'PHP', version: '8.2', installedVers: ['8.1', '8.2'], activeVersion: '8.2' },
      { name: 'OpenSSL', version: '3.0', installedVers: ['3.0'] },
    ]);

    const { result } = renderHook(() => usePrerequisites({ addToast, addLog }));

    await act(async () => {
      await result.current.refreshPrerequisites();
    });

    expect(result.current.prerequisites).toHaveLength(2);
    expect(result.current.selectedVersions['PHP']).toBe('8.2');
    expect(result.current.selectedVersions['OpenSSL']).toBe('3.0');
  });

  it('updates download progress and adds toast on error', () => {
    const { result } = renderHook(() => usePrerequisites({ addToast, addLog }));

    act(() => {
      result.current.handleDownloadProgress({
        name: 'Node.js',
        status: 'Error: Connection timeout',
        percentage: 15,
      });
    });

    expect(result.current.downloadProgress['Node.js'].status).toBe('Error: Connection timeout');
    expect(addToast).toHaveBeenCalledWith('Installation Failed', expect.stringContaining('Error'), 'error');
    expect(addLog).toHaveBeenCalledWith(expect.stringContaining('Installation Failed'), 'error');
  });

  it('calls InstallPrerequisite with resolved task', async () => {
    const { result } = renderHook(() => usePrerequisites({ addToast, addLog }));

    const task = { name: 'PHP', version: '8.2', versionUrls: { '8.2': 'http://example.com' } };

    await act(async () => {
      await result.current.handleInstallSingle(task);
    });

    expect(AppBackend.InstallPrerequisite).toHaveBeenCalledWith(
      expect.objectContaining({
        name: 'PHP',
        version: '8.2',
        target: 'php/php-8.2',
      })
    );
  });
});
