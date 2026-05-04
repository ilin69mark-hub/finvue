import { BrowserRouter, Routes, Route, Link } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import AssetDetail from './pages/AssetDetail'
import Alerts from './pages/Alerts'

function App() {
  return (
    <BrowserRouter>
      <div className="min-h-screen bg-white">
        <nav className="bg-gray-800 text-white p-4">
          <div className="container mx-auto flex gap-4">
            <Link to="/" className="hover:text-gray-300">Dashboard</Link>
            <Link to="/alerts" className="hover:text-gray-300">Alerts</Link>
          </div>
        </nav>

        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/asset/:id" element={<AssetDetail />} />
          <Route path="/alerts" element={<Alerts />} />
        </Routes>
      </div>
    </BrowserRouter>
  )
}

export default App