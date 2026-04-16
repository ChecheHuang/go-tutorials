package handler

import (
	"log/slog"
	"net/http"
	"slices"

	"ticket-system/internal/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WSHandler struct {
	hub      *ws.Hub
	upgrader websocket.Upgrader
}

func NewWSHandler(hub *ws.Hub, allowedOrigins []string) *WSHandler {
	return &WSHandler{
		hub: hub,
		upgrader: websocket.Upgrader{
			CheckOrigin: newOriginChecker(allowedOrigins),
		},
	}
}

// newOriginChecker 根據白名單建立 Origin 檢查函式
// 白名單為空或包含 "*" 時允許所有來源（僅限開發環境使用）
func newOriginChecker(allowedOrigins []string) func(r *http.Request) bool {
	allowAll := len(allowedOrigins) == 0 || slices.Contains(allowedOrigins, "*")

	return func(r *http.Request) bool {
		if allowAll {
			return true
		}
		origin := r.Header.Get("Origin")
		return slices.Contains(allowedOrigins, origin)
	}
}

// HandleWebSocket 處理 WebSocket 連線
// GET /ws
func (h *WSHandler) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
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
