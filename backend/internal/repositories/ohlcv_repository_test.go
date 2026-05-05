package repositories

import (
	"context"
	"testing"
	"time"

	"finvue/internal/models"
)

func TestOHLCVRepository_BatchInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропуск интеграционного теста в short-режиме")
	}

	ctx := context.Background()
	assetRepo := NewAssetRepository()
	repo := NewOHLCVRepository()

	asset, err := assetRepo.UpsertFromSymbol(ctx, "TESTUSDT", "Test Coin", models.AssetTypeCrypto)
	if err != nil {
		t.Fatalf("Создание актива: %v", err)
	}

	candles := []models.OHLCV{
		{
			AssetID:   asset.ID,
			Timestamp: time.Now().Truncate(time.Minute),
			Open:      100.0,
			High:      110.0,
			Low:       99.0,
			Close:     105.0,
			Volume:    1000.0,
		},
		{
			AssetID:   asset.ID,
			Timestamp: time.Now().Truncate(time.Minute).Add(time.Minute),
			Open:      105.0,
			High:      115.0,
			Low:       104.0,
			Close:     110.0,
			Volume:    1500.0,
		},
	}

	err = repo.BatchInsert(ctx, "ohlcv_1m", candles)
	if err != nil {
		t.Fatalf("BatchInsert() error = %v", err)
	}
}

func TestOHLCVRepository_GetByAssetAndTimeframe(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропуск интеграционного теста в short-режиме")
	}

	ctx := context.Background()
	assetRepo := NewAssetRepository()
	repo := NewOHLCVRepository()

	asset, _ := assetRepo.UpsertFromSymbol(ctx, "ETHUSDT", "Ethereum", models.AssetTypeCrypto)

	req := models.OHLCVRequest{
		AssetID:   asset.ID,
		Timeframe: models.Timeframe1M,
		Limit:     10,
	}

	candles, err := repo.GetByAssetAndTimeframe(ctx, req)
	if err != nil {
		t.Fatalf("GetByAssetAndTimeframe() error = %v", err)
	}

	t.Logf("Получено свечей: %d", len(candles))
}

func TestTableNameFromTimeframe(t *testing.T) {
	tests := []struct {
		tf     models.Timeframe
		expect string
	}{
		{models.Timeframe1M, "ohlcv_1m"},
		{models.Timeframe1H, "ohlcv_1h"},
		{models.Timeframe1D, "ohlcv_1d"},
	}

	for _, tt := range tests {
		result := TableNameFromTimeframe(tt.tf)
		if result != tt.expect {
			t.Errorf("Ожидалось %s, получили %s", tt.expect, result)
		}
	}
}

func TestParseTimeframe(t *testing.T) {
	tests := []struct {
		input  string
		expect models.Timeframe
	}{
		{"1m", models.Timeframe1M},
		{"1M", models.Timeframe1M},
		{"1h", models.Timeframe1H},
		{"1H", models.Timeframe1H},
		{"1d", models.Timeframe1D},
		{"1D", models.Timeframe1D},
		{"unknown", models.Timeframe1H},
	}

	for _, tt := range tests {
		result := ParseTimeframe(tt.input)
		if result != tt.expect {
			t.Errorf("Ожидалось %s, получили %s", tt.expect, result)
		}
	}
}
