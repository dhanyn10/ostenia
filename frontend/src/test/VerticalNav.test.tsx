import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import VerticalNav from '../components/VerticalNav'
import React from 'react'

// Mock Icons because it might use things that need complex setup
vi.mock('../components/Icons', () => ({
  default: {
    Plugins: () => <div data-testid="plugins-icon" />
  }
}))

describe('VerticalNav Component', () => {
  const mockProps = {
    activeTab: 'activity',
    setActiveTab: vi.fn(),
    toggleTheme: vi.fn(),
    theme: 'light'
  }

  it('renders all navigation buttons', () => {
    render(<VerticalNav {...mockProps} />)
    expect(screen.getByTitle('Activity Center')).toBeInTheDocument()
    expect(screen.getByTitle('Proxy Management')).toBeInTheDocument()
    expect(screen.getByTitle('SSH & Remote Files')).toBeInTheDocument()
    expect(screen.getByTitle('Plugin Management')).toBeInTheDocument()
    expect(screen.getByTitle('System Activity Logs')).toBeInTheDocument()
  })

  it('calls setActiveTab when a button is clicked', () => {
    render(<VerticalNav {...mockProps} />)
    fireEvent.click(screen.getByTitle('Proxy Management'))
    expect(mockProps.setActiveTab).toHaveBeenCalledWith('proxy')
  })

  it('calls toggleTheme when theme button is clicked', () => {
    render(<VerticalNav {...mockProps} />)
    fireEvent.click(screen.getByTitle('Switch to Dark Mode'))
    expect(mockProps.toggleTheme).toHaveBeenCalled()
  })

  it('highlights the active tab', () => {
    const { rerender } = render(<VerticalNav {...mockProps} />)
    expect(screen.getByTitle('Activity Center')).toHaveClass('bg-blue-600')

    rerender(<VerticalNav {...mockProps} activeTab="proxy" />)
    expect(screen.getByTitle('Proxy Management')).toHaveClass('bg-blue-600')
    expect(screen.getByTitle('Activity Center')).not.toHaveClass('bg-blue-600')
  })
})
