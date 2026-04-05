package handler

import (
	"log/slog"
	"net/http"

	"ticket-system/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 開發用，生產環境要檢查 Origin
	},
}

type WSHandler struct {
	hub *ws.Hub
}

func NewWSHandler(hub *ws.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// HandleWebSocket 處理 WebSocket 連線
// GET /ws
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("WebSocket 升級失敗", "error", err)
		return
	}

	h.hub.Register(conn)

	// 讀取迴圈（保持連線，偵測斷線）
	go func() {
		defer h.hub.Unregister(conn)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}
