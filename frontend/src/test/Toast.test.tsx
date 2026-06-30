import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import React from 'react';
import Toast from '../components/Toast';

describe('Toast Component', () => {
  const mockToasts = [
    { id: '1', title: 'Success', message: 'Operation successful', type: 'success' },
    { id: '2', title: 'Error', message: 'Something went wrong', type: 'error' },
    { id: '3', title: 'Info', message: 'Just information', type: 'info' },
  ];

  it('renders all toasts', () => {
    const removeToast = vi.fn();
    render(<Toast toasts={mockToasts} removeToast={removeToast} />);

    expect(screen.getByText('Success')).toBeInTheDocument();
    expect(screen.getByText('Operation successful')).toBeInTheDocument();
    expect(screen.getByText('Error')).toBeInTheDocument();
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.getByText('Info')).toBeInTheDocument();
    expect(screen.getByText('Just information')).toBeInTheDocument();
  });

  it('calls removeToast when close button is clicked', () => {
    const removeToast = vi.fn();
    render(<Toast toasts={mockToasts} removeToast={removeToast} />);

    const closeButtons = screen.getAllByRole('button');
    fireEvent.click(closeButtons[0]);

    expect(removeToast).toHaveBeenCalledWith('1');
  });

  it('applies correct classes based on toast type', () => {
    const removeToast = vi.fn();
    const { container } = render(<Toast toasts={mockToasts} removeToast={removeToast} />);

    const toastElements = container.firstChild?.childNodes;
    // Check if the element exists and contains the expected class string in its className
    expect((toastElements?.[0] as HTMLElement).className).toContain('bg-emerald-50');
    expect((toastElements?.[1] as HTMLElement).className).toContain('bg-rose-50');
  });
});
