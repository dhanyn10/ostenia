import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import StatusBar from '../components/StatusBar'
import React from 'react'

describe('StatusBar Component', () => {
  const mockServices = [
    { name: 'Apache', status: 'Running', port: 80 },
    { name: 'MySQL', status: 'Stopped', port: 0 },
  ]

  it('renders correctly with running services count', () => {
    render(<StatusBar services={mockServices} />)
    expect(screen.getByText('1 Services Running')).toBeInTheDocument()
  })

  it('toggles dropdown when clicked', () => {
    render(<StatusBar services={mockServices} />)
    const button = screen.getByText('1 Services Running').closest('button')

    fireEvent.click(button)
    expect(screen.getByText('Environment Services')).toBeInTheDocument()
    expect(screen.getByText('Apache')).toBeInTheDocument()
    expect(screen.getByText('MySQL')).toBeInTheDocument()

    fireEvent.click(button)
    expect(screen.queryByText('Environment Services')).not.toBeInTheDocument()
  })

  it('displays service status and ports in dropdown', () => {
    render(<StatusBar services={mockServices} />)
    const button = screen.getByText('1 Services Running').closest('button')
    fireEvent.click(button)

    expect(screen.getByText(':80')).toBeInTheDocument()
    expect(screen.getByText('Running')).toBeInTheDocument()
    expect(screen.getByText('Stopped')).toBeInTheDocument()
  })
})
