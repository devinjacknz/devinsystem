import { useEffect, useRef, useState } from 'react'

interface WebSocketOptions {
  onOpen?: () => void
  onMessage?: (data: any) => void
  onClose?: () => void
  onError?: (error: any) => void
}

export function useWebSocket(url: string, options: WebSocketOptions = {}) {
  const ws = useRef<WebSocket | null>(null)
  const [readyState, setReadyState] = useState<number>(WebSocket.CONNECTING)

  useEffect(() => {
    ws.current = new WebSocket(url)

    ws.current.onopen = () => {
      setReadyState(WebSocket.OPEN)
      options.onOpen?.()
    }

    ws.current.onmessage = (event) => {
      let data = event.data
      try {
        if (typeof data === 'string') {
          data = JSON.parse(data)
        }
      } catch (e) {
        // Keep raw data if parsing fails
      }
      options.onMessage?.(data)
    }

    ws.current.onclose = () => {
      setReadyState(WebSocket.CLOSED)
      options.onClose?.()
    }

    ws.current.onerror = (error) => {
      options.onError?.(error)
    }

    return () => {
      if (ws.current) {
        ws.current.close()
      }
    }
  }, [url])

  const send = (data: any) => {
    if (ws.current?.readyState === WebSocket.OPEN) {
      ws.current.send(JSON.stringify(data))
    }
  }

  return {
    send,
    readyState
  }
}
