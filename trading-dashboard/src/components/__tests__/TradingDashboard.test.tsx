import { render, screen, fireEvent, waitFor, within, act, cleanup } from '@testing-library/react'
import { TradingDashboard } from '../TradingDashboard'
import { useWebSocket } from '../../hooks/use-websocket'
import '@testing-library/jest-dom'
import { ThemeProvider } from '../theme-provider'
import { ErrorBoundary } from '../ErrorBoundary'
import type { PriceData, Position, TradingPair } from '../../types'
import { TRADING_PAIRS } from '../../types'

jest.mock('../../hooks/use-websocket')

import type { ReactNode } from 'react'

const renderWithTheme = (component: ReactNode) => {
  return render(
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      {component}
    </ThemeProvider>
  )
}

jest.mock('../../hooks/use-websocket')

describe('TradingDashboard', () => {
  const mockPriceHistory: PriceData[] = [
    { timestamp: Date.now() - 1000, price: 100 },
    { timestamp: Date.now(), price: 101 }
  ]

  const mockPositions: Position[] = [
    {
      symbol: TRADING_PAIRS[0] as TradingPair,
      size: 10,
      entryPrice: 100,
      currentPrice: 102.5,
      pnl: 25
    }
  ]

  beforeEach(() => {
    jest.clearAllMocks()
    ;(useWebSocket as jest.Mock).mockImplementation((url: string) => ({
      data: url.includes('prices') ? mockPriceHistory : mockPositions,
      error: null,
      isConnected: true,
      retryCount: 0,
      handleError: jest.fn(),
      reconnect: jest.fn(),
      disconnect: jest.fn(),
      send: jest.fn()
    }))
  })

  afterEach(() => {
    cleanup()
    jest.clearAllMocks()
    ;(useWebSocket as jest.Mock).mockReset()
  })

  it('should render loading state initially', async () => {
    const mockWebSocket = {
      data: undefined,
      error: null,
      isConnected: true
    }
    ;(useWebSocket as jest.Mock).mockReturnValue(mockWebSocket)

    const { unmount } = renderWithTheme(<TradingDashboard />)

    const loadingStatus = screen.getByRole('status', { name: /Loading Trading System/i })
    expect(loadingStatus).toHaveTextContent(/Loading Trading System/i)
    expect(screen.getByRole('status', { name: /Loading status/i })).toHaveTextContent(/Fetching market data/i)

    // Cleanup
    unmount()
  })

  it('should render error state when connection fails', async () => {
    const mockWebSocket = {
      data: undefined,
      error: new Error('Connection failed'),
      isConnected: false
    }
    ;(useWebSocket as jest.Mock).mockReturnValue(mockWebSocket)

    const { unmount } = renderWithTheme(<TradingDashboard />)
    
    const errorHeading = screen.getAllByRole('alert')[0]
    expect(errorHeading).toHaveTextContent(/Connection Error/i)
    
    const errorMessage = screen.getAllByRole('alert')[1]
    expect(errorMessage).toHaveTextContent(/Connection lost to trading servers/i)
    
    expect(screen.getByRole('status', { name: /connection status/i })).toHaveClass('bg-red-500')

    unmount()
  })

  it('should render price chart and positions when data is available', async () => {
    // Mock both price and position data
    ;(useWebSocket as jest.Mock).mockImplementation((url: string) => ({
      data: url.includes('prices') ? mockPriceHistory : mockPositions,
      error: null,
      isConnected: true,
      retryCount: 0
    }))

    const { unmount } = renderWithTheme(<TradingDashboard />)
    
    // Verify price chart is rendered
    await waitFor(() => {
      expect(screen.getByTestId('price-chart-title')).toHaveTextContent(/Price Chart/i)
      expect(screen.getByRole('region', { name: /price chart/i })).toBeInTheDocument()
    })
    
    // Click on positions tab and verify positions
    const positionsTab = screen.getByRole('tab', { name: /positions/i })
    expect(positionsTab).toBeInTheDocument()
    fireEvent.click(positionsTab)
    
    await waitFor(() => {
      const positionsTitle = screen.getByRole('heading', { name: /Open Positions/i })
      expect(positionsTitle).toHaveTextContent(/Open Positions/i)
      
      // Look for trading pair in the positions section specifically
      const positionsSection = screen.getByRole('region', { name: /positions/i })
      const tradingPairElement = within(positionsSection).getByRole('cell', { name: new RegExp(TRADING_PAIRS[0], 'i') })
      expect(tradingPairElement.textContent).toContain(TRADING_PAIRS[0])
    }, { timeout: 3000 })

    unmount()
  })

  it('should handle time range changes and chart updates', async () => {
    const timeRanges = ['1H', '24H', '7D', '30D']
    const mockPriceData = {
      '1H': [{ timestamp: Date.now() - 3600000, price: 100 }, { timestamp: Date.now(), price: 110 }],
      '24H': [{ timestamp: Date.now() - 86400000, price: 95 }, { timestamp: Date.now(), price: 110 }],
      '7D': [{ timestamp: Date.now() - 604800000, price: 90 }, { timestamp: Date.now(), price: 110 }],
      '30D': [{ timestamp: Date.now() - 2592000000, price: 85 }, { timestamp: Date.now(), price: 110 }]
    }

    ;(useWebSocket as jest.Mock).mockImplementation((url: string) => {
      const timeRange = (url.match(/timeRange=(\w+)/)?.[1] || '1H') as '1H' | '24H' | '7D' | '30D'
      return {
        data: url.includes('prices') ? mockPriceData[timeRange] : mockPositions,
        error: null,
        isConnected: true,
        retryCount: 0
      }
    })

    renderWithTheme(<TradingDashboard />)

    for (const range of timeRanges) {
      await act(async () => {
        const timeRangeButton = screen.getByRole('button', { name: new RegExp(range, 'i') })
        fireEvent.click(timeRangeButton)
        jest.runOnlyPendingTimers()
      })
      
      expect(useWebSocket).toHaveBeenCalledWith(
        expect.stringContaining(`timeRange=${range}`)
      )
      
      // Verify button state changes
      const timeRangeButton = screen.getByRole('button', { name: new RegExp(range, 'i') })
      expect(timeRangeButton).toHaveClass('bg-primary')
      expect(timeRangeButton).toHaveClass('text-primary-foreground')
      
      // Other buttons should not have primary classes
      const otherRanges = timeRanges.filter(r => r !== range)
      for (const otherRange of otherRanges) {
        const otherButton = screen.getByRole('button', { name: new RegExp(otherRange, 'i') })
        expect(otherButton).not.toHaveClass('bg-primary')
      }

      // Verify chart data updates
      await waitFor(() => {
        const chartContainer = screen.getByRole('region', { name: /price chart/i })
        expect(chartContainer).toBeInTheDocument()
        const lineChart = within(chartContainer).getByRole('img', { name: /price chart/i })
        expect(lineChart).toBeInTheDocument()
        const xAxis = within(chartContainer).getByRole('graphics-symbol', { name: /x-axis/i })
        expect(xAxis).toBeInTheDocument()
        const yAxis = within(chartContainer).getByRole('graphics-symbol', { name: /y-axis/i })
        expect(yAxis).toBeInTheDocument()
        const tooltip = within(chartContainer).getByRole('tooltip')
        expect(tooltip).toBeInTheDocument()
      }, { timeout: 3000 })
    }
  })

  it('should handle symbol changes', async () => {
    const mockSymbolData = {
      'SOL/USD': [{ timestamp: Date.now(), price: 100 }],
      'SOL/USDC': [{ timestamp: Date.now(), price: 99.5 }],
      'BONK/USD': [{ timestamp: Date.now(), price: 0.00001 }]
    }

    ;(useWebSocket as jest.Mock).mockImplementation((url: string) => {
      const symbol = decodeURIComponent(url.match(/symbol=([^&]+)/)?.[1] || TRADING_PAIRS[0]) as TradingPair
      return {
        data: url.includes('prices') ? mockSymbolData[symbol] : mockPositions,
        error: null,
        isConnected: true,
        retryCount: 0
      }
    })

    renderWithTheme(<TradingDashboard />)

    // Wait for initial render
    await waitFor(() => {
      expect(screen.queryByRole('heading', { name: /Loading Trading System/i })).not.toBeInTheDocument()
    })

    // Test each trading pair
    for (const symbol of TRADING_PAIRS) {
      await act(async () => {
        const symbolSelect = screen.getByRole('combobox', { name: /trading pair/i })
        fireEvent.change(symbolSelect, { target: { value: symbol } })
        jest.runOnlyPendingTimers()
      })
      
      // Verify WebSocket connection
      expect(useWebSocket).toHaveBeenCalledWith(
        expect.stringMatching(new RegExp(`symbol=${encodeURIComponent(symbol)}`))
      )

      // Verify chart updates
      const chartSubtitle = screen.getByRole('heading', { name: new RegExp(`${symbol}`, 'i') })
      expect(chartSubtitle).toHaveTextContent(new RegExp(`${symbol}`, 'i'))
      
      const chartContainer = screen.getByRole('region', { name: /price chart/i })
      expect(chartContainer).toBeInTheDocument()
      const lineChart = within(chartContainer).getByRole('img', { name: /price chart/i })
      expect(lineChart).toBeInTheDocument()
      const priceLine = within(chartContainer).getByRole('graphics-symbol', { name: /price line/i })
      expect(priceLine).toBeInTheDocument()
    }
  })

  it('should calculate and display total PnL correctly', () => {
    renderWithTheme(<TradingDashboard />)
    const pnlValue = screen.getByRole('status', { name: /total profit and loss/i })
    expect(pnlValue).toHaveTextContent(/\$25\.00/)
    expect(pnlValue).toHaveClass('text-green-500')
  })

  it('should handle retry attempts and max retries', async () => {
    const mockError = new Error('Connection error')
    let retryCount = 0
    const disconnect = jest.fn()
    
    ;(useWebSocket as jest.Mock).mockImplementation(() => ({
      data: null,
      error: mockError,
      isConnected: false,
      retryCount,
      retryTimeout: 0,
      disconnect
    }))

    renderWithTheme(<TradingDashboard />)
    
    // Initial error state
    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(/Connection lost to trading servers/i)
    })
    
    // Test retry mechanism
    const maxRetries = 3
    for (let i = 0; i < maxRetries; i++) {
      const retryButton = screen.getByRole('button', { name: /Try Again/i })
      fireEvent.click(retryButton)
      retryCount++
      
      ;(useWebSocket as jest.Mock).mockImplementation(() => ({
        data: null,
        error: mockError,
        isConnected: false,
        retryCount,
        retryTimeout: 0
      }))

      await waitFor(() => {
        const retryButton = screen.getByRole('button', { name: /Try Again/i })
        expect(retryButton).toBeInTheDocument()
        expect(screen.getByRole('alert')).toHaveTextContent(new RegExp(`Connection lost to trading servers \\(Attempt ${retryCount} of ${maxRetries}\\)`, 'i'))
      })
    }

    // Test max retries reached
    const retryButton = screen.getByRole('button', { name: /Try Again/i })
    fireEvent.click(retryButton)
    retryCount++
    
    ;(useWebSocket as jest.Mock).mockImplementation(() => ({
      data: null,
      error: mockError,
      isConnected: false,
      retryCount,
      retryTimeout: 0
    }))

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(/Max Retries Reached/i)
      expect(retryButton).toBeDisabled()
    })

    // Verify useWebSocket was called correct number of times
    expect(useWebSocket).toHaveBeenCalledTimes(maxRetries + 1) // Initial + retries
    
    // Test reset functionality
    await act(async () => {
      const resetButton = screen.getByRole('button', { name: /Reset Connection/i })
      fireEvent.click(resetButton)
      jest.runOnlyPendingTimers()
    })
    
    expect(screen.getByRole('status')).toHaveTextContent(/Connecting/i)
  })

  it('should handle partial connection states', async () => {
    ;(useWebSocket as jest.Mock)
      .mockImplementationOnce(() => ({
        data: mockPriceHistory,
        error: null,
        isConnected: true
      }))
      .mockImplementationOnce(() => ({
        data: null,
        error: new Error('Position feed error'),
        isConnected: false
      }))

    renderWithTheme(<TradingDashboard />)
    
    await waitFor(() => {
      const priceChart = screen.getByTestId('price-chart-card')
      expect(within(priceChart).getByRole('heading', { name: /Price Chart/i })).toBeInTheDocument()
      expect(screen.getByRole('alert')).toHaveTextContent(/Position feed unavailable/i)
      expect(screen.getByRole('status', { name: /connection status/i })).toHaveClass('bg-yellow-500')
      expect(screen.getByRole('button', { name: /Try Again/i })).toBeInTheDocument()
    }, { timeout: 5000 })
    
    // Test recovery
    ;(useWebSocket as jest.Mock)
      .mockImplementation(() => ({
        data: mockPriceHistory,
        error: null,
        isConnected: true
      }))

    const retryButton = screen.getByRole('button', { name: /Try Again/i })
    fireEvent.click(retryButton)

    await waitFor(() => {
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
      expect(screen.getByRole('status', { name: /connection status/i })).toHaveClass('bg-green-500')
    })
  })

  it('should handle successful data updates', async () => {
    const mockPositionsData = [
      {
        symbol: TRADING_PAIRS[0] as TradingPair,
        size: 10,
        entryPrice: 100,
        currentPrice: 110,
        pnl: 100
      },
      {
        symbol: TRADING_PAIRS[1] as TradingPair,
        size: 5,
        entryPrice: 200,
        currentPrice: 180,
        pnl: -100
      }
    ]

    ;(useWebSocket as jest.Mock)
      .mockImplementationOnce(() => ({
        data: mockPriceHistory,
        error: null,
        isConnected: true
      }))
      .mockImplementationOnce(() => ({
        data: mockPositionsData,
        error: null,
        isConnected: true
      }))

    renderWithTheme(<TradingDashboard />)
    
    // Verify portfolio stats
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /Portfolio Value/i })).toBeInTheDocument()
      expect(screen.getByRole('heading', { name: /Total P&L/i })).toBeInTheDocument()
      expect(screen.getByRole('heading', { name: /Active Positions/i })).toBeInTheDocument()
      expect(screen.getByRole('status', { name: /portfolio value/i })).toHaveTextContent(/\$2,050\.00/)
      expect(screen.getByRole('status', { name: /total p&l/i })).toHaveTextContent(/\$25\.00/)
      expect(screen.getByRole('status', { name: /profitable positions/i })).toHaveTextContent('1 profitable')
    }, { timeout: 3000 })

    // Verify chart controls
    await act(async () => {
      const timeRangeButton = screen.getByRole('button', { name: /24H/i })
      fireEvent.click(timeRangeButton)
      jest.runOnlyPendingTimers()
    })
    expect(useWebSocket).toHaveBeenCalledWith(expect.stringContaining('timeRange=24H'))

    // Verify symbol selection
    await act(async () => {
      const symbolSelect = screen.getByRole('combobox', { name: /trading pair/i })
      fireEvent.change(symbolSelect, { target: { value: TRADING_PAIRS[1] } })
      jest.runOnlyPendingTimers()
    })
    expect(useWebSocket).toHaveBeenCalledWith(expect.stringContaining(TRADING_PAIRS[1]))

    // Verify positions tab
    await act(async () => {
      const positionsTab = screen.getByRole('tab', { name: /positions/i })
      fireEvent.click(positionsTab)
      jest.runOnlyPendingTimers()
    })
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /Open Positions/i })).toBeInTheDocument()
      const positionsSection = screen.getByRole('region', { name: /positions/i })
      const tradingPairElements = within(positionsSection).getAllByRole('cell', { name: new RegExp(TRADING_PAIRS[0], 'i') })
      expect(tradingPairElements.length).toBeGreaterThan(0)
    })
  })

  beforeEach(() => {
    jest.clearAllMocks()
  })

  afterEach(() => {
    cleanup()
  })

  it('should handle loading states correctly', async () => {
    const disconnect = jest.fn()
    
    ;(useWebSocket as jest.Mock).mockImplementation((url: string) => ({
      data: url.includes('prices') ? undefined : undefined,
      error: null,
      isConnected: true,
      disconnect
    }))
    
    const { rerender, unmount } = renderWithTheme(<TradingDashboard />)
    
    // Initial loading state
    expect(screen.getByRole('status', { name: /Loading Trading System/i })).toBeInTheDocument()
    expect(screen.getByRole('status', { name: /Loading status/i })).toBeInTheDocument()
    
    // Update mock with data
    ;(useWebSocket as jest.Mock).mockImplementation((url: string) => ({
      data: url.includes('prices') ? mockPriceHistory : mockPositions,
      error: null,
      isConnected: true,
      disconnect
    }))
    rerender(<TradingDashboard />)
    
    // Wait for loading state to clear
    await waitFor(() => {
      expect(screen.queryByRole('status', { name: /Loading Trading System/i })).not.toBeInTheDocument()
    })
    
    const chartContainer = screen.getByTestId('price-chart-card')
    expect(chartContainer).toBeInTheDocument()
    expect(within(chartContainer).getByRole('heading', { name: /Price Chart/i })).toBeInTheDocument()
    
    // Cleanup and verify disconnect was called
    unmount()
    expect(disconnect).toHaveBeenCalled()
  })

  it('should calculate portfolio statistics correctly', async () => {
    const mockPositionsData = [
      {
        symbol: TRADING_PAIRS[0] as TradingPair,
        size: 10,
        entryPrice: 100,
        currentPrice: 120,
        pnl: 200
      },
      {
        symbol: TRADING_PAIRS[1] as TradingPair,
        size: 5,
        entryPrice: 200,
        currentPrice: 180,
        pnl: -100
      }
    ]

    ;(useWebSocket as jest.Mock)
      .mockImplementationOnce(() => ({
        data: mockPriceHistory,
        error: null,
        isConnected: true
      }))
      .mockImplementationOnce(() => ({
        data: mockPositionsData,
        error: null,
        isConnected: true
      }))

    renderWithTheme(<TradingDashboard />)

    // Initial portfolio stats
    await waitFor(() => {
      const portfolioValue = screen.getByRole('status', { name: /portfolio value/i })
      const totalValue = (10 * 120) + (5 * 180) // positions value (1200 + 900)
      expect(portfolioValue).toHaveTextContent(`$${totalValue.toFixed(2)}`) // Total portfolio value (2100)
      expect(screen.getByRole('status', { name: /total p&l/i })).toHaveTextContent(/\$100\.00/) // Net PnL (200 - 100)
      expect(screen.getByRole('status', { name: /profitable positions/i })).toHaveTextContent('1')
      expect(screen.getByRole('status', { name: /total positions/i })).toHaveTextContent('2')
      expect(screen.getByRole('status', { name: /connection status/i })).toHaveClass('bg-green-500')
    }, { timeout: 3000 })

    // Test filtering by symbol
    const symbolSelect = screen.getByTestId('symbol-select')
    fireEvent.change(symbolSelect, { target: { value: TRADING_PAIRS[0] } })

    // Verify filtered portfolio stats
    await waitFor(() => {
      expect(screen.getByRole('status', { name: /portfolio value/i })).toHaveTextContent(/\$1,200\.00/) // Filtered portfolio value
      expect(screen.getByRole('status', { name: /total p&l/i })).toHaveTextContent(/\$200\.00/) // Filtered PnL
      expect(screen.getByRole('status', { name: /profitable positions/i })).toHaveTextContent('1 profitable')
    }, { timeout: 3000 })

    // Test percentage change calculation
    const percentageChange = ((1200 - 1000) / 1000) * 100 // (currentValue - initialValue) / initialValue * 100
    await waitFor(() => {
      expect(screen.getByRole('status', { name: /percentage change/i })).toHaveTextContent(`↑ ${percentageChange.toFixed(2)}%`)
    }, { timeout: 3000 })
  })

  it('should handle connection state transitions and retries', async () => {
    let mockRetryCount = 0
    const maxRetries = 3

    // Initial error state
    ;(useWebSocket as jest.Mock).mockImplementation(() => ({
      data: null,
      error: new Error('Connection lost to trading servers'),
      isConnected: false,
      retryCount: mockRetryCount,
      handleError: jest.fn(),
      reconnect: jest.fn(),
      disconnect: jest.fn()
    }))

    renderWithTheme(<TradingDashboard />)
    
    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(/Connection Error/i)
      expect(screen.getByRole('alert')).toHaveTextContent(/Connection lost to trading servers/i)
      expect(screen.getByRole('status', { name: /connection status/i })).toHaveClass('bg-red-500')
    }, { timeout: 2000 })

    // Test retry mechanism with max retries
    for (let i = 0; i < maxRetries; i++) {
      const retryButton = screen.getByRole('button', { name: /Try Again/i })
      fireEvent.click(retryButton)
      mockRetryCount++

      ;(useWebSocket as jest.Mock).mockImplementation(() => ({
        data: null,
        error: new Error('Connection lost to trading servers'),
        isConnected: false,
        retryCount: mockRetryCount
      }))

      await waitFor(() => {
        expect(screen.getByRole('alert')).toHaveTextContent(/Connection Error/i)
        expect(screen.getByRole('alert')).toHaveTextContent(/Connection lost to trading servers/i)
        expect(screen.getByRole('alert')).toHaveTextContent(new RegExp(`Attempt ${mockRetryCount} of ${maxRetries}`, 'i'))
      })
    }

    // Max retries reached
    ;(useWebSocket as jest.Mock).mockImplementation(() => ({
      data: null,
      error: new Error('Connection lost to trading servers'),
      isConnected: false,
      retryCount: maxRetries
    }))

    await waitFor(() => {
      const tryAgainButton = screen.getByRole('button', { name: /Try Again/i })
      expect(tryAgainButton).toBeDisabled()
      expect(screen.getByRole('alert')).toHaveTextContent(/Max Retries Reached/i)
    })

    // Test reset functionality
    const resetButton = screen.getByRole('button', { name: /Reset Connection/i })
    fireEvent.click(resetButton)

    // Verify reset state
    await waitFor(() => {
      const tryAgainButton = screen.getByRole('button', { name: /Try Again/i })
      expect(tryAgainButton).not.toBeDisabled()
    })

    // Successful reconnection after reset
    ;(useWebSocket as jest.Mock).mockImplementation(() => ({
      data: mockPriceHistory,
      error: null,
      isConnected: true,
      retryCount: 0
    }))

    const tryAgainButton = screen.getByRole('button', { name: /Try Again/i })
    fireEvent.click(tryAgainButton)

    await waitFor(() => {
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
      expect(screen.getByRole('heading', { name: /Price Chart/i })).toBeInTheDocument()
    })
  })

  it('should handle partial connection states', async () => {
    // Price feed connected, positions feed disconnected
    ;(useWebSocket as jest.Mock).mockImplementation((url: string) => {
      if (url.includes('prices')) {
        return {
          data: mockPriceHistory,
          error: null,
          isConnected: true,
          retryCount: 0
        }
      }
      return {
        data: null,
        error: new Error('Position feed unavailable'),
        isConnected: false,
        retryCount: 0
      }
    })

    renderWithTheme(<TradingDashboard />)

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(/Position feed unavailable/i)
      expect(screen.getByRole('status', { name: /connection status/i })).toHaveClass('bg-yellow-500')
      const retryButton = screen.getByRole('button', { name: /Try Again/i })
      expect(retryButton).toBeInTheDocument()
    }, { timeout: 3000 })

    // Test retry functionality
    await act(async () => {
      const retryButton = screen.getByRole('button', { name: /Try Again/i })
      fireEvent.click(retryButton)
      jest.runOnlyPendingTimers()
    })

    // Both feeds connected after retry
    ;(useWebSocket as jest.Mock).mockImplementation((url: string) => ({
      data: url.includes('prices') ? mockPriceHistory : mockPositions,
      error: null,
      isConnected: true,
      retryCount: 0
    }))

    await waitFor(() => {
      expect(screen.getByRole('status', { name: /connection status/i })).toHaveClass('bg-green-500')
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
      const chartContainer = screen.getByTestId('price-chart-card')
      expect(within(chartContainer).getByRole('heading', { name: /Price Chart/i })).toBeInTheDocument()
      expect(screen.getByRole('heading', { name: /Portfolio Value/i })).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('should handle error boundary functionality and recovery', async () => {
    const mockError = new Error('Critical trading system error')
    const mockReload = jest.fn()
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { reload: mockReload }
    })

    ;(useWebSocket as jest.Mock).mockImplementation(() => {
      throw mockError
    })

    const { rerender } = renderWithTheme(
      <ErrorBoundary>
        <TradingDashboard />
      </ErrorBoundary>
    )

    // Verify error state and UI
    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /An error occurred in the trading dashboard/i })).toBeInTheDocument()
      expect(screen.getByText(/Critical trading system error/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /Reload Application/i })).toBeInTheDocument()
      expect(screen.getByTestId('status-indicator')).toHaveClass('bg-red-500')
      expect(screen.queryByRole('status', { name: /connection status/i })).not.toBeInTheDocument()
    }, { timeout: 3000 })

    // Test error boundary reset
    await act(async () => {
      const errorContainer = screen.getByTestId('error-container')
      const reloadButton = within(errorContainer).getByRole('button', { name: /Reload Application/i })
      fireEvent.click(reloadButton)
      jest.advanceTimersByTime(1000)
    })
    expect(mockReload).toHaveBeenCalled()

    // Test recovery by re-rendering without error component
    await act(async () => {
      ;(useWebSocket as jest.Mock).mockImplementation(() => ({
        data: mockPriceHistory,
        error: null,
        isConnected: true,
        retryCount: 0
      }))
      rerender(<TradingDashboard />)
      jest.advanceTimersByTime(1000)
    })

    // Mock successful WebSocket connections
    ;(useWebSocket as jest.Mock)
      .mockImplementation(() => ({
        data: mockPriceHistory,
        error: null,
        isConnected: true,
        retryCount: 0
      }))

    // Verify recovery state
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    expect(screen.getByRole('heading', { name: /Price Chart/i })).toBeInTheDocument()
    expect(screen.getByRole('status', { name: /connection status/i })).toHaveClass('bg-green-500')

    // Verify trading functionality is restored
    await act(async () => {
      const symbolSelect = screen.getByRole('combobox', { name: /trading pair/i })
      fireEvent.change(symbolSelect, { target: { value: TRADING_PAIRS[0] } })
      jest.runOnlyPendingTimers()
    })

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /Price Chart/i })).toBeInTheDocument()
      expect(screen.getByRole('heading', { name: /Portfolio Value/i })).toBeInTheDocument()
      expect(screen.getByRole('img', { name: /price chart/i })).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('should handle specific WebSocket error types', async () => {
    const errorTypes = [
      { message: 'Rate limit exceeded', expectation: /Rate limit exceeded/i },
      { message: 'Critical error', expectation: /Critical trading system error/i },
      { message: 'Market data unavailable', expectation: /Connection lost to trading servers/i }
    ]

    for (const { message, expectation } of errorTypes) {
      ;(useWebSocket as jest.Mock).mockImplementation(() => ({
        data: null,
        error: new Error(message),
        isConnected: false,
        retryCount: 0
      }))

      const { unmount } = renderWithTheme(<TradingDashboard />)

      await waitFor(() => {
        const errorContainer = screen.getByTestId('error-container')
        const errorMessage = within(errorContainer).getByRole('alert')
        expect(errorMessage).toHaveTextContent(expectation)
        expect(screen.getByRole('status', { name: /connection status/i })).toHaveClass('bg-red-500')
        const retryButton = within(errorContainer).getByRole('button', { name: /Try Again/i })
        expect(retryButton).toBeInTheDocument()
      }, { timeout: 3000 })

      await act(async () => {
        unmount()
        jest.advanceTimersByTime(1000)
      })
    }
  })

  it('should handle concurrent WebSocket connections and data synchronization', async () => {
    jest.useFakeTimers()
    const mockPriceUpdates = [
      { timestamp: Date.now(), price: 100 },
      { timestamp: Date.now() + 1000, price: 102 },
      { timestamp: Date.now() + 2000, price: 98 }
    ]

    const mockPositionUpdates = [
      {
        symbol: TRADING_PAIRS[0] as TradingPair,
        size: 10,
        entryPrice: 100,
        currentPrice: 102,
        pnl: 20
      },
      {
        symbol: TRADING_PAIRS[0] as TradingPair,
        size: 15,
        entryPrice: 102,
        currentPrice: 98,
        pnl: -60
      }
    ]

    // Initial state with first updates
    await act(async () => {
      ;(useWebSocket as jest.Mock).mockImplementation((url: string) => ({
        data: url.includes('prices') ? mockPriceUpdates[0] : mockPositionUpdates[0],
        error: null,
        isConnected: true,
        retryCount: 0
      }))

      renderWithTheme(<TradingDashboard />)
      jest.advanceTimersByTime(1000)
    })

    // Verify initial state
    await waitFor(() => {
      const chartContainer = screen.getByTestId('price-chart-card')
      expect(within(chartContainer).getByRole('heading', { name: /Price Chart/i })).toBeInTheDocument()
      expect(screen.getByRole('status', { name: /connection status/i })).toHaveClass('bg-green-500')
      const portfolioValue = screen.getByRole('status', { name: /portfolio value/i })
      expect(portfolioValue).toHaveTextContent(/\$1,020\.00/) // 10 shares * $102 current price
      expect(screen.getByRole('status', { name: /total p&l/i })).toHaveTextContent(/\$20\.00/) // (102 - 100) * 10 shares
    }, { timeout: 3000 })

    // Update to second state
    await act(async () => {
      ;(useWebSocket as jest.Mock).mockImplementation((url: string) => ({
        data: url.includes('prices') ? mockPriceUpdates[1] : mockPositionUpdates[1],
        error: null,
        isConnected: true,
        retryCount: 0
      }))
    })
    
    await act(async () => {
      jest.advanceTimersByTime(1000)
    })

    // Verify synchronized updates
    await waitFor(() => {
      expect(screen.getByRole('status', { name: /portfolio value/i })).toHaveTextContent(/\$1,470\.00/)
      expect(screen.getByRole('status', { name: /total p&l/i })).toHaveTextContent(/\$-60\.00/)
      expect(screen.getByRole('status', { name: /profitable positions/i })).toHaveTextContent('0 profitable')
    }, { timeout: 3000 })
  }, 10000) // Increase timeout to 10 seconds

  it('should handle WebSocket reconnection edge cases', async () => {
    let retryCount = 0
    const maxRetries = 3
    
    // Initial connection failure
    ;(useWebSocket as jest.Mock).mockImplementation(() => ({
      data: null,
      error: new Error('Connection lost to trading servers'),
      isConnected: false,
      retryCount,
      maxRetries,
      disconnect: jest.fn()
    }))

    const { unmount } = renderWithTheme(<TradingDashboard />)

    // Verify disconnected state
    await waitFor(() => {
      expect(screen.getByRole('status', { name: /connection status/i })).toHaveClass('bg-red-500')
      const errorContainer = screen.getByTestId('error-container')
      expect(within(errorContainer).getByRole('alert')).toHaveTextContent(/Connection Error/i)
      expect(within(errorContainer).getByText(/Connection lost to trading servers/i)).toBeInTheDocument()
      const retryButton = within(errorContainer).getByRole('button', { name: /Try Again/i })
      expect(retryButton).toBeInTheDocument()
      expect(retryButton).toBeEnabled()
    }, { timeout: 3000 })

    // Simulate partial reconnection (price feed only)
    let mockDisconnect = jest.fn()
    ;(useWebSocket as jest.Mock)
      .mockImplementationOnce(() => ({
        data: mockPriceHistory,
        error: null,
        isConnected: true,
        retryCount: 0,
        maxRetries: 3,
        disconnect: mockDisconnect
      }))
      .mockImplementationOnce(() => ({
        data: null,
        error: new Error('Position feed unavailable'),
        isConnected: false,
        retryCount: 1,
        maxRetries: 3,
        disconnect: mockDisconnect
      }))

    await act(async () => {
      const retryButton = screen.getByRole('button', { name: /Try Again/i })
      fireEvent.click(retryButton)
      await Promise.resolve()
      jest.runOnlyPendingTimers()
    })

    // Verify partial connection state
    await waitFor(() => {
      const statusIndicator = screen.getByRole('status', { name: /connection status/i })
      expect(statusIndicator).toHaveClass('bg-yellow-500')
      const errorContainer = screen.getByTestId('error-container')
      expect(within(errorContainer).getByRole('alert')).toHaveTextContent(/Connection Warning/i)
      expect(within(errorContainer).getByText(/Position feed unavailable/i)).toBeInTheDocument()
      expect(within(errorContainer).getByText(/Attempt 1 of 3/i)).toBeInTheDocument()
      expect(screen.getByRole('heading', { name: /Price Chart/i })).toBeInTheDocument()
    }, { timeout: 3000 })

    // Simulate full reconnection
    ;(useWebSocket as jest.Mock)
      .mockImplementation(() => ({
        data: mockPriceHistory,
        error: null,
        isConnected: true,
        retryCount: 0,
        maxRetries: 3,
        disconnect: mockDisconnect
      }))

    await act(async () => {
      const retryButton = screen.getByRole('button', { name: /Try Again/i })
      fireEvent.click(retryButton)
      await Promise.resolve()
      jest.runOnlyPendingTimers()
    })

    // Verify fully connected state
    await waitFor(() => {
      expect(screen.getByRole('status', { name: /connection status/i })).toHaveClass('bg-green-500')
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
      const chartContainer = screen.getByTestId('price-chart-card')
      expect(within(chartContainer).getByRole('heading', { name: /Price Chart/i })).toBeInTheDocument()
      expect(screen.getByRole('heading', { name: /Portfolio Value/i })).toBeInTheDocument()

      // Test connection stability
      expect(screen.queryByRole('button', { name: /Try Again/i })).not.toBeInTheDocument()
      expect(screen.queryByRole('status', { name: /connecting/i })).not.toBeInTheDocument()
      expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    }, { timeout: 3000 })

    // Cleanup
    await act(async () => {
      unmount()
      jest.advanceTimersByTime(1000)
    })
    expect(mockDisconnect).toHaveBeenCalled()
  })
})
