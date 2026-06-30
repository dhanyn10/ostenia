import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock scrollIntoView
if (typeof Element !== 'undefined') {
  Element.prototype.scrollIntoView = vi.fn();
}

// Mock matchMedia
if (typeof window !== 'undefined') {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation(query => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(), // deprecated
      removeListener: vi.fn(), // deprecated
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  });
}

// Mock ResizeObserver
if (typeof globalThis !== 'undefined' && !globalThis.ResizeObserver) {
    globalThis.ResizeObserver = vi.fn().mockImplementation(() => ({
        observe: vi.fn(),
        unobserve: vi.fn(),
        disconnect: vi.fn(),
    }));
}

// Mock crypto.randomUUID
if (typeof crypto !== 'undefined' && !crypto.randomUUID) {
  crypto.randomUUID = () => Math.random().toString(36).substring(2, 15);
} else if (typeof globalThis !== 'undefined' && !globalThis.crypto) {
  globalThis.crypto = {
    randomUUID: () => Math.random().toString(36).substring(2, 15)
  };
}

// Mock Wails runtime
globalThis.go = {
  main: {
    App: {
      GetConfig: vi.fn().mockResolvedValue({}),
      SaveConfig: vi.fn().mockResolvedValue(true),
      GetSSHSessions: vi.fn().mockResolvedValue([]),
      ConnectSSH: vi.fn().mockResolvedValue(null),
      DisconnectSSH: vi.fn().mockResolvedValue(null),
      DeleteSSHSession: vi.fn().mockResolvedValue(null),
      StartAllServices: vi.fn().mockResolvedValue(null),
      StopAllServices: vi.fn().mockResolvedValue(null),
      OpenTerminal: vi.fn().mockResolvedValue(null),
      OpenServerRootFolder: vi.fn().mockResolvedValue(null),
      OpenAppsLocationFolder: vi.fn().mockResolvedValue(null),
      CancelDownload: vi.fn().mockResolvedValue(null),
      OpenPluginFolder: vi.fn().mockResolvedValue(null),
      DeleteVersion: vi.fn().mockResolvedValue(null),
      StartService: vi.fn().mockResolvedValue(null),
      StopService: vi.fn().mockResolvedValue(null),
      SetApacheHTTPS: vi.fn().mockResolvedValue(null),
      SetNginxHTTPS: vi.fn().mockResolvedValue(null),
      InstallPrerequisite: vi.fn().mockResolvedValue(null),
      InstallPluginModule: vi.fn().mockResolvedValue(null),
      UninstallPluginModule: vi.fn().mockResolvedValue(null),
      SelectServerRoot: vi.fn().mockResolvedValue(""),
      SelectWWWRoot: vi.fn().mockResolvedValue(""),
      GetPrerequisites: vi.fn().mockResolvedValue([]),
      GetServiceStatus: vi.fn().mockResolvedValue({ status: 'Stopped', pid: 0, port: 0 }),
      Minimize: vi.fn().mockResolvedValue(null),
      Maximize: vi.fn().mockResolvedValue(null),
      Unmaximize: vi.fn().mockResolvedValue(null),
      Close: vi.fn().mockResolvedValue(null),
      ToggleDevTools: vi.fn().mockResolvedValue(null),
      GetProxyApps: vi.fn().mockResolvedValue([]),
      SaveProxyPort: vi.fn().mockResolvedValue(null),
      OpenProxyTerminal: vi.fn().mockResolvedValue(null),
    }
  },
  plugins: {
    Manager: {
      DownloadAndExtract: vi.fn().mockResolvedValue(true),
    }
  }
}

globalThis.runtime = {
  EventsOn: vi.fn(),
  EventsOff: vi.fn(),
  LogInfo: vi.fn(),
  LogError: vi.fn(),
  WindowSetTitle: vi.fn(),
}
