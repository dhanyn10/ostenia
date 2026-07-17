import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import React from 'react';
import ProfileCategory from './ProfileCategory';

// Mock AppBackend
vi.mock('../../../wailsjs/go/backend/App', () => ({
  ExportProfile: vi.fn(),
  ImportProfile: vi.fn(),
}));

import * as AppBackendRaw from '../../../wailsjs/go/backend/App';
const AppBackend = AppBackendRaw as any;

describe('ProfileCategory Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('handles Export All action on click', () => {
    AppBackend.ExportProfile.mockResolvedValue(true);

    render(<ProfileCategory initApp={vi.fn()} />);

    const exportBtn = screen.getByRole('button', { name: /export all/i });
    fireEvent.click(exportBtn);

    expect(AppBackend.ExportProfile).toHaveBeenCalledWith(true, true);
  });

  it('handles Import Profile action on click', async () => {
    const initAppMock = vi.fn();
    AppBackend.ImportProfile.mockResolvedValue(true);

    render(<ProfileCategory initApp={initAppMock} />);

    const importBtn = screen.getByRole('button', { name: /import profile/i });
    fireEvent.click(importBtn);

    expect(AppBackend.ImportProfile).toHaveBeenCalled();
    await waitFor(() => {
      expect(initAppMock).toHaveBeenCalled();
    });
  });

  it('handles granular exports (Config Only, SSH Sessions Only) on click', () => {
    AppBackend.ExportProfile.mockResolvedValue(true);

    render(<ProfileCategory initApp={vi.fn()} />);

    const configBtn = screen.getByRole('button', { name: /config only/i });
    fireEvent.click(configBtn);
    expect(AppBackend.ExportProfile).toHaveBeenCalledWith(true, false);

    const sshBtn = screen.getByRole('button', { name: /ssh sessions only/i });
    fireEvent.click(sshBtn);
    expect(AppBackend.ExportProfile).toHaveBeenCalledWith(false, true);
  });

  it('handles exceptions in Import and Export actions gracefully', async () => {
    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    AppBackend.ImportProfile.mockRejectedValue(new Error('Import error'));
    AppBackend.ExportProfile.mockRejectedValue(new Error('Export error'));

    render(<ProfileCategory initApp={vi.fn()} />);

    const importBtn = screen.getByRole('button', { name: /import profile/i });
    fireEvent.click(importBtn);
    await waitFor(() => {
      expect(consoleErrorSpy).toHaveBeenCalled();
    });

    const exportBtn = screen.getByRole('button', { name: /export all/i });
    fireEvent.click(exportBtn);
    await waitFor(() => {
      expect(consoleErrorSpy).toHaveBeenCalled();
    });

    consoleErrorSpy.mockRestore();
  });
});
