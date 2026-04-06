// ==========================================================================
// 第二十七課：WebSocket 即時通訊
// ==========================================================================
//
// 什麼是 WebSocket？
//   一般的 HTTP 請求是「一問一答」：
//     瀏覽器：「伺服器，給我文章列表」
//     伺服器：「好，給你」（連線結束）
//
//   WebSocket 是「長期通話」：
//     瀏覽器：「我想建立長連線」
//     伺服器：「好，保持連線」
//     之後：雙方都可以隨時主動傳訊息，不需要再問一次
//
// 生活比喻：
//   HTTP = 發簡訊（問一句答一句，然後掛掉）
//   WebSocket = 打電話（通話持續，雙方可以隨時說話）
//
// 什麼時候用 WebSocket？
//   - 聊天室（即時訊息）
//   - 即時通知（「你有新的回覆」）
//   - 多人線上遊戲
//   - 股票即時報價
//   - 協作編輯（Google Docs 那種）
//
// 這節課學什麼？
//   1. WebSocket 連線建立和關閉
//   2. 訊息的讀取和發送
//   3. Hub 模式（廣播給所有連線的客戶端）
//   4. 完整的聊天室伺服器
//
// 執行方式：go run ./tutorials/23-websocket
// 然後用瀏覽器開啟：http://localhost:8080
// （瀏覽器有內建 WebSocket 支援，可以直接用！）
// ==========================================================================

package main // 宣告這是 main 套件

import (
	"encoding/json" // JSON 編解碼（訊息格式）
	"fmt"           // 格式化輸出
	"log"           // 日誌輸出
	"net/http"      // HTTP 伺服器
	"sync"          // 並發控制（Mutex）
	"time"          // 時間處理

	"github.com/gorilla/websocket" // WebSocket 套件
)

// ==========================================================================
// 1. WebSocket 基礎概念
// ==========================================================================
//
// WebSocket 連線流程：
//
//   客戶端（瀏覽器）              伺服器
//       │                          │
//       │──── HTTP Upgrade 請求 ──▶│  // 先用 HTTP 建立連線
//       │◀─── 101 Switching ───────│  // 伺服器同意升級
//       │                          │
//       │◀══════ WebSocket ════════│  // 升級成功！雙向通道建立
//       │                          │
//       │──── 訊息 A ─────────────▶│  // 客戶端主動發送
//       │◀─── 訊息 B ──────────────│  // 伺服器主動推送
//       │──── 訊息 C ─────────────▶│  // 可以同時收發
//       │                          │
//       │──── Close ──────────────▶│  // 任一方可以主動關閉

// ==========================================================================
// 2. 訊息格式定義
// ==========================================================================

// Message 聊天訊息格式（傳輸用的結構）
type Message struct {
	Type      string    `json:"type"`      // 訊息類型：chat、join、leave、error
	Username  string    `json:"username"`  // 發送者名稱
	Content   string    `json:"content"`   // 訊息內容
	Timestamp time.Time `json:"timestamp"` // 發送時間
	RoomID    string    `json:"room_id"`   // 聊天室 ID（支援多個聊天室）
}

// ==========================================================================
// 3. Client — 代表一個 WebSocket 連線
// ==========================================================================

// Client 代表一個連線的客戶端
type Client struct {
	hub      *Hub            // 所屬的 Hub（廣播中心）
	conn     *websocket.Conn // WebSocket 連線物件
	send     chan []byte     // 發送緩衝 channel：要發給這個客戶端的訊息
	username string          // 使用者名稱
	roomID   string          // 所在聊天室
}

// readPump 負責從 WebSocket 讀取訊息（每個連線一個 goroutine）
//
// 這個函式會一直阻塞，持續讀取客戶端發來的訊息
// 當連線斷開時，會清理資源並通知 Hub
func (c *Client) readPump() { // 從 WebSocket 讀取訊息
	defer func() { // 函式結束時執行清理
		c.hub.unregister <- c // 通知 Hub：這個客戶端斷線了
		c.conn.Close()        // 關閉 WebSocket 連線
	}()

	// 設定讀取大小上限（防止惡意客戶端傳送超大訊息）
	c.conn.SetReadLimit(4096) // 最多讀取 4KB

	// 設定 Pong Handler（用於心跳機制，確認連線還活著）
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // 60 秒沒收到訊息則斷線
	c.conn.SetPongHandler(func(string) error {               // 收到 Pong 時更新 deadline
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for { // 無限迴圈：持續讀取訊息
		_, rawMsg, err := c.conn.ReadMessage() // 讀取下一條訊息（阻塞直到有訊息）
		if err != nil {                        // 如果讀取失敗（連線斷開）
			if websocket.IsUnexpectedCloseError(err, // 非正常關閉才記錄錯誤
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 異常斷開 [%s]: %v", c.username, err)
			}
			break // 退出迴圈（defer 會清理）
		}

		// 解析 JSON 訊息
		var msg Message                                      // 儲存解析後的訊息
		if err := json.Unmarshal(rawMsg, &msg); err != nil { // 解析 JSON
			log.Printf("JSON 解析失敗 [%s]: %v", c.username, err) // 印出錯誤
			continue                                          // 繼續讀下一條
		}

		// 補充訊息元資料（服務端添加，客戶端無法偽造）
		msg.Username = c.username  // 使用者名稱由伺服器設定
		msg.Timestamp = time.Now() // 時間戳由伺服器設定
		msg.RoomID = c.roomID      // 聊天室 ID 由伺服器設定

		// 把訊息轉回 JSON，送到 Hub 廣播
		if jsonData, err := json.Marshal(msg); err == nil { // 轉回 JSON
			c.hub.broadcast <- jsonData // 送到 Hub 的廣播 channel
		}
	}
}

// writePump 負責把訊息寫入 WebSocket（每個連線一個 goroutine）
//
// 從 send channel 取訊息，寫入 WebSocket 連線
// 同時負責發送 Ping（心跳），確認連線還活著
func (c *Client) writePump() { // 把訊息寫入 WebSocket
	ticker := time.NewTicker(30 * time.Second) // 每 30 秒發一次 Ping（心跳）
	defer func() {                             // 函式結束時清理
		ticker.Stop()  // 停止心跳計時器
		c.conn.Close() // 關閉連線
	}()

	for { // 無限迴圈：持續等待要發送的訊息
		select { // 同時監聽多個 channel
		case message, ok := <-c.send: // 有訊息要發送
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)) // 10 秒寫入超時
			if !ok {                                                  // send channel 已關閉（Hub 要求斷線）
				c.conn.WriteMessage(websocket.CloseMessage, []byte{}) // 發送關閉訊號
				return                                                // 退出
			}

			w, err := c.conn.NextWriter(websocket.TextMessage) // 取得寫入器
			if err != nil {                                    // 如果無法寫入
				return // 退出（readPump 會做清理）
			}
			w.Write(message) // 寫入訊息

			// 把 send channel 裡積累的訊息一次全部寫入（減少系統呼叫次數）
			n := len(c.send)         // 還有幾條待發訊息
			for i := 0; i < n; i++ { // 逐一取出
				w.Write([]byte{'\n'}) // 訊息之間加換行
				w.Write(<-c.send)     // 寫入下一條
			}

			if err := w.Close(); err != nil { // 關閉寫入器（確保訊息被送出）
				return // 退出
			}

		case <-ticker.C: // 每 30 秒發一次 Ping（心跳機制）
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)) // 10 秒超時
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return // 如果 Ping 失敗，連線可能斷了，退出
			}
		}
	}
}

// ==========================================================================
// 4. Hub — 廣播中心（所有連線的管理者）
// ==========================================================================
//
// Hub 的工作：
//   - 管理所有目前連線的客戶端
//   - 新客戶端連線 → 加入 clients 列表
//   - 客戶端斷線 → 從列表移除
//   - 有新訊息 → 廣播給所有客戶端
//
// 為什麼要用 Hub？
//   假設有 100 個客戶端，其中一個發了訊息
//   需要把這條訊息發給另外 99 個人
//   不能在 readPump goroutine 直接寫入別人的連線（並發問題！）
//   Hub 作為唯一的「廣播員」，在單一 goroutine 中管理所有連線
//
// Hub 是一個典型的 Actor 模式：
//   所有狀態（clients map）只在 Hub 的 goroutine 中修改，不需要 Mutex

// Hub 管理所有 WebSocket 連線的廣播中心
type Hub struct {
	clients    map[*Client]bool // 所有連線中的客戶端（用 map 方便 O(1) 刪除）
	broadcast  chan []byte      // 廣播 channel：收到訊息後發給所有人
	register   chan *Client     // 註冊 channel：新客戶端連線
	unregister chan *Client     // 登出 channel：客戶端斷線
	mu         sync.RWMutex     // 保護 clients 的讀寫（在 Hub 外讀取時用）
}

// newHub 建立並回傳一個新的 Hub
func newHub() *Hub { // 建立 Hub
	return &Hub{
		clients:    make(map[*Client]bool), // 初始化客戶端 map（空的）
		broadcast:  make(chan []byte, 256), // 廣播 channel（有緩衝，256 條訊息容量）
		register:   make(chan *Client),     // 無緩衝 channel（同步）
		unregister: make(chan *Client),     // 無緩衝 channel（同步）
	}
}

// run 啟動 Hub（在獨立的 goroutine 中執行）
//
// 這個函式是 Hub 的核心：
//
//	用 select 同時監聽 register、unregister、broadcast 三個 channel
//	一次只做一件事（單執行緒），所以不需要 Mutex
func (h *Hub) run() { // 啟動 Hub
	for { // 無限迴圈：持續處理事件
		select { // 同時監聽三個 channel

		case client := <-h.register: // 有新客戶端連線
			h.clients[client] = true // 加入客戶端 map
			log.Printf("✅ 新客戶端連線: %s（目前 %d 人）", client.username, len(h.clients))

			// 廣播「有人加入」的通知
			joinMsg, _ := json.Marshal(Message{
				Type:      "join",                                    // 訊息類型
				Username:  "系統",                                      // 發送者
				Content:   fmt.Sprintf("%s 加入了聊天室", client.username), // 內容
				Timestamp: time.Now(),                                // 時間
			})
			h.broadcastToAll(joinMsg) // 廣播給所有人

		case client := <-h.unregister: // 有客戶端斷線
			if _, ok := h.clients[client]; ok { // 確認客戶端存在
				delete(h.clients, client) // 從 map 中移除
				close(client.send)        // 關閉發送 channel（通知 writePump 結束）
				log.Printf("❌ 客戶端斷線: %s（剩餘 %d 人）", client.username, len(h.clients))

				// 廣播「有人離開」的通知
				leaveMsg, _ := json.Marshal(Message{
					Type:      "leave",
					Username:  "系統",
					Content:   fmt.Sprintf("%s 離開了聊天室", client.username),
					Timestamp: time.Now(),
				})
				h.broadcastToAll(leaveMsg) // 廣播給剩餘的人
			}

		case message := <-h.broadcast: // 有訊息要廣播
			h.broadcastToAll(message) // 發送給所有客戶端
		}
	}
}

// broadcastToAll 把訊息發給所有連線的客戶端
func (h *Hub) broadcastToAll(message []byte) { // 廣播給所有人
	for client := range h.clients { // 遍歷所有客戶端
		select {
		case client.send <- message: // 把訊息放入客戶端的發送 channel
			// 成功放入（非阻塞）
		default:
			// send channel 滿了（客戶端太慢消費訊息）
			// 移除這個客戶端，避免阻塞其他客戶端
			delete(h.clients, client)
			close(client.send)
			log.Printf("⚠️ 客戶端 %s 的 send channel 已滿，強制斷線", client.username)
		}
	}
}

// ClientCount 回傳目前連線的客戶端數量
func (h *Hub) ClientCount() int { // 取得連線數
	h.mu.RLock()         // 讀取鎖（允許多個 goroutine 同時讀）
	defer h.mu.RUnlock() // 解鎖
	return len(h.clients)
}

// ==========================================================================
// 5. Upgrader — HTTP 升級成 WebSocket
// ==========================================================================

// upgrader 把 HTTP 連線升級成 WebSocket 連線
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024, // 讀取緩衝區大小（1KB）
	WriteBufferSize: 1024, // 寫入緩衝區大小（1KB）

	// CheckOrigin 檢查請求來源（防止 CSRF 攻擊）
	// 在正式環境中，應該檢查 origin 是否在白名單內
	// 這裡為了示範方便，允許所有來源
	CheckOrigin: func(r *http.Request) bool {
		return true // 允許所有來源（示範用；正式環境要做限制！）
	},
}

// ==========================================================================
// 6. HTTP Handlers
// ==========================================================================

// wsHandler 處理 WebSocket 連線請求
func wsHandler(hub *Hub, w http.ResponseWriter, r *http.Request) { // WebSocket 處理器
	// 從 URL 查詢參數取得使用者名稱和聊天室
	username := r.URL.Query().Get("username") // 例：?username=Alice
	if username == "" {                       // 如果沒有提供名稱
		username = fmt.Sprintf("匿名用戶%d", time.Now().UnixMilli()%1000) // 產生隨機名稱
	}
	roomID := r.URL.Query().Get("room") // 例：?room=general
	if roomID == "" {                   // 如果沒有指定聊天室
		roomID = "general" // 預設聊天室
	}

	// 把 HTTP 連線升級成 WebSocket 連線
	conn, err := upgrader.Upgrade(w, r, nil) // 升級（nil = 不加額外 header）
	if err != nil {                          // 如果升級失敗
		log.Printf("WebSocket 升級失敗: %v", err) // 印出錯誤
		return                                // 提前返回
	}

	// 建立 Client 物件
	client := &Client{
		hub:      hub,                    // 關聯到 Hub
		conn:     conn,                   // WebSocket 連線
		send:     make(chan []byte, 256), // 發送 channel（有緩衝）
		username: username,               // 使用者名稱
		roomID:   roomID,                 // 聊天室
	}

	// 把新客戶端註冊到 Hub
	hub.register <- client // 送到 Hub 的 register channel

	// 啟動兩個 goroutine：一個讀、一個寫
	// 這是標準的 WebSocket 並發模式：讀寫分離
	go client.writePump() // 啟動寫入 goroutine（負責把訊息寫給客戶端）
	go client.readPump()  // 啟動讀取 goroutine（負責從客戶端讀取訊息）
	// ↑ 這裡用 go，所以 wsHandler 立刻返回
	// readPump 和 writePump 在背景持續運行，直到連線斷開
}

// statusHandler 回傳伺服器狀態（REST API，不是 WebSocket）
func statusHandler(hub *Hub, w http.ResponseWriter, r *http.Request) { // 狀態 API
	w.Header().Set("Content-Type", "application/json") // 設定回應格式
	fmt.Fprintf(w, `{"status":"ok","clients":%d,"time":"%s"}`,
		hub.ClientCount(),               // 目前連線人數
		time.Now().Format(time.RFC3339), // 目前時間
	)
}

// homeHandler 回傳 HTML 聊天室頁面（內嵌在 Go 程式中）
func homeHandler(w http.ResponseWriter, r *http.Request) { // 首頁 HTML
	w.Header().Set("Content-Type", "text/html; charset=utf-8") // 設定 HTML 格式
	fmt.Fprint(w, chatHTML)                                    // 輸出 HTML 頁面
}

// ==========================================================================
// 7. 聊天室 HTML 頁面（內嵌在 Go 程式中）
// ==========================================================================
//
// 這是一個完整的聊天室前端，使用瀏覽器內建的 WebSocket API
// 不需要安裝任何前端套件！

const chatHTML = `<!DOCTYPE html>
<html lang="zh-TW">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Go WebSocket 聊天室</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: sans-serif; background: #f0f2f5; height: 100vh; display: flex; flex-direction: column; }
        .header { background: #1a73e8; color: white; padding: 16px; text-align: center; }
        .header h1 { font-size: 20px; }
        .header p { font-size: 13px; opacity: 0.8; margin-top: 4px; }
        .chat-container { flex: 1; display: flex; flex-direction: column; max-width: 800px; margin: 16px auto; width: 100%; padding: 0 16px; }
        #messages { flex: 1; background: white; border-radius: 12px; padding: 16px; overflow-y: auto; min-height: 300px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .message { margin-bottom: 12px; }
        .message.system { text-align: center; color: #888; font-size: 13px; font-style: italic; }
        .message .meta { font-size: 12px; color: #888; margin-bottom: 2px; }
        .message .bubble { background: #e8f0fe; border-radius: 12px; padding: 8px 12px; display: inline-block; max-width: 80%; }
        .message.mine .meta { text-align: right; }
        .message.mine .bubble { background: #1a73e8; color: white; float: right; clear: both; }
        .status-bar { background: white; border-radius: 8px; padding: 8px 12px; margin: 8px 0; font-size: 13px; color: #666; display: flex; justify-content: space-between; }
        .status-dot { width: 8px; height: 8px; border-radius: 50%; background: #ccc; display: inline-block; margin-right: 6px; }
        .status-dot.connected { background: #34a853; }
        .input-area { background: white; border-radius: 12px; padding: 12px; display: flex; gap: 8px; margin-top: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        #username { width: 120px; padding: 8px; border: 1px solid #ddd; border-radius: 8px; font-size: 14px; }
        #messageInput { flex: 1; padding: 8px 12px; border: 1px solid #ddd; border-radius: 8px; font-size: 14px; }
        #sendBtn { padding: 8px 20px; background: #1a73e8; color: white; border: none; border-radius: 8px; cursor: pointer; font-size: 14px; }
        #sendBtn:hover { background: #1557b0; }
        #sendBtn:disabled { background: #ccc; cursor: not-allowed; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 Go WebSocket 聊天室</h1>
        <p>第二十三課示範 — 即時雙向通訊</p>
    </div>
    <div class="chat-container">
        <div class="status-bar">
            <span><span class="status-dot" id="statusDot"></span><span id="statusText">連線中...</span></span>
            <span id="clientCount">0 人在線</span>
        </div>
        <div id="messages"></div>
        <div class="input-area">
            <input type="text" id="username" placeholder="你的名字" value="訪客">
            <input type="text" id="messageInput" placeholder="輸入訊息..." disabled>
            <button id="sendBtn" onclick="connect()" >連線</button>
        </div>
    </div>

    <script>
        let ws = null;
        let myUsername = '';

        function connect() {
            myUsername = document.getElementById('username').value || '訪客';
            const wsUrl = 'ws://' + location.host + '/ws?username=' + encodeURIComponent(myUsername);

            ws = new WebSocket(wsUrl);

            ws.onopen = function() {
                document.getElementById('statusDot').className = 'status-dot connected';
                document.getElementById('statusText').textContent = '已連線';
                document.getElementById('messageInput').disabled = false;
                document.getElementById('sendBtn').textContent = '發送';
                document.getElementById('sendBtn').onclick = sendMessage;
                document.getElementById('username').disabled = true;
                addSystemMessage('✅ 已連線到聊天室！');
            };

            ws.onmessage = function(event) {
                const msg = JSON.parse(event.data);
                if (msg.type === 'join' || msg.type === 'leave') {
                    addSystemMessage(msg.content);
                } else {
                    addMessage(msg);
                }
            };

            ws.onclose = function() {
                document.getElementById('statusDot').className = 'status-dot';
                document.getElementById('statusText').textContent = '已斷線';
                document.getElementById('messageInput').disabled = true;
                document.getElementById('sendBtn').textContent = '重新連線';
                document.getElementById('sendBtn').onclick = connect;
                addSystemMessage('❌ 已斷線');
            };

            ws.onerror = function(err) {
                addSystemMessage('連線錯誤：' + err.message);
            };
        }

        function sendMessage() {
            const input = document.getElementById('messageInput');
            const content = input.value.trim();
            if (!content || !ws || ws.readyState !== WebSocket.OPEN) return;

            ws.send(JSON.stringify({ type: 'chat', content: content }));
            input.value = '';
        }

        document.getElementById('messageInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') sendMessage();
        });

        function addMessage(msg) {
            const isMine = msg.username === myUsername;
            const div = document.createElement('div');
            div.className = 'message' + (isMine ? ' mine' : '');
            const time = new Date(msg.timestamp).toLocaleTimeString('zh-TW', {hour: '2-digit', minute: '2-digit'});
            div.innerHTML = '<div class="meta">' + msg.username + ' · ' + time + '</div>' +
                            '<div class="bubble">' + escapeHtml(msg.content) + '</div>';
            document.getElementById('messages').appendChild(div);
            document.getElementById('messages').scrollTop = 999999;
        }

        function addSystemMessage(content) {
            const div = document.createElement('div');
            div.className = 'message system';
            div.textContent = content;
            document.getElementById('messages').appendChild(div);
            document.getElementById('messages').scrollTop = 999999;
        }

        function escapeHtml(text) {
            return text.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
        }

        // 定期更新在線人數
        setInterval(function() {
            fetch('/status').then(r => r.json()).then(data => {
                document.getElementById('clientCount').textContent = data.clients + ' 人在線';
            }).catch(() => {});
        }, 3000);
    </script>
</body>
</html>`

// ==========================================================================
// 主程式
// ==========================================================================

func main() { // 程式進入點
	fmt.Println("==========================================") // 分隔線
	fmt.Println(" 第二十三課：WebSocket 即時通訊")                      // 標題
	fmt.Println("==========================================") // 分隔線

	// 建立 Hub（廣播中心）
	hub := newHub() // 建立 Hub
	go hub.run()    // 在獨立 goroutine 中啟動 Hub

	// 設定路由
	http.HandleFunc("/", homeHandler) // 首頁：聊天室 HTML 頁面

	// WebSocket 端點（/ws）
	// 注意：WebSocket 路由不能用 http.HandleFunc 直接傳 hub，要用 closure
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		wsHandler(hub, w, r) // 把 hub 傳進去
	})

	// 狀態 API
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		statusHandler(hub, w, r) // 把 hub 傳進去
	})

	// 啟動 HTTP 伺服器
	addr := ":8080"                                  // 監聽 8080 埠
	fmt.Printf("🚀 伺服器啟動：http://localhost%s\n", addr) // 印出啟動訊息
	fmt.Println()
	fmt.Println("使用方式：")
	fmt.Println("  1. 用瀏覽器開啟 http://localhost:8080")
	fmt.Println("  2. 輸入名字，點「連線」")
	fmt.Println("  3. 再開一個瀏覽器分頁，用不同名字連線")
	fmt.Println("  4. 在兩個分頁之間即時聊天！")
	fmt.Println()
	fmt.Println("WebSocket 端點：ws://localhost:8080/ws?username=你的名字")
	fmt.Println("狀態 API：     http://localhost:8080/status")
	fmt.Println()
	fmt.Println("按 Ctrl+C 停止伺服器")

	if err := http.ListenAndServe(addr, nil); err != nil { // 啟動伺服器（阻塞）
		log.Fatalf("伺服器啟動失敗: %v", err) // 如果失敗，印出錯誤並結束
	}
}
