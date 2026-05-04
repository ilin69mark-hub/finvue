package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"finvue/internal/fetchers"
	"finvue/internal/models"
	"finvue/internal/repositories"
	"finvue/internal/pkg/logger"
	"finvue/internal/websocket"

	"go.uber.org/zap"
)

type FetcherService struct {
	fetcher         fetchers.PriceFetcher
	assetRepo       *repositories.AssetRepository
	ohlcvRepo       *repositories.OHLCVRepository
	alertRepo       *repositories.AlertRepository
	indicatorService *IndicatorService
	interval        time.Duration
	stopCh          chan struct{}
	wg              sync.WaitGroup
	isRunning       bool
	mu              sync.RWMutex
	lastSyncTime    time.Time
}

func NewFetcherService(
	fetcher fetchers.PriceFetcher,
	assetRepo *repositories.AssetRepository,
	ohlcvRepo *repositories.OHLCVRepository,
	alertRepo *repositories.AlertRepository,
	interval time.Duration,
) *FetcherService {
	if interval == 0 {
		interval = time.Minute
	}

	indicatorService := NewIndicatorService(ohlcvRepo, assetRepo)

	return &FetcherService{
		fetcher:          fetcher,
		assetRepo:        assetRepo,
		ohlcvRepo:        ohlcvRepo,
		alertRepo:        alertRepo,
		indicatorService: indicatorService,
		interval:         interval,
		stopCh:           make(chan struct{}),
	}
}

func (s *FetcherService) Start(ctx context.Context) {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		logger.Warn("FetcherService уже запущен")
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run(ctx)

	logger.Info("FetcherService запущен", zap.Duration("interval", s.interval))
}

func (s *FetcherService) Stop() {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return
	}
	s.mu.Unlock()

	close(s.stopCh)
	s.wg.Wait()

	s.mu.Lock()
	s.isRunning = false
	s.mu.Unlock()

	logger.Info("FetcherService остановлен")
}

func (s *FetcherService) run(ctx context.Context) {
	defer s.wg.Done()

	s.syncOnce(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("FetcherService: контекст отменён, выходим")
			return
		case <-s.stopCh:
			logger.Info("FetcherService: получен сигнал остановки")
			return
		case <-ticker.C:
			s.syncOnce(ctx)
		}
	}
}

const maxAssetsToProcess = 30

func (s *FetcherService) syncOnce(ctx context.Context) {
	logger.Debug("Начало синхронизации данных")

	assets, err := s.fetcher.GetSupportedAssets(ctx)
	if len(assets) > maxAssetsToProcess {
		assets = assets[:maxAssetsToProcess]
	}
	if err != nil {
		logger.Error("Ошибка получения списка активов", zap.Error(err))
		return
	}

	if len(assets) == 0 {
		logger.Warn("Не получено активов от Binance")
		return
	}

	logger.Debug("Получены активы от Binance", zap.Int("count", len(assets)))

	savedCount := 0
	priceUpdateCount := 0

	for _, asset := range assets {
		select {
		case <-ctx.Done():
			logger.Info("FetcherService: контекст отменён во время синхронизации")
			return
		case <-s.stopCh:
			logger.Info("FetcherService: остановка во время синхронизации")
			return
		default:
		}

		savedAsset, err := s.assetRepo.UpsertFromSymbol(ctx, asset.Symbol, asset.Name, asset.AssetType)
		if err != nil {
			logger.Error("Ошибка сохранения актива", zap.String("symbol", asset.Symbol), zap.Error(err))
			continue
		}

		savedCount++

		ticker, err := s.fetcher.GetCurrentPrice(ctx, asset.Symbol)
		if err != nil {
			logger.Debug("Ошибка получения цены", zap.String("symbol", asset.Symbol), zap.Error(err))
			continue
		}

		savedAsset.SetLastPrice(ticker.Price)
		if err := s.assetRepo.Update(ctx, savedAsset); err != nil {
			logger.Error("Ошибка обновления цены актива", zap.String("symbol", asset.Symbol), zap.Error(err))
		} else {
			priceUpdateCount++
		}

		if hub := websocket.GetGlobalHub(); hub != nil {
			hub.BroadcastPriceUpdate(asset.Symbol, ticker.Price)
		}

		candles, err := s.fetcher.GetRecentCandles(ctx, asset.Symbol, models.Timeframe1M, 60)
		if err != nil {
			logger.Debug("Ошибка получения свечей", zap.String("symbol", asset.Symbol), zap.Error(err))
			continue
		}

		if len(candles) > 0 {
			for i := range candles {
				candles[i].AssetID = savedAsset.ID
			}

			if err := s.ohlcvRepo.BatchInsert(ctx, "ohlcv_1m", candles); err != nil {
				logger.Error("Ошибка сохранения свечей", zap.String("symbol", asset.Symbol), zap.Error(err))
			}

			s.aggregateHigherTimeframes(ctx, savedAsset.ID, asset.Symbol)

			s.checkCrossoverAndAlert(ctx, savedAsset.ID)
		}
	}

	s.lastSyncTime = time.Now()
	logger.Info("Синхронизация завершена",
		zap.Int("assets_found", len(assets)),
		zap.Int("assets_saved", savedCount),
		zap.Int("prices_updated", priceUpdateCount))
}

func (s *FetcherService) GetLastSyncTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastSyncTime
}

func (s *FetcherService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

func (s *FetcherService) ForceSync(ctx context.Context) error {
	if !s.IsRunning() {
		return fmt.Errorf("сервис не запущен")
	}
	s.syncOnce(ctx)
	return nil
}

func (s *FetcherService) aggregateHigherTimeframes(ctx context.Context, assetID int64, symbol string) {
	if err := s.aggregateTimeframe(ctx, assetID, "ohlcv_1m", "ohlcv_1h", time.Hour); err != nil {
		logger.Error("Ошибка агрегации в часовые свечи", zap.String("symbol", symbol), zap.Error(err))
	}

	if err := s.aggregateTimeframe(ctx, assetID, "ohlcv_1m", "ohlcv_1d", 24*time.Hour); err != nil {
		logger.Error("Ошибка агрегации в дневные свечи", zap.String("symbol", symbol), zap.Error(err))
	}
}

func (s *FetcherService) aggregateTimeframe(ctx context.Context, assetID int64, sourceTable, targetTable string, interval time.Duration) error {
	latestTarget, err := s.ohlcvRepo.GetLatestCandle(ctx, assetID, targetTable)
	if err != nil {
		return fmt.Errorf("ошибка получения последней целевой свечи: %w", err)
	}

	var from time.Time
	if latestTarget != nil {
		from = latestTarget.Timestamp.Add(interval)
	} else {
		firstCandle, err := s.ohlcvRepo.GetFirstCandle(ctx, assetID, sourceTable)
		if err != nil {
			return fmt.Errorf("ошибка получения первой исходной свечи: %w", err)
		}
		if firstCandle == nil {
			return nil
		}
		from = firstCandle.Timestamp
	}

	to := from.Add(interval * 2)
	if to.After(time.Now()) {
		to = time.Now()
	}

	if from.After(to) {
		return nil
	}

	sourceCandles, err := s.ohlcvRepo.GetCandlesInRange(ctx, sourceTable, assetID, from, to)
	if err != nil {
		return fmt.Errorf("ошибка получения исходных свечей: %w", err)
	}

	if len(sourceCandles) == 0 {
		return nil
	}

	var targetCandles []models.OHLCV
	bucketStart := sourceCandles[0].Timestamp.Truncate(interval)

	for _, c := range sourceCandles {
		candleTime := c.Timestamp.Truncate(interval)
		if !candleTime.Equal(bucketStart) {
			if built := s.buildCandleFromBucket(targetCandles); built != nil {
				built.AssetID = assetID
				built.Timestamp = bucketStart
				targetCandles = append(targetCandles, *built)
			}
			bucketStart = candleTime
			targetCandles = []models.OHLCV{}
		}
		targetCandles = append(targetCandles, c)
	}

	if len(targetCandles) > 0 {
		if built := s.buildCandleFromBucket(targetCandles); built != nil {
			built.AssetID = assetID
			built.Timestamp = bucketStart
			targetCandles = append(targetCandles, *built)
		}
	}

	if len(targetCandles) > 0 {
		if err := s.ohlcvRepo.BatchInsert(ctx, targetTable, targetCandles); err != nil {
			return fmt.Errorf("ошибка вставки агрегированных свечей: %w", err)
		}
		logger.Debug("Агрегированные свечи сохранены",
			zap.String("target_table", targetTable),
			zap.Int("count", len(targetCandles)))
	}

	return nil
}

func (s *FetcherService) buildCandleFromBucket(candles []models.OHLCV) *models.OHLCV {
	if len(candles) == 0 {
		return nil
	}

	open := candles[0].Open
	high := candles[0].High
	low := candles[0].Low
	close := candles[len(candles)-1].Close
	var volume float64

	for _, c := range candles {
		if c.High > high {
			high = c.High
		}
		if c.Low < low {
			low = c.Low
		}
		volume += c.Volume
	}

	return &models.OHLCV{
		Open:   open,
		High:   high,
		Low:    low,
		Close:  close,
		Volume: volume,
	}
}

func (s *FetcherService) checkCrossoverAndAlert(ctx context.Context, assetID int64) {
	smaResp, err := s.indicatorService.CalculateSMA(ctx, SMARequest{
		AssetID:   assetID,
		Timeframe: "1d",
		FastPeriod: 20,
		SlowPeriod: 50,
	})
	if err != nil {
		logger.Debug("Ошибка расчёта SMA для алерта", zap.Error(err))
		return
	}

	if smaResp == nil || smaResp.Crossover == "insufficient_data" || smaResp.Crossover == "neutral" {
		return
	}

	if smaResp.Crossover == "bullish" || smaResp.Crossover == "bearish" {
		lastAlert, err := s.alertRepo.GetLastByAssetAndType(ctx, assetID, models.AlertTypeSMACrossover)
		if err == nil && lastAlert != nil && time.Since(lastAlert.CreatedAt) < 24*time.Hour {
			return
		}

		crossoverType := "бычье"
		alertType := models.AlertTypeSMACrossover
		if smaResp.Crossover == "bearish" {
			crossoverType = "медвежье"
			alertType = models.AlertTypeSMACrossover
		}

		alert := &models.Alert{
			AssetID:   assetID,
			AlertType: alertType,
			Message:   fmt.Sprintf("%s пересечение SMA: SMA20=$%.2f, SMA50=$%.2f", crossoverType, smaResp.FastSMA, smaResp.SlowSMA),
			Value:     smaResp.FastSMA,
			Threshold: smaResp.SlowSMA,
			IsRead:    false,
		}

		if err := s.alertRepo.Create(ctx, alert); err != nil {
			logger.Error("Ошибка создания алерта", zap.Error(err))
		} else {
			logger.Info("Создан алерт crossover", zap.String("symbol", smaResp.Symbol), zap.String("type", smaResp.Crossover))
		}
	}
}