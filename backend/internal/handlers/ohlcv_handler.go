package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"finvue/internal/models"
	"finvue/internal/repositories"
	"finvue/internal/pkg/logger"

	"go.uber.org/zap"
)

type OHLCVHandler struct {
	repo      *repositories.OHLCVRepository
	assetRepo *repositories.AssetRepository
}

func NewOHLCVHandler(repo *repositories.OHLCVRepository, assetRepo *repositories.AssetRepository) *OHLCVHandler {
	return &OHLCVHandler{repo: repo, assetRepo: assetRepo}
}

type OHLCVResponse struct {
	ID        int64     `json:"id"`
	AssetID   int64     `json:"asset_id"`
	Timestamp string    `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
}

func (h *OHLCVHandler) GetOHLCV(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	assetIDStr := r.URL.Query().Get("asset_id")
	if assetIDStr == "" {
		http.Error(w, `{"error":"Параметр asset_id обязателен"}`, http.StatusBadRequest)
		return
	}

	assetID, err := strconv.ParseInt(assetIDStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Неверный формат asset_id"}`, http.StatusBadRequest)
		return
	}

	asset, err := h.assetRepo.GetByID(ctx, assetID)
	if err != nil {
		logger.Error("Ошибка получения актива", zap.Error(err))
		http.Error(w, `{"error":"Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
		return
	}
	if asset == nil {
		http.Error(w, `{"error":"Актив не найден"}`, http.StatusNotFound)
		return
	}

	timeframeStr := r.URL.Query().Get("timeframe")
	if timeframeStr == "" {
		timeframeStr = "1h"
	}
	timeframe := repositories.ParseTimeframe(timeframeStr)

	limit := 100
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	var from, to *time.Time
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			from = &t
		}
	}
	if toStr := r.URL.Query().Get("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			to = &t
		}
	}

	req := models.OHLCVRequest{
		AssetID:   assetID,
		Timeframe: timeframe,
		From:      from,
		To:        to,
		Limit:     limit,
	}

	candles, err := h.repo.GetByAssetAndTimeframe(ctx, req)
	if err != nil {
		logger.Error("Ошибка получения свечей", zap.Error(err))
		http.Error(w, `{"error":"Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
		return
	}

	response := struct {
		Asset   AssetResponse   `json:"asset"`
		Timeframe string         `json:"timeframe"`
		Candles []OHLCVResponse `json:"candles"`
	}{
		Asset: AssetResponse{
			ID:        asset.ID,
			Symbol:    asset.Symbol,
			Name:      asset.Name,
			AssetType: string(asset.AssetType),
			IsActive:  asset.IsActive,
		},
		Timeframe: string(timeframe),
		Candles:   make([]OHLCVResponse, 0, len(candles)),
	}

	for _, c := range candles {
		response.Candles = append(response.Candles, OHLCVResponse{
			ID:        c.ID,
			AssetID:   c.AssetID,
			Timestamp: c.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			Open:      c.Open,
			High:      c.High,
			Low:       c.Low,
			Close:     c.Close,
			Volume:    c.Volume,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Ошибка кодирования ответа", zap.Error(err))
	}
}