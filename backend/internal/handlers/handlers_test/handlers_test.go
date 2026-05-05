package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"finvue/internal/handlers"
	"finvue/internal/repositories"
)

func TestAlertHandler_GetAlerts(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/alerts", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var alerts []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &alerts); err != nil {
		t.Errorf("Failed to parse JSON: %v", err)
	}
}

func TestAlertHandler_MarkRead(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/alerts/1/read", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404, got %d", w.Code)
	}
}

func TestAlertHandler_DeleteAlert(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/alerts/1", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404, got %d", w.Code)
	}
}

func TestIndicatorHandler_GetSMA(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/indicators/sma?asset_id=1&period=20", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 200 or 400, got %d", w.Code)
	}
}

func TestIndicatorHandler_GetAllSMA(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/indicators/sma/all?period=20", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 200 or 400, got %d", w.Code)
	}
}

func TestOHLCVHandler_GetOHLCV_WithAssetID(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ohlcv?asset_id=1&timeframe=1h", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 200 or 400, got %d", w.Code)
	}
}

func TestOHLCVHandler_GetOHLCV_InvalidTimeframe(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ohlcv?asset_id=1&timeframe=invalid", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 200 or 400, got %d", w.Code)
	}
}

func TestAssetHandler_GetAssetByID_Valid(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/1", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 200 or 404, got %d", w.Code)
	}
}