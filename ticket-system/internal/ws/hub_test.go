package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestServer 建立 WebSocket 測試伺服器
func setupTestServer(hub *Hub) *httptest.Server {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		hub.Register(conn)
		// 讀取迴圈保持連線
		go func() {
			defer hub.Unregister(conn)
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}()
	}))
}

// dialWS 建立 WebSocket 客戶端連線
func dialWS(t *testing.T, server *httptest.Server) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	return conn
}

func TestHub_RegisterAndBroadcast(t *testing.T) {
	hub := NewHub()
	server := setupTestServer(hub)
	defer server.Close()

	client := dialWS(t, server)
	defer client.Close()

	// 等待 register 完成
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, hub.ClientCount())

	// 廣播訊息
	hub.Broadcast(map[string]string{"type": "test", "msg": "hello"})

	// 客戶端應收到訊息
	_, msg, err := client.ReadMessage()
	require.NoError(t, err)
	assert.Contains(t, string(msg), "hello")
}

func TestHub_MultipleClients(t *testing.T) {
	hub := NewHub()
	server := setupTestServer(hub)
	defer server.Close()

	const clientCount = 5
	clients := make([]*websocket.Conn, clientCount)
	for i := 0; i < clientCount; i++ {
		clients[i] = dialWS(t, server)
		defer clients[i].Close()
	}

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, clientCount, hub.ClientCount())

	// 廣播，所有客戶端都應收到
	hub.Broadcast(map[string]string{"event": "stock_update"})

	for i, client := range clients {
		_, msg, err := client.ReadMessage()
		require.NoError(t, err, "客戶端 %d 應收到訊息", i)
		assert.Contains(t, string(msg), "stock_update")
	}
}

func TestHub_UnregisterRemovesClient(t *testing.T) {
	hub := NewHub()
	server := setupTestServer(hub)
	defer server.Close()

	client := dialWS(t, server)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, hub.ClientCount())

	// 客戶端斷線
	client.Close()
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, hub.ClientCount())
}

func TestHub_BroadcastToEmpty_NoPanic(t *testing.T) {
	hub := NewHub()
	assert.NotPanics(t, func() {
		hub.Broadcast(map[string]string{"msg": "nobody home"})
	})
}

func TestHub_ConcurrentBroadcast_NoPanic(t *testing.T) {
	hub := NewHub()
	server := setupTestServer(hub)
	defer server.Close()

	// 建立多個客戶端
	for i := 0; i < 10; i++ {
		c := dialWS(t, server)
		defer c.Close()
	}
	time.Sleep(50 * time.Millisecond)

	// 並發廣播
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			hub.Broadcast(map[string]int{"seq": n})
		}(i)
	}
	wg.Wait()
}
