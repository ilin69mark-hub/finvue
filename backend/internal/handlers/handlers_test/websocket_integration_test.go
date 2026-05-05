package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"finvue/internal/handlers"
	"finvue/internal/repositories"
	ws "finvue/internal/websocket"
)

func TestWebSocket_Integration(t *testing.T) {
	ws.InitGlobalHub()

	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	hub := router.Hub

	server := httptest.NewServer(router.Setup())
	defer server.Close()

	wsURL := "ws://" + server.Listener.Addr().String() + "/ws?symbol=BTCUSDT"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Ошибка подключения к WebSocket: %v", err)
	}
	defer conn.Close()

	hub.BroadcastPriceUpdate("BTCUSDT", 42000.0)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Ошибка чтения сообщения: %v", err)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(msg, &resp); err != nil {
		t.Fatalf("Ошибка парсинга JSON: %v", err)
	}

	if resp["type"] != "price_update" {
		t.Errorf("Ожидался тип price_update, получили %v", resp["type"])
	}
}

func TestWebSocket_MultipleClients(t *testing.T) {
	ws.InitGlobalHub()

	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	hub := router.Hub

	server := httptest.NewServer(router.Setup())
	defer server.Close()

	wsURL := "ws://" + server.Listener.Addr().String() + "/ws?symbol=BTCUSDT"

	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Ошибка подключения первого клиента: %v", err)
	}
	defer conn1.Close()

	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Ошибка подключения второго клиента: %v", err)
	}
	defer conn2.Close()

	hub.BroadcastPriceUpdate("BTCUSDT", 42000.0)

	var wg sync.WaitGroup
	wg.Add(2)

	var received1, received2 bool

	go func() {
		defer wg.Done()
		conn1.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, msg, err := conn1.ReadMessage()
		if err == nil {
			var resp map[string]interface{}
			if json.Unmarshal(msg, &resp) == nil {
				if resp["type"] == "price_update" {
					received1 = true
				}
			}
		}
	}()

	go func() {
		defer wg.Done()
		conn2.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, msg, err := conn2.ReadMessage()
		if err == nil {
			var resp map[string]interface{}
			if json.Unmarshal(msg, &resp) == nil {
				if resp["type"] == "price_update" {
					received2 = true
				}
			}
		}
	}()

	wg.Wait()

	if !received1 {
		t.Error("Первый клиент не получил сообщение")
	}
	if !received2 {
		t.Error("Второй клиент не получил сообщение")
	}
}

func TestWebSocket_DifferentSymbols(t *testing.T) {
	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)
	hub := router.Hub

	server := httptest.NewServer(router.Setup())
	defer server.Close()

	addr := server.Listener.Addr().String()

	connBTC, _, err := websocket.DefaultDialer.Dial("ws://"+addr+"/ws?symbol=BTCUSDT", nil)
	if err != nil {
		t.Fatalf("Ошибка подключения BTC клиента: %v", err)
	}
	defer connBTC.Close()

	connETH, _, err := websocket.DefaultDialer.Dial("ws://"+addr+"/ws?symbol=ETHUSDT", nil)
	if err != nil {
		t.Fatalf("Ошибка подключения ETH клиента: %v", err)
	}
	defer connETH.Close()

	hub.BroadcastPriceUpdate("BTCUSDT", 42000.0)

	connBTC.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msgBTC, err := connBTC.ReadMessage()
	if err != nil {
		t.Fatalf("Ошибка чтения от BTC: %v", err)
	}

	var respBTC map[string]interface{}
	if err := json.Unmarshal(msgBTC, &respBTC); err != nil {
		t.Fatalf("Ошибка парсинга: %v", err)
	}
	if respBTC["type"] != "price_update" {
		t.Errorf("Ожидался price_update для BTC")
	}

	connETH.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err = connETH.ReadMessage()
	if err == nil {
		t.Error("ETH клиент не должен был получить сообщение BTC")
	}
}

func TestWebSocket_InvalidUpgrade(t *testing.T) {
	ws.InitGlobalHub()

	assetRepo := repositories.NewAssetRepository()
	ohlcvRepo := repositories.NewOHLCVRepository()

	router := handlers.NewRouter(assetRepo, ohlcvRepo)

	server := httptest.NewServer(router.Setup())
	defer server.Close()

	resp, err := http.Post("http://"+server.Listener.Addr().String()+"/ws?symbol=BTCUSDT", "text/plain", nil)
	if err != nil {
		t.Fatalf("Ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Ожидался статус 405, получили %d", resp.StatusCode)
	}
}
