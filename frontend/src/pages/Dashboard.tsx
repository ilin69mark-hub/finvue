import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'

interface Asset {
  id: number
  symbol: string
  name: string
  asset_type: string
  last_price: number | null
}

export default function Dashboard() {
  const [assets, setAssets] = useState<Asset[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetch('/api/v1/assets')
      .then(res => {
        if (!res.ok) throw new Error('API error')
        return res.json()
      })
      .then(data => {
        setAssets(data)
        setLoading(false)
      })
      .catch(err => {
        console.error('Error loading assets:', err)
        setError(err.message)
        setLoading(false)
      })
  }, [])

  if (loading) return <div className="p-6 text-center">Загрузка активов...</div>
  if (error) return <div className="p-6 text-center text-red-500">Ошибка: {error}</div>

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-6 text-gray-800">Мониторинг криптовалют</h1>
      <p className="text-gray-500 mb-4">Выберите актив для просмотра графика:</p>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {assets.map(asset => (
          <Link
            key={asset.id}
            to={`/asset/${asset.id}`}
            className="block p-4 border border-gray-200 rounded-lg hover:bg-blue-50 hover:border-blue-300 transition"
          >
            <div className="flex justify-between items-center">
              <div>
                <span className="font-semibold text-lg text-gray-800">{asset.symbol}</span>
                <span className="ml-2 text-gray-500 text-sm">{asset.name}</span>
              </div>
              {asset.last_price ? (
                <span className="text-green-600 font-medium">${asset.last_price.toFixed(2)}</span>
              ) : (
                <span className="text-gray-400">—</span>
              )}
            </div>
          </Link>
        ))}
      </div>

      {assets.length === 0 && (
        <div className="text-center text-gray-500 mt-8">
          Нет активов. Подождите, идет загрузка данных...
        </div>
      )}
    </div>
  )
}