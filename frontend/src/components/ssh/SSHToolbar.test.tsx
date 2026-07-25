import { render, screen, fireEvent, act } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import React from "react";
import SSHToolbar from "./SSHToolbar";

describe("SSHToolbar Component", () => {
  const setExplorerVisibleMock = vi.fn();
  const onReconnectMock = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders toolbar correctly", () => {
    render(
      <SSHToolbar
        explorerVisible={true}
        setExplorerVisible={setExplorerVisibleMock}
        onReconnect={onReconnectMock}
        connecting={false}
      />,
    );

    expect(screen.getByTitle("Toggle Explorer")).toBeInTheDocument();
    expect(screen.getByTitle("Reconnect")).toBeInTheDocument();
  });

  it("triggers events when clicking buttons", () => {
    render(
      <SSHToolbar
        explorerVisible={false}
        setExplorerVisible={setExplorerVisibleMock}
        onReconnect={onReconnectMock}
        connecting={false}
      />,
    );

    // Toggle explorer
    fireEvent.click(screen.getByTitle("Toggle Explorer"));
    expect(setExplorerVisibleMock).toHaveBeenCalledWith(true);

    // Reconnect
    fireEvent.click(screen.getByTitle("Reconnect"));
    expect(onReconnectMock).toHaveBeenCalled();
  });

  it("disables reconnect button when connecting", () => {
    render(
      <SSHToolbar
        explorerVisible={false}
        setExplorerVisible={setExplorerVisibleMock}
        onReconnect={onReconnectMock}
        connecting={true}
      />,
    );

    const reconnectBtn = screen.getByTitle("Reconnect");
    expect(reconnectBtn).toBeDisabled();

    // Clicking should not trigger reconnect
    fireEvent.click(reconnectBtn);
    expect(onReconnectMock).not.toHaveBeenCalled();
  });
});
