package infrastructure

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type Logger interface {
	Info(msg string, keysAndValues ...any)
}

type WsHub struct {
	mu          sync.RWMutex
	connections map[int64]*websocket.Conn
	log         Logger
}

func NewWsHub(log Logger) *WsHub {
	return &WsHub{
		connections: make(map[int64]*websocket.Conn),
		log:         log,
	}
}

func (h *WsHub) Register(userID int64, conn *websocket.Conn) {
	h.mu.Lock()
	h.connections[userID] = conn
	h.mu.Unlock()
	h.log.Info("WebSocket connected", "user_id", userID)
}

func (h *WsHub) Unregister(userID int64) {
	h.mu.Lock()
	delete(h.connections, userID)
	h.mu.Unlock()
	h.log.Info("WebSocket disconnected", "user_id", userID)
}

func (h *WsHub) NotifyUser(userID int64, message interface{}) {
	h.mu.RLock()
	conn, exists := h.connections[userID]
	h.mu.RUnlock()

	if exists {
		_ = conn.WriteJSON(message)
	}
}
