package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"finvue/internal/fetchers"
	"finvue/internal/handlers"
	"finvue/internal/pkg/config"
	"finvue/internal/pkg/database"
	"finvue/internal/pkg/logger"
	"finvue/internal/repositories"
	"finvue/internal/services"
	"finvue/internal/websocket"

	"go.uber.org/zap"
)

const (
	shutdownTimeout    = 30 * time.Second
	fetcherStopTimeout = 10 * time.Second
)

func runMigrations(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS assets (
			id SERIAL PRIMARY KEY,
			symbol VARCHAR(20) NOT NULL UNIQUE,
			name VARCHAR(100) NOT NULL,
			asset_type VARCHAR(20) NOT NULL DEFAULT 'crypto',
			is_active BOOLEAN NOT NULL DEFAULT true,
			last_price DECIMAL(20, 8),
			last_price_updated TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS ohlcv_1m (
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
		)`,
		`CREATE TABLE IF NOT EXISTS ohlcv_1h (
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
		)`,
		`CREATE TABLE IF NOT EXISTS ohlcv_1d (
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
		)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id SERIAL PRIMARY KEY,
			asset_id INTEGER NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
			alert_type VARCHAR(50) NOT NULL,
			message TEXT NOT NULL,
			value DECIMAL(20, 8),
			threshold DECIMAL(20, 8),
			is_read BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_assets_symbol ON assets(symbol)`,
		`CREATE INDEX IF NOT EXISTS idx_ohlcv_1m_asset_time ON ohlcv_1m(asset_id, timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_ohlcv_1h_asset_time ON ohlcv_1h(asset_id, timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_ohlcv_1d_asset_time ON ohlcv_1d(asset_id, timestamp DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_asset ON alerts(asset_id)`,
	}

	for _, q := range queries {
		if _, err := database.Pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("ошибка миграции: %w", err)
		}
	}
	return nil
}

func main() {
	if err := logger.Init(true); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Запуск FinVue приложения")

	cfg := config.Load()
	logger.Info("Конфигурация загружена", zap.String("port", cfg.Server.Port))

	if err := database.Connect(&cfg.Database); err != nil {
		logger.Fatal("Не удалось подключиться к БД", zap.Error(err))
	}
	defer database.Close()

	ctx := context.Background()
	logger.Info("Применение миграций...")
	if err := runMigrations(ctx); err != nil {
		logger.Warn("Ошибка миграций", zap.Error(err))
	} else {
		logger.Info("Миграции применены")
	}

	websocket.InitGlobalHub()
	logger.Info("WebSocket хаб инициализирован")

	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()
	alertRepo := repositories.NewAlertRepository()

	binanceFetcher := fetchers.NewBinanceFetcher()

	fetcherService := services.NewFetcherService(
		binanceFetcher,
		assetRepo,
		ohlcvRepo,
		alertRepo,
		time.Minute,
	)

	appCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fetcherService.Start(appCtx)
	logger.Info("FetcherService запущен")

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	go func() {
		logger.Info("HTTP сервер запущен", zap.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Ошибка запуска HTTP сервера", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Info("Получен сигнал завершения", zap.String("signal", sig.String()))

	logger.Info("Начинаем graceful shutdown...")

	logger.Info("Отмена контекста...")
	cancel()

	logger.Info("Остановка FetcherService...", zap.Duration("timeout", fetcherStopTimeout))
	stopped := make(chan struct{})
	go func() {
		fetcherService.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
		logger.Info("FetcherService остановлен")
	case <-time.After(fetcherStopTimeout):
		logger.Warn("FetcherService не остановлен вовремя")
	}

	logger.Info("Остановка HTTP сервера...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Ошибка при shutdown HTTP сервера", zap.Error(err))
	}

	logger.Info("Закрытие подключения к БД...")
	database.Close()

	logger.Info("Graceful shutdown завершён")
}