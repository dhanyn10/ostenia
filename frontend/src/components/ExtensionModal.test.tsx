import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import ExtensionModal from './ExtensionModal';
import React from 'react';

describe('ExtensionModal Component', () => {
  const extensions = [
    { name: 'mysqli', enabled: true },
    { name: 'curl', enabled: false },
    { name: 'gd', enabled: true },
  ];

  it('does not render when isOpen is false', () => {
    const { container } = render(
      <ExtensionModal
        isOpen={false}
        onClose={() => {}}
        extensions={extensions}
        onToggle={() => {}}
        serviceName="PHP"
      />
    );
    expect(container.firstChild).toBeNull();
  });

  it('renders extensions and service name', () => {
    render(
      <ExtensionModal
        isOpen={true}
        onClose={() => {}}
        extensions={extensions}
        onToggle={() => {}}
        serviceName="PHP"
      />
    );
    expect(screen.getByText(/PHP Extensions/i)).toBeInTheDocument();
    expect(screen.getByText('mysqli')).toBeInTheDocument();
    expect(screen.getByText('curl')).toBeInTheDocument();
    expect(screen.getByText('gd')).toBeInTheDocument();
  });

  it('filters extensions based on search term', () => {
    render(
      <ExtensionModal
        isOpen={true}
        onClose={() => {}}
        extensions={extensions}
        onToggle={() => {}}
        serviceName="PHP"
      />
    );

    const searchInput = screen.getByPlaceholderText(/Search extensions/i);
    fireEvent.change(searchInput, { target: { value: 'curl' } });

    expect(screen.getByText('curl')).toBeInTheDocument();
    expect(screen.queryByText('mysqli')).not.toBeInTheDocument();
  });

  it('calls onToggle when extension is clicked', () => {
    const onToggle = vi.fn();
    render(
      <ExtensionModal
        isOpen={true}
        onClose={() => {}}
        extensions={extensions}
        onToggle={onToggle}
        serviceName="PHP"
      />
    );

    fireEvent.click(screen.getByText('curl'));
    expect(onToggle).toHaveBeenCalledWith('curl', true);

    fireEvent.click(screen.getByText('mysqli'));
    expect(onToggle).toHaveBeenCalledWith('mysqli', false);
  });

  it('calls onClose when close buttons are clicked', () => {
    const onClose = vi.fn();
    render(
      <ExtensionModal
        isOpen={true}
        onClose={onClose}
        extensions={extensions}
        onToggle={() => {}}
        serviceName="PHP"
      />
    );

    // There are multiple close buttons (X in header, Close in footer)
    const closeBtn = screen.getByRole('button', { name: /^Close$/i });
    fireEvent.click(closeBtn);
    expect(onClose).toHaveBeenCalled();

    // Backdrop click
    const backdrop = screen.getAllByRole('button')[0];
    fireEvent.click(backdrop);
    expect(onClose).toHaveBeenCalled();
  });
});
