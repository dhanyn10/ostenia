import { useState, useEffect } from "react";
import * as AppBackend from "@/wailsjs/go/backend/App";

export function useSSHSettings() {
  const [sshSessions, setSshSessions] = useState<any[]>([]);
  const [showPasswords, setShowPasswords] = useState(false);

  const [zoomEnabled, setZoomEnabled] = useState<boolean>(() => {
    if (typeof localStorage === "undefined") return true;
    return localStorage.getItem("ostenia_ssh_zoom_enabled") !== "false";
  });

  const [monitorEnabled, setMonitorEnabled] = useState<boolean>(() => {
    if (typeof localStorage === "undefined") return true;
    return localStorage.getItem("ostenia_ssh_monitor_enabled") !== "false";
  });

  const [monitorInterval, setMonitorInterval] = useState<number>(() => {
    if (typeof localStorage === "undefined") return 3;
    const val = Number.parseInt(
      localStorage.getItem("ostenia_ssh_monitor_interval") || "3",
      10,
    );
    return Number.isNaN(val) || val < 1 ? 3 : Math.min(Math.max(val, 1), 60);
  });

  const [displayMode, setDisplayMode] = useState<string>(() => {
    if (typeof localStorage === "undefined") return "tooltip";
    const mode =
      localStorage.getItem("ostenia_ssh_monitor_display_mode") || "tooltip";
    return ["tooltip", "hover-inline", "always"].includes(mode)
      ? mode
      : "tooltip";
  });

  const [maxTimeout, setMaxTimeout] = useState<number>(() => {
    if (typeof localStorage === "undefined") return 10;
    const val = Number.parseInt(
      localStorage.getItem("ostenia_ssh_max_timeout") || "10",
      10,
    );
    return Number.isNaN(val) || val < 1 ? 10 : val;
  });

  const [maxRetries, setMaxRetries] = useState<number>(() => {
    if (typeof localStorage === "undefined") return 3;
    const val = Number.parseInt(
      localStorage.getItem("ostenia_ssh_max_retries") || "3",
      10,
    );
    return Number.isNaN(val) || val < 1 ? 3 : val;
  });

  const loadSSHSessions = async () => {
    try {
      const sessions = await AppBackend.GetSSHSessions();
      setSshSessions(sessions || []);
    } catch (err) {
      console.error(err);
    }
  };

  useEffect(() => {
    void loadSSHSessions();
  }, []);

  const handleToggleZoom = (enabled: boolean) => {
    setZoomEnabled(enabled);
    if (typeof localStorage !== "undefined") {
      localStorage.setItem("ostenia_ssh_zoom_enabled", enabled ? "true" : "false");
    }
    if (typeof window !== "undefined") {
      window.dispatchEvent(new Event("ostenia_ssh_zoom_settings_changed"));
    }
  };

  const handleToggleMonitorEnabled = (enabled: boolean) => {
    setMonitorEnabled(enabled);
    if (typeof localStorage !== "undefined") {
      localStorage.setItem(
        "ostenia_ssh_monitor_enabled",
        enabled ? "true" : "false",
      );
    }
    if (typeof window !== "undefined") {
      window.dispatchEvent(new Event("ostenia_ssh_monitor_settings_changed"));
    }
  };

  const handleIntervalChange = (rawValue: string) => {
    const parsed = Number.parseInt(rawValue, 10);
    if (!Number.isNaN(parsed)) {
      const clamped = Math.min(Math.max(parsed, 1), 60);
      setMonitorInterval(clamped);
      if (typeof localStorage !== "undefined") {
        localStorage.setItem("ostenia_ssh_monitor_interval", String(clamped));
      }
      if (typeof window !== "undefined") {
        window.dispatchEvent(new Event("ostenia_ssh_monitor_settings_changed"));
      }
    }
  };

  const handleDisplayModeChange = (rawMode: string) => {
    const sanitizedMode = ["tooltip", "hover-inline", "always"].includes(rawMode)
      ? rawMode
      : "tooltip";
    setDisplayMode(sanitizedMode);
    if (typeof localStorage !== "undefined") {
      localStorage.setItem("ostenia_ssh_monitor_display_mode", sanitizedMode);
    }
    if (typeof window !== "undefined") {
      window.dispatchEvent(new Event("ostenia_ssh_monitor_settings_changed"));
    }
  };

  const handleMaxTimeoutChange = (rawValue: string) => {
    const parsed = Number.parseInt(rawValue, 10);
    if (!Number.isNaN(parsed)) {
      const clamped = Math.max(parsed, 1);
      setMaxTimeout(clamped);
      if (typeof localStorage !== "undefined") {
        localStorage.setItem("ostenia_ssh_max_timeout", String(clamped));
      }
    }
  };

  const handleMaxRetriesChange = (rawValue: string) => {
    const parsed = Number.parseInt(rawValue, 10);
    if (!Number.isNaN(parsed)) {
      const clamped = Math.max(parsed, 1);
      setMaxRetries(clamped);
      if (typeof localStorage !== "undefined") {
        localStorage.setItem("ostenia_ssh_max_retries", String(clamped));
      }
    }
  };

  return {
    sshSessions,
    setSshSessions,
    showPasswords,
    setShowPasswords,
    zoomEnabled,
    monitorEnabled,
    monitorInterval,
    displayMode,
    maxTimeout,
    maxRetries,
    loadSSHSessions,
    handleToggleZoom,
    handleToggleMonitorEnabled,
    handleIntervalChange,
    handleDisplayModeChange,
    handleMaxTimeoutChange,
    handleMaxRetriesChange,
  };
}
