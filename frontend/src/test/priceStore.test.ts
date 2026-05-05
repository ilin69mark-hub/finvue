import { describe, it, expect, beforeEach } from 'vitest'
import { usePriceStore, useAlertStore } from '../stores/priceStore'

describe('usePriceStore', () => {
  beforeEach(() => {
    usePriceStore.getState().clear()
  })

  it('should set and get price', () => {
    const ticker = { symbol: 'BTCUSDT', price: 42000 }
    usePriceStore.getState().setPrice(ticker)
    const price = usePriceStore.getState().getPrice('BTCUSDT')
    expect(price).toEqual(ticker)
  })

  it('should set multiple prices', () => {
    const tickers = [
      { symbol: 'BTCUSDT', price: 42000 },
      { symbol: 'ETHUSDT', price: 3000 },
    ]
    usePriceStore.getState().setPrices(tickers)
    expect(usePriceStore.getState().prices['BTCUSDT'].price).toBe(42000)
    expect(usePriceStore.getState().prices['ETHUSDT'].price).toBe(3000)
  })

  it('should update lastUpdate on setPrice', () => {
    const ticker = { symbol: 'BTCUSDT', price: 42000 }
    usePriceStore.getState().setPrice(ticker)
    expect(usePriceStore.getState().lastUpdate).not.toBeNull()
  })

  it('should clear all prices', () => {
    usePriceStore.getState().setPrice({ symbol: 'BTCUSDT', price: 42000 })
    usePriceStore.getState().clear()
    expect(usePriceStore.getState().prices).toEqual({})
    expect(usePriceStore.getState().lastUpdate).toBeNull()
  })
})

describe('useAlertStore', () => {
  beforeEach(() => {
    useAlertStore.getState().clear()
  })

  it('should add alert', () => {
    const alert = { id: 1, assetId: 1, type: 'warning', message: 'Test', isRead: false, createdAt: '2024-01-01' }
    useAlertStore.getState().addAlert(alert)
    expect(useAlertStore.getState().alerts).toHaveLength(1)
  })

  it('should mark alert as read', () => {
    const alert = { id: 1, assetId: 1, type: 'warning', message: 'Test', isRead: false, createdAt: '2024-01-01' }
    useAlertStore.getState().addAlert(alert)
    useAlertStore.getState().markRead(1)
    expect(useAlertStore.getState().alerts[0].isRead).toBe(true)
  })

  it('should remove alert', () => {
    const alert = { id: 1, assetId: 1, type: 'warning', message: 'Test', isRead: false, createdAt: '2024-01-01' }
    useAlertStore.getState().addAlert(alert)
    useAlertStore.getState().removeAlert(1)
    expect(useAlertStore.getState().alerts).toHaveLength(0)
  })

  it('should clear all alerts', () => {
    useAlertStore.getState().addAlert({ id: 1, assetId: 1, type: 'warning', message: 'Test', isRead: false, createdAt: '2024-01-01' })
    useAlertStore.getState().clear()
    expect(useAlertStore.getState().alerts).toHaveLength(0)
  })
})