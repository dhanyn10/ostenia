import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import React from 'react';
import SSHMonitorCategory from './SSHMonitorCategory';

describe('SSHMonitorCategory Component', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
  });

  it('renders with default localstorage values', () => {
    render(<SSHMonitorCategory />);

    const checkbox = screen.getByLabelText('Enable Real-Time Resource Monitoring');
    expect(checkbox).toBeInTheDocument();
    expect(checkbox).toBeChecked();

    const intervalInput = screen.getByLabelText('Refresh Interval (seconds)');
    expect(intervalInput).toBeInTheDocument();
    expect(intervalInput).toHaveValue(3);
  });

  it('handles toggle monitoring state', () => {
    render(<SSHMonitorCategory />);

    const checkbox = screen.getByLabelText('Enable Real-Time Resource Monitoring');
    fireEvent.click(checkbox);

    expect(checkbox).not.toBeChecked();
    expect(localStorage.getItem('ostenia_ssh_monitor_enabled')).toBe('false');
  });

  it('handles custom interval values and changes refresh interval correctly', () => {
    render(<SSHMonitorCategory />);

    const intervalInput = screen.getByLabelText('Refresh Interval (seconds)');
    fireEvent.change(intervalInput, { target: { value: '10' } });

    expect(intervalInput).toHaveValue(10);
    expect(localStorage.getItem('ostenia_ssh_monitor_interval')).toBe('10');
  });
});
