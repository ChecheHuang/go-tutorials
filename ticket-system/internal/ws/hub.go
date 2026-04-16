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
	mu      sync.RWMutex           // 保護 clients map
	writeMu sync.Mutex             // 序列化所有 Broadcast 寫入（gorilla/websocket 不支援併發寫入同一個 conn）
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
	if _, ok := h.clients[conn]; !ok {
		return // 已經被移除，避免重複 Close
	}
	delete(h.clients, conn)
	conn.Close()
	slog.Debug("WebSocket 連線移除", "total", len(h.clients))
}

// Broadcast 廣播 JSON 訊息給所有連線的客戶端
//
// 設計要點：
//  1. writeMu 確保同時間只有一個 Broadcast 在寫入（gorilla/websocket 禁止併發寫入同一個 conn）
//  2. 先在 RLock 內複製連線清單，然後立即釋放鎖
//  3. 在 mu 鎖外逐一發送，避免慢連線阻塞 Register/Unregister
//  4. 發送失敗的連線收集起來，最後統一 Unregister（此時不持有任何鎖）
func (h *Hub) Broadcast(data any) {
	msg, err := json.Marshal(data)
	if err != nil {
		slog.Error("WebSocket 序列化失敗", "error", err)
		return
	}

	h.writeMu.Lock()
	defer h.writeMu.Unlock()

	// 複製連線清單，最小化 mu 鎖持有時間
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.clients))
	for conn := range h.clients {
		conns = append(conns, conn)
	}
	h.mu.RUnlock()

	// 在 mu 鎖外發送訊息（仍在 writeMu 保護下，確保不會併發寫入同一個 conn）
	var failed []*websocket.Conn
	for _, conn := range conns {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			slog.Warn("WebSocket 發送失敗", "error", err)
			failed = append(failed, conn)
		}
	}

	// 統一清除失敗的連線（此時不持有 mu，不會 deadlock）
	for _, conn := range failed {
		h.Unregister(conn)
	}
}

// ClientCount 回傳目前連線數
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
