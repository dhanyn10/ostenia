import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import React from 'react';
import SSHManagementCategory from './SSHManagementCategory';

// Mock AppBackend
vi.mock('../../../wailsjs/go/backend/App', () => ({
  GetSSHSessions: vi.fn(),
}));

import * as AppBackendRaw from '../../../wailsjs/go/backend/App';
const AppBackend = AppBackendRaw as any;

describe('SSHManagementCategory Component', () => {
  const mockSessions = [
    {
      id: 'session-1',
      name: 'Server 1',
      host: '1.2.3.4',
      user: 'root',
      password: 'supersecretpassword',
      passphrase: 'supersecretpassphrase',
    },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders masked passwords by default', async () => {
    AppBackend.GetSSHSessions.mockResolvedValue(mockSessions);

    render(<SSHManagementCategory />);

    await waitFor(() => {
      expect(AppBackend.GetSSHSessions).toHaveBeenCalled();
    });

    // Check that sensitive fields are masked with "***"
    const preElement = screen.getByText((content, element) => {
      return element?.tagName.toLowerCase() === 'pre' && content.includes('"password": "***"');
    });
    expect(preElement).toBeInTheDocument();
    expect(preElement.textContent).toContain('"passphrase": "***"');
    expect(preElement.textContent).not.toContain('supersecretpassword');
  });

  it('unmasks passwords when toggle button is clicked', async () => {
    AppBackend.GetSSHSessions.mockResolvedValue(mockSessions);

    render(<SSHManagementCategory />);

    await waitFor(() => {
      expect(AppBackend.GetSSHSessions).toHaveBeenCalled();
    });

    const toggleBtn = screen.getByRole('button', { name: /show passwords/i });

    fireEvent.click(toggleBtn);

    // Should now show actual password and passphrase, and label should change to "Mask Passwords"
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /mask passwords/i })).toBeInTheDocument();
    });

    const preElement = screen.getByText((content, element) => {
      return element?.tagName.toLowerCase() === 'pre' && content.includes('"password": "supersecretpassword"');
    });
    expect(preElement).toBeInTheDocument();
    expect(preElement.textContent).toContain('"passphrase": "supersecretpassphrase"');
  });

  it('handles empty response or failure in loading sessions', async () => {
    AppBackend.GetSSHSessions.mockRejectedValue(new Error('Failed to load sessions'));

    render(<SSHManagementCategory />);

    await waitFor(() => {
      expect(AppBackend.GetSSHSessions).toHaveBeenCalled();
    });

    // Should show empty array empty state JSON
    const preElement = screen.getByText((content, element) => {
      return element?.tagName.toLowerCase() === 'pre' && content.includes('[]');
    });
    expect(preElement).toBeInTheDocument();
  });

  it('renders and toggles zoom setting in local storage and dispatches event', async () => {
    AppBackend.GetSSHSessions.mockResolvedValue([]);
    localStorage.clear();

    const dispatchSpy = vi.spyOn(window, 'dispatchEvent');

    render(<SSHManagementCategory />);

    // Verify checkbox is rendered and checked by default
    const checkbox = screen.getByRole('checkbox', { name: /Enable Terminal Zoom/i }) as HTMLInputElement;
    expect(checkbox).toBeInTheDocument();
    expect(checkbox.checked).toBe(true);

    // Toggle setting off
    fireEvent.click(checkbox);
    await waitFor(() => {
      expect(checkbox.checked).toBe(false);
    });
    expect(localStorage.getItem('ostenia_ssh_zoom_enabled')).toBe('false');
    expect(dispatchSpy).toHaveBeenCalled();

    // Toggle setting back on
    fireEvent.click(checkbox);
    await waitFor(() => {
      expect(checkbox.checked).toBe(true);
    });
    expect(localStorage.getItem('ostenia_ssh_zoom_enabled')).toBe('true');
  });

  it('renders and toggles monitoring settings in local storage and dispatches events', async () => {
    AppBackend.GetSSHSessions.mockResolvedValue([]);
    localStorage.clear();

    const dispatchSpy = vi.spyOn(window, 'dispatchEvent');

    render(<SSHManagementCategory />);

    // 1. Enable Resource Tracking checkbox
    const trackingCheckbox = screen.getByRole('checkbox', { name: /Enable Resource Tracking/i }) as HTMLInputElement;
    expect(trackingCheckbox).toBeInTheDocument();
    expect(trackingCheckbox.checked).toBe(true); // default true

    fireEvent.click(trackingCheckbox);
    await waitFor(() => {
      expect(trackingCheckbox.checked).toBe(false);
    });
    expect(localStorage.getItem('ostenia_ssh_monitor_enabled')).toBe('false');
    expect(dispatchSpy).toHaveBeenCalled();

    // 2. Interval input
    fireEvent.click(trackingCheckbox); // re-enable so input is active
    await waitFor(() => {
      expect(trackingCheckbox.checked).toBe(true);
    });
    const intervalInput = screen.getByLabelText(/Interval/i) as HTMLInputElement;
    expect(intervalInput).toBeInTheDocument();
    expect(intervalInput.disabled).toBe(false);

    fireEvent.change(intervalInput, { target: { value: '5' } });
    expect(localStorage.getItem('ostenia_ssh_monitor_interval')).toBe('5');

    // Interval should be clamped
    fireEvent.change(intervalInput, { target: { value: '100' } });
    expect(localStorage.getItem('ostenia_ssh_monitor_interval')).toBe('60');

    // 3. Display style select
    const displaySelect = screen.getByLabelText(/Display Style/i) as HTMLSelectElement;
    expect(displaySelect).toBeInTheDocument();
    expect(displaySelect.disabled).toBe(false);

    fireEvent.change(displaySelect, { target: { value: 'always' } });
    expect(localStorage.getItem('ostenia_ssh_monitor_display_mode')).toBe('always');
  });
});
