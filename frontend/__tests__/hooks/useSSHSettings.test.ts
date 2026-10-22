import { describe, it, expect, beforeEach, vi } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { useSSHSettings } from "@/hooks/useSSHSettings";
import * as AppBackend from "@/wailsjs/go/backend/App";

vi.mock("@/wailsjs/go/backend/App", () => ({
  GetSSHSessions: vi.fn(),
}));

describe("useSSHSettings hook", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorage.clear();
    vi.mocked(AppBackend.GetSSHSessions).mockResolvedValue([]);
  });

  const renderHookAndWait = async () => {
    const rendered = renderHook(() => useSSHSettings());
    await waitFor(() => {
      expect(AppBackend.GetSSHSessions).toHaveBeenCalled();
    });
    return rendered;
  };

  it("initializes with default settings when localStorage is empty", async () => {
    const { result } = await renderHookAndWait();

    expect(result.current.zoomEnabled).toBe(true);
    expect(result.current.monitorEnabled).toBe(true);
    expect(result.current.monitorInterval).toBe(3);
    expect(result.current.displayMode).toBe("tooltip");
    expect(result.current.maxTimeout).toBe(10);
    expect(result.current.maxRetries).toBe(3);
    expect(result.current.showPasswords).toBe(false);
  });

  it("initializes with values persisted in localStorage", async () => {
    localStorage.setItem("ostenia_ssh_zoom_enabled", "false");
    localStorage.setItem("ostenia_ssh_monitor_enabled", "false");
    localStorage.setItem("ostenia_ssh_monitor_interval", "15");
    localStorage.setItem("ostenia_ssh_monitor_display_mode", "always");
    localStorage.setItem("ostenia_ssh_max_timeout", "25");
    localStorage.setItem("ostenia_ssh_max_retries", "5");

    const { result } = await renderHookAndWait();

    expect(result.current.zoomEnabled).toBe(false);
    expect(result.current.monitorEnabled).toBe(false);
    expect(result.current.monitorInterval).toBe(15);
    expect(result.current.displayMode).toBe("always");
    expect(result.current.maxTimeout).toBe(25);
    expect(result.current.maxRetries).toBe(5);
  });

  it("loads and sets SSH sessions from backend on mount", async () => {
    const mockSessions = [
      { id: "1", host: "192.168.1.100", user: "root" },
      { id: "2", host: "example.com", user: "admin" },
    ];
    vi.mocked(AppBackend.GetSSHSessions).mockResolvedValue(mockSessions);

    const { result } = await renderHookAndWait();

    expect(result.current.sshSessions).toEqual(mockSessions);
  });

  it("toggles zoom and dispatches event", async () => {
    const dispatchSpy = vi.spyOn(window, "dispatchEvent");
    const { result } = await renderHookAndWait();

    act(() => {
      result.current.handleToggleZoom(false);
    });

    expect(result.current.zoomEnabled).toBe(false);
    expect(localStorage.getItem("ostenia_ssh_zoom_enabled")).toBe("false");
    expect(dispatchSpy).toHaveBeenCalledWith(
      expect.objectContaining({ type: "ostenia_ssh_zoom_settings_changed" }),
    );
  });

  it("toggles monitor and dispatches event", async () => {
    const dispatchSpy = vi.spyOn(window, "dispatchEvent");
    const { result } = await renderHookAndWait();

    act(() => {
      result.current.handleToggleMonitorEnabled(false);
    });

    expect(result.current.monitorEnabled).toBe(false);
    expect(localStorage.getItem("ostenia_ssh_monitor_enabled")).toBe("false");
    expect(dispatchSpy).toHaveBeenCalledWith(
      expect.objectContaining({ type: "ostenia_ssh_monitor_settings_changed" }),
    );
  });

  it("handles interval changes with clamping between 1 and 60", async () => {
    const { result } = await renderHookAndWait();

    act(() => {
      result.current.handleIntervalChange("10");
    });
    expect(result.current.monitorInterval).toBe(10);
    expect(localStorage.getItem("ostenia_ssh_monitor_interval")).toBe("10");

    // Clamps below 1 to 1
    act(() => {
      result.current.handleIntervalChange("0");
    });
    expect(result.current.monitorInterval).toBe(1);

    // Clamps above 60 to 60
    act(() => {
      result.current.handleIntervalChange("999");
    });
    expect(result.current.monitorInterval).toBe(60);
  });

  it("handles display mode changes and validates permitted options", async () => {
    const { result } = await renderHookAndWait();

    act(() => {
      result.current.handleDisplayModeChange("hover-inline");
    });
    expect(result.current.displayMode).toBe("hover-inline");
    expect(localStorage.getItem("ostenia_ssh_monitor_display_mode")).toBe(
      "hover-inline",
    );

    // Invalid mode falls back to tooltip
    act(() => {
      result.current.handleDisplayModeChange("invalid-mode");
    });
    expect(result.current.displayMode).toBe("tooltip");
  });

  it("handles timeout and retries changes", async () => {
    const { result } = await renderHookAndWait();

    act(() => {
      result.current.handleMaxTimeoutChange("45");
      result.current.handleMaxRetriesChange("8");
    });

    expect(result.current.maxTimeout).toBe(45);
    expect(result.current.maxRetries).toBe(8);
    expect(localStorage.getItem("ostenia_ssh_max_timeout")).toBe("45");
    expect(localStorage.getItem("ostenia_ssh_max_retries")).toBe("8");
  });
});
