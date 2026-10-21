package ws

import (
	"GoPower/models"
	"encoding/json"
	"sync"
	"time"
)

type WebSocketHub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	quit       chan struct{}
	wg         sync.WaitGroup
	isRunning  bool
}

func NewWebSocketHub() *WebSocketHub {
	h := &WebSocketHub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		quit:       make(chan struct{}),
		isRunning:  true,
	}
	h.wg.Add(1)
	return h
}

func (h *WebSocketHub) Run() {
	defer h.wg.Done()

	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			h.clients[c] = true
			h.mu.Unlock()
		case c := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.Send)
			}
			h.mu.Unlock()
		case msg := <-h.broadcast:
			h.mu.Lock()
			for c := range h.clients {
				select {
				case c.Send <- msg:
				default:
					close(c.Send)
					delete(h.clients, c)
				}
			}
			h.mu.Unlock()
		case <-h.quit:
			h.mu.Lock()
			for c := range h.clients {
				delete(h.clients, c)
				close(c.Send)
			}
			h.mu.Unlock()
			return
		}
	}
}

func (h *WebSocketHub) RegisterClient(client *Client) {
	h.mu.RLock()
	running := h.isRunning
	h.mu.RUnlock()
	if running {
		select {
		case <-h.quit:
			return
		case h.register <- client:
		}
	}
}

func (h *WebSocketHub) UnregisterClient(client *Client) {
	h.mu.RLock()
	running := h.isRunning
	h.mu.RUnlock()
	if running {
		select {
		case <-h.quit:
			return
		case h.unregister <- client:
		}
	}
}

func (h *WebSocketHub) Stop() {
	h.mu.Lock()
	if !h.isRunning {
		h.mu.Unlock()
		return
	}
	h.isRunning = false
	close(h.quit)
	h.mu.Unlock()

	h.wg.Wait()
}

func (h *WebSocketHub) Broadcast(message []byte) error {
	h.mu.RLock()
	if !h.isRunning {
		h.mu.RUnlock()
		return models.ErrHubStopped
	}
	h.mu.RUnlock()
	select {
	case <-h.quit:
		return models.ErrHubStopped
	case h.broadcast <- message:
		return nil
	}
}

func (h *WebSocketHub) BroadcastTelemetry(summary *models.GridSummary) error {
	msg := models.TelemetryMessage{
		Type:      "TELEMETRY",
		Summary:   summary,
		Timestamp: time.Now(),
	}
	bMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if err := h.Broadcast(bMsg); err != nil {
		return err
	}
	return nil
}

func (h *WebSocketHub) BroadcastAlert(severity, message string) error {
	msg := models.GridAlert{
		Type:      "ALERT",
		Severity:  severity,
		Message:   message,
		Timestamp: time.Now(),
	}
	bMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if err := h.Broadcast(bMsg); err != nil {
		return err
	}
	return nil
}

func (h *WebSocketHub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
