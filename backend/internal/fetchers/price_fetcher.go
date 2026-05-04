package fetchers

import (
	"context"
	"time"

	"finvue/internal/models"
)

type Ticker struct {
	Symbol          string    `json:"symbol"`
	Price           float64   `json:"price"`
	PriceChange24h  float64   `json:"price_change_24h"`
	Volume24h       float64   `json:"volume_24h"`
	High24h         float64   `json:"high_24h"`
	Low24h          float64   `json:"low_24h"`
	LastUpdateTime  time.Time `json:"last_update_time"`
}

type PriceFetcher interface {
	GetSupportedAssets(ctx context.Context) ([]models.Asset, error)
	GetCurrentPrice(ctx context.Context, symbol string) (*Ticker, error)
	GetAllPrices(ctx context.Context) ([]Ticker, error)
	GetRecentCandles(ctx context.Context, symbol string, timeframe models.Timeframe, limit int) ([]models.OHLCV, error)
}

type PriceService struct {
	fetcher PriceFetcher
}

func NewPriceService(fetcher PriceFetcher) *PriceService {
	return &PriceService{fetcher: fetcher}
}

func (s *PriceService) GetSupportedAssets(ctx context.Context) ([]models.Asset, error) {
	return s.fetcher.GetSupportedAssets(ctx)
}

func (s *PriceService) GetCurrentPrice(ctx context.Context, symbol string) (*Ticker, error) {
	return s.fetcher.GetCurrentPrice(ctx, symbol)
}

func (s *PriceService) GetAllPrices(ctx context.Context) ([]Ticker, error) {
	return s.fetcher.GetAllPrices(ctx)
}

func (s *PriceService) GetRecentCandles(ctx context.Context, symbol string, timeframe models.Timeframe, limit int) ([]models.OHLCV, error) {
	return s.fetcher.GetRecentCandles(ctx, symbol, timeframe, limit)
}