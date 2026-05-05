package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"finvue/internal/pkg/logger"
	"finvue/internal/repositories"

	"go.uber.org/zap"
)

type AssetHandler struct {
	repo *repositories.AssetRepository
}

func NewAssetHandler(repo *repositories.AssetRepository) *AssetHandler {
	return &AssetHandler{repo: repo}
}

type AssetResponse struct {
	ID               int64    `json:"id"`
	Symbol           string   `json:"symbol"`
	Name             string   `json:"name"`
	AssetType        string   `json:"asset_type"`
	IsActive         bool     `json:"is_active"`
	LastPrice        *float64 `json:"last_price,omitempty"`
	LastPriceUpdated *string  `json:"last_price_updated,omitempty"`
}

func (h *AssetHandler) GetAssets(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	ctx := r.Context()
	assets, err := h.repo.GetAll(ctx, includeInactive)
	if err != nil {
		logger.Error("Ошибка получения активов", zap.Error(err))
		http.Error(w, `{"error":"Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
		return
	}

	response := make([]AssetResponse, 0, len(assets))
	for _, a := range assets {
		var lastPrice *float64
		if a.LastPrice > 0 {
			lastPrice = &a.LastPrice
		}

		var lastPriceUpdated *string
		if a.LastPriceUpdated != nil {
			ts := a.LastPriceUpdated.Format("2006-01-02T15:04:05Z07:00")
			lastPriceUpdated = &ts
		}

		response = append(response, AssetResponse{
			ID:               a.ID,
			Symbol:           a.Symbol,
			Name:             a.Name,
			AssetType:        string(a.AssetType),
			IsActive:         a.IsActive,
			LastPrice:        lastPrice,
			LastPriceUpdated: lastPriceUpdated,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Ошибка кодирования ответа", zap.Error(err))
	}
}

func (h *AssetHandler) GetAssetByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, `{"error":"ID актива обязателен"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Неверный формат ID"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	asset, err := h.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Ошибка получения актива", zap.Error(err))
		http.Error(w, `{"error":"Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
		return
	}

	if asset == nil {
		http.Error(w, `{"error":"Актив не найден"}`, http.StatusNotFound)
		return
	}

	var lastPrice *float64
	if asset.LastPrice > 0 {
		lastPrice = &asset.LastPrice
	}

	var lastPriceUpdated *string
	if asset.LastPriceUpdated != nil {
		ts := asset.LastPriceUpdated.Format("2006-01-02T15:04:05Z07:00")
		lastPriceUpdated = &ts
	}

	response := AssetResponse{
		ID:               asset.ID,
		Symbol:           asset.Symbol,
		Name:             asset.Name,
		AssetType:        string(asset.AssetType),
		IsActive:         asset.IsActive,
		LastPrice:        lastPrice,
		LastPriceUpdated: lastPriceUpdated,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Ошибка кодирования ответа", zap.Error(err))
	}
}
