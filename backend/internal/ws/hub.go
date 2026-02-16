package ws

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu       sync.RWMutex
	byRoomID map[string]map[*websocket.Conn]struct{}
}

func NewHub() *Hub {
	return &Hub{byRoomID: make(map[string]map[*websocket.Conn]struct{})}
}

func (h *Hub) Add(roomID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.byRoomID[roomID]; !ok {
		h.byRoomID[roomID] = make(map[*websocket.Conn]struct{})
	}
	h.byRoomID[roomID][conn] = struct{}{}
}

func (h *Hub) Remove(roomID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns, ok := h.byRoomID[roomID]
	if !ok {
		return
	}
	delete(conns, conn)
	if len(conns) == 0 {
		delete(h.byRoomID, roomID)
	}
}

func (h *Hub) Broadcast(roomID string, payload any) {
	h.mu.RLock()
	conns := h.byRoomID[roomID]
	targets := make([]*websocket.Conn, 0, len(conns))
	for conn := range conns {
		targets = append(targets, conn)
	}
	h.mu.RUnlock()

	for _, conn := range targets {
		if err := conn.WriteJSON(payload); err != nil {
			_ = conn.Close()
			h.Remove(roomID, conn)
		}
	}
}
