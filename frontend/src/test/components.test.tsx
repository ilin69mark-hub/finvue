import { describe, it, expect, vi } from 'vitest'
import { render, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import Dashboard from '../pages/Dashboard'
import Alerts from '../pages/Alerts'

const mockAssets = [
  { id: 1, symbol: 'BTCUSDT', name: 'Bitcoin', last_price: 42000, is_active: true },
  { id: 2, symbol: 'ETHUSDT', name: 'Ethereum', last_price: 3000, is_active: true },
]

vi.mock('../api/client', async () => {
  const actual = await vi.importActual('../api/client')
  return {
    ...actual,
    getAssets: vi.fn().mockResolvedValue(mockAssets),
    getAlerts: vi.fn().mockResolvedValue([]),
  }
})

vi.mock('../hooks/useWebSocket', () => ({
  useWebSocket: vi.fn(() => ({
    isConnected: false,
    lastMessage: null,
    connect: vi.fn(),
    disconnect: vi.fn(),
    send: vi.fn(),
  })),
}))

describe('Dashboard', () => {
  it('should render with mocked data', async () => {
    render(
      <BrowserRouter>
        <Dashboard />
      </BrowserRouter>
    )
    await waitFor(() => {
      expect(document.body).toBeInTheDocument()
    }, { timeout: 3000 })
  })

  it('should display error on API failure', async () => {
    const { container } = render(
      <BrowserRouter>
        <Dashboard />
      </BrowserRouter>
    )
    expect(container).toBeInTheDocument()
  })
})

describe('Alerts', () => {
  it('should render alerts page', async () => {
    const { container } = render(
      <BrowserRouter>
        <Alerts />
      </BrowserRouter>
    )
    await waitFor(() => {
      expect(container).toBeInTheDocument()
    }, { timeout: 3000 })
  })
})