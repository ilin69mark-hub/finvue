import axios from 'axios'
import type { Asset, OHLCVResponse, Alert, OHLCVParams, AssetListParams } from '../types/api'

const API_BASE_URL = ''

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.response.use(
  response => response,
  error => {
    console.error('API Error:', error.response?.data || error.message)
    return Promise.reject(error)
  }
)

export const getAssets = async (params?: AssetListParams): Promise<Asset[]> => {
  const response = await api.get('/api/v1/assets', { params })
  return response.data
}

export const getAssetById = async (id: number): Promise<Asset> => {
  const response = await api.get(`/api/v1/assets/${id}`)
  return response.data
}

export const getOHLCV = async (params: OHLCVParams): Promise<OHLCVResponse> => {
  const response = await api.get('/api/v1/ohlcv', { params })
  return response.data
}

export const getAlerts = async (): Promise<Alert[]> => {
  const response = await api.get('/api/v1/alerts')
  return response.data
}

export const createAlert = async (alert: Partial<Alert>): Promise<Alert> => {
  const response = await api.post('/api/v1/alerts', alert)
  return response.data
}

export const markAlertRead = async (id: number): Promise<void> => {
  await api.patch(`/api/v1/alerts/${id}/read`)
}

export const deleteAlert = async (id: number): Promise<void> => {
  await api.delete(`/api/v1/alerts/${id}`)
}

export default api