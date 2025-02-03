import { useState, useEffect, useCallback } from 'react'
import { generateMockPriceHistory } from '../utils/mock-data'
import { PriceData, Position } from '../types'

const RECONNECT_DELAY = 3000

const mockPositions = [
  {
    symbol: 'SOL/USD',
    size: 10,
    entryPrice: 100,
    currentPrice: 102.5,
    pnl: 25
  }
]

type WebSocketState<T> = {
  data: T | undefined
  error: Error | null
  isConnected: boolean
}

type WebSocketCleanup = {
  ws: WebSocket | null
  timeoutId: NodeJS.Timeout | null
  intervalId: NodeJS.Timeout | null
  reconnectTimeoutId: NodeJS.Timeout | null
  isCleanedUp: boolean
  reconnectAttempts: number
}

export function useWebSocket<T>(url: string) {
  const [state, setState] = useState<WebSocketState<T>>({
    data: undefined,
    error: null,
    isConnected: false
  })
  const [shouldReconnect, setShouldReconnect] = useState(true)

  const setData = useCallback((data: T | ((prev: T | undefined) => T)) => {
    setState(prev => ({
      ...prev,
      data: typeof data === 'function' ? (data as ((prev: T | undefined) => T))(prev.data) : data
    }))
  }, [])

  const setError = useCallback((error: Error | null) => 
    setState(prev => ({ ...prev, error })), [])

  const setIsConnected = useCallback((isConnected: boolean) => 
    setState(prev => ({ ...prev, isConnected })), [])

  const connect = useCallback(() => {
    const cleanup: WebSocketCleanup = {
      ws: null,
      timeoutId: null,
      intervalId: null,
      reconnectTimeoutId: null,
      isCleanedUp: false,
      reconnectAttempts: 0
    }
    const MAX_RECONNECT_ATTEMPTS = 5
    
    const cleanupFn = () => {
      if (cleanup.isCleanedUp) return
      cleanup.isCleanedUp = true
      if (cleanup.timeoutId) clearTimeout(cleanup.timeoutId)
      if (cleanup.intervalId) clearInterval(cleanup.intervalId)
      if (cleanup.reconnectTimeoutId) clearTimeout(cleanup.reconnectTimeoutId)
      if (cleanup.ws) {
        cleanup.ws.onclose = null
        cleanup.ws.onerror = null
        cleanup.ws.onmessage = null
        cleanup.ws.onopen = null
        cleanup.ws.close()
        cleanup.ws = null
      }
      setIsConnected(false)
    }

    if (url.includes('prices')) {
      const timeRange = new URL(url).searchParams.get('timeRange') || '24H'
      setIsConnected(true)
      
      // Initial data load
      cleanup.timeoutId = setTimeout(() => {
        if (!cleanup.isCleanedUp) {
          setData(generateMockPriceHistory(timeRange) as T)
          
          // Simulate real-time updates
          cleanup.intervalId = setInterval(() => {
            if (!cleanup.isCleanedUp) {
              setData((prevData) => {
                if (!prevData || !Array.isArray(prevData)) return prevData as T
                const typedData = prevData as PriceData[]
                const lastPrice = typedData[typedData.length - 1].price
                const newPrice = lastPrice * (1 + (Math.random() - 0.5) * 0.002)
                return [...typedData.slice(1), { timestamp: Date.now(), price: newPrice }] as T
              })
            }
          }, 2000)
        }
      }, 100)

      return cleanupFn
    }

    if (url.includes('positions')) {
      setIsConnected(true)
      
      // Initial data load
      cleanup.timeoutId = setTimeout(() => {
        if (!cleanup.isCleanedUp) {
          setData(mockPositions as T)
          
          // Simulate position updates
          cleanup.intervalId = setInterval(() => {
            if (!cleanup.isCleanedUp) {
              setData((prevData) => {
                if (!prevData || !Array.isArray(prevData)) return prevData as T
                const typedData = prevData as Position[]
                return typedData.map(pos => ({
                  ...pos,
                  currentPrice: pos.currentPrice * (1 + (Math.random() - 0.5) * 0.001),
                  pnl: pos.size * (pos.currentPrice * (1 + (Math.random() - 0.5) * 0.001) - pos.entryPrice)
                })) as T
              })
            }
          }, 3000)
        }
      }, 100)

      return cleanupFn
    }

    // Real WebSocket connection
    const setupWebSocket = () => {
      try {
        cleanup.ws = new WebSocket(url)

        cleanup.ws.onopen = () => {
          if (!cleanup.isCleanedUp) {
            setIsConnected(true)
            setError(null)
            cleanup.reconnectAttempts = 0
          }
        }

        cleanup.ws.onclose = () => {
          setIsConnected(false)
          if (shouldReconnect && !cleanup.isCleanedUp) {
            cleanup.reconnectAttempts++
            if (cleanup.reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
              const delay = Math.min(1000 * Math.pow(2, cleanup.reconnectAttempts), 30000)
              cleanup.reconnectTimeoutId = setTimeout(() => {
                if (!cleanup.isCleanedUp) {
                  setupWebSocket()
                }
              }, delay)
            } else {
              setError(new Error('Maximum reconnection attempts reached'))
            }
          }
        }

        cleanup.ws.onmessage = (event: MessageEvent) => {
          if (!cleanup.isCleanedUp) {
            try {
              const parsed = JSON.parse(event.data)
              setData(parsed as T)
              setError(null)
            } catch (err) {
              setError(new Error('Failed to parse WebSocket data'))
              if (cleanup.ws) {
                cleanup.ws.close()
              }
            }
          }
        }

        cleanup.ws.onerror = () => {
          if (!cleanup.isCleanedUp) {
            setError(new Error('WebSocket connection error'))
            setIsConnected(false)
            if (cleanup.ws) {
              cleanup.ws.close()
            }
          }
        }
      } catch (err) {
        setError(err instanceof Error ? err : new Error('Failed to connect to WebSocket'))
        if (shouldReconnect) {
          setTimeout(setupWebSocket, RECONNECT_DELAY)
        }
      }
    }

    setupWebSocket()

    return () => {
      if (cleanup.ws) {
        setShouldReconnect(false)
        cleanup.ws.close()
      }
      cleanupFn()
    }
  }, [url, shouldReconnect])

  const disconnect = useCallback(() => {
    setShouldReconnect(false)
    const cleanup = connect()
    cleanup()
    setData(undefined as T | ((prev: T | undefined) => T))
    setError(null)
    setIsConnected(false)
  }, [connect, setData, setError, setIsConnected])

  useEffect(() => {
    setShouldReconnect(true)
    const cleanup = connect()
    return () => {
      setShouldReconnect(false)
      cleanup()
      setData(undefined as T | ((prev: T | undefined) => T))
      setError(null)
      setIsConnected(false)
    }
  }, [connect, setData, setError, setIsConnected])

  return { 
    data: state.data, 
    error: state.error, 
    isConnected: state.isConnected,
    disconnect
  }
}
