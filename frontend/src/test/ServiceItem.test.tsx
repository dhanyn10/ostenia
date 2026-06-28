import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import ServiceItem from '../components/ServiceItem';
import React from 'react';

describe('ServiceItem', () => {
  const defaultProps = {
    service: { name: 'Apache', status: 'Stopped', port: 80, ports: [] },
    task: { installedVers: ['2.4.54'] },
    isExpanded: false,
    onToggleAccordion: vi.fn(),
    renderIcon: () => <div data-testid="icon" />,
    handleToggleService: vi.fn(),
    handleRemoveFromHome: vi.fn(),
    handleSwitchVersion: vi.fn(),
    handleOpenLocalTerminal: vi.fn(),
    handleToggleHttps: vi.fn(),
    openTerminalDropdown: null,
    setOpenTerminalDropdown: vi.fn(),
    setIsModalOpen: vi.fn(),
    apacheHttps: false,
    nginxHttps: false,
    isOpenSslEnabled: true,
    setActiveTab: vi.fn(),
    handleOpenPluginFolder: vi.fn(),
  };

  it('renders service name', () => {
    render(<ServiceItem {...defaultProps} />);
    // Use getAllByText if there are multiple or be more specific
    expect(screen.getByRole('heading', { name: /APACHE/i })).toBeDefined();
  });

  it('renders status', () => {
    render(<ServiceItem {...defaultProps} />);
    expect(screen.getByText('Stopped')).toBeDefined();
  });

  it('calls onToggleAccordion when clicked', () => {
    render(<ServiceItem {...defaultProps} />);
    const accordionButton = screen.getByRole('heading', { name: /APACHE/i }).closest('button');
    fireEvent.click(accordionButton!);
    expect(defaultProps.onToggleAccordion).toHaveBeenCalledWith('Apache', true);
  });

  it('calls handleToggleService when toggle button is clicked', () => {
    render(<ServiceItem {...defaultProps} />);
    // The toggle button is the one before the trash button
    const buttons = screen.getAllByRole('button');
    // MainActions has toggle and trash.
    // Index 0 is the accordion button.
    // Index 1 should be the toggle.
    // Index 2 should be the trash.
    fireEvent.click(buttons[1]);
    expect(defaultProps.handleToggleService).toHaveBeenCalledWith('Apache', 'Stopped');
  });

  it('shows version switcher for PHP', () => {
    const phpProps = {
      ...defaultProps,
      service: { name: 'PHP', status: 'Running', activeVersion: '8.2.0' },
      task: { installedVers: ['8.2.0', '8.3.0'] }
    };
    render(<ServiceItem {...phpProps} />);
    expect(screen.getByText('8.2.0')).toBeDefined();
    expect(screen.getByText('8.3.0')).toBeDefined();
  });
});
