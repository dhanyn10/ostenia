import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import SSHSessionView from './SSHSessionView';
import * as AppBackend from '../../wailsjs/go/backend/App';

// Mock xterm
vi.mock('@xterm/xterm', () => {
  return {
    Terminal: vi.fn().mockImplementation(() => ({
      loadAddon: vi.fn(),
      open: vi.fn(),
      dispose: vi.fn(),
      onData: vi.fn(),
      onTitleChange: vi.fn(),
      write: vi.fn(),
      options: { theme: {} },
      focus: vi.fn(),
    })),
  };
});

vi.mock('@xterm/addon-fit', () => {
  return {
    FitAddon: vi.fn().mockImplementation(() => ({
      fit: vi.fn(),
      proposeDimensions: vi.fn().mockReturnValue({ cols: 120, rows: 24 }),
    })),
  };
});

// Mock Wails backend functions
vi.mock('../../wailsjs/go/backend/App', () => ({
  ConnectSSH: vi.fn().mockResolvedValue(null),
  GetRemoteFiles: vi.fn().mockResolvedValue([
    { name: 'test.txt', isDir: false, size: 1024 },
    { name: 'folder', isDir: true, size: 0 }
  ]),
  GetRemoteCurrentPath: vi.fn().mockResolvedValue('/home/user'),
  SendSSHInput: vi.fn().mockResolvedValue(null),
  ResizeSSHTerminal: vi.fn().mockResolvedValue(null),
  DownloadRemoteFile: vi.fn().mockResolvedValue(null),
  UploadRemoteFile: vi.fn().mockResolvedValue(null),
  EditRemoteFile: vi.fn().mockResolvedValue(null),
  ExecuteSFTPAction: vi.fn().mockResolvedValue(null),
}));

describe('SSHSessionView Component', () => {
  const mockSession = {
    id: 'session-1',
    name: 'Test Server',
    host: '1.2.3.4',
    user: 'root'
  };

  const mockProps = {
    session: mockSession,
    onClose: vi.fn(),
    addToast: vi.fn(),
    isActive: true,
    theme: 'light',
    onOpenSettings: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    window.confirm = vi.fn().mockReturnValue(true);
    window.prompt = vi.fn().mockReturnValue('new-name');
  });

  it('renders and connects to SSH', async () => {
    render(<SSHSessionView {...mockProps} />);

    expect(AppBackend.ConnectSSH).toHaveBeenCalledWith(mockSession);

    await waitFor(() => {
      expect(screen.queryByText(/Connecting\.\.\./i)).not.toBeInTheDocument();
    });
  });

  it('lists remote files after connection', async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(AppBackend.GetRemoteFiles).toHaveBeenCalled();
    });

    expect(screen.getByText('test.txt')).toBeInTheDocument();
    expect(screen.getByText('folder')).toBeInTheDocument();
  });

  it('handles toolbar actions', async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
        expect(screen.queryByText(/Connecting\.\.\./i)).not.toBeInTheDocument();
    });

    const toggleExplorerBtn = screen.getByTitle(/Toggle Explorer/i);
    fireEvent.click(toggleExplorerBtn);
    expect(screen.queryByText('test.txt')).not.toBeInTheDocument();

    const closeBtn = screen.getByTitle(/Close/i);
    fireEvent.click(closeBtn);
    expect(mockProps.onClose).toHaveBeenCalled();
  });

  it('handles file deletion', async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText('test.txt')).toBeInTheDocument();
    });

    const fileItem = screen.getByText('test.txt');
    fireEvent.contextMenu(fileItem);

    const deleteBtn = screen.getByText(/Delete/i);
    fireEvent.click(deleteBtn);

    expect(window.confirm).toHaveBeenCalled();
    expect(AppBackend.ExecuteSFTPAction).toHaveBeenCalledWith(
        mockSession.id, 'delete', expect.stringContaining('test.txt'), ''
    );
  });

  it('handles directory navigation', async () => {
    render(<SSHSessionView {...mockProps} />);

    await waitFor(() => {
      expect(screen.getByText('folder')).toBeInTheDocument();
    });

    const folderItem = screen.getByText('folder');
    fireEvent.doubleClick(folderItem);

    expect(AppBackend.GetRemoteFiles).toHaveBeenCalledWith(mockSession.id, expect.stringContaining('folder'));
  });
});
