import { renderHook, act } from '@testing-library/react'
import { useWebSocket } from '../websocket/useWebSocket'

class MockWebSocket {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSING = 2
  static CLOSED = 3

  public readyState: number
  public onopen: ((ev: any) => void) | null
  public onclose: ((ev: any) => void) | null
  public onmessage: ((ev: any) => void) | null
  public onerror: ((ev: any) => void) | null
  
  constructor(public url: string) {
    this.readyState = MockWebSocket.CONNECTING
    this.onopen = null
    this.onclose = null
    this.onmessage = null
    this.onerror = null
  }

  send = jest.fn()
  close = jest.fn()
}

describe('useWebSocket', () => {
  const mockUrl = 'ws://localhost:8080'
  let mockWebSocket: MockWebSocket

  beforeEach(() => {
    mockWebSocket = new MockWebSocket(mockUrl)
    ;(global as any).WebSocket = jest.fn(() => mockWebSocket)
    ;(global as any).WebSocket.CONNECTING = MockWebSocket.CONNECTING
    ;(global as any).WebSocket.OPEN = MockWebSocket.OPEN
    ;(global as any).WebSocket.CLOSING = MockWebSocket.CLOSING
    ;(global as any).WebSocket.CLOSED = MockWebSocket.CLOSED
  })

  afterEach(() => {
    jest.clearAllMocks()
  })

  it('establishes WebSocket connection', () => {
    const { result } = renderHook(() => useWebSocket(mockUrl))
    expect(global.WebSocket).toHaveBeenCalledWith(mockUrl)
    expect(result.current.readyState).toBe(MockWebSocket.CONNECTING)
  })

  it('handles connection open', () => {
    const onOpen = jest.fn()
    const { result } = renderHook(() => useWebSocket(mockUrl, { onOpen }))

    act(() => {
      mockWebSocket.readyState = MockWebSocket.OPEN
      mockWebSocket.onopen?.({})
    })

    expect(onOpen).toHaveBeenCalled()
    expect(result.current.readyState).toBe(MockWebSocket.OPEN)
  })

  it('handles messages', () => {
    const onMessage = jest.fn()
    renderHook(() => useWebSocket(mockUrl, { onMessage }))

    const message = { type: 'price', data: { value: 100 } }
    act(() => {
      mockWebSocket.onmessage?.({ data: JSON.stringify(message) })
    })

    expect(onMessage).toHaveBeenCalledWith(message)
  })

  it('handles connection close', () => {
    const onClose = jest.fn()
    const { result } = renderHook(() => useWebSocket(mockUrl, { onClose }))

    act(() => {
      mockWebSocket.readyState = MockWebSocket.CLOSED
      mockWebSocket.onclose?.({})
    })

    expect(onClose).toHaveBeenCalled()
    expect(result.current.readyState).toBe(MockWebSocket.CLOSED)
  })

  it('handles errors', () => {
    const onError = jest.fn()
    renderHook(() => useWebSocket(mockUrl, { onError }))

    const error = new Error('Connection failed')
    act(() => {
      mockWebSocket.onerror?.(error)
    })

    expect(onError).toHaveBeenCalledWith(error)
  })

  it('sends messages correctly', () => {
    const { result } = renderHook(() => useWebSocket(mockUrl))
    const message = { type: 'subscribe', data: { symbol: 'SOL/USDC' } }

    act(() => {
      mockWebSocket.readyState = MockWebSocket.OPEN
      mockWebSocket.onopen?.({})
      result.current.send(message)
    })

    expect(mockWebSocket.send).toHaveBeenCalledWith(JSON.stringify(message))
  })

  it('cleans up on unmount', () => {
    const { unmount } = renderHook(() => useWebSocket(mockUrl))
    unmount()
    expect(mockWebSocket.close).toHaveBeenCalled()
  })

  it('handles reconnection', () => {
    jest.useFakeTimers()
    const { result } = renderHook(() => useWebSocket(mockUrl))

    act(() => {
      mockWebSocket.readyState = MockWebSocket.CLOSED
      mockWebSocket.onclose?.({})
      jest.advanceTimersByTime(5000)
    })

    expect(global.WebSocket).toHaveBeenCalledTimes(2)
    expect(result.current.readyState).toBe(MockWebSocket.CONNECTING)
    jest.useRealTimers()
  })

  it('handles binary messages', () => {
    const onMessage = jest.fn()
    renderHook(() => useWebSocket(mockUrl, { onMessage }))

    const binaryData = new Blob(['test data'])
    act(() => {
      mockWebSocket.onmessage?.({ data: binaryData })
    })

    expect(onMessage).toHaveBeenCalledWith(binaryData)
  })
})
