// Package ws 提供 WebSocket Hub 管理（第 27 課）
package ws

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

// Hub 管理所有 WebSocket 連線
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]bool
}

// NewHub 建立 Hub
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]bool),
	}
}

// Register 註冊新的連線
func (h *Hub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[conn] = true
	slog.Debug("WebSocket 連線註冊", "total", len(h.clients))
}

// Unregister 移除連線
func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, conn)
	conn.Close()
	slog.Debug("WebSocket 連線移除", "total", len(h.clients))
}

// Broadcast 廣播 JSON 訊息給所有連線的客戶端
func (h *Hub) Broadcast(data any) {
	msg, err := json.Marshal(data)
	if err != nil {
		slog.Error("WebSocket 序列化失敗", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			slog.Warn("WebSocket 發送失敗", "error", err)
			go h.Unregister(conn) // 非同步移除避免死鎖
		}
	}
}

// ClientCount 回傳目前連線數
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
