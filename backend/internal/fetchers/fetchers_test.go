package fetchers

import (
	"context"
	"testing"

	"finvue/internal/models"
)

type mockFetcher struct{}

func (m *mockFetcher) GetSupportedAssets(ctx context.Context) ([]models.Asset, error) {
	return []models.Asset{
		{Symbol: "BTCUSDT", Name: "Bitcoin", AssetType: models.AssetTypeCrypto},
		{Symbol: "ETHUSDT", Name: "Ethereum", AssetType: models.AssetTypeCrypto},
	}, nil
}

func (m *mockFetcher) GetCurrentPrice(ctx context.Context, symbol string) (*Ticker, error) {
	return &Ticker{
		Symbol:         symbol,
		Price:          42000,
		PriceChange24h: 2.5,
		Volume24h:      1000000,
		High24h:        43000,
		Low24h:         41000,
	}, nil
}

func (m *mockFetcher) GetAllPrices(ctx context.Context) ([]Ticker, error) {
	return []Ticker{
		{Symbol: "BTCUSDT", Price: 42000},
		{Symbol: "ETHUSDT", Price: 3000},
	}, nil
}

func (m *mockFetcher) GetRecentCandles(ctx context.Context, symbol string, timeframe models.Timeframe, limit int) ([]models.OHLCV, error) {
	return []models.OHLCV{
		{AssetID: 1, Open: 42000, High: 43000, Low: 41000, Close: 42500},
	}, nil
}

func TestPriceService_New(t *testing.T) {
	fetcher := &mockFetcher{}
	service := NewPriceService(fetcher)
	if service == nil {
		t.Error("Expected non-nil service")
	}
}

func TestPriceService_GetSupportedAssets(t *testing.T) {
	fetcher := &mockFetcher{}
	service := NewPriceService(fetcher)

	ctx := context.Background()
	assets, err := service.GetSupportedAssets(ctx)
	if err != nil {
		t.Errorf("GetSupportedAssets error: %v", err)
	}
	if len(assets) != 2 {
		t.Errorf("Expected 2 assets, got %d", len(assets))
	}
}

func TestPriceService_GetCurrentPrice(t *testing.T) {
	fetcher := &mockFetcher{}
	service := NewPriceService(fetcher)

	ctx := context.Background()
	ticker, err := service.GetCurrentPrice(ctx, "BTCUSDT")
	if err != nil {
		t.Errorf("GetCurrentPrice error: %v", err)
	}
	if ticker == nil {
		t.Error("Expected non-nil ticker")
	}
	if ticker.Price != 42000 {
		t.Errorf("Expected price 42000, got %f", ticker.Price)
	}
}

func TestPriceService_GetAllPrices(t *testing.T) {
	fetcher := &mockFetcher{}
	service := NewPriceService(fetcher)

	ctx := context.Background()
	tickers, err := service.GetAllPrices(ctx)
	if err != nil {
		t.Errorf("GetAllPrices error: %v", err)
	}
	if len(tickers) != 2 {
		t.Errorf("Expected 2 tickers, got %d", len(tickers))
	}
}

func TestPriceService_GetRecentCandles(t *testing.T) {
	fetcher := &mockFetcher{}
	service := NewPriceService(fetcher)

	ctx := context.Background()
	candles, err := service.GetRecentCandles(ctx, "BTCUSDT", models.Timeframe1H, 100)
	if err != nil {
		t.Errorf("GetRecentCandles error: %v", err)
	}
	if len(candles) != 1 {
		t.Errorf("Expected 1 candle, got %d", len(candles))
	}
}

func TestTicker_Fields(t *testing.T) {
	ticker := Ticker{
		Symbol:         "BTCUSDT",
		Price:          42000,
		PriceChange24h: 2.5,
		Volume24h:      1000000,
		High24h:        43000,
		Low24h:         41000,
	}

	if ticker.Symbol != "BTCUSDT" {
		t.Errorf("Expected Symbol BTCUSDT, got %s", ticker.Symbol)
	}
	if ticker.Price != 42000 {
		t.Errorf("Expected Price 42000, got %f", ticker.Price)
	}
}

func TestPriceService_ImplementsInterface(t *testing.T) {
	var fetcher PriceFetcher = &mockFetcher{}
	if fetcher == nil {
		t.Error("mockFetcher should implement PriceFetcher")
	}
}