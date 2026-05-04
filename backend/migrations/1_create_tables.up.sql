-- Таблица активов (криптовалюты, акции и т.д.)
CREATE TABLE IF NOT EXISTS assets (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    asset_type VARCHAR(20) NOT NULL DEFAULT 'crypto',
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_price DECIMAL(20, 8),
    last_price_updated TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assets_symbol ON assets(symbol);
CREATE INDEX idx_assets_type ON assets(asset_type);
CREATE INDEX idx_assets_active ON assets(is_active) WHERE is_active = true;

-- Таблица минутных свечей (1m)
CREATE TABLE IF NOT EXISTS ohlcv_1m (
    id BIGSERIAL PRIMARY KEY,
    asset_id INTEGER NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL,
    open DECIMAL(20, 8) NOT NULL,
    high DECIMAL(20, 8) NOT NULL,
    low DECIMAL(20, 8) NOT NULL,
    close DECIMAL(20, 8) NOT NULL,
    volume DECIMAL(30, 8) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(asset_id, timestamp)
);

CREATE INDEX idx_ohlcv_1m_asset_time ON ohlcv_1m(asset_id, timestamp DESC);
CREATE INDEX idx_ohlcv_1m_timestamp ON ohlcv_1m(timestamp DESC);

-- Таблица часовых свечей (1h)
CREATE TABLE IF NOT EXISTS ohlcv_1h (
    id BIGSERIAL PRIMARY KEY,
    asset_id INTEGER NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL,
    open DECIMAL(20, 8) NOT NULL,
    high DECIMAL(20, 8) NOT NULL,
    low DECIMAL(20, 8) NOT NULL,
    close DECIMAL(20, 8) NOT NULL,
    volume DECIMAL(30, 8) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(asset_id, timestamp)
);

CREATE INDEX idx_ohlcv_1h_asset_time ON ohlcv_1h(asset_id, timestamp DESC);
CREATE INDEX idx_ohlcv_1h_timestamp ON ohlcv_1h(timestamp DESC);

-- Таблица дневных свечей (1d)
CREATE TABLE IF NOT EXISTS ohlcv_1d (
    id BIGSERIAL PRIMARY KEY,
    asset_id INTEGER NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    timestamp TIMESTAMP NOT NULL,
    open DECIMAL(20, 8) NOT NULL,
    high DECIMAL(20, 8) NOT NULL,
    low DECIMAL(20, 8) NOT NULL,
    close DECIMAL(20, 8) NOT NULL,
    volume DECIMAL(30, 8) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(asset_id, timestamp)
);

CREATE INDEX idx_ohlcv_1d_asset_time ON ohlcv_1d(asset_id, timestamp DESC);
CREATE INDEX idx_ohlcv_1d_timestamp ON ohlcv_1d(timestamp DESC);

-- Таблица алертов/уведомлений
CREATE TABLE IF NOT EXISTS alerts (
    id SERIAL PRIMARY KEY,
    asset_id INTEGER NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    alert_type VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    value DECIMAL(20, 8),
    threshold DECIMAL(20, 8),
    is_read BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alerts_asset ON alerts(asset_id);
CREATE INDEX idx_alerts_read ON alerts(is_read) WHERE is_read = false;
CREATE INDEX idx_alerts_created ON alerts(created_at DESC);