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
  const mockSessions = [
    { id: '1', name: 'Server 1', host: '1.2.3.4', authMethod: 'password' },
    { id: '2', name: 'Server 2', host: '5.6.7.8', authMethod: 'key' },
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
      expect(screen.getByText('1.2.3.4')).toBeInTheDocument()
      expect(screen.getByText('5.6.7.8')).toBeInTheDocument()
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

    await waitFor(() => screen.getByText('1.2.3.4'))

    const card = screen.getByText('1.2.3.4').closest('div').parentElement
    fireEvent.doubleClick(card)

    expect(screen.getByTestId('ssh-session-view')).toBeInTheDocument()
    expect(screen.getByText('Dashboard')).toBeInTheDocument()
  })
})
