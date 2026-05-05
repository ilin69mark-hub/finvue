package websocket

import (
	"net/http"

	"finvue/internal/pkg/logger"
	"github.com/gorilla/websocket"

	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:80",
			"http://localhost",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
			"http://127.0.0.1:80",
			"http://127.0.0.1",
		}
		for _, o := range allowedOrigins {
			if origin == o {
				return true
			}
		}
		return false
	},
}

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	symbol := r.URL.Query().Get("symbol")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Ошибка апгрейда до WebSocket", zap.Error(err))
		return
	}

	client := h.hub.Subscribe(symbol)
	client.conn = conn

	h.hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (h *Handler) HandleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head><title>FinVue WebSocket</title></head>
<body>
<h1>FinVue WebSocket</h1>
<p>Use: /ws?symbol=BTCUSDT</p>
</body>
</html>
	`))
}
