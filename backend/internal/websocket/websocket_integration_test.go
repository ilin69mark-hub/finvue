package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestWebSocket_ConnectAndSubscribe(t *testing.T) {
	InitGlobalHub()
	hub := GetGlobalHub()

	wsURL := "ws://localhost:8080/ws?symbol=BTCUSDT"

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

func TestWebSocket_BroadcastToSubscribers(t *testing.T) {
	InitGlobalHub()
	hub := GetGlobalHub()

	conn1, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws?symbol=BTCUSDT", nil)
	if err != nil {
		t.Fatalf("Ошибка подключения первого клиента: %v", err)
	}
	defer conn1.Close()

	conn2, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws?symbol=BTCUSDT", nil)
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
	InitGlobalHub()
	hub := GetGlobalHub()

	connBTC, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws?symbol=BTCUSDT", nil)
	if err != nil {
		t.Fatalf("Ошибка подключения BTC клиента: %v", err)
	}
	defer connBTC.Close()

	connETH, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws?symbol=ETHUSDT", nil)
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

func TestWebSocket_UpgradeError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	req, _ := http.NewRequest("POST", server.URL+"/ws?symbol=BTCUSDT", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Ошибка запроса: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Ожидался статус 405, получили %d", resp.StatusCode)
	}
}

func TestHub_BroadcastPriceUpdate(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	hub.BroadcastPriceUpdate("BTCUSDT", 50000.0)

	select {
	case msg := <-hub.broadcast:
		if msg.Symbol != "BTCUSDT" {
			t.Errorf("Ожидался symbol=BTCUSDT, получили %s", msg.Symbol)
		}
		if msg.Type != "price_update" {
			t.Errorf("Ожидался type=price_update, получили %s", msg.Type)
		}
	case <-time.After(1 * time.Second):
		t.Error("Таймаут ожидания сообщения")
	}
}

func TestClient_Subscribe(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := hub.Subscribe("BTCUSDT")
	if client.symbol != "BTCUSDT" {
		t.Errorf("Ожидался symbol=BTCUSDT, получили %s", client.symbol)
	}
	if client.hub != hub {
		t.Error("Клиент не привязан к хабу")
	}
}