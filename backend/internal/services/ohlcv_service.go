package services

import (
	"context"

	"finvue/internal/dto"
	"finvue/internal/models"
	"finvue/internal/repositories"
)

type OHLCVService struct {
	repo      *repositories.OHLCVRepository
	assetRepo *repositories.AssetRepository
}

func NewOHLCVService(repo *repositories.OHLCVRepository, assetRepo *repositories.AssetRepository) *OHLCVService {
	return &OHLCVService{repo: repo, assetRepo: assetRepo}
}

func (s *OHLCVService) GetCandles(ctx context.Context, query dto.OHLCVQueryDTO) (*dto.OHLCVListResponse, error) {
	query.SetDefaults()

	asset, err := s.assetRepo.GetByID(ctx, query.AssetID)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, nil
	}

	timeframe := repositories.ParseTimeframe(query.Timeframe)

	req := models.OHLCVRequest{
		AssetID:   query.AssetID,
		Timeframe: timeframe,
		From:      query.From,
		To:        query.To,
		Limit:     query.Limit,
	}

	candles, err := s.repo.GetByAssetAndTimeframe(ctx, req)
	if err != nil {
		return nil, err
	}

	assetService := NewAssetService(s.assetRepo)
	assetDTO := assetService.modelToDTO(*asset)

	response := &dto.OHLCVListResponse{
		Asset:     assetDTO,
		Timeframe: query.Timeframe,
		Candles:   make([]dto.OHLCVDTO, 0, len(candles)),
		Total:     len(candles),
	}

	for _, c := range candles {
		response.Candles = append(response.Candles, s.candleToDTO(c))
	}

	return response, nil
}

func (s *OHLCVService) candleToDTO(candle models.OHLCV) dto.OHLCVDTO {
	return dto.OHLCVDTO{
		ID:        candle.ID,
		AssetID:   candle.AssetID,
		Timestamp: candle.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		Open:      candle.Open,
		High:      candle.High,
		Low:       candle.Low,
		Close:     candle.Close,
		Volume:    candle.Volume,
	}
}