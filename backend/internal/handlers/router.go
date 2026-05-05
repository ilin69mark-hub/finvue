package handlers

import (
	"net/http"

	"finvue/internal/repositories"
	"finvue/internal/services"
	"finvue/internal/websocket"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	assetHandler     *AssetHandler
	ohlcvHandler     *OHLCVHandler
	indicatorHandler *IndicatorHandler
	alertHandler     *AlertHandler
	wsHandler        *websocket.Handler
	Hub              *websocket.Hub
}

func NewRouter(assetRepo *repositories.AssetRepository, ohlcvRepo *repositories.OHLCVRepository) *Router {
	hub := websocket.GetGlobalHub()
	wsHandler := websocket.NewHandler(hub)

	indicatorService := services.NewIndicatorService(ohlcvRepo, assetRepo)
	indicatorHandler := NewIndicatorHandler(indicatorService)

	alertRepo := repositories.NewAlertRepository()
	alertHandler := NewAlertHandler(alertRepo)

	return &Router{
		assetHandler:     NewAssetHandler(assetRepo),
		ohlcvHandler:     NewOHLCVHandler(ohlcvRepo, assetRepo),
		indicatorHandler: indicatorHandler,
		alertHandler:     alertHandler,
		wsHandler:        wsHandler,
		Hub:              hub,
	}
}

func (r *Router) Setup() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(CORS)
	mux.Use(JSONErrorHandler)

	mux.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.Get("/ws", r.wsHandler.Handle)

	mux.Route("/api/v1", func(api chi.Router) {
		api.Get("/assets", r.assetHandler.GetAssets)
		api.Get("/assets/{id}", r.assetHandler.GetAssetByID)
		api.Get("/ohlcv", r.ohlcvHandler.GetOHLCV)
		api.Get("/indicators/sma", r.indicatorHandler.GetSMA)
		api.Get("/indicators/sma/all", r.indicatorHandler.GetAllSMA)
		api.Get("/alerts", r.alertHandler.GetAlerts)
		api.Patch("/alerts/{id}/read", r.alertHandler.MarkRead)
		api.Delete("/alerts/{id}", r.alertHandler.DeleteAlert)
	})

	return mux
}
