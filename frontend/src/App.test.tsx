import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import App from './App';
import * as AppBackend from '../wailsjs/go/backend/App';

// Mock Wails backend functions
vi.mock('../wailsjs/go/backend/App', () => ({
  GetPrerequisites: vi.fn().mockResolvedValue([]),
  GetConfig: vi.fn().mockResolvedValue({
    wwwRoot: '/var/www',
    baseDir: '/opt/ostenia',
    apacheHttps: false,
    nginxHttps: false,
    defaultEditor: 'code'
  }),
  GetServiceStatus: vi.fn().mockResolvedValue({ status: 'Stopped', pid: 0, port: 0 }),
  StartAllServices: vi.fn().mockResolvedValue(null),
  StopAllServices: vi.fn().mockResolvedValue(null),
  OpenTerminal: vi.fn().mockResolvedValue(null),
  OpenServerRootFolder: vi.fn().mockResolvedValue(null),
  OpenAppsLocationFolder: vi.fn().mockResolvedValue(null),
  CancelDownload: vi.fn().mockResolvedValue(null),
  OpenPluginFolder: vi.fn().mockResolvedValue(null),
  DeleteVersion: vi.fn().mockResolvedValue(null),
  StartService: vi.fn().mockResolvedValue(null),
  StopService: vi.fn().mockResolvedValue(null),
  SetApacheHTTPS: vi.fn().mockResolvedValue(null),
  SetNginxHTTPS: vi.fn().mockResolvedValue(null),
  InstallPrerequisite: vi.fn().mockResolvedValue(null),
  InstallPluginModule: vi.fn().mockResolvedValue(null),
  UninstallPluginModule: vi.fn().mockResolvedValue(null),
  SelectServerRoot: vi.fn().mockResolvedValue('/new/apps'),
  SelectWWWRoot: vi.fn().mockResolvedValue('/new/www'),
  Maximize: vi.fn(),
  Unmaximize: vi.fn(),
  Minimize: vi.fn(),
  Close: vi.fn(),
  ToggleDevTools: vi.fn(),
  GetProxyApps: vi.fn().mockResolvedValue([]),
  GetSSHSessions: vi.fn().mockResolvedValue([]),
}));

vi.mock('../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
  EventsOff: vi.fn(),
  EventsOnMultiple: vi.fn(),
  EventsEmit: vi.fn(),
  LogInfo: vi.fn(),
  LogError: vi.fn(),
  WindowSetTitle: vi.fn(),
}));

describe('App Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
  });

  const waitForLoadingToFinish = async () => {
    await waitFor(() => {
      expect(screen.queryByText(/Scanning Plugins/i)).not.toBeInTheDocument();
    }, { timeout: 2000 });
  };

  it('renders and initializes with data', async () => {
    render(<App />);

    await waitForLoadingToFinish();

    expect(AppBackend.GetConfig).toHaveBeenCalled();
    expect(screen.getByRole('heading', { name: /Activity Center/i })).toBeInTheDocument();
  });

  it('switches tabs correctly', async () => {
    (AppBackend.GetPrerequisites as any).mockResolvedValue([
      { name: 'PHP', version: '8.2', isInstalled: true }
    ]);

    render(<App />);

    await waitForLoadingToFinish();

    const pluginsTabButton = screen.getByTitle(/Plugin Management/i);
    fireEvent.click(pluginsTabButton);

    expect(screen.getByRole('heading', { name: /Plugin Management/i })).toBeInTheDocument();
    expect(screen.getAllByText('PHP').length).toBeGreaterThan(0);
  });

  it('toggles theme', async () => {
    render(<App />);

    await waitForLoadingToFinish();

    const themeToggle = screen.getByTitle(/Switch to Dark Mode/i);
    expect(document.documentElement.classList.contains('dark')).toBe(false);

    fireEvent.click(themeToggle);

    expect(document.documentElement.classList.contains('dark')).toBe(true);
    expect(localStorage.getItem('theme')).toBe('dark');
  });

  it('opens and closes settings modal', async () => {
    render(<App />);

    await waitForLoadingToFinish();

    const settingsMenuButton = screen.getByRole('button', { name: /^Settings$/ });
    fireEvent.click(settingsMenuButton);

    const profileButton = screen.getByRole('button', { name: /^Profile$/ });
    fireEvent.click(profileButton);

    expect(screen.getByRole('heading', { name: /^Settings$/i })).toBeInTheDocument();

    const closeButton = screen.getByRole('button', { name: /^Close$/i });
    fireEvent.click(closeButton);

    await waitFor(() => {
      expect(screen.queryByRole('heading', { name: /^Settings$/i })).not.toBeInTheDocument();
    });
  });

  it('handles service start/stop', async () => {
    (AppBackend.GetPrerequisites as any).mockResolvedValue([
      { name: 'Apache', version: '2.4', installedVers: ['2.4'], status: 'Ready' }
    ]);

    render(<App />);

    await waitForLoadingToFinish();

    const buttons = screen.getAllByRole('button');
    const toggleButton = buttons.find(b => b.className.includes('w-12 h-6'));

    expect(toggleButton).toBeInTheDocument();

    fireEvent.click(toggleButton!);

    expect(AppBackend.StartService).toHaveBeenCalled();
  });

  it('adds a toast message on error', async () => {
     (AppBackend.StartService as any).mockRejectedValue(new Error('Failed to start'));
     (AppBackend.GetPrerequisites as any).mockResolvedValue([
        { name: 'Apache', version: '2.4', installedVers: ['2.4'], status: 'Ready' }
      ]);

     render(<App />);

     await waitForLoadingToFinish();

     const buttons = screen.getAllByRole('button');
     const toggleButton = buttons.find(b => b.className.includes('w-12 h-6'));

     fireEvent.click(toggleButton!);

     expect(await screen.findByText(/Failed to start Apache/i)).toBeInTheDocument();
  });
});
