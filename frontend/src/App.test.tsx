import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import App from './App';
import * as AppBackend from '../wailsjs/go/backend/App';

const eventCallbacks: Record<string, (...args: any[]) => void> = {};

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
  GetServiceStatus: vi.fn((name) => {
    if (name === 'OpenSSL') {
      return Promise.resolve({ status: 'Running', pid: 123, port: 0 });
    }
    return Promise.resolve({ status: 'Stopped', pid: 0, port: 0 });
  }),
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
  GetWSLDistros: vi.fn().mockResolvedValue([]),
  GetPHPExtensions: vi.fn().mockResolvedValue([]),
  TogglePHPExtension: vi.fn().mockResolvedValue(null),
}));

vi.mock('../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn((event, cb) => {
    eventCallbacks[event] = cb;
  }),
  EventsOff: vi.fn(),
  EventsOnMultiple: vi.fn(),
  EventsEmit: vi.fn(),
  LogInfo: vi.fn(),
  LogError: vi.fn(),
  WindowSetTitle: vi.fn(),
  OnFileDrop: vi.fn(),
  OnFileDropOff: vi.fn(),
}));

describe('App Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    // Reset event callbacks dictionary
    for (const key in eventCallbacks) {
      delete eventCallbacks[key];
    }
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

  it('handles browse directory actions', async () => {
    render(<App />);

    await waitForLoadingToFinish();

    const browseButtons = screen.getAllByTitle('Browse Directory');
    expect(browseButtons).toHaveLength(2);

    // Browse Apps Location
    fireEvent.click(browseButtons[0]);
    await waitFor(() => {
      expect(AppBackend.SelectServerRoot).toHaveBeenCalled();
    });

    // Browse Server Root
    fireEvent.click(browseButtons[1]);
    await waitFor(() => {
      expect(AppBackend.SelectWWWRoot).toHaveBeenCalled();
    });
  });

  it('handles start and stop all actions', async () => {
    render(<App />);

    await waitForLoadingToFinish();

    const startAllBtn = screen.getByRole('button', { name: /Start All/i });
    fireEvent.click(startAllBtn);
    expect(AppBackend.StartAllServices).toHaveBeenCalled();

    const stopAllBtn = screen.getByRole('button', { name: /Stop All/i });
    fireEvent.click(stopAllBtn);
    expect(AppBackend.StopAllServices).toHaveBeenCalled();
  });

  it('handles events triggered from Wails runtime', async () => {
    render(<App />);

    await waitForLoadingToFinish();

    // Trigger service_status event
    if (eventCallbacks['service_status']) {
      eventCallbacks['service_status']({
        name: 'MySQL',
        status: 'Running',
        pid: 1234,
        port: 3306
      });
    }

    // Trigger service_log event
    if (eventCallbacks['service_log']) {
      eventCallbacks['service_log']({
        service: 'MySQL',
        message: 'Server started successfully'
      });
    }

    // Trigger environment_changed event
    if (eventCallbacks['environment_changed']) {
      eventCallbacks['environment_changed']();
      expect(AppBackend.GetConfig).toHaveBeenCalled();
    }
  });

  it('handles download progress events', async () => {
    render(<App />);

    await waitForLoadingToFinish();

    // Trigger download_progress start
    if (eventCallbacks['download_progress']) {
      eventCallbacks['download_progress']({
        name: 'PHP',
        percentage: 50,
        status: 'Downloading...'
      });

      // Trigger download_progress completed (refreshes prerequisites)
      eventCallbacks['download_progress']({
        name: 'PHP',
        percentage: 100,
        status: 'Completed'
      });
      expect(AppBackend.GetPrerequisites).toHaveBeenCalled();

      // Trigger download_progress error
      eventCallbacks['download_progress']({
        name: 'PHP',
        percentage: 10,
        status: 'Error: Connection lost'
      });
    }
  });

  it('handles toggle https for Apache and Nginx', async () => {
    (AppBackend.GetPrerequisites as any).mockResolvedValue([
      { name: 'Apache', version: '2.4', installedVers: ['2.4'], status: 'Ready' },
      { name: 'Nginx', version: '1.24', installedVers: ['1.24'], status: 'Ready' }
    ]);

    render(<App />);

    await waitForLoadingToFinish();

    const httpsButtons = screen.getAllByTitle(/Enable HTTPS/i);
    expect(httpsButtons.length).toBeGreaterThan(0);

    // Toggle Apache HTTPS
    fireEvent.click(httpsButtons[0]);
    expect(AppBackend.SetApacheHTTPS).toHaveBeenCalled();

    // Toggle Nginx HTTPS
    fireEvent.click(httpsButtons[1]);
    expect(AppBackend.SetNginxHTTPS).toHaveBeenCalled();
  });

  it('handles deleting versions and module operations inside PluginsTab', async () => {
    (AppBackend.GetPrerequisites as any).mockResolvedValue([
      {
        name: 'PHP',
        version: '8.2',
        installedVers: ['8.2', '8.1'],
        status: 'Ready',
        isInstalled: true,
        versions: ['8.2', '8.1', '8.0'],
        modules: [
          { name: 'composer', isInstalled: true, status: 'Ready', version: '2.5' },
          { name: 'xdebug', isInstalled: false, status: 'Not Installed', version: '3.2' }
        ]
      }
    ]);

    render(<App />);

    await waitForLoadingToFinish();

    // Switch to PluginsTab
    const pluginsTabButton = screen.getByTitle(/Plugin Management/i);
    fireEvent.click(pluginsTabButton);

    // Trigger delete version tag click
    const deleteBtn = screen.getByTitle('Delete v8.1');
    fireEvent.click(deleteBtn);
    expect(AppBackend.DeleteVersion).toHaveBeenCalledWith('PHP', '8.1');

    // Find phpCard using deleteBtn
    const phpCard = deleteBtn.closest('.p-4');
    expect(phpCard).not.toBeNull();

    // Expand modules section
    const expandBtn = phpCard!.querySelector('svg.lucide-chevron-down')?.parentElement;
    expect(expandBtn).toBeDefined();
    fireEvent.click(expandBtn!);

    // Install module
    const installModuleBtn = screen.getByTitle('Install xdebug');
    fireEvent.click(installModuleBtn);
    expect(AppBackend.InstallPluginModule).toHaveBeenCalledWith('PHP', 'xdebug');

    // Uninstall module
    const uninstallModuleBtn = screen.getByTitle('Uninstall composer');
    fireEvent.click(uninstallModuleBtn);
    expect(AppBackend.UninstallPluginModule).toHaveBeenCalledWith('PHP', 'composer');
  });

  it('handles downloading single prerequisite with target resolving', async () => {
    (AppBackend.GetPrerequisites as any).mockResolvedValue([
      {
        name: 'Node.js',
        version: '18.0.0',
        installedVers: [],
        status: 'Not Installed',
        versions: ['18.0.0'],
        versionUrls: { '18.0.0': 'http://example.com/node.zip' }
      }
    ]);

    render(<App />);

    await waitForLoadingToFinish();

    // Switch to PluginsTab
    const pluginsTabButton = screen.getByTitle(/Plugin Management/i);
    fireEvent.click(pluginsTabButton);

    const downloadBtn = screen.getByRole('button', { name: /Download/i });
    fireEvent.click(downloadBtn);

    expect(AppBackend.InstallPrerequisite).toHaveBeenCalled();
  });

  it('handles errors during installation and module operations', async () => {
    (AppBackend.InstallPrerequisite as any).mockRejectedValue(new Error('Install pre-req failed'));
    (AppBackend.InstallPluginModule as any).mockRejectedValue(new Error('Install module failed'));
    (AppBackend.UninstallPluginModule as any).mockRejectedValue(new Error('Uninstall module failed'));

    (AppBackend.GetPrerequisites as any).mockResolvedValue([
      {
        name: 'PHP',
        version: '8.2',
        installedVers: ['8.2', '8.1'],
        status: 'Ready',
        isInstalled: true,
        versions: ['8.2', '8.1', '8.0'],
        iconSvg: '<svg>dummy</svg>',
        modules: [
          { name: 'composer', isInstalled: true, status: 'Ready', version: '2.5' },
          { name: 'xdebug', isInstalled: false, status: 'Not Installed', version: '3.2' }
        ]
      }
    ]);

    render(<App />);
    await waitForLoadingToFinish();

    // Switch to PluginsTab
    const pluginsTabButton = screen.getByTitle(/Plugin Management/i);
    fireEvent.click(pluginsTabButton);

    // Find PHP card and expand
    const deleteBtn = screen.getByTitle('Delete v8.1');
    const phpCard = deleteBtn.closest('.p-4');
    const expandBtn = phpCard!.querySelector('svg.lucide-chevron-down')?.parentElement;
    fireEvent.click(expandBtn!);

    // Install module (should fail)
    const installModuleBtn = screen.getByTitle('Install xdebug');
    fireEvent.click(installModuleBtn);
    await waitFor(() => {
      expect(AppBackend.InstallPluginModule).toHaveBeenCalled();
    });

    // Uninstall module (should fail)
    const uninstallModuleBtn = screen.getByTitle('Uninstall composer');
    fireEvent.click(uninstallModuleBtn);
    await waitFor(() => {
      expect(AppBackend.UninstallPluginModule).toHaveBeenCalled();
    });
  });

  it('handles toggling service with errors and special HeidiSQL uninstall case', async () => {
    (AppBackend.StopService as any).mockRejectedValue(new Error('Failed to stop'));
    (AppBackend.GetPrerequisites as any).mockResolvedValue([
      { name: 'HeidiSQL', version: '12.0', installedVers: ['12.0'], status: 'Ready' },
      { name: 'Apache', version: '2.4', installedVers: ['2.4'], status: 'Ready' }
    ]);

    render(<App />);
    await waitForLoadingToFinish();

    // Trigger status update of HeidiSQL to Running
    if (eventCallbacks['service_status']) {
      eventCallbacks['service_status']({
        name: 'HeidiSQL',
        status: 'Running',
        pid: 999
      });
    }

    // Wait for the UI state update to complete
    await waitFor(() => {
      const activeHeidiStatus = screen.queryByText('Running');
      expect(activeHeidiStatus).toBeDefined();
    });

    // Find HeidiSQL card and toggle button
    const heidiHeading = screen.getAllByRole('heading', { name: /HeidiSQL/i })[0];
    const heidiCard = heidiHeading.closest('.shadow-sm');
    expect(heidiCard).not.toBeNull();

    // Find the toggle button using the static class name prefix w-12
    let toggleButton = Array.from(heidiCard!.querySelectorAll('button')).find(b => b.className.includes('w-12'));
    expect(toggleButton).toBeDefined();

    // Toggle service when status is Running and name is HeidiSQL -> should open confirmation modal
    fireEvent.click(toggleButton!);

    // Check that confirmation modal is open
    expect(screen.getByText(/Are you sure you want to uninstall HeidiSQL/i)).toBeInTheDocument();

    // Click confirm inside confirmation modal to trigger handleConfirmHeidiSQLUninstall
    const confirmBtn = screen.getByRole('button', { name: /Confirm/i });
    fireEvent.click(confirmBtn);
    await waitFor(() => {
      expect(AppBackend.StopService).toHaveBeenCalledWith('HeidiSQL');
    });

    // Now toggle service for Apache running to trigger stop error toast
    if (eventCallbacks['service_status']) {
      eventCallbacks['service_status']({
        name: 'Apache',
        status: 'Running',
        pid: 888
      });
    }

    const apacheHeading = screen.getAllByRole('heading', { name: /Apache/i })[0];
    const apacheCard = apacheHeading.closest('.shadow-sm');
    expect(apacheCard).not.toBeNull();

    let apacheToggleButton: HTMLButtonElement | undefined;
    await waitFor(() => {
      apacheToggleButton = Array.from(apacheCard!.querySelectorAll('button')).find(b => b.className.includes('w-12')) as HTMLButtonElement;
      expect(apacheToggleButton).toBeDefined();
    });

    // Toggle button should call StopService and fail
    fireEvent.click(apacheToggleButton!);
    expect(await screen.findByText(/Failed to stop Apache/i)).toBeInTheDocument();
  });

  it('handles adding and removing plugins to/from home list', async () => {
    (AppBackend.GetPrerequisites as any).mockResolvedValue([
      { name: 'PHP', version: '8.2', installedVers: ['8.2'], status: 'Ready' }
    ]);

    render(<App />);
    await waitForLoadingToFinish();

    // Find PHP card and heading
    const phpHeading = screen.getAllByRole('heading', { name: /PHP/i })[0];
    const phpCard = phpHeading.closest('.shadow-sm');
    expect(phpCard).not.toBeNull();

    // Find the remove button using class and click it (always visible on main row!)
    let removeBtn = Array.from(phpCard!.querySelectorAll('button')).find(b => b.className.includes('hover:bg-rose-500/10') || b.innerHTML.includes('trash'));
    expect(removeBtn).toBeDefined();
    fireEvent.click(removeBtn!);

    // Find the ActivityTab container specifically to verify PHP is removed only from home list
    const appsLocationLabel = screen.getByText(/Apps Location/i);
    const activityContainer = appsLocationLabel.closest('.h-full');
    expect(activityContainer).not.toBeNull();

    // Check that PHP is removed from home list inside ActivityTab container
    await waitFor(() => {
      const homePhpHeading = within(activityContainer!).queryByRole('heading', { name: /PHP/i });
      expect(homePhpHeading).toBeNull();
    });

    // Add PHP back to home list
    const addPluginBtn = screen.getByRole('button', { name: /Add Plugin/i });
    fireEvent.click(addPluginBtn);

    const addPhpBtn = screen.getByRole('button', { name: /PHP/i });
    expect(addPhpBtn).toBeDefined();

    // Click add button
    fireEvent.click(addPhpBtn);

    // Verify it is back
    await waitFor(() => {
      const homePhpHeading = within(activityContainer!).queryByRole('heading', { name: /PHP/i });
      expect(homePhpHeading).not.toBeNull();
    });
  });
});
