package dto

import (
	"testing"
	"time"
)

func TestAssetDTO_JSON(t *testing.T) {
	price := 42000.0
	dto := AssetDTO{
		ID:        1,
		Symbol:    "BTCUSDT",
		Name:      "Bitcoin",
		AssetType: "crypto",
		IsActive:  true,
		LastPrice: &price,
	}

	if dto.Symbol != "BTCUSDT" {
		t.Errorf("Expected Symbol BTCUSDT, got %s", dto.Symbol)
	}
	if *dto.LastPrice != 42000 {
		t.Errorf("Expected LastPrice 42000, got %f", *dto.LastPrice)
	}
}

func TestCreateAssetDTO_Validation(t *testing.T) {
	dto := CreateAssetDTO{
		Symbol:    "BTCUSDT",
		Name:      "Bitcoin",
		AssetType: "crypto",
		IsActive:  true,
	}

	if dto.Symbol == "" {
		t.Error("Symbol should not be empty")
	}
	if dto.AssetType != "crypto" && dto.AssetType != "stock" && dto.AssetType != "forex" {
		t.Errorf("Invalid asset type: %s", dto.AssetType)
	}
}

func TestUpdateAssetDTO(t *testing.T) {
	isActive := true
	dto := UpdateAssetDTO{
		Name:      "Bitcoin Updated",
		AssetType: "crypto",
		IsActive:  &isActive,
		LastPrice: 50000,
	}

	if dto.Name != "Bitcoin Updated" {
		t.Errorf("Expected Name Bitcoin Updated, got %s", dto.Name)
	}
	if !*dto.IsActive {
		t.Error("Expected IsActive true")
	}
}

func TestAssetListResponse(t *testing.T) {
	assets := []AssetDTO{
		{ID: 1, Symbol: "BTCUSDT", Name: "Bitcoin"},
		{ID: 2, Symbol: "ETHUSDT", Name: "Ethereum"},
	}

	response := AssetListResponse{
		Assets:   assets,
		Total:    2,
		Page:     1,
		PageSize: 10,
	}

	if len(response.Assets) != 2 {
		t.Errorf("Expected 2 assets, got %d", len(response.Assets))
	}
	if response.Total != 2 {
		t.Errorf("Expected Total 2, got %d", response.Total)
	}
}

func TestOHLCVQueryDTO_SetDefaults(t *testing.T) {
	query := OHLCVQueryDTO{
		AssetID:  1,
		Timeframe: "",
		Limit:    0,
	}

	query.SetDefaults()

	if query.Limit != 100 {
		t.Errorf("Expected Limit 100, got %d", query.Limit)
	}
	if query.Timeframe != "1h" {
		t.Errorf("Expected Timeframe 1h, got %s", query.Timeframe)
	}
}

func TestOHLCVDTO(t *testing.T) {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	dto := OHLCVDTO{
		Timestamp: timestamp,
		Open:      42000,
		High:      43000,
		Low:       41000,
		Close:     42500,
		Volume:    1000,
	}

	if dto.Open != 42000 {
		t.Errorf("Expected Open 42000, got %f", dto.Open)
	}
}

func TestOHLCVListResponse(t *testing.T) {
	response := OHLCVListResponse{
		Asset:     AssetDTO{ID: 1, Symbol: "BTCUSDT"},
		Timeframe: "1h",
		Candles:   []OHLCVDTO{},
		Total:     0,
	}

	if response.Asset.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %s", response.Asset.Symbol)
	}
}

func TestOHLCVQueryDTO(t *testing.T) {
	query := OHLCVQueryDTO{
		AssetID:   1,
		Timeframe: "1d",
		From:      nil,
		To:        nil,
		Limit:     50,
	}

	if query.AssetID != 1 {
		t.Errorf("Expected AssetID 1, got %d", query.AssetID)
	}
	if query.Timeframe != "1d" {
		t.Errorf("Expected Timeframe 1d, got %s", query.Timeframe)
	}
}