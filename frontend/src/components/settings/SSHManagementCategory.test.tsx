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

    await act(async () => {
      render(<SSHManagementCategory />);
    });

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

    await act(async () => {
      render(<SSHManagementCategory />);
    });

    await waitFor(() => {
      expect(AppBackend.GetSSHSessions).toHaveBeenCalled();
    });

    const toggleBtn = screen.getByRole('button', { name: /show passwords/i });

    await act(async () => {
      fireEvent.click(toggleBtn);
    });

    // Should now show actual password and passphrase, and label should change to "Mask Passwords"
    expect(screen.getByRole('button', { name: /mask passwords/i })).toBeInTheDocument();

    const preElement = screen.getByText((content, element) => {
      return element?.tagName.toLowerCase() === 'pre' && content.includes('"password": "supersecretpassword"');
    });
    expect(preElement).toBeInTheDocument();
    expect(preElement.textContent).toContain('"passphrase": "supersecretpassphrase"');
  });

  it('handles empty response or failure in loading sessions', async () => {
    AppBackend.GetSSHSessions.mockRejectedValue(new Error('Failed to load sessions'));

    await act(async () => {
      render(<SSHManagementCategory />);
    });

    await waitFor(() => {
      expect(AppBackend.GetSSHSessions).toHaveBeenCalled();
    });

    // Should show empty array empty state JSON
    const preElement = screen.getByText((content, element) => {
      return element?.tagName.toLowerCase() === 'pre' && content.includes('[]');
    });
    expect(preElement).toBeInTheDocument();
  });
});
