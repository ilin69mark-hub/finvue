# FinVue - Financial Monitoring & Advisory Platform

Платформа мониторинга криптовалют с индикаторами и алертами в реальном времени.

## Быстрый старт

```bash
# Сборка и запуск всех сервисов
docker compose up --build
```

После запуска:
- **Frontend**: http://localhost
- **Backend API**: http://localhost/api
- **PostgreSQL**: localhost:5432 (логин: finvue / finvue_secret)

## Возможности

- ✅ Подключение к Binance Public API
- ✅ Сохранение OHLCV свечей (1m, 1h, 1d) в PostgreSQL
- ✅ REST API для активов, свечей, индикаторов, алертов
- ✅ WebSocket для live цен в реальном времени
- ✅ SMA индикатор (20/50 период) с автоматическими алертами при crossover
- ✅ Интерактивный график TradingView Lightweight Charts
- ✅ Docker Compose для быстрого deployment

## API Endpoints

```
GET /api/v1/assets              # Список активов
GET /api/v1/assets/{id}         # Актив по ID
GET /api/v1/ohlcv?asset_id=1&timeframe=1h&limit=50  # Свечи
GET /api/v1/indicators/sma?asset_id=1  # SMA индикатор
GET /api/v1/alerts               # Список алертов
GET /ws?symbol=BTCUSDT          # WebSocket stream цен
```

## Архитектура

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Nginx     │────▶│   Backend   │────▶│ PostgreSQL  │
│  (Frontend) │     │   (Go 1.26) │     │             │
└─────────────┘     └──────┬──────┘     └─────────────┘
                          │
                    ┌─────┴─────┐
                    │ WebSocket │
                    │   (prices)│
                    └───────────┘
```

## Разработка

### Требования
- Go 1.26+
- Node.js 18+
- Docker & Docker Compose

### Локальный запуск без Docker

**Backend:**
```bash
cd backend
go run cmd/server/main.go
```

**Frontend:**
```bash
cd frontend
npm install
npm run dev
```

## Структура проекта

```
.
├── docker-compose.yml          # Все сервисы
├── backend/
│   ├── Dockerfile              # Go приложение
│   ├── cmd/server/main.go      # Точка входа
│   └── internal/
│       ├── handlers/          # HTTP handlers
│       ├── services/           # Бизнес-логика
│       ├── repositories/       # Работа с БД
│       ├── fetchers/           # Binance API клиент
│       └── websocket/          # WebSocket хаб
└── frontend/
│   ├── Dockerfile              # Nginx + React
│   └── src/
│       ├── pages/              # Dashboard, AssetDetail, Alerts
│       ├── components/         # CandleChart
│       ├── hooks/              # useWebSocket
│       ├── api/                # API клиент
│       └── stores/             # Zustand store
```

## Технологический стек

**Backend:**
- Go 1.26
- chi (роутер)
- pgx (PostgreSQL драйвер)
- gorilla/websocket
- zap (логирование)

**Frontend:**
- React 19
- Vite
- TypeScript
- Tailwind CSS
- lightweight-charts (TradingView)
- Zustand (state management)
- React Router

## Мониторинг

После запуска данные загружаются автоматически:
1. FetcherService раз в минуту получает цены с Binance
2. Сохраняет минутные свечи и агрегирует в часовые/дневные
3. Рассчитывает SMA(20)/SMA(50) для дневных свечей
4. Создаёт алерты при bullish/bearish crossover

Логи смотри в `docker logs finvue-backend`