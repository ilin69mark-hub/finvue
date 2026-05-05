package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"finvue/internal/handlers"
	"finvue/internal/pkg/database"
	"finvue/internal/repositories"
)

func init() {
	if err := database.InitForTests(); err != nil {
		panic("Failed to init database for tests: " + err.Error())
	}
}

func TestAssetHandler_GetAssets(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропуск интеграционного теста в short-режиме")
	}

	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200, получили %d", w.Code)
	}

	var response []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Ошибка парсинга JSON: %v", err)
	}

	t.Logf("Получено активов: %d", len(response))
}

func TestAssetHandler_GetAssets_IncludeInactive(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропуск интеграционного теста в short-режиме")
	}

	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets?include_inactive=true", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200, получили %d", w.Code)
	}
}

func TestAssetHandler_GetAssetByID_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Пропуск интеграционного теста в short-режиме")
	}

	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/999999999", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Ожидался статус 404, получили %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Ошибка парсинга JSON: %v", err)
	}

	if response["error"] != "Актив не найден" {
		t.Errorf("Ожидалось сообщение 'Актив не найден', получили %s", response["error"])
	}
}

func TestAssetHandler_GetAssetByID_InvalidID(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/assets/invalid", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидался статус 400, получили %d", w.Code)
	}
}

func TestOHLCVHandler_GetOHLCV_MissingAssetID(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ohlcv", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Ожидался статус 400, получили %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("Ошибка парсинга JSON: %v", err)
	}

	if !strings.Contains(response["error"], "asset_id") {
		t.Errorf("Ожидалось сообщение с 'asset_id', получили %s", response["error"])
	}
}

func TestHealthEndpoint(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Ожидался статус 200, получили %d", w.Code)
	}

	if strings.TrimSpace(w.Body.String()) != "OK" {
		t.Errorf("Ожидалось 'OK', получили %s", w.Body.String())
	}
}

func TestCORS_Preflight(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/assets", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")

	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Ожидался статус 204, получили %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Errorf("Ожидался Access-Control-Allow-Origin: http://localhost:5173")
	}
}

func TestNotFound(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Ожидался статус 404, получили %d", w.Code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	handler := router.Setup()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/assets", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Ожидался статус 405, получили %d", w.Code)
	}
}
