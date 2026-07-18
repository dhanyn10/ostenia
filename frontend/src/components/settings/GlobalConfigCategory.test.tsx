import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import React from 'react';
import GlobalConfigCategory from './GlobalConfigCategory';

// Mock AppBackend
vi.mock('../../../wailsjs/go/backend/App', () => ({
  GetInstalledApps: vi.fn(),
  SelectServerRoot: vi.fn(),
  SelectWWWRoot: vi.fn(),
  SetDefaultEditor: vi.fn(),
  SelectDefaultEditor: vi.fn(),
}));

import * as AppBackendRaw from '../../../wailsjs/go/backend/App';
const AppBackend = AppBackendRaw as any;

describe('GlobalConfigCategory Component', () => {
  const mockConfig = {
    baseDir: '/path/to/ostenia',
    wwwRoot: '/path/to/www',
    defaultEditor: '/usr/bin/code',
  };

  const mockApps = [
    { name: 'VS Code', path: '/usr/bin/code' },
    { name: 'Sublime', path: '/usr/bin/subl' },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    AppBackend.GetInstalledApps.mockResolvedValue(mockApps);
  });

  it('renders initial config values and loads installed apps', async () => {
    render(<GlobalConfigCategory appConfig={mockConfig} initApp={vi.fn()} />);

    expect(screen.getByDisplayValue('/path/to/ostenia')).toBeInTheDocument();
    expect(screen.getByDisplayValue('/path/to/www')).toBeInTheDocument();

    await waitFor(() => {
      expect(AppBackend.GetInstalledApps).toHaveBeenCalled();
    });

    expect(screen.getByText('VS Code')).toBeInTheDocument();
    expect(screen.getByText('Sublime')).toBeInTheDocument();
  });

  it('handles empty apps response', async () => {
    AppBackend.GetInstalledApps.mockRejectedValue(new Error('Failed to load apps'));
    render(<GlobalConfigCategory appConfig={mockConfig} initApp={vi.fn()} />);

    await waitFor(() => {
      expect(AppBackend.GetInstalledApps).toHaveBeenCalled();
    });
  });

  it('handles SelectServerRoot (Browse Ostenia Home)', async () => {
    const initAppMock = vi.fn();
    AppBackend.SelectServerRoot.mockResolvedValue('/new/path');

    render(<GlobalConfigCategory appConfig={mockConfig} initApp={initAppMock} />);

    const browseButtons = screen.getAllByRole('button', { name: /browse/i });
    // First browse button is for Apps Location
    fireEvent.click(browseButtons[0]);

    expect(AppBackend.SelectServerRoot).toHaveBeenCalled();
    await waitFor(() => {
      expect(initAppMock).toHaveBeenCalled();
    });
  });

  it('handles SelectWWWRoot (Browse WWW Root)', async () => {
    const initAppMock = vi.fn();
    AppBackend.SelectWWWRoot.mockResolvedValue('/new/www');

    render(<GlobalConfigCategory appConfig={mockConfig} initApp={initAppMock} />);

    const browseButtons = screen.getAllByRole('button', { name: /browse/i });
    // Second browse button is for Server Root
    fireEvent.click(browseButtons[1]);

    expect(AppBackend.SelectWWWRoot).toHaveBeenCalled();
    await waitFor(() => {
      expect(initAppMock).toHaveBeenCalled();
    });
  });

  it('handles changing default editor selection', async () => {
    const initAppMock = vi.fn();
    AppBackend.SetDefaultEditor.mockResolvedValue(true);

    render(<GlobalConfigCategory appConfig={mockConfig} initApp={initAppMock} />);

    await waitFor(() => {
      expect(screen.getByText('VS Code')).toBeInTheDocument();
    });

    const select = screen.getByRole('combobox');
    fireEvent.change(select, { target: { value: '/usr/bin/subl' } });

    expect(AppBackend.SetDefaultEditor).toHaveBeenCalledWith('/usr/bin/subl');
    await waitFor(() => {
      expect(initAppMock).toHaveBeenCalled();
    });
  });

  it('handles custom browse for default editor', async () => {
    const initAppMock = vi.fn();
    AppBackend.SelectDefaultEditor.mockResolvedValue(true);

    render(<GlobalConfigCategory appConfig={mockConfig} initApp={initAppMock} />);

    const customBrowseBtn = screen.getByRole('button', { name: /custom browse/i });
    fireEvent.click(customBrowseBtn);

    expect(AppBackend.SelectDefaultEditor).toHaveBeenCalled();
    await waitFor(() => {
      expect(initAppMock).toHaveBeenCalled();
    });
  });

  it('handles removing the current custom editor', async () => {
    const initAppMock = vi.fn();
    AppBackend.SetDefaultEditor.mockResolvedValue(true);

    render(<GlobalConfigCategory appConfig={mockConfig} initApp={initAppMock} />);

    // The trash icon is inside a button. Since it has Trash2 icon, let's find the button.
    // The button has classes: "p-1.5 hover:bg-rose-500/10 rounded text-mui-grey-400 hover:text-rose-500 transition-colors"
    // Let's grab it via a query or querySelector.
    const trashBtn = screen.getByRole('button', { name: '' }); // and containing Trash2
    expect(trashBtn).toBeInTheDocument();

    fireEvent.click(trashBtn);

    expect(AppBackend.SetDefaultEditor).toHaveBeenCalledWith('');
    await waitFor(() => {
      expect(initAppMock).toHaveBeenCalled();
    });
  });
});
