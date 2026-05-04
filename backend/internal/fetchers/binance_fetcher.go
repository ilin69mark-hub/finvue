package fetchers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"finvue/internal/models"
	"finvue/internal/pkg/logger"

	"go.uber.org/zap"
)

const (
	BinanceBaseURL   = "https://api.binance.com"
	BinanceAPITimeout = 30 * time.Second
)

type BinanceFetcher struct {
	client *http.Client
}

func NewBinanceFetcher() *BinanceFetcher {
	return &BinanceFetcher{
		client: &http.Client{
			Timeout: BinanceAPITimeout,
		},
	}
}

type BinanceSymbol struct {
	Symbol      string `json:"symbol"`
	BaseAsset   string `json:"baseAsset"`
	QuoteAsset  string `json:"quoteAsset"`
	Status      string `json:"status"`
}

type BinanceTicker struct {
	Symbol             string `json:"symbol"`
	LastPrice          string `json:"lastPrice"`
	PriceChange        string `json:"priceChange"`
	Volume             string `json:"volume"`
	HighPrice          string `json:"highPrice"`
	LowPrice           string `json:"lowPrice"`
}

type BinanceKline struct {
	OpenTime        int64    `json:"OpenTime"`
	Open            string   `json:"Open"`
	High            string   `json:"High"`
	Low             string   `json:"Low"`
	Close           string   `json:"Close"`
	Volume          string   `json:"Volume"`
	CloseTime       int64    `json:"CloseTime"`
}

func (f *BinanceFetcher) GetSupportedAssets(ctx context.Context) ([]models.Asset, error) {
	url := BinanceBaseURL + "/api/v3/exchangeInfo"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неверный статус ответа: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела ответа: %w", err)
	}

	var info struct {
		Symbols []BinanceSymbol `json:"symbols"`
	}

	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	var assets []models.Asset
	seenSymbols := make(map[string]bool)

	for _, sym := range info.Symbols {
		if sym.Status != "TRADING" {
			continue
		}
		if sym.QuoteAsset != "USDT" {
			continue
		}
		if sym.BaseAsset == "USDT" {
			continue
		}

		symbol := sym.BaseAsset + sym.QuoteAsset
		if seenSymbols[symbol] {
			continue
		}
		seenSymbols[symbol] = true

		assets = append(assets, models.Asset{
			Symbol:    symbol,
			Name:      sym.BaseAsset,
			AssetType: models.AssetTypeCrypto,
			IsActive:  true,
		})
	}

	logger.Debug("Загружены активы с Binance", zap.Int("count", len(assets)))
	return assets, nil
}

func (f *BinanceFetcher) GetCurrentPrice(ctx context.Context, symbol string) (*Ticker, error) {
	url := BinanceBaseURL + "/api/v3/ticker/24hr?symbol=" + symbol

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неверный статус ответа: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела ответа: %w", err)
	}

	var ticker BinanceTicker
	if err := json.Unmarshal(body, &ticker); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	price, _ := strconv.ParseFloat(ticker.LastPrice, 64)
	priceChange, _ := strconv.ParseFloat(ticker.PriceChange, 64)
	volume, _ := strconv.ParseFloat(ticker.Volume, 64)
	high, _ := strconv.ParseFloat(ticker.HighPrice, 64)
	low, _ := strconv.ParseFloat(ticker.LowPrice, 64)

	return &Ticker{
		Symbol:          ticker.Symbol,
		Price:           price,
		PriceChange24h:  priceChange,
		Volume24h:       volume,
		High24h:         high,
		Low24h:          low,
		LastUpdateTime:  time.Now(),
	}, nil
}

func (f *BinanceFetcher) GetAllPrices(ctx context.Context) ([]Ticker, error) {
	url := BinanceBaseURL + "/api/v3/ticker/24hr"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неверный статус ответа: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела ответа: %w", err)
	}

	var tickers []BinanceTicker
	if err := json.Unmarshal(body, &tickers); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	var result []Ticker
	for _, t := range tickers {
		if !strings.HasSuffix(t.Symbol, "USDT") {
			continue
		}

		price, _ := strconv.ParseFloat(t.LastPrice, 64)
		priceChange, _ := strconv.ParseFloat(t.PriceChange, 64)
		volume, _ := strconv.ParseFloat(t.Volume, 64)
		high, _ := strconv.ParseFloat(t.HighPrice, 64)
		low, _ := strconv.ParseFloat(t.LowPrice, 64)

		result = append(result, Ticker{
			Symbol:          t.Symbol,
			Price:           price,
			PriceChange24h:  priceChange,
			Volume24h:       volume,
			High24h:         high,
			Low24h:          low,
			LastUpdateTime:  time.Now(),
		})
	}

	logger.Debug("Получены цены с Binance", zap.Int("count", len(result)))
	return result, nil
}

func (f *BinanceFetcher) GetRecentCandles(ctx context.Context, symbol string, timeframe models.Timeframe, limit int) ([]models.OHLCV, error) {
	interval := f.timeframeToInterval(timeframe)
	url := BinanceBaseURL + "/api/v3/klines?symbol=" + symbol + "&interval=" + interval + "&limit=" + strconv.Itoa(limit)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неверный статус ответа: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения тела ответа: %w", err)
	}

	var rawKlines [][]interface{}
	if err := json.Unmarshal(body, &rawKlines); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	var candles []models.OHLCV
	for _, k := range rawKlines {
		if len(k) < 6 {
			continue
		}

		open, _ := strconv.ParseFloat(k[1].(string), 64)
		high, _ := strconv.ParseFloat(k[2].(string), 64)
		low, _ := strconv.ParseFloat(k[3].(string), 64)
		close, _ := strconv.ParseFloat(k[4].(string), 64)
		volume, _ := strconv.ParseFloat(k[5].(string), 64)

		openTime := time.UnixMilli(int64(k[0].(float64)))

		candles = append(candles, models.OHLCV{
			Timestamp: openTime,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		})
	}

	return candles, nil
}

func (f *BinanceFetcher) timeframeToInterval(tf models.Timeframe) string {
	switch tf {
	case models.Timeframe1M:
		return "1m"
	case models.Timeframe1H:
		return "1h"
	case models.Timeframe1D:
		return "1d"
	default:
		return "1h"
	}
}