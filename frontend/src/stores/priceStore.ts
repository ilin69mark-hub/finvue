import { create } from 'zustand'
import type { Ticker } from '../types/api'

interface PriceState {
  prices: Record<string, Ticker>
  lastUpdate: string | null

  setPrice: (ticker: Ticker) => void
  setPrices: (tickers: Ticker[]) => void
  getPrice: (symbol: string) => Ticker | undefined
  clear: () => void
}

export const usePriceStore = create<PriceState>((set, get) => ({
  prices: {},
  lastUpdate: null,

  setPrice: (ticker: Ticker) => {
    set((state) => ({
      prices: {
        ...state.prices,
        [ticker.symbol]: ticker,
      },
      lastUpdate: new Date().toISOString(),
    }))
  },

  setPrices: (tickers: Ticker[]) => {
    const prices: Record<string, Ticker> = {}
    tickers.forEach((ticker) => {
      prices[ticker.symbol] = ticker
    })
    set({
      prices,
      lastUpdate: new Date().toISOString(),
    })
  },

  getPrice: (symbol: string) => {
    return get().prices[symbol]
  },

  clear: () => {
    set({ prices: {}, lastUpdate: null })
  },
}))

interface AlertState {
  alerts: Array<{
    id: number
    assetId: number
    type: string
    message: string
    isRead: boolean
    createdAt: string
  }>

  addAlert: (alert: AlertState['alerts'][0]) => void
  markRead: (id: number) => void
  removeAlert: (id: number) => void
  clear: () => void
}

export const useAlertStore = create<AlertState>((set) => ({
  alerts: [],

  addAlert: (alert) => {
    set((state) => ({
      alerts: [alert, ...state.alerts],
    }))
  },

  markRead: (id) => {
    set((state) => ({
      alerts: state.alerts.map((a) =>
        a.id === id ? { ...a, isRead: true } : a
      ),
    }))
  },

  removeAlert: (id) => {
    set((state) => ({
      alerts: state.alerts.filter((a) => a.id !== id),
    }))
  },

  clear: () => {
    set({ alerts: [] })
  },
}))