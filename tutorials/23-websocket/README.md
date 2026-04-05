# 第二十三課：WebSocket 即時通訊

> **一句話總結**：HTTP 是「發簡訊」（問一句答一句），WebSocket 是「打電話」（保持通話，雙方隨時可以說話）。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解 WebSocket 協議，實作即時通訊功能 |
| 🔴 資深工程師 | Hub 模式、連線管理、Heartbeat、大規模連線（水平擴展） |

## 你會學到什麼？

- WebSocket 和 HTTP 的根本差異
- `gorilla/websocket`：Go 最流行的 WebSocket 套件
- **Upgrader**：把 HTTP 連線升級成 WebSocket
- **Client**：代表單一 WebSocket 連線（readPump + writePump）
- **Hub 模式**：管理所有連線、廣播訊息的核心模式
- 心跳機制（Ping/Pong）：偵測斷線
- 完整可運行的聊天室（前後端都在這個程式裡）

## 執行方式

```bash
go run ./tutorials/23-websocket
```

然後：
1. 用瀏覽器開啟 `http://localhost:8080`
2. 輸入名字，點「連線」
3. **再開一個新分頁**，用不同名字連線
4. 在兩個分頁之間即時聊天！

## HTTP vs WebSocket

```
HTTP（發簡訊）：
  瀏覽器 ──「給我文章列表」──▶ 伺服器
  瀏覽器 ◀──「好，給你」────── 伺服器
  （連線結束）

  缺點：伺服器「無法主動」通知瀏覽器
  → 想要即時更新？只能每秒問一次（輪詢，浪費資源）


WebSocket（打電話）：
  瀏覽器 ──「我要建立長連線」──▶ 伺服器
  瀏覽器 ◀──「好，保持連線」──── 伺服器
  （通道持續開著）

  瀏覽器 ────── 「你好」 ─────▶ 伺服器
  瀏覽器 ◀───── 「你也好」 ───── 伺服器
  伺服器 ──「有新訊息通知你」──▶ 瀏覽器  ← 伺服器主動推送！
```

## 連線建立流程

```
瀏覽器                           伺服器
  │                                │
  │── GET /ws HTTP/1.1 ───────────▶│  ← 先用 HTTP
  │   Upgrade: websocket           │
  │   Connection: Upgrade          │
  │                                │
  │◀── HTTP/1.1 101 Switching ─────│  ← 同意升級
  │    Upgrade: websocket          │
  │                                │
  │════════ WebSocket ═════════════│  ← 升級成功！
  │                                │
  │── {"type":"chat","content":"hi"}▶│
  │◀── {"type":"chat","username":"Bob",...} │
```

## 架構設計：Hub 模式

```
                 ┌─────────────────────────────┐
                 │            Hub              │
                 │  clients: map[*Client]bool  │
                 │  ┌─────────────────────┐    │
  Alice 連線 ──▶ │  │ register channel    │    │
  Bob 斷線  ──▶ │  │ unregister channel  │    │
  Alice 說話──▶ │  │ broadcast channel   │    │
                 │  └─────────────────────┘    │
                 │        ↓（只有 Hub 的        │
                 │         goroutine 修改 map） │
                 └─────────────────────────────┘
                          ↓ 廣播
              ┌───────────┼───────────┐
           Client A    Client B    Client C
           (Alice)      (Bob)      (Carol)
          readPump    readPump    readPump
          writePump   writePump   writePump
```

**為什麼要用 Hub？**

如果 Alice 說了一句話，需要發給 Bob 和 Carol。
如果直接在 Alice 的 goroutine 裡寫入 Bob 的連線，會有並發問題（多個 goroutine 同時寫同一個 WebSocket 連線）。

Hub 是唯一的「廣播員」，在單一 goroutine 中管理所有操作，不需要 Mutex。

## 每個連線的生命週期

```go
// 新連線進來
conn, _ := upgrader.Upgrade(w, r, nil)
client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
hub.register <- client  // 通知 Hub

// 兩個 goroutine 並行運行
go client.readPump()   // 讀：從 WebSocket 讀訊息 → 發到 Hub
go client.writePump()  // 寫：從 send channel 讀 → 寫入 WebSocket

// 斷線時
hub.unregister <- client  // 通知 Hub 移除
close(client.send)         // 通知 writePump 退出
conn.Close()               // 關閉連線
```

## 程式碼結構

```
main.go
├── Message struct      — 訊息格式（JSON）
├── Client struct        — 代表一個 WebSocket 連線
│   ├── readPump()       — 持續讀取客戶端訊息
│   └── writePump()      — 持續把訊息發給客戶端
├── Hub struct           — 廣播中心
│   ├── run()            — Hub 的主迴圈
│   └── broadcastToAll() — 廣播給所有人
├── upgrader             — HTTP → WebSocket 升級器
├── wsHandler()          — WebSocket 路由處理器
├── statusHandler()      — 狀態 API
├── homeHandler()        — 回傳 HTML 頁面
├── chatHTML             — 內嵌的聊天室前端
└── main()               — 設定路由、啟動伺服器
```

## 關鍵 API

### 伺服器端（gorilla/websocket）

```go
// 升級 HTTP → WebSocket
upgrader := websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}
conn, err := upgrader.Upgrade(w, r, nil)

// 讀取訊息
messageType, data, err := conn.ReadMessage()

// 寫入訊息
err = conn.WriteMessage(websocket.TextMessage, data)

// 使用 Writer（批量寫入，更有效率）
w, _ := conn.NextWriter(websocket.TextMessage)
w.Write(data)
w.Close()

// 關閉
conn.WriteMessage(websocket.CloseMessage, []byte{})
conn.Close()
```

### 瀏覽器端（原生 WebSocket API）

```javascript
// 建立連線
const ws = new WebSocket('ws://localhost:8080/ws?username=Alice')

// 事件監聽
ws.onopen    = () => console.log('已連線')
ws.onmessage = (event) => console.log('收到：', event.data)
ws.onclose   = () => console.log('已斷線')
ws.onerror   = (err) => console.error('錯誤', err)

// 發送訊息
ws.send(JSON.stringify({ type: 'chat', content: '你好！' }))

// 主動關閉
ws.close()
```

## 心跳機制（Ping/Pong）

```
為什麼需要心跳？

  用戶把手機放進口袋 → 網路靜默 → 伺服器以為連線正常
  但其實手機已經沒網路了！

心跳流程（每 30 秒）：
  伺服器 ──── Ping ────▶ 客戶端
  伺服器 ◀─── Pong ───── 客戶端  ← 更新 read deadline

  如果 60 秒都沒收到 Pong → SetReadDeadline 觸發 → 連線斷開
```

```go
// 發送 Ping（在 writePump 的 ticker 中）
conn.WriteMessage(websocket.PingMessage, nil)

// 處理 Pong（更新 deadline）
conn.SetPongHandler(func(string) error {
    conn.SetReadDeadline(time.Now().Add(60 * time.Second))
    return nil
})
```

## 常見問題 FAQ

### Q: 為什麼要兩個 goroutine（readPump + writePump）？

WebSocket 連線的讀和寫必須分開，因為它們都是阻塞操作：
- `ReadMessage()` 會阻塞到有新訊息
- `WriteMessage()` 會阻塞到訊息發出

如果只用一個 goroutine，讀的時候就沒辦法寫，或寫的時候就沒辦法讀。

### Q: 為什麼 send channel 要有緩衝（256）？

```go
send: make(chan []byte, 256)
```

如果 send 是無緩衝的，廣播時 Hub 的 goroutine 要等每個客戶端都接收完才能繼續。某個慢客戶端會卡住所有廣播。

有緩衝後，Hub 把訊息放進 channel 就繼續，客戶端的 writePump 自己慢慢消費。如果 channel 滿了（256 條都還沒發），就強制斷開這個慢客戶端。

### Q: CheckOrigin 在正式環境要怎麼設定？

```go
// 示範用（允許所有）
CheckOrigin: func(r *http.Request) bool { return true }

// 正式環境（只允許指定網域）
CheckOrigin: func(r *http.Request) bool {
    origin := r.Header.Get("Origin")
    return origin == "https://yourdomain.com"
}
```

### Q: 如何支援多個聊天室？

目前 Hub 是一個全域廣播。要支援多個聊天室：
1. 用 `map[string]*Hub`，每個房間一個 Hub
2. 廣播時只發給同房間的客戶端（在 `broadcastToAll` 加 roomID 過濾）

## 練習

1. **多聊天室**：修改 Hub，讓 `broadcastToAll` 只廣播給同一 `roomID` 的客戶端
2. **私訊**：加入 `to` 欄位，若 `type == "private"`，只發給指定 username 的客戶端
3. **歷史訊息**：用 Redis List（第二十課）儲存最近 50 條訊息，新客戶端連線時把歷史訊息發過去
4. **上線名單**：加一個 `/users` API，回傳目前在線的所有使用者名稱

## 下一課預告

**第二十四課：Clean Architecture 進階** —— 把前面所有技術（Gin + GORM + Redis + zap + JWT）整合進一個完整的 Clean Architecture 專案，實作依賴注入（Dependency Injection）。
