import { render, screen } from '@testing-library/react'
import { ErrorBoundary } from '../ErrorBoundary'
import '@testing-library/jest-dom'

const ThrowError = () => {
  throw new Error('Test error')
}

describe('ErrorBoundary', () => {
  const mockReload = jest.fn()
  
  beforeEach(() => {
    jest.spyOn(console, 'error').mockImplementation(() => {})
    const mockLocation = { reload: mockReload }
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: mockLocation,
      writable: true
    })
  })

  afterEach(() => {
    jest.restoreAllMocks()
    // Restore original window.location
    delete (window as any).location
  })

  it('should render children when no error occurs', () => {
    render(
      <ErrorBoundary>
        <div>Test Content</div>
      </ErrorBoundary>
    )
    expect(screen.getByText('Test Content')).toBeInTheDocument()
  })

  it('should render error UI when error occurs', () => {
    render(
      <ErrorBoundary>
        <ThrowError />
      </ErrorBoundary>
    )

    expect(screen.getByText(/An error occurred in the trading dashboard/i)).toBeInTheDocument()
    const reloadButton = screen.getByRole('button', { name: /Reload Application/i })
    expect(reloadButton).toBeInTheDocument()
    
    reloadButton.click()
    expect(mockReload).toHaveBeenCalled()
  })

  it('should handle nested errors', () => {
    const NestedError = () => (
      <div>
        <ErrorBoundary>
          <ThrowError />
        </ErrorBoundary>
      </div>
    )

    render(
      <ErrorBoundary>
        <NestedError />
      </ErrorBoundary>
    )

    expect(screen.getByText(/An error occurred in the trading dashboard/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Reload Application/i })).toBeInTheDocument()
  })
})
