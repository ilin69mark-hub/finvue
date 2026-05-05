package models

import (
	"testing"
	"time"
)

func TestAsset_SetLastPrice(t *testing.T) {
	asset := &Asset{
		ID:        1,
		Symbol:    "BTCUSDT",
		Name:      "Bitcoin",
		AssetType: AssetTypeCrypto,
		IsActive:  true,
	}

	asset.SetLastPrice(50000)

	if asset.LastPrice != 50000 {
		t.Errorf("Expected LastPrice 50000, got %f", asset.LastPrice)
	}
	if asset.LastPriceUpdated == nil {
		t.Error("Expected LastPriceUpdated to be set")
	}
}

func TestAsset_GetLastPrice(t *testing.T) {
	asset := &Asset{
		ID:        1,
		Symbol:    "BTCUSDT",
		Name:      "Bitcoin",
		AssetType: AssetTypeCrypto,
	}

	asset.SetLastPrice(42000)
	price := asset.GetLastPrice()

	if price != 42000 {
		t.Errorf("Expected price 42000, got %f", price)
	}
}

func TestAsset_IsActive_Getter(t *testing.T) {
	asset := &Asset{
		ID:        1,
		Symbol:    "BTCUSDT",
		Name:      "Bitcoin",
		AssetType: AssetTypeCrypto,
		IsActive:  true,
	}

	if !asset.IsActive_Getter() {
		t.Error("Expected IsActive to be true")
	}

	asset.IsActive = false
	if asset.IsActive_Getter() {
		t.Error("Expected IsActive to be false")
	}
}

func TestOHLCV_TableName(t *testing.T) {
	tests := []struct {
		timeframe string
		expected  string
	}{
		{"1m", "ohlcv_1m"},
		{"5m", "ohlcv_1m"},
		{"15m", "ohlcv_1m"},
		{"1h", "ohlcv_1h"},
		{"4h", "ohlcv_1h"},
		{"1d", "ohlcv_1d"},
		{"1w", "ohlcv_1d"},
		{"invalid", "ohlcv_1d"},
	}

	for _, tt := range tests {
		ohlcv := &OHLCV{}
		result := ohlcv.TableName(tt.timeframe)
		if result != tt.expected {
			t.Errorf("TableName(%s) = %s, expected %s", tt.timeframe, result, tt.expected)
		}
	}
}

func TestAlert_SetRead(t *testing.T) {
	alert := &Alert{
		ID:        1,
		AssetID:   1,
		AlertType: AlertTypePriceAbove,
		Message:   "Test alert",
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	alert.SetRead()

	if !alert.IsRead {
		t.Error("Expected IsRead to be true")
	}
}

func TestAlert_SetUnread(t *testing.T) {
	alert := &Alert{
		ID:        1,
		AssetID:   1,
		AlertType: AlertTypePriceAbove,
		Message:   "Test alert",
		IsRead:    true,
		CreatedAt: time.Now(),
	}

	alert.SetUnread()

	if alert.IsRead {
		t.Error("Expected IsRead to be false")
	}
}

func TestAssetType_String(t *testing.T) {
	tests := []struct {
		assetType AssetType
		expected  string
	}{
		{AssetTypeCrypto, "crypto"},
		{AssetTypeStock, "stock"},
		{AssetTypeForex, "forex"},
		{"", ""},
	}

	for _, tt := range tests {
		result := tt.assetType.String()
		if result != tt.expected {
			t.Errorf("AssetType.String() = %s, expected %s", result, tt.expected)
		}
	}
}