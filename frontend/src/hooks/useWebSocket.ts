import { useEffect, useRef, useCallback, useState } from 'react'
import type { Ticker } from '../types/api'

interface UseWebSocketOptions {
  symbol?: string
  onMessage?: (data: Ticker) => void
  onConnect?: () => void
  onDisconnect?: () => void
  onError?: (error: Event) => void
  reconnectAttempts?: number
  reconnectInterval?: number
}

interface UseWebSocketReturn {
  isConnected: boolean
  lastMessage: Ticker | null
  connect: () => void
  disconnect: () => void
  send: (message: object) => void
}

export function useWebSocket(options: UseWebSocketOptions = {}): UseWebSocketReturn {
  const {
    symbol,
    onMessage,
    onConnect,
    onDisconnect,
    onError,
    reconnectAttempts = 5,
    reconnectInterval = 3000,
  } = options

  const [isConnected, setIsConnected] = useState(false)
  const [lastMessage, setLastMessage] = useState<Ticker | null>(null)

  const wsRef = useRef<WebSocket | null>(null)
  const attemptsRef = useRef(0)
  const reconnectTimeoutRef = useRef<number | null>(null)
  const connectRef = useRef<() => void>(() => {})

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      return
    }

    const wsURL = `ws://${window.location.host}/ws${symbol ? `?symbol=${symbol}` : ''}`
    const ws = new WebSocket(wsURL)

    ws.onopen = () => {
      setIsConnected(true)
      attemptsRef.current = 0
      onConnect?.()
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'price_update' && data.payload) {
          const ticker: Ticker = {
            symbol: data.payload.symbol,
            price: data.payload.price,
            price_change_24h: 0,
            volume_24h: 0,
            high_24h: 0,
            low_24h: 0,
            last_update_time: new Date().toISOString(),
          }
          setLastMessage(ticker)
          onMessage?.(ticker)
        }
      } catch (err) {
        console.error('Error parsing WebSocket message:', err)
      }
    }

    ws.onclose = () => {
      setIsConnected(false)
      onDisconnect?.()

      if (attemptsRef.current < reconnectAttempts) {
        attemptsRef.current++
        console.log(`Reconnecting... (attempt ${attemptsRef.current}/${reconnectAttempts})`)
        reconnectTimeoutRef.current = window.setTimeout(() => {
          connectRef.current()
        }, reconnectInterval)
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      onError?.(error)
    }

    wsRef.current = ws
  }, [symbol, onMessage, onConnect, onDisconnect, onError, reconnectAttempts, reconnectInterval])

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current)
    }
    attemptsRef.current = reconnectAttempts

    if (wsRef.current) {
      wsRef.current.close()
      wsRef.current = null
    }
    setIsConnected(false)
  }, [reconnectAttempts])

  const send = useCallback((message: object) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message))
    }
  }, [])

  useEffect(() => {
    return () => {
      disconnect()
    }
  }, [disconnect])

  useEffect(() => {
    connectRef.current = connect
  }, [connect])

  useEffect(() => {
    connect()
  }, [connect])

  return {
    isConnected,
    lastMessage,
    connect,
    disconnect,
    send,
  }
}