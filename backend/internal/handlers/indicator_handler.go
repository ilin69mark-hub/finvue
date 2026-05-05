package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"finvue/internal/pkg/logger"
	"finvue/internal/services"

	"go.uber.org/zap"
)

type IndicatorHandler struct {
	service *services.IndicatorService
}

func NewIndicatorHandler(service *services.IndicatorService) *IndicatorHandler {
	return &IndicatorHandler{service: service}
}

func (h *IndicatorHandler) GetSMA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

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

	timeframe := r.URL.Query().Get("timeframe")
	if timeframe == "" {
		timeframe = "1d"
	}

	fastPeriod := 20
	if fp := r.URL.Query().Get("fast_period"); fp != "" {
		if parsed, err := strconv.Atoi(fp); err == nil && parsed > 0 {
			fastPeriod = parsed
		}
	}

	slowPeriod := 50
	if sp := r.URL.Query().Get("slow_period"); sp != "" {
		if parsed, err := strconv.Atoi(sp); err == nil && parsed > 0 {
			slowPeriod = parsed
		}
	}

	ctx := r.Context()
	resp, err := h.service.CalculateSMA(ctx, services.SMARequest{
		AssetID:    assetID,
		Timeframe:  timeframe,
		FastPeriod: fastPeriod,
		SlowPeriod: slowPeriod,
	})
	if err != nil {
		logger.Error("Ошибка расчёта SMA", zap.Error(err))
		http.Error(w, `{"error":"Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
		return
	}

	if resp == nil {
		http.Error(w, `{"error":"Актив не найден"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		logger.Error("Ошибка кодирования ответа", zap.Error(err))
	}
}

func (h *IndicatorHandler) GetAllSMA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	results, err := h.service.GetAllAssetsSMA(ctx)
	if err != nil {
		logger.Error("Ошибка получения SMA для всех активов", zap.Error(err))
		http.Error(w, `{"error":"Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		logger.Error("Ошибка кодирования ответа", zap.Error(err))
	}
}
