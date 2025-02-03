import React, { useState, useCallback, useEffect, memo, useMemo } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from './ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from './ui/tabs'
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { useWebSocket } from '../hooks/use-websocket'
import type { PriceData, Position, TimeRange, TradingPair, ActiveSymbol } from '../types'
import { TRADING_PAIRS, TIME_RANGES } from '../types'
import { ErrorBoundary } from './ErrorBoundary'
import { clsx } from 'clsx'

interface TradingDashboardProps {
  initialSymbol?: ActiveSymbol;
  initialTimeRange?: TimeRange;
  onSymbolChange?: (symbol: TradingPair) => void;
  onTimeRangeChange?: (range: TimeRange) => void;
  onError?: (error: Error) => void;
  onConnectionStateChange?: (isConnected: boolean) => void;
}

const formatCurrency = (value: number) => 
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(value)

const formatTimestamp = (timestamp: number) => 
  new Date(timestamp).toLocaleString()

export const TradingDashboard: React.FC<TradingDashboardProps> = memo(({ 
  initialSymbol = TRADING_PAIRS[0],
  initialTimeRange = '24H',
  onSymbolChange,
  onTimeRangeChange,
  onError,
  onConnectionStateChange
}) => {
  const [activeSymbol, setActiveSymbol] = useState<ActiveSymbol>(initialSymbol)
  const [timeRange, setTimeRange] = useState<TimeRange>(initialTimeRange)
  const [isChartLoading, setIsChartLoading] = useState(false)
  const [isPositionsLoading, setIsPositionsLoading] = useState(false)
  const [retryCount, setRetryCount] = useState(0)
  const [hasReachedMaxRetries, setHasReachedMaxRetries] = useState(false)
  const [retryTimeout, setRetryTimeout] = useState(0)
  const MAX_RETRIES = 3
  const RETRY_TIMEOUT = 30
  
  const { data: priceHistory, error: priceError, isConnected: isPriceConnected, disconnect: disconnectPrice } = 
    useWebSocket<PriceData[]>(`ws://localhost:8080/ws/prices?timeRange=${timeRange}&symbol=${encodeURIComponent(activeSymbol)}&retry=${retryCount}`)
  const { data: positions, error: positionsError, isConnected: isPositionsConnected, disconnect: disconnectPositions } = 
    useWebSocket<Position[]>(`ws://localhost:8080/ws/positions?retry=${retryCount}`)

  useEffect(() => {
    return () => {
      disconnectPrice?.()
      disconnectPositions?.()
    }
  }, [disconnectPrice, disconnectPositions])

  useEffect(() => {
    let timeoutId: NodeJS.Timeout | undefined
    const hasError = priceError || positionsError || !isPriceConnected || !isPositionsConnected

    if (hasError) {
      setIsChartLoading(false)
      setIsPositionsLoading(false)
      return
    }

    setIsChartLoading(!hasError)
    setIsPositionsLoading(!hasError)
    
    if (hasError) {
      if (retryCount >= MAX_RETRIES && !hasReachedMaxRetries) {
        setHasReachedMaxRetries(true)
        timeoutId = setTimeout(() => {
          setRetryTimeout(RETRY_TIMEOUT)
        }, 100)
      } else if (!hasReachedMaxRetries) {
        timeoutId = setTimeout(() => {
          setRetryCount(prev => prev + 1)
        }, 1000)
      }
    } else {
      setHasReachedMaxRetries(false)
      setRetryCount(0)
      setRetryTimeout(0)
    }

    return () => {
      if (timeoutId) clearTimeout(timeoutId)
    }
  }, [priceError, positionsError, isPriceConnected, isPositionsConnected, retryCount, hasReachedMaxRetries, MAX_RETRIES, RETRY_TIMEOUT])

  useEffect(() => {
    let timer: NodeJS.Timeout | undefined
    
    if (!isPriceConnected || !isPositionsConnected || priceError || positionsError) {
      setIsChartLoading(false)
    } else {
      setIsChartLoading(true)
      timer = setTimeout(() => setIsChartLoading(false), 100)
    }
    
    return () => {
      if (timer) clearTimeout(timer)
    }
  }, [activeSymbol, timeRange, isPriceConnected, isPositionsConnected, priceError, positionsError])

  useEffect(() => {
    let timer: NodeJS.Timeout | undefined
    const isFullyConnected = isPriceConnected && isPositionsConnected
    
    onConnectionStateChange?.(isFullyConnected)
    
    if (priceError || positionsError) {
      const error = priceError || positionsError
      if (error) onError?.(error)
    }
    
    if (!isFullyConnected || priceError || positionsError) {
      setIsPositionsLoading(false)
    } else {
      setIsPositionsLoading(true)
      timer = setTimeout(() => setIsPositionsLoading(false), 500)
    }
    
    return () => {
      if (timer) clearTimeout(timer)
      setIsPositionsLoading(false)
      setRetryCount(0)
      setHasReachedMaxRetries(false)
      setRetryTimeout(0)
    }
  }, [activeSymbol, isPriceConnected, isPositionsConnected, priceError, positionsError, onConnectionStateChange, onError])

  const handleTimeRangeChange = useCallback((range: TimeRange): void => {
    setTimeRange(range)
    onTimeRangeChange?.(range)
  }, [onTimeRangeChange])

  const handleSymbolChange = useCallback((symbol: TradingPair): void => {
    setActiveSymbol(symbol)
    onSymbolChange?.(symbol)
  }, [onSymbolChange])

  const handleRetry = useCallback(() => {
    if (!hasReachedMaxRetries && retryTimeout === 0) {
      setRetryCount(prev => {
        const newCount = prev + 1
        if (newCount >= MAX_RETRIES) {
          setHasReachedMaxRetries(true)
          setRetryTimeout(RETRY_TIMEOUT)
          onError?.(new Error('Maximum retry attempts reached'))
        }
        return newCount
      })

      setIsChartLoading(true)
      setIsPositionsLoading(true)
      
      if (disconnectPrice) disconnectPrice()
      if (disconnectPositions) disconnectPositions()
    }
  }, [hasReachedMaxRetries, retryTimeout, MAX_RETRIES, RETRY_TIMEOUT, onError, disconnectPrice, disconnectPositions])

  useEffect(() => {
    let timer: NodeJS.Timeout | undefined
    if (retryTimeout > 0) {
      timer = setInterval(() => {
        setRetryTimeout(t => {
          if (t <= 1) {
            return 0
          }
          return t - 1
        })
      }, 1000)
    }
    return () => {
      if (timer) clearInterval(timer)
    }
  }, [retryTimeout])

  const hasError = priceError || positionsError
  const error = priceError || positionsError
  const isConnectionError = !isPriceConnected && !isPositionsConnected
  const isPartialConnection = (isPriceConnected && !isPositionsConnected) || (!isPriceConnected && isPositionsConnected)

  const statusColor = useMemo(() => {
    if (hasReachedMaxRetries || error) return 'bg-red-500'
    if (!isPriceConnected && !isPositionsConnected) return 'bg-red-500'
    if (!isPriceConnected || !isPositionsConnected) return 'bg-yellow-500'
    return 'bg-green-500'
  }, [hasReachedMaxRetries, error, isPriceConnected, isPositionsConnected])
  
  const titleColor = error?.message?.includes('critical') ? 'text-red-500' :
                    !isPriceConnected && !isPositionsConnected ? 'text-red-500' :
                    isPartialConnection ? 'text-yellow-500' : 
                    'text-red-500'

  const getErrorMessage = () => {
    if (hasReachedMaxRetries) {
      return `Maximum Retry Attempts Reached (${MAX_RETRIES}/${MAX_RETRIES})\nPlease try again later`
    }
    
    if (isPartialConnection) {
      return `Connection Warning\n${!isPositionsConnected ? 'Position feed unavailable' : 'Price feed unavailable'}`
    }

    if (error?.message?.includes('rate limit')) {
      return 'Connection Error\nRate limit exceeded'
    }
    
    return `Connection Error\n${retryCount > 0 ? `Reconnecting to trading servers (Attempt ${retryCount}/${MAX_RETRIES})` : 'Connection lost to trading servers'}`
  }

  const renderError = () => {
    if (!hasError) return null
    
    return (
      <Card className="p-4">
        <CardContent className="p-6 pt-0 space-y-4">
          <div aria-live="assertive" className="flex flex-col items-center space-y-2" data-testid="error-container">
              <div 
                data-testid="error-status-indicator" 
                className={`w-2 h-2 rounded-full ${statusColor}`}
                aria-hidden="true"
              />
              <div role="alert" className={`font-medium ${titleColor}`}>
                {getErrorMessage().split('\n').map((line, i) => (
                  <div key={i} className={i === 0 ? '' : 'text-sm text-muted-foreground text-center mt-1'}>
                    {line}
                  </div>
                ))}
              </div>
              <div className="flex flex-col items-center space-y-2">
                <button
                  onClick={handleRetry}
                  className="px-4 py-2 bg-primary text-primary-foreground rounded hover:bg-primary/90"
                  data-testid="retry-button"
                  disabled={retryTimeout > 0 || hasReachedMaxRetries}
                >
                  {retryTimeout > 0 ? `Try Again (${retryTimeout}s)` : 'Try Again'}
                </button>
                {hasReachedMaxRetries && (
                  <>
                    <div className="text-sm text-red-500">Max Retries Reached</div>
                    <button
                      onClick={() => {
                        setRetryCount(0)
                        setHasReachedMaxRetries(false)
                        handleRetry()
                      }}
                      className="px-4 py-2 bg-secondary text-secondary-foreground rounded hover:bg-secondary/90"
                    >
                      Reset Connection
                    </button>
                  </>
                )}
              </div>
            </div>
          </CardContent>
        </Card>
      )
  }



  const isLoading = (!priceHistory && !hasError && !hasReachedMaxRetries && !retryCount) || (!positions && !hasError && !hasReachedMaxRetries && !retryCount) || isChartLoading || isPositionsLoading
  const loadingMessage = !isPriceConnected && !isPositionsConnected ? 'Connecting to trading servers...' :
                        !isPriceConnected ? 'Connecting to price feed...' :
                        !isPositionsConnected ? 'Loading positions...' :
                        isChartLoading ? 'Fetching market data...' :
                        isPositionsLoading ? 'Loading positions...' :
                        !priceHistory ? 'Fetching market data...' : 'Loading Trading System'

  const renderLoading = () => {
    if (!isLoading || !loadingMessage) return null
    
    return (
      <Card className="p-4">
        <CardContent className="flex flex-col items-center justify-center space-y-4">
          <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin" role="progressbar" aria-label="Loading indicator" />
          <div className="text-center">
            <h2 className="text-sm font-medium" role="status" aria-label="Loading Trading System">Loading Trading System</h2>
            <div className="text-xs text-muted-foreground" role="status" aria-label="Loading status">
              {loadingMessage}
            </div>
          </div>
        </CardContent>
      </Card>
    )
  }

  const renderContent = () => {
    if (isLoading && !isPartialConnection) return null

    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <select 
              value={activeSymbol}
              onChange={(e: React.ChangeEvent<HTMLSelectElement>) => handleSymbolChange(e.target.value as TradingPair)}
              className="p-1 rounded bg-background border border-input text-sm"
              data-testid="symbol-select"
              aria-label="Select trading pair"
            >
              {TRADING_PAIRS.map((symbol) => (
                <option key={symbol} value={symbol}>{symbol}</option>
              ))}
            </select>
            <select 
              value={timeRange}
              onChange={(e: React.ChangeEvent<HTMLSelectElement>) => handleTimeRangeChange(e.target.value as TimeRange)}
              className="p-1 rounded bg-background border border-input text-sm"
              data-testid="timerange-select"
              aria-label="Select time range"
            >
              {TIME_RANGES.map((range) => (
                <option key={range} value={range}>{range}</option>
              ))}
            </select>
          </div>
          <div className="flex items-center space-x-2">
            <div className="text-sm text-muted-foreground">
              Last Updated: {priceHistory && priceHistory.length > 0 ? formatTimestamp(priceHistory[priceHistory.length - 1].timestamp) : 'Never'}
            </div>
            <div 
              className={`w-2 h-2 rounded-full ${statusColor}`}
              role="status" aria-label="Connection status"
            />
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <Card className="col-span-4" data-testid="price-chart-card">
            <div className="flex flex-col space-y-1.5 p-6">
              <h2 id="price-chart-title" data-testid="price-chart-title" className="text-2xl font-semibold leading-none tracking-tight">
                Price Chart - {activeSymbol}
              </h2>
              <div className="text-sm text-muted-foreground">
                Historical price data
              </div>
            </div>
            <CardContent>
              <div className="h-[400px]" role="region" aria-label="price chart" aria-labelledby="price-chart-title">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={priceHistory} role="img" aria-label="price chart">
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis 
                      dataKey="timestamp" 
                      tickFormatter={formatTimestamp}
                    />
                    <YAxis 
                      tickFormatter={(value) => formatCurrency(value)}
                    />
                    <Tooltip />
                    <Line
                      type="monotone"
                      dataKey="price"
                      stroke="#2563eb"
                      dot={false}
                      strokeWidth={2}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>

          <Card className="col-span-2">
            <CardHeader>
              <CardTitle>Open Positions</CardTitle>
              <div className="text-sm text-muted-foreground">
                Currently active trading positions
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-8">
                <div className="flex items-center">
                  <div className="ml-4 space-y-1">
                    <p className="text-sm font-medium leading-none">
                      Total Positions
                    </p>
                    <p className="text-sm text-muted-foreground">
                      {Array.isArray(positions) ? positions.length : 0} active
                    </p>
                  </div>
                  <div className="ml-auto font-medium">
                    {Array.isArray(positions) ? positions.length : 0}
                  </div>
                </div>
                <div>
                  <div className="flex items-center">
                    <div className="ml-4 space-y-1">
                      <p className="text-sm font-medium leading-none">
                        Total Value
                      </p>
                      <p className="text-sm text-muted-foreground">
                        Current position values
                      </p>
                    </div>
                    <div className="ml-auto font-medium">
                      {formatCurrency(totalValue)}
                    </div>
                  </div>
                </div>
                <div>
                  <div className="flex items-center">
                    <div className="ml-4 space-y-1">
                      <p className="text-sm font-medium leading-none">
                        Average Position Size
                      </p>
                      <p className="text-sm text-muted-foreground">
                        Mean value per position
                      </p>
                    </div>
                    <div className="ml-auto font-medium">
                      {formatCurrency(filteredPositions.reduce((sum, pos) => sum + (pos.size * pos.currentPrice), 0) / filteredPositions.length)}
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card className="col-span-2">
            <CardHeader>
              <CardTitle>Performance Metrics</CardTitle>
              <div className="text-sm text-muted-foreground">
                Trading performance statistics
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-8">
                <div>
                  <div className="flex items-center">
                    <div className="ml-4 space-y-1">
                      <p className="text-sm font-medium leading-none">
                        Win Rate
                      </p>
                      <p className="text-sm text-muted-foreground">
                        Percentage of profitable trades
                      </p>
                    </div>
                    <div className="ml-auto font-medium">
                      {((filteredPositions.filter(p => p.pnl > 0).length / filteredPositions.length) * 100).toFixed(1)}%
                    </div>
                  </div>
                </div>
                <div>
                  <div className="flex items-center">
                    <div className="ml-4 space-y-1">
                      <p className="text-sm font-medium leading-none">
                        Average Return
                      </p>
                      <p className="text-sm text-muted-foreground">
                        Mean return per trade
                      </p>
                    </div>
                    <div className="ml-auto font-medium">
                      {(filteredPositions.reduce((sum, pos) => sum + ((pos.currentPrice - pos.entryPrice) / pos.entryPrice) * 100, 0) / filteredPositions.length).toFixed(2)}%
                    </div>
                  </div>
                </div>
                <div>
                  <div className="flex items-center">
                    <div className="ml-4 space-y-1">
                      <h3 className="text-sm font-medium text-muted-foreground">Performance</h3>
                      <p 
                        className={clsx(
                          "text-2xl font-bold",
                          totalPnL >= 0 ? "text-green-500" : "text-red-500"
                        )}
                        data-testid="total-pnl-display"
                        role="status"
                        aria-label="Total profit and loss"
                      >
                        {formatCurrency(totalPnL)}
                      </p>
                      <p className="text-sm text-muted-foreground" data-testid="profitable-positions">
                        {Array.isArray(positions) ? positions.filter(p => p.pnl > 0).length : 0} profitable
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    )
  }

  const filteredPositions = useMemo(() => {
    if (!positions || !Array.isArray(positions)) return []
    return activeSymbol === 'all' ? positions : positions.filter(p => p.symbol === activeSymbol)
  }, [positions, activeSymbol])

  interface Stats {
    totalPnL: number;
    totalValue: number;
    percentageChange: number;
  }

  const stats = useMemo<Stats>(() => {
    const baseValue = 2000 // Base portfolio value
    if (!Array.isArray(filteredPositions) || filteredPositions.length === 0) {
      return { totalPnL: 0, totalValue: baseValue, percentageChange: 0 }
    }
    
    const totalValue = filteredPositions.reduce((sum, pos) => 
      sum + (pos.size * pos.currentPrice), baseValue)
    const totalPnL = filteredPositions.reduce((sum, pos) => 
      sum + (pos.size * (pos.currentPrice - pos.entryPrice)), 0)
    const initialValue = totalValue - totalPnL
    const percentageChange = initialValue > 0 ? (totalPnL / initialValue) * 100 : 0
    
    return { 
      totalPnL: Number(totalPnL.toFixed(2)),
      totalValue: Number(totalValue.toFixed(2)),
      percentageChange: Number(percentageChange.toFixed(2))
    }
  }, [filteredPositions])

  const { totalPnL = 0, totalValue = 0, percentageChange = 0 } = stats || {}

  return (
    <ErrorBoundary>
      <div className="space-y-6">
        {renderError()}
        {renderLoading()}
        {renderContent()}
        <div className="grid grid-cols-3 gap-4">
          <Card>
            <CardContent className="pt-6">
              <div className="flex flex-col space-y-2">
                <h3 className="text-sm font-medium text-muted-foreground" role="heading" aria-level={3}>Portfolio Value</h3>
                <p className="text-2xl font-bold" role="status" aria-label="Portfolio value">{formatCurrency(totalValue)}</p>
                <p className={clsx(
                  "text-sm",
                  percentageChange >= 0 ? "text-green-500" : "text-red-500"
                )} role="status" aria-label="Portfolio change">
                  {percentageChange >= 0 ? "↑" : "↓"} {Math.abs(percentageChange).toFixed(2)}%
                </p>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="flex flex-col space-y-2">
                <h3 className="text-sm font-medium text-muted-foreground" role="heading" aria-level={3}>Total P&L</h3>
                <p className={clsx(
                  "text-2xl font-bold",
                  totalPnL >= 0 ? "text-green-500" : "text-red-500"
                )} role="status" aria-label="Total profit and loss">
                  {formatCurrency(totalPnL)}
                </p>
                <p className="text-sm text-muted-foreground" role="status" aria-label="Profitable positions">
                  {Array.isArray(positions) ? positions.filter(p => p.pnl > 0).length : 0} profitable trades
                </p>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <div className="flex flex-col space-y-2">
                <h3 className="text-sm font-medium text-muted-foreground" role="heading" aria-level={3}>Active Positions</h3>
                <p className="text-2xl font-bold" role="status" aria-label="Number of active positions">{positions?.length ?? 0}</p>
                <p className="text-sm text-muted-foreground" role="status" aria-label="Available trading pairs">
                  {TRADING_PAIRS.length} pairs available
                </p>
              </div>
            </CardContent>
          </Card>
        </div>
        <Tabs defaultValue="trading" className="w-full" aria-label="Trading dashboard sections">
          <TabsList className="grid w-full grid-cols-2" aria-label="Dashboard views">
            <TabsTrigger value="trading" aria-controls="trading-tab">Trading</TabsTrigger>
            <TabsTrigger value="positions" aria-controls="positions-tab">Positions</TabsTrigger>
          </TabsList>
          <TabsContent value="trading" id="trading-tab" role="tabpanel">
            <Card>
              <CardHeader>
                <CardTitle>
                  <div className="flex flex-col space-y-4 w-full">
                    <div className="flex justify-between items-center">
                      <div className="flex items-center space-x-4">
                        <span>Price Chart - {activeSymbol}</span>
                        <select 
                          value={activeSymbol}
                          onChange={(e: React.ChangeEvent<HTMLSelectElement>) => handleSymbolChange(e.target.value as TradingPair)}
                          className="p-1 rounded bg-background border border-input text-sm"
                          aria-label="Select trading pair"
                        >
                          {TRADING_PAIRS.map(symbol => (
                            <option key={symbol} value={symbol}>{symbol}</option>
                          ))}
                        </select>
                      </div>
                      <div className="flex items-center space-x-2">
                        <button
                          onClick={handleRetry}
                          className="p-2 rounded hover:bg-accent"
                          aria-label="Refresh data"
                        >
                          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" role="img" aria-hidden="true">
                            <path d="M21 2v6h-6"></path><path d="M3 12a9 9 0 0 1 15-6.7L21 8"></path>
                            <path d="M3 22v-6h6"></path><path d="M21 12a9 9 0 0 1-15 6.7L3 16"></path>
                          </svg>
                        </button>
                        <span className="text-xs text-muted-foreground" role="status" aria-label="Last update time">
                          Last updated: {priceHistory && priceHistory.length > 0 ? formatTimestamp(priceHistory[priceHistory.length - 1].timestamp) : 'Never'}
                        </span>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2" role="group" aria-label="Time range selection">
                      {(['1H', '24H', '7D', '30D'] as const).map((range) => (
                        <button
                          key={range}
                          onClick={() => handleTimeRangeChange(range)}
                          className={`px-2 py-1 text-sm rounded ${
                            timeRange === range 
                              ? 'bg-primary text-primary-foreground' 
                              : 'hover:bg-accent'
                          }`}
                          aria-pressed={timeRange === range}
                          aria-label={`${range} time range`}
                        >
                          {range}
                        </button>
                      ))}
                    </div>
                  </div>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="h-[400px] relative">
                  {(!priceHistory || isChartLoading) && (
                    <div className="absolute inset-0 flex items-center justify-center bg-background/50">
                      <div className="flex flex-col items-center space-y-4">
                        <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin" />
                        <p className="text-sm text-muted-foreground">Loading price data...</p>
                      </div>
                    </div>
                  )}
                  <div className="w-full h-[400px]" data-testid="chart-container">
                    <ResponsiveContainer width="100%" height="100%" data-testid="responsive-container">
                      <LineChart data={priceHistory || []} data-testid="line-chart">
                        <CartesianGrid strokeDasharray="3 3" strokeOpacity={0.5} data-testid="cartesian-grid" />
                        <XAxis 
                          dataKey="timestamp" 
                          tickFormatter={formatTimestamp}
                          stroke="#888888"
                          data-testid="x-axis"
                        />
                        <YAxis 
                          tickFormatter={formatCurrency}
                          stroke="#888888"
                          data-testid="y-axis"
                        />
                        <Tooltip 
                          formatter={(value: number) => formatCurrency(value)}
                          labelFormatter={formatTimestamp}
                          data-testid="tooltip"
                        />
                        <Line 
                          type="monotone" 
                          dataKey="price" 
                          stroke="#2563eb"
                          strokeWidth={2}
                          dot={false}
                          data-testid="price-line"
                        />
                      </LineChart>
                    </ResponsiveContainer>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
          <TabsContent value="positions" forceMount id="positions-tab" role="tabpanel">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle role="heading" aria-level={2}>Open Positions</CardTitle>
                <select 
                  role="combobox"
                  value={activeSymbol}
                  onChange={(e: React.ChangeEvent<HTMLSelectElement>) => handleSymbolChange(e.target.value as TradingPair)}
                  className="p-1 rounded bg-background border border-input text-sm"
                  aria-label="Filter trading pairs"
                >
                  <option value="all">All Pairs</option>
                  {TRADING_PAIRS.map(symbol => (
                    <option key={symbol} value={symbol}>{symbol}</option>
                  ))}
                </select>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4" role="list" aria-label="Trading positions">
                  {isPositionsLoading ? (
                    <div className="text-center p-4" role="status" aria-label="Loading positions">
                      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto" role="progressbar" />
                    </div>
                  ) : filteredPositions?.map((position) => (
                    <Card key={position.symbol} role="listitem">
                      <CardContent className="p-4">
                        <div className="flex justify-between items-center">
                          <div>
                            <h3 className="font-bold" role="heading" aria-level={3}>{position.symbol}</h3>
                            <p className="text-sm text-muted-foreground" role="status" aria-label={`Position size for ${position.symbol}`}>
                              Size: {position.size}
                            </p>
                          </div>
                          <div className="text-right space-y-1">
                            <p className="text-sm" role="status" aria-label={`Entry price for ${position.symbol}`}>Entry: {formatCurrency(position.entryPrice)}</p>
                            <p className="text-sm" role="status" aria-label={`Current price for ${position.symbol}`}>Current: {formatCurrency(position.currentPrice)}</p>
                            <p className={`font-bold ${position.pnl >= 0 ? 'text-green-500' : 'text-red-500'}`} role="status" aria-label={`Profit and loss for ${position.symbol}`}>
                              PnL: {formatCurrency(position.pnl)}
                            </p>
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                  {!isPositionsLoading && (!filteredPositions || filteredPositions.length === 0) && (
                    <div className="text-center p-4 text-muted-foreground" role="status" aria-label="No positions">
                      No open positions found
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </ErrorBoundary>
  )
})
