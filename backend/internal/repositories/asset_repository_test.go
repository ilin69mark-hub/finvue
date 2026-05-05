package repositories

import (
	"context"
	"testing"

	"finvue/internal/models"
	"finvue/internal/pkg/database"
)

func init() {
	if err := database.InitForTests(); err != nil {
		panic("Failed to init database for tests: " + err.Error())
	}
}

func TestAssetRepository_UpsertFromSymbol(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропуск интеграционного теста в short-режиме")
	}

	ctx := context.Background()
	repo := NewAssetRepository()

	asset, err := repo.UpsertFromSymbol(ctx, "BTCUSDT", "Bitcoin", models.AssetTypeCrypto)
	if err != nil {
		t.Fatalf("UpsertFromSymbol() error = %v", err)
	}

	if asset.Symbol != "BTCUSDT" {
		t.Errorf("Ожидался symbol=BTCUSDT, получили %s", asset.Symbol)
	}
	if asset.Name != "Bitcoin" {
		t.Errorf("Ожидался name=Bitcoin, получили %s", asset.Name)
	}
	if asset.AssetType != models.AssetTypeCrypto {
		t.Errorf("Ожидался asset_type=crypto, получили %s", asset.AssetType)
	}
	if !asset.IsActive {
		t.Error("Ожидался is_active=true")
	}
	if asset.ID == 0 {
		t.Error("Ожидался id > 0")
	}

	asset2, err := repo.UpsertFromSymbol(ctx, "BTCUSDT", "Bitcoin Updated", models.AssetTypeCrypto)
	if err != nil {
		t.Fatalf("UpsertFromSymbol() second call error = %v", err)
	}

	if asset2.ID != asset.ID {
		t.Errorf("Ожидался тот же id=%d, получили %d", asset.ID, asset2.ID)
	}
}

func TestAssetRepository_GetAll(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропуск интеграционного теста в short-режиме")
	}

	ctx := context.Background()
	repo := NewAssetRepository()

	_, _ = repo.UpsertFromSymbol(ctx, "ETHUSDT", "Ethereum", models.AssetTypeCrypto)
	_, _ = repo.UpsertFromSymbol(ctx, "BNBUSDT", "BNB", models.AssetTypeCrypto)

	assets, err := repo.GetAll(ctx, false)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(assets) == 0 {
		t.Error("Ожидался непустой список активов")
	}
}
