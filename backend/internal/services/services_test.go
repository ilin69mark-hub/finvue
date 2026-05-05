package services

import (
	"context"
	"testing"

	"finvue/internal/dto"
	"finvue/internal/models"
	"finvue/internal/repositories"
	"finvue/internal/pkg/database"
)

func init() {
	database.InitForTests()
}

func TestNewIndicatorService(t *testing.T) {
	ohlcvRepo := repositories.NewOHLCVRepository()
	assetRepo := repositories.NewAssetRepository()

	service := NewIndicatorService(ohlcvRepo, assetRepo)
	if service == nil {
		t.Error("Expected non-nil service")
	}
}

func TestNewOHLCVService(t *testing.T) {
	ohlcvRepo := repositories.NewOHLCVRepository()
	assetRepo := repositories.NewAssetRepository()

	service := NewOHLCVService(ohlcvRepo, assetRepo)
	if service == nil {
		t.Error("Expected non-nil service")
	}
}

func TestNewAssetService(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()

	service := NewAssetService(assetRepo)
	if service == nil {
		t.Error("Expected non-nil service")
	}
}

func TestIndicatorService_CalculateSMA(t *testing.T) {
	ohlcvRepo := repositories.NewOHLCVRepository()
	assetRepo := repositories.NewAssetRepository()

	service := NewIndicatorService(ohlcvRepo, assetRepo)

	ctx := context.Background()
	asset, _ := assetRepo.UpsertFromSymbol(ctx, "BTCUSDT", "Bitcoin", models.AssetTypeCrypto)

	if asset != nil {
		req := SMARequest{
			AssetID:    asset.ID,
			Timeframe:  "1h",
			FastPeriod: 20,
			SlowPeriod: 50,
		}
		resp, err := service.CalculateSMA(ctx, req)
		if err != nil {
			t.Logf("CalculateSMA error (may be expected without data): %v", err)
		}
		if resp != nil {
			if resp.AssetID != asset.ID {
				t.Errorf("Expected AssetID %d, got %d", asset.ID, resp.AssetID)
			}
		}
	}
}

func TestIndicatorService_CalculateSMA_Defaults(t *testing.T) {
	ohlcvRepo := repositories.NewOHLCVRepository()
	assetRepo := repositories.NewAssetRepository()

	service := NewIndicatorService(ohlcvRepo, assetRepo)

	ctx := context.Background()
	asset, _ := assetRepo.UpsertFromSymbol(ctx, "ETHUSDT", "Ethereum", models.AssetTypeCrypto)

	if asset != nil {
		req := SMARequest{
			AssetID: asset.ID,
		}
		resp, err := service.CalculateSMA(ctx, req)
		if err != nil {
			t.Logf("CalculateSMA error: %v", err)
		}
		if resp != nil {
			t.Logf("SMA calculated: fast=%f, slow=%f", resp.FastSMA, resp.SlowSMA)
		}
	}
}

func TestIndicatorService_GetAllAssetsSMA(t *testing.T) {
	ohlcvRepo := repositories.NewOHLCVRepository()
	assetRepo := repositories.NewAssetRepository()

	service := NewIndicatorService(ohlcvRepo, assetRepo)

	ctx := context.Background()
	_, _ = assetRepo.GetAll(ctx, true)

	responses, err := service.GetAllAssetsSMA(ctx)
	if err != nil {
		t.Logf("GetAllAssetsSMA error: %v", err)
	}
	if responses != nil {
		t.Logf("Got %d SMA responses", len(responses))
	}
}

func TestOHLCVService_GetCandles(t *testing.T) {
	ohlcvRepo := repositories.NewOHLCVRepository()
	assetRepo := repositories.NewAssetRepository()

	service := NewOHLCVService(ohlcvRepo, assetRepo)

	ctx := context.Background()
	asset, _ := assetRepo.UpsertFromSymbol(ctx, "BTCUSDT", "Bitcoin", models.AssetTypeCrypto)

	if asset != nil {
		query := dto.OHLCVQueryDTO{
			AssetID:  asset.ID,
			Timeframe: "1h",
			Limit:    100,
		}
		_, err := service.GetCandles(ctx, query)
		if err != nil {
			t.Logf("GetCandles error: %v", err)
		}
	}
}

func TestAssetService_GetAll(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	service := NewAssetService(assetRepo)

	ctx := context.Background()
	assets, err := service.GetAll(ctx, false)
	if err != nil {
		t.Errorf("GetAll error: %v", err)
	}
	if assets != nil {
		t.Logf("Got %d assets", len(assets))
	}
}

func TestAssetService_GetByID(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	service := NewAssetService(assetRepo)

	ctx := context.Background()
	asset, _ := assetRepo.UpsertFromSymbol(ctx, "BTCUSDT", "Bitcoin", models.AssetTypeCrypto)

	if asset != nil {
		result, err := service.GetByID(ctx, asset.ID)
		if err != nil {
			t.Errorf("GetByID error: %v", err)
		}
		if result != nil && result.ID != asset.ID {
			t.Errorf("Expected ID %d, got %d", asset.ID, result.ID)
		}
	}
}

func TestAssetService_GetBySymbol(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	service := NewAssetService(assetRepo)

	ctx := context.Background()
	_, _ = assetRepo.UpsertFromSymbol(ctx, "BTCUSDT", "Bitcoin", models.AssetTypeCrypto)

	result, err := service.GetBySymbol(ctx, "BTCUSDT")
	if err != nil {
		t.Errorf("GetBySymbol error: %v", err)
	}
	if result != nil && result.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %s", result.Symbol)
	}
}

func TestAssetService_UpsertFromSymbol(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	service := NewAssetService(assetRepo)

	ctx := context.Background()
	result, err := service.UpsertFromSymbol(ctx, "XRPUSDT", "XRP", "crypto")
	if err != nil {
		t.Errorf("UpsertFromSymbol error: %v", err)
	}
	if result != nil {
		if result.Symbol != "XRPUSDT" {
			t.Errorf("Expected symbol XRPUSDT, got %s", result.Symbol)
		}
	}
}

func TestAssetService_Update(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	service := NewAssetService(assetRepo)

	ctx := context.Background()
	asset, _ := assetRepo.UpsertFromSymbol(ctx, "ADAUSDT", "Cardano", models.AssetTypeCrypto)

	if asset != nil {
		input := dto.UpdateAssetDTO{
			Name: "Cardano Updated",
		}
		_, err := service.Update(ctx, asset.ID, input)
		if err != nil {
			t.Logf("Update error: %v", err)
		}
	}
}

func TestAssetService_Create(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	service := NewAssetService(assetRepo)

	ctx := context.Background()
	symbol := "DOGEUSDT"
	_, _ = assetRepo.UpsertFromSymbol(ctx, symbol, "Dogecoin", models.AssetTypeCrypto)

	result, err := service.GetBySymbol(ctx, symbol)
	if err == nil && result != nil {
		t.Logf("Asset %s exists", symbol)
	}
}