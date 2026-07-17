import { render, screen, fireEvent, act } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import React from 'react';
import SSHToolbar from './SSHToolbar';

describe('SSHToolbar Component', () => {
  const setExplorerVisibleMock = vi.fn();
  const onFitMock = vi.fn();
  const onReconnectMock = vi.fn();
  const onCloseMock = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders toolbar correctly', async () => {
    await act(async () => {
      render(
        <SSHToolbar
          explorerVisible={true}
          setExplorerVisible={setExplorerVisibleMock}
          onFit={onFitMock}
          onReconnect={onReconnectMock}
          onClose={onCloseMock}
          connecting={false}
        />
      );
    });

    expect(screen.getByTitle('Toggle Explorer')).toBeInTheDocument();
    expect(screen.getByTitle('Fit Terminal')).toBeInTheDocument();
    expect(screen.getByTitle('Reconnect')).toBeInTheDocument();
    expect(screen.getByTitle('Close')).toBeInTheDocument();
  });

  it('triggers events when clicking buttons', async () => {
    await act(async () => {
      render(
        <SSHToolbar
          explorerVisible={false}
          setExplorerVisible={setExplorerVisibleMock}
          onFit={onFitMock}
          onReconnect={onReconnectMock}
          onClose={onCloseMock}
          connecting={false}
        />
      );
    });

    // Toggle explorer
    fireEvent.click(screen.getByTitle('Toggle Explorer'));
    expect(setExplorerVisibleMock).toHaveBeenCalledWith(true);

    // Fit terminal
    fireEvent.click(screen.getByTitle('Fit Terminal'));
    expect(onFitMock).toHaveBeenCalled();

    // Reconnect
    fireEvent.click(screen.getByTitle('Reconnect'));
    expect(onReconnectMock).toHaveBeenCalled();

    // Close
    fireEvent.click(screen.getByTitle('Close'));
    expect(onCloseMock).toHaveBeenCalled();
  });

  it('disables reconnect button when connecting', async () => {
    await act(async () => {
      render(
        <SSHToolbar
          explorerVisible={false}
          setExplorerVisible={setExplorerVisibleMock}
          onFit={onFitMock}
          onReconnect={onReconnectMock}
          onClose={onCloseMock}
          connecting={true}
        />
      );
    });

    const reconnectBtn = screen.getByTitle('Reconnect');
    expect(reconnectBtn).toBeDisabled();

    // Clicking should not trigger reconnect
    fireEvent.click(reconnectBtn);
    expect(onReconnectMock).not.toHaveBeenCalled();
  });
});
