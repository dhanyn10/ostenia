import { render, screen } from '@testing-library/react';
import { describe, it, expect } from 'vitest';
import LogViewer from './LogViewer';
import React from 'react';

describe('LogViewer Component', () => {
  it('renders empty state when no logs are provided', () => {
    render(<LogViewer logs={[]} />);
    expect(screen.getByText(/No activity recorded yet/i)).toBeInTheDocument();
  });

  it('renders logs correctly', () => {
    const logs = [
      { id: 1, time: '10:00:00', msg: 'Normal message' },
      { id: 2, time: '10:00:01', msg: 'Error: something failed' },
      { id: 3, time: '10:00:02', msg: 'Ready success' },
      { id: 4, time: '10:00:03', msg: '[WRN] warning' },
    ];
    render(<LogViewer logs={logs} />);

    expect(screen.getByText('Normal message')).toBeInTheDocument();
    expect(screen.getByText('Error: something failed')).toBeInTheDocument();
    expect(screen.getByText('Ready success')).toBeInTheDocument();
    expect(screen.getByText('[WRN] warning')).toBeInTheDocument();
  });

  it('applies correct color classes based on log content', () => {
    const logs = [
      { id: 1, time: '10:00:01', msg: 'Error: something failed' },
      { id: 2, time: '10:00:02', msg: 'Ready success' },
      { id: 3, time: '10:00:03', msg: '[WRN] warning' },
    ];
    render(<LogViewer logs={logs} />);

    const errorLog = screen.getByText('Error: something failed');
    expect(errorLog).toHaveClass('text-rose-500');

    const successLog = screen.getByText('Ready success');
    expect(successLog).toHaveClass('text-emerald-500');

    const warningLog = screen.getByText('[WRN] warning');
    expect(warningLog).toHaveClass('text-amber-500');
  });
});
