import { renderHook, act } from '@testing-library/react'
import { useWebSocket } from '../use-websocket'
import '@testing-library/jest-dom'

declare global {
  interface WebSocket {
    close(): void;
    send(data: string): void;
    onopen: ((this: WebSocket, ev: Event) => any) | null;
    onclose: ((this: WebSocket, ev: CloseEvent) => any) | null;
    onmessage: ((this: WebSocket, ev: MessageEvent<any>) => any) | null;
    onerror: ((this: WebSocket, ev: Event) => any) | null;
  }
  var WebSocket: {
    prototype: WebSocket;
    new(url: string | URL, protocols?: string | string[]): WebSocket;
    readonly CONNECTING: 0;
    readonly OPEN: 1;
    readonly CLOSING: 2;
    readonly CLOSED: 3;
  };
}

describe('useWebSocket', () => {
  let mockWebSocket: any
  const mockUrl = 'ws://localhost:8080/ws/test'

  beforeEach(() => {
    jest.useFakeTimers()
    mockWebSocket = {
      close: jest.fn(),
      send: jest.fn(),
      onopen: null,
      onclose: null,
      onmessage: null,
      onerror: null,
    }
    global.WebSocket = jest.fn(() => mockWebSocket) as any
  })

  afterEach(() => {
    jest.useRealTimers()
    jest.clearAllMocks()
  })

  it('should initialize with default state', () => {
    const { result } = renderHook(() => useWebSocket(mockUrl))
    expect(result.current.data).toBeUndefined()
    expect(result.current.error).toBeNull()
    expect(result.current.isConnected).toBeFalsy()
  })

  it('should handle successful connection', () => {
    const { result } = renderHook(() => useWebSocket(mockUrl))
    
    act(() => {
      mockWebSocket.onopen()
    })

    expect(result.current.isConnected).toBeTruthy()
    expect(result.current.error).toBeNull()
  })

  it('should handle connection error', () => {
    const { result } = renderHook(() => useWebSocket(mockUrl))
    
    act(() => {
      mockWebSocket.onerror(new Event('error'))
    })

    expect(result.current.isConnected).toBeFalsy()
    expect(result.current.error).toBeTruthy()
  })

  it('should handle message reception', () => {
    const mockData = { test: 'data' }
    const { result } = renderHook(() => useWebSocket(mockUrl))
    
    act(() => {
      mockWebSocket.onmessage({ data: JSON.stringify(mockData) })
    })

    expect(result.current.data).toEqual(mockData)
    expect(result.current.error).toBeNull()
  })

  it('should handle reconnection attempts', () => {
    const url = 'ws://localhost:8080/ws/test'
    const { result } = renderHook(() => useWebSocket(url))

    act(() => {
      mockWebSocket.onclose(new CloseEvent('close'))
    })

    expect(result.current.isConnected).toBeFalsy()
    
    act(() => {
      jest.advanceTimersByTime(5000)
    })

    expect(global.WebSocket).toHaveBeenCalledTimes(2)
  })

  it('should handle message parsing errors', () => {
    const url = 'ws://localhost:8080/ws/test'
    const { result } = renderHook(() => useWebSocket(url))

    act(() => {
      mockWebSocket.onmessage({ data: 'invalid json' })
    })

    expect(result.current.error).toBeTruthy()
    expect(result.current.data).toBeUndefined()
  })

  it('should cleanup WebSocket on unmount', () => {
    const url = 'ws://localhost:8080/ws/test'
    const { unmount } = renderHook(() => useWebSocket(url))

    unmount()

    expect(mockWebSocket.close).toHaveBeenCalled()
  })

  it('should handle connection timeout', () => {
    const url = 'ws://localhost:8080/ws/test'
    const { result } = renderHook(() => useWebSocket(url))

    act(() => {
      jest.advanceTimersByTime(10000)
      mockWebSocket.onerror(new Event('error'))
    })

    expect(result.current.error).toBeTruthy()
    expect(result.current.error?.message).toContain('WebSocket connection error')
  })

  it('should cleanup on unmount', () => {
    const { unmount } = renderHook(() => useWebSocket(mockUrl))
    unmount()
    expect(mockWebSocket.close).toHaveBeenCalled()
  })
})
