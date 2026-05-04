package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"finvue/internal/repositories"
	"finvue/internal/pkg/logger"

	"go.uber.org/zap"
)

type AlertHandler struct {
	repo *repositories.AlertRepository
}

func NewAlertHandler(repo *repositories.AlertRepository) *AlertHandler {
	return &AlertHandler{repo: repo}
}

type AlertResponse struct {
	ID         int64     `json:"id"`
	AssetID    int64     `json:"asset_id"`
	AlertType  string    `json:"alert_type"`
	Message    string    `json:"message"`
	Value      float64   `json:"value,omitempty"`
	Threshold  float64   `json:"threshold,omitempty"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  string    `json:"created_at"`
}

func (h *AlertHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	unreadOnly := r.URL.Query().Get("unread_only") == "true"

	ctx := r.Context()
	alerts, err := h.repo.GetAll(ctx, unreadOnly)
	if err != nil {
		logger.Error("Ошибка получения алертов", zap.Error(err))
		http.Error(w, `{"error":"Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
		return
	}

	response := make([]AlertResponse, 0, len(alerts))
	for _, a := range alerts {
		response = append(response, AlertResponse{
			ID:        a.ID,
			AssetID:   a.AssetID,
			AlertType: string(a.AlertType),
			Message:   a.Message,
			Value:     a.Value,
			Threshold: a.Threshold,
			IsRead:    a.IsRead,
			CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Ошибка кодирования ответа", zap.Error(err))
	}
}

func (h *AlertHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, `{"error":"ID алерта обязателен"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Неверный формат ID"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.repo.MarkRead(ctx, id); err != nil {
		logger.Error("Ошибка обновления алерта", zap.Error(err))
		http.Error(w, `{"error":"Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}

func (h *AlertHandler) DeleteAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, `{"error":"ID алерта обязателен"}`, http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"Неверный формат ID"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	if err := h.repo.Delete(ctx, id); err != nil {
		logger.Error("Ошибка удаления алерта", zap.Error(err))
		http.Error(w, `{"error":"Внутренняя ошибка сервера"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success":true}`))
}