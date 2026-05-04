package services

import (
	"context"
	"math"

	"finvue/internal/models"
	"finvue/internal/repositories"
)

type IndicatorService struct {
	ohlcvRepo *repositories.OHLCVRepository
	assetRepo *repositories.AssetRepository
}

func NewIndicatorService(ohlcvRepo *repositories.OHLCVRepository, assetRepo *repositories.AssetRepository) *IndicatorService {
	return &IndicatorService{
		ohlcvRepo: ohlcvRepo,
		assetRepo: assetRepo,
	}
}

type SMARequest struct {
	AssetID    int64  `json:"asset_id"`
	Timeframe  string `json:"timeframe"`
	FastPeriod int    `json:"fast_period"`
	SlowPeriod int    `json:"slow_period"`
}

type SMAResponse struct {
	AssetID   int64     `json:"asset_id"`
	Symbol    string    `json:"symbol"`
	Timeframe string    `json:"timeframe"`
	FastSMA   float64   `json:"fast_sma"`
	SlowSMA   float64   `json:"slow_sma"`
	Crossover string    `json:"crossover"`
	LastPrice float64   `json:"last_price"`
	UpdatedAt string    `json:"updated_at"`
}

func (s *IndicatorService) CalculateSMA(ctx context.Context, req SMARequest) (*SMAResponse, error) {
	if req.FastPeriod == 0 {
		req.FastPeriod = 20
	}
	if req.SlowPeriod == 0 {
		req.SlowPeriod = 50
	}
	if req.Timeframe == "" {
		req.Timeframe = "1d"
	}

	asset, err := s.assetRepo.GetByID(ctx, req.AssetID)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, nil
	}

	ohlcvReq := models.OHLCVRequest{
		AssetID:   req.AssetID,
		Timeframe: repositories.ParseTimeframe(req.Timeframe),
		Limit:     req.SlowPeriod + 10,
	}

	candles, err := s.ohlcvRepo.GetByAssetAndTimeframe(ctx, ohlcvReq)
	if err != nil {
		return nil, err
	}

	if len(candles) < req.SlowPeriod {
		return &SMAResponse{
			AssetID:   req.AssetID,
			Symbol:    asset.Symbol,
			Timeframe: req.Timeframe,
			FastSMA:   0,
			SlowSMA:   0,
			Crossover: "insufficient_data",
			LastPrice: asset.LastPrice,
		}, nil
	}

	fastSMA := s.calculateSMA(candles, req.FastPeriod)
	slowSMA := s.calculateSMA(candles, req.SlowPeriod)

	var crossover string
	if len(candles) >= 2 {
		prevFast := s.calculateSMAFromSlice(candles[:len(candles)-1], req.FastPeriod)
		prevSlow := s.calculateSMAFromSlice(candles[:len(candles)-1], req.SlowPeriod)

		if prevFast <= prevSlow && fastSMA > slowSMA {
			crossover = "bullish"
		} else if prevFast >= prevSlow && fastSMA < slowSMA {
			crossover = "bearish"
		} else if fastSMA > slowSMA {
			crossover = "bullish_continues"
		} else if fastSMA < slowSMA {
			crossover = "bearish_continues"
		} else {
			crossover = "neutral"
		}
	} else {
		crossover = "neutral"
	}

	return &SMAResponse{
		AssetID:   req.AssetID,
		Symbol:    asset.Symbol,
		Timeframe: req.Timeframe,
		FastSMA:   roundTo2Decimals(fastSMA),
		SlowSMA:   roundTo2Decimals(slowSMA),
		Crossover: crossover,
		LastPrice: asset.LastPrice,
	}, nil
}

func (s *IndicatorService) calculateSMA(candles []models.OHLCV, period int) float64 {
	if len(candles) < period {
		return 0
	}

	startIdx := len(candles) - period
	slice := candles[startIdx : startIdx+period]

	return s.calculateSMAFromSlice(slice, period)
}

func (s *IndicatorService) calculateSMAFromSlice(candles []models.OHLCV, period int) float64 {
	if len(candles) < period {
		return 0
	}

	var sum float64
	for i := len(candles) - period; i < len(candles); i++ {
		sum += candles[i].Close
	}

	return sum / float64(period)
}

func roundTo2Decimals(val float64) float64 {
	return math.Round(val*100) / 100
}

func (s *IndicatorService) GetAllAssetsSMA(ctx context.Context) ([]SMAResponse, error) {
	assets, err := s.assetRepo.GetAll(ctx, false)
	if err != nil {
		return nil, err
	}

	var results []SMAResponse
	for _, asset := range assets {
		resp, err := s.CalculateSMA(ctx, SMARequest{
			AssetID:   asset.ID,
			Timeframe: "1d",
			FastPeriod: 20,
			SlowPeriod: 50,
		})
		if err != nil {
			continue
		}
		if resp != nil && resp.Crossover != "insufficient_data" {
			results = append(results, *resp)
		}
	}

	return results, nil
}