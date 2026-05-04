export interface Asset {
  id: number
  symbol: string
  name: string
  asset_type: string
  is_active: boolean
  last_price: number | null
  last_price_updated: string | null
}

export interface OHLCV {
  id: number
  asset_id: number
  timestamp: string
  open: number
  high: number
  low: number
  close: number
  volume: number
}

export interface OHLCVResponse {
  asset: Asset
  timeframe: string
  candles: OHLCV[]
}

export interface Alert {
  id: number
  asset_id: number
  alert_type: string
  message: string
  value: number | null
  threshold: number | null
  is_read: boolean
  created_at: string
}

export interface Ticker {
  symbol: string
  price: number
  price_change_24h: number
  volume_24h: number
  high_24h: number
  low_24h: number
  last_update_time: string
}

export type Timeframe = '1m' | '1h' | '1d'

export interface OHLCVParams {
  asset_id: number
  timeframe?: Timeframe
  from?: string
  to?: string
  limit?: number
}

export interface AssetListParams {
  include_inactive?: boolean
}