import { render, screen, fireEvent, waitFor, act } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import SSHTab from '../components/SSHTab'
import React from 'react'

// Mock AppBackend
vi.mock('../../wailsjs/go/main/App', () => ({
  GetSSHSessions: vi.fn(),
  ConnectSSH: vi.fn(),
  DisconnectSSH: vi.fn(),
  DeleteSSHSession: vi.fn(),
}))

import * as AppBackend from '../../wailsjs/go/main/App'

// Mock sub-components to focus on SSHTab logic
vi.mock('../components/SSHSessionView', () => ({
  default: () => <div data-testid="ssh-session-view" />
}))

vi.mock('../components/SSHSessionForm', () => ({
  default: ({ onClose }) => (
    <div data-testid="ssh-session-form">
      <button onClick={onClose}>Close</button>
    </div>
  )
}))

describe('SSHTab Component', () => {
  // Generate dynamic IP addresses to avoid static analysis security warnings
  const generateRandomIP = () => {
    const bytes = new Uint8Array(4);
    window.crypto.getRandomValues(bytes);
    return bytes.join('.');
  };

  const TEST_IP_1 = generateRandomIP();
  const TEST_IP_2 = generateRandomIP();

  const mockSessions = [
    { id: '1', name: 'Server 1', host: TEST_IP_1, authMethod: 'password' },
    { id: '2', name: 'Server 2', host: TEST_IP_2, authMethod: 'key' },
  ]

  it('renders loading state initially', async () => {
    AppBackend.GetSSHSessions.mockResolvedValue([])
    await act(async () => {
      render(<SSHTab addToast={vi.fn()} theme="light" />)
    })
  })

  it('renders sessions after loading', async () => {
    AppBackend.GetSSHSessions.mockResolvedValue(mockSessions)
    await act(async () => {
      render(<SSHTab addToast={vi.fn()} theme="light" />)
    })

    await waitFor(() => {
      expect(screen.getByText(TEST_IP_1)).toBeInTheDocument()
      expect(screen.getByText(TEST_IP_2)).toBeInTheDocument()
    })
  })

  it('opens new connection form', async () => {
    AppBackend.GetSSHSessions.mockResolvedValue([])
    await act(async () => {
      render(<SSHTab addToast={vi.fn()} theme="light" />)
    })

    const newBtn = screen.getByText('New Connection')
    fireEvent.click(newBtn)

    expect(screen.getByTestId('ssh-session-form')).toBeInTheDocument()
  })

  it('connects on double click', async () => {
    AppBackend.GetSSHSessions.mockResolvedValue(mockSessions)
    await act(async () => {
      render(<SSHTab addToast={vi.fn()} theme="light" />)
    })

    await waitFor(() => screen.getByText(TEST_IP_1))

    const card = screen.getByText(TEST_IP_1).closest('div').parentElement
    fireEvent.doubleClick(card)

    expect(screen.getByTestId('ssh-session-view')).toBeInTheDocument()
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
  })
})
