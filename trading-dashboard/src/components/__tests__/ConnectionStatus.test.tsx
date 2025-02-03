import { render, screen } from '@testing-library/react'
import { ConnectionStatus } from '../../components/ConnectionStatus'
import '@testing-library/jest-dom'

describe('ConnectionStatus', () => {
  it('should show connected status when all connections are active', () => {
    render(
      <ConnectionStatus 
        isPriceConnected={true}
        isPositionsConnected={true}
      />
    )
    
    expect(screen.getByText(/Connected/i)).toBeInTheDocument()
    expect(screen.getByTestId('status-indicator')).toHaveClass('bg-green-500')
  })

  it('should show partial connection status when only price feed is connected', () => {
    render(
      <ConnectionStatus 
        isPriceConnected={true}
        isPositionsConnected={false}
      />
    )
    
    expect(screen.getByText(/Partial Connection/i)).toBeInTheDocument()
    expect(screen.getByTestId('status-indicator')).toHaveClass('bg-yellow-500')
    expect(screen.getByText(/Position tracking disconnected/i)).toBeInTheDocument()
  })

  it('should show partial connection status when only positions feed is connected', () => {
    render(
      <ConnectionStatus 
        isPriceConnected={false}
        isPositionsConnected={true}
      />
    )
    
    expect(screen.getByText(/Partial Connection/i)).toBeInTheDocument()
    expect(screen.getByTestId('status-indicator')).toHaveClass('bg-yellow-500')
    expect(screen.getByText(/Price feed disconnected/i)).toBeInTheDocument()
  })

  it('should show disconnected status when no connections are active', () => {
    render(
      <ConnectionStatus 
        isPriceConnected={false}
        isPositionsConnected={false}
      />
    )
    
    expect(screen.getByText(/Disconnected/i)).toBeInTheDocument()
    expect(screen.getByTestId('status-indicator')).toHaveClass('bg-red-500')
  })

  it('should show loading state when connections are being established', () => {
    render(
      <ConnectionStatus 
        isPriceConnected={undefined}
        isPositionsConnected={undefined}
      />
    )
    
    expect(screen.getByText(/Connecting/i)).toBeInTheDocument()
    expect(screen.getByTestId('status-indicator')).toHaveClass('animate-pulse')
  })

  it('should show tooltip with detailed connection status', () => {
    render(
      <ConnectionStatus 
        isPriceConnected={true}
        isPositionsConnected={false}
      />
    )
    
    const statusIndicator = screen.getByTestId('status-indicator')
    expect(statusIndicator).toHaveAttribute('title', expect.stringContaining('Price feed: Connected'))
    expect(statusIndicator).toHaveAttribute('title', expect.stringContaining('Position tracking: Disconnected'))
  })
})
