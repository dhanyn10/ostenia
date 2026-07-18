import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import ActivityTab from './ActivityTab';
import React from 'react';
import * as AppBackend from '../../wailsjs/go/backend/App';

// Mock Wails backend functions
vi.mock('../../wailsjs/go/backend/App', () => ({
  GetPHPExtensions: vi.fn().mockResolvedValue([{ name: 'mysqli', enabled: true }]),
  TogglePHPExtension: vi.fn().mockResolvedValue(null),
  SwitchServiceVersion: vi.fn().mockResolvedValue(null),
  OpenServiceTerminal: vi.fn().mockResolvedValue(null),
}));

describe('ActivityTab Component', () => {
  const defaultProps = {
    serverRoot: '/server/root',
    appsLocation: '/apps/location',
    handleBrowseAppsLocation: vi.fn(),
    handleBrowseServerRoot: vi.fn(),
    isAddingPlugin: false,
    setIsAddingPlugin: vi.fn(),
    prerequisites: [{ name: 'PHP', version: '8.2', installedVers: ['8.2.0'], status: 'Ready' }],
    services: [{ name: 'PHP', status: 'Running', port: 9000 }],
    handleAddToHome: vi.fn(),
    renderIcon: () => <span>Icon</span>,
    handleToggleService: vi.fn(),
    handleRemoveFromHome: vi.fn(),
    setActiveTab: vi.fn(),
    handleOpenPluginFolder: vi.fn(),
    handleOpenServerRootFolder: vi.fn(),
    handleOpenAppsLocationFolder: vi.fn(),
    apacheHttps: false,
    nginxHttps: false,
    handleToggleHttps: vi.fn(),
    isLoading: false,
    transitioningServices: new Set(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders correctly with services', () => {
    render(<ActivityTab {...defaultProps} />);
    expect(screen.getByText('Apps Location')).toBeInTheDocument();
    expect(screen.getByText('Server Root Directory')).toBeInTheDocument();
    expect(screen.getByText('PHP')).toBeInTheDocument();
  });

  it('calls browse handlers when buttons are clicked', () => {
    render(<ActivityTab {...defaultProps} />);

    const browseButtons = screen.getAllByTitle('Browse Directory');
    fireEvent.click(browseButtons[0]);
    expect(defaultProps.handleBrowseAppsLocation).toHaveBeenCalled();

    fireEvent.click(browseButtons[1]);
    expect(defaultProps.handleBrowseServerRoot).toHaveBeenCalled();
  });

  it('calls open folder handlers', () => {
    render(<ActivityTab {...defaultProps} />);

    const openButtons = screen.getAllByTitle('Open in Explorer');
    fireEvent.click(openButtons[0]);
    expect(defaultProps.handleOpenAppsLocationFolder).toHaveBeenCalled();

    fireEvent.click(openButtons[1]);
    expect(defaultProps.handleOpenServerRootFolder).toHaveBeenCalled();
  });

  it('fetches PHP extensions when PHP accordion is expanded', async () => {
    render(<ActivityTab {...defaultProps} />);

    const phpItem = screen.getByText('PHP');
    fireEvent.click(phpItem);

    await waitFor(() => {
      expect(AppBackend.GetPHPExtensions).toHaveBeenCalled();
    });
  });

  it('shows loading state', () => {
    render(<ActivityTab {...defaultProps} isLoading={true} />);
    expect(screen.getByText(/Scanning Plugins/i)).toBeInTheDocument();
  });

  it('shows empty state when no services', () => {
    render(<ActivityTab {...defaultProps} services={[]} />);
    expect(screen.getByText('No services active')).toBeInTheDocument();
  });
});
