import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'

interface Alert {
  id: number
  asset_id: number
  alert_type: string
  message: string
  is_read: boolean
  created_at: string
}

export default function Alerts() {
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('http://localhost:8080/api/v1/alerts')
      .then(res => {
        if (res.ok) return res.json()
        return []
      })
      .then(data => {
        setAlerts(data)
        setLoading(false)
      })
      .catch(() => {
        setLoading(false)
      })
  }, [])

  if (loading) return <div className="p-4">Загрузка...</div>

  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">Alerts</h1>

      {alerts.length === 0 ? (
        <div className="text-gray-500">
          <p>Уведомлений пока нет</p>
          <p className="mt-2">Алерты появятся при пересечении SMA (Этап 30)</p>
        </div>
      ) : (
        <div className="space-y-2">
          {alerts.map(alert => (
            <div
              key={alert.id}
              className={`p-4 border rounded ${alert.is_read ? 'bg-gray-50' : 'bg-white border-blue-500'}`}
            >
              <div className="flex justify-between">
                <span className="font-semibold">{alert.alert_type}</span>
                <span className="text-sm text-gray-500">
                  {new Date(alert.created_at).toLocaleString()}
                </span>
              </div>
              <p className="mt-1">{alert.message}</p>
              <Link to={`/asset/${alert.asset_id}`} className="text-blue-500 text-sm">
                Перейти к активу →
              </Link>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}