package websocket

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"finvue/internal/pkg/logger"

	"go.uber.org/zap"
)

type Hub struct {
	clients    map[*Client]bool
	subs      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
	mu         sync.RWMutex
}

type Message struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Symbol    string      `json:"symbol,omitempty"`
}

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	symbol string
}

var globalHub *Hub

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		subs:       make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
	}
}

func InitGlobalHub() {
	globalHub = NewHub()
	go globalHub.Run()
}

func GetGlobalHub() *Hub {
	return globalHub
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if client.symbol != "" {
				if h.subs[client.symbol] == nil {
					h.subs[client.symbol] = make(map[*Client]bool)
				}
				h.subs[client.symbol][client] = true
			}
			h.mu.Unlock()
			logger.Debug("Клиент подключён", zap.String("symbol", client.symbol))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if client.symbol != "" {
					delete(h.subs[client.symbol], client)
				}
				close(client.send)
			}
			h.mu.Unlock()
			logger.Debug("Клиент отключён", zap.String("symbol", client.symbol))

		case message := <-h.broadcast:
			h.mu.RLock()
			if message.Symbol != "" {
				if subs, ok := h.subs[message.Symbol]; ok {
					for client := range subs {
						select {
						case client.send <- h.encodeMessage(message):
						default:
							close(client.send)
							delete(h.clients, client)
							delete(h.subs[message.Symbol], client)
						}
					}
				}
			} else {
				for client := range h.clients {
					select {
					case client.send <- h.encodeMessage(message):
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) encodeMessage(msg *Message) []byte {
	data, _ := json.Marshal(msg)
	return data
}

func (h *Hub) BroadcastPriceUpdate(symbol string, price float64) {
	h.broadcast <- &Message{
		Type:   "price_update",
		Symbol: symbol,
		Payload: map[string]interface{}{
			"symbol": symbol,
			"price":  price,
		},
	}
}

func (h *Hub) Subscribe(symbol string) *Client {
	return &Client{
		hub:    h,
		symbol: symbol,
		send:   make(chan []byte, 256),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("Ошибка чтения из WebSocket", zap.Error(err))
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		if msg.Type == "subscribe" {
			if symbol, ok := msg.Payload.(string); ok {
				c.hub.mu.Lock()
				delete(c.hub.subs[c.symbol], c)
				c.symbol = symbol
				if c.hub.subs[symbol] == nil {
					c.hub.subs[symbol] = make(map[*Client]bool)
				}
				c.hub.subs[symbol][c] = true
				c.hub.mu.Unlock()
				logger.Debug("Клиент подписался", zap.String("symbol", symbol))
			}
		}
	}
}

func (c *Client) WritePump() {
	defer c.conn.Close()

	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}