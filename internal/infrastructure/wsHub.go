package infrastructure

import (
	"sync"

	"github.com/gofiber/contrib/websocket"
)

type WsHub struct {
	mu          sync.RWMutex
	connections map[int64]*websocket.Conn
}

func NewWsHub() *WsHub {
	return &WsHub{
		connections: make(map[int64]*websocket.Conn),
	}
}

// Register добавляет активное соединение пользователя в хаб
func (h *WsHub) Register(userID int64, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[userID] = conn
}

func (h *WsHub) Unregister(userID int64) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.connections, userID)
}

func (h *WsHub) NotifyUser(userID int64, message interface{}) {
	h.mu.RLock()
	conn, exists := h.connections[userID]
	h.mu.RUnlock()

	if exists {
		_ = conn.WriteJSON(message)
	}
}
