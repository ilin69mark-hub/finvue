import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import CandleChart from '../components/CandleChart'
import { useWebSocket } from '../hooks/useWebSocket'
import { getAssetById, getOHLCV } from '../api/client'
import type { Asset, OHLCV, Timeframe } from '../types/api'

export default function AssetDetail() {
  const { id } = useParams<{ id: string }>()
  const [asset, setAsset] = useState<Asset | null>(null)
  const [candles, setCandles] = useState<OHLCV[]>([])
  const [timeframe, setTimeframe] = useState<Timeframe>('1h')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [currentPrice, setCurrentPrice] = useState<number | undefined>(undefined)

  const { lastMessage } = useWebSocket({
    symbol: asset?.symbol,
    onMessage: (ticker) => {
      setCurrentPrice(ticker.price)
    },
  })

  useEffect(() => {
    if (!id) return

    let cancelled = false

    const fetchData = async () => {
      try {
        const [assetData, ohlcvData] = await Promise.all([
          getAssetById(Number(id)),
          getOHLCV({ asset_id: Number(id), timeframe, limit: 50 })
        ])

        if (cancelled) return

        if (!assetData || Object.keys(assetData).length === 0) {
          setError('Актив не найден в базе данных')
        } else {
          setAsset(assetData)
          setCandles(ohlcvData.candles || [])
          if (assetData.last_price) {
            setCurrentPrice(assetData.last_price)
          }
        }
      } catch (err) {
        if (!cancelled) {
          console.error('Error loading data:', err)
          setError(err instanceof Error ? err.message : 'Unknown error')
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    // eslint-disable-next-line react-hooks/set-state-in-effect
    setLoading(true)
     
    setError(null)
    fetchData()

    return () => {
      cancelled = true
    }
  }, [id, timeframe])

  if (loading) return <div className="p-6 text-center">Загрузка...</div>
  if (error) return (
    <div className="p-6">
      <Link to="/" className="text-blue-500 hover:underline mb-4 inline-block">← Назад</Link>
      <div className="text-red-500 p-4 bg-red-50 rounded">{error}</div>
    </div>
  )
  if (!asset) return (
    <div className="p-6">
      <Link to="/" className="text-blue-500 hover:underline mb-4 inline-block">← Назад</Link>
      <div className="text-gray-500">Актив не найден</div>
    </div>
  )

  return (
    <div className="p-6">
      <Link to="/" className="text-blue-500 hover:underline mb-4 inline-block">← Назад к Dashboard</Link>

      <div className="mb-4">
        <h1 className="text-3xl font-bold text-gray-800">{asset.symbol}</h1>
        <p className="text-gray-500">{asset.name}</p>
        <p className="text-2xl mt-2 font-medium">
          Цена: ${currentPrice?.toFixed(2) || '—'}
          {lastMessage && <span className="ml-2 text-green-500 text-sm">(live)</span>}
        </p>
      </div>

      <div className="mb-4">
        <label className="mr-2 text-gray-600">Таймфрейм:</label>
        <select
          value={timeframe}
          onChange={e => setTimeframe(e.target.value as Timeframe)}
          className="border p-2 rounded bg-white"
        >
          <option value="1m">1 минута</option>
          <option value="1h">1 час</option>
          <option value="1d">1 день</option>
        </select>
      </div>

      <div className="border rounded-lg overflow-hidden bg-white">
        {candles.length > 0 ? (
          <CandleChart data={candles} height={400} lastPrice={currentPrice} />
        ) : (
          <div className="p-8 text-center text-gray-500">
            Нет данных для графика
          </div>
        )}
      </div>

      <div className="mt-6">
        <h2 className="text-lg font-semibold mb-2 text-gray-700">Последние свечи</h2>
        <div className="overflow-x-auto">
          <table className="min-w-full border bg-white">
            <thead>
              <tr className="bg-gray-100">
                <th className="border p-2 text-left">Время</th>
                <th className="border p-2 text-right">Open</th>
                <th className="border p-2 text-right">High</th>
                <th className="border p-2 text-right">Low</th>
                <th className="border p-2 text-right">Close</th>
                <th className="border p-2 text-right">Volume</th>
              </tr>
            </thead>
            <tbody>
              {candles.slice(0, 10).map(c => (
                <tr key={c.id} className="hover:bg-gray-50">
                  <td className="border p-2 text-sm">
                    {new Date(c.timestamp).toLocaleString()}
                  </td>
                  <td className="border p-2 text-right">{c.open.toFixed(2)}</td>
                  <td className="border p-2 text-right">{c.high.toFixed(2)}</td>
                  <td className="border p-2 text-right">{c.low.toFixed(2)}</td>
                  <td className="border p-2 text-right">{c.close.toFixed(2)}</td>
                  <td className="border p-2 text-right">{c.volume.toFixed(2)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}