import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock Wails runtime
global.window.go = {
  main: {
    App: {
      GetConfig: vi.fn().mockResolvedValue({}),
      SaveConfig: vi.fn().mockResolvedValue(true),
      GetSSHSessions: vi.fn().mockResolvedValue([]),
      ConnectSSH: vi.fn().mockResolvedValue(null),
      DisconnectSSH: vi.fn().mockResolvedValue(null),
      DeleteSSHSession: vi.fn().mockResolvedValue(null),
    }
  },
  plugins: {
    Manager: {
      DownloadAndExtract: vi.fn().mockResolvedValue(true),
    }
  }
}

global.window.runtime = {
  EventsOn: vi.fn(),
  EventsOff: vi.fn(),
  LogInfo: vi.fn(),
  LogError: vi.fn(),
}
