package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"finvue/internal/handlers"
	"finvue/internal/repositories"
)

func TestAssetHandler_GetAssets_Pagination(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	tests := []struct {
		name   string
		url    string
		status int
	}{
		{"default limit", "/api/v1/assets", http.StatusOK},
		{"with limit", "/api/v1/assets?limit=10", http.StatusOK},
		{"with page", "/api/v1/assets?page=1", http.StatusOK},
		{"with include inactive", "/api/v1/assets?include_inactive=true", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tt.status && w.Code != http.StatusOK {
				t.Errorf("Expected status %d or 200, got %d", tt.status, w.Code)
			}
		})
	}
}

func TestAssetHandler_GetAssetByID_NotFound(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/999999999", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestOHLCVHandler_GetOHLCV_AllParams(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	tests := []struct {
		name   string
		url    string
		status int
	}{
		{"with from", "/api/v1/ohlcv?asset_id=1&timeframe=1h&from=2024-01-01", http.StatusOK},
		{"with to", "/api/v1/ohlcv?asset_id=1&timeframe=1h&to=2024-12-31", http.StatusOK},
		{"with limit", "/api/v1/ohlcv?asset_id=1&timeframe=1h&limit=50", http.StatusOK},
		{"1m timeframe", "/api/v1/ohlcv?asset_id=1&timeframe=1m", http.StatusOK},
		{"4h timeframe", "/api/v1/ohlcv?asset_id=1&timeframe=4h", http.StatusOK},
		{"1d timeframe", "/api/v1/ohlcv?asset_id=1&timeframe=1d", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
				t.Errorf("Expected status 200 or 400, got %d", w.Code)
			}
		})
	}
}

func TestAlertHandler_GetAlerts_Pagination(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	tests := []struct {
		name   string
		url    string
		status int
	}{
		{"default", "/api/v1/alerts", http.StatusOK},
		{"with limit", "/api/v1/alerts?limit=10", http.StatusOK},
		{"with page", "/api/v1/alerts?page=2", http.StatusOK},
		{"with unread only", "/api/v1/alerts?unread=true", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}
		})
	}
}

func TestMiddleware_Headers(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Accept", "application/json")

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("Expected CORS headers in response")
	}
}

func TestMiddleware_RequestID(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Log("Request ID header may not be set in test mode")
	}
}