# 第三十課：gRPC 基礎

> **一句話總結**：gRPC 讓你像呼叫本地函式一樣呼叫遠端服務，比 REST 快 5-10 倍，是微服務標配。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解 gRPC 概念，與 REST API 的差異 |
| 🔴 資深工程師 | **重點**：微服務架構中 gRPC 是服務間通訊的主流 |

## 你會學到什麼？

- Protocol Buffers（Protobuf）訊息定義語法
- `protoc` 如何生成 Go 程式碼（完整步驟）
- 四種 RPC 通訊模式（Unary、Server Streaming、Client Streaming、Bidirectional）
- gRPC 錯誤處理與狀態碼對應
- Interceptor（攔截器）— gRPC 版的 Middleware
- gRPC vs REST 的完整比較與選擇時機

## 執行方式

```bash
go run ./tutorials/30-grpc
```

---

## Protocol Buffers（Protobuf）訊息定義

gRPC 使用 `.proto` 檔案定義服務介面和訊息格式。這是一種**介面描述語言（IDL）**，用來取代 JSON Schema 或 Swagger。

### .proto 檔案語法

```protobuf
// article.proto
syntax = "proto3";                    // 使用 proto3 語法（目前主流）
package article;                      // 套件名稱（避免命名衝突）
option go_package = "./pb/article";   // 生成的 Go 程式碼放在哪個目錄

// 訊息定義（類似 Go 的 struct）
message Article {
  uint64 id       = 1;    // 欄位編號（不是值！用於二進位編碼）
  string title    = 2;    // 每個欄位有 type + name + number
  string content  = 3;
  string status   = 4;
  string author_id = 5;
}

// 請求和回應也是 message
message GetArticleRequest {
  uint64 id = 1;
}

message ListArticlesResponse {
  repeated Article articles = 1;   // repeated = 陣列
  int64 total = 2;
}

// 服務定義（定義有哪些 RPC 方法）
service ArticleService {
  rpc GetArticle(GetArticleRequest) returns (Article);               // Unary
  rpc ListArticles(ListArticlesRequest) returns (ListArticlesResponse);
  rpc WatchArticles(WatchRequest) returns (stream Article);          // Server Streaming
  rpc BatchCreate(stream CreateArticleRequest) returns (BatchResult); // Client Streaming
  rpc Chat(stream ChatMessage) returns (stream ChatMessage);         // Bidirectional
}
```

### Protobuf 欄位編號規則

| 規則 | 說明 |
|------|------|
| 編號 1-15 | 用 1 byte 編碼，給最常用的欄位 |
| 編號 16-2047 | 用 2 bytes 編碼 |
| 不能重複使用 | 一旦分配就不能改，否則破壞向後相容 |
| 不能用 19000-19999 | protobuf 內部保留 |

### 常見 Protobuf 型別對照

| Protobuf | Go | 說明 |
|----------|-----|------|
| `string` | `string` | UTF-8 字串 |
| `int32` / `int64` | `int32` / `int64` | 整數 |
| `uint64` | `uint64` | 無符號整數 |
| `bool` | `bool` | 布林值 |
| `bytes` | `[]byte` | 二進位資料 |
| `float` / `double` | `float32` / `float64` | 浮點數 |
| `repeated T` | `[]T` | 陣列 |
| `map<K, V>` | `map[K]V` | 映射 |
| `google.protobuf.Timestamp` | `timestamppb.Timestamp` | 時間戳 |

---

## protoc 生成 Go 程式碼（完整步驟）

```
                 article.proto
                      │
                      ▼
               ┌─────────────┐
               │   protoc     │  Protocol Buffer Compiler
               │  (編譯器)    │
               └──────┬──────┘
                      │
          ┌───────────┼───────────┐
          ▼                       ▼
  protoc-gen-go            protoc-gen-go-grpc
  (生成訊息型別)           (生成服務介面)
          │                       │
          ▼                       ▼
  article.pb.go            article_grpc.pb.go
  - Article struct         - ArticleServiceServer interface
  - GetArticleRequest      - ArticleServiceClient interface
  - Marshal/Unmarshal      - RegisterArticleServiceServer()
```

### 步驟一：安裝工具

```bash
# 安裝 protoc（Protocol Buffer Compiler）
# macOS
brew install protobuf

# Linux
apt install -y protobuf-compiler

# 安裝 Go 的 protoc 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 步驟二：執行 protoc

```bash
protoc \
  --go_out=. \              # 生成 Go 訊息型別 → article.pb.go
  --go_opt=paths=source_relative \
  --go-grpc_out=. \         # 生成 gRPC 服務介面 → article_grpc.pb.go
  --go-grpc_opt=paths=source_relative \
  article.proto
```

### 步驟三：生成的程式碼長什麼樣

```go
// article.pb.go（自動生成，不要手動修改）
type Article struct {
    Id       uint64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
    Title    string `protobuf:"bytes,2,opt,name=title,proto3" json:"title,omitempty"`
    Content  string `protobuf:"bytes,3,opt,name=content,proto3" json:"content,omitempty"`
    // ... Marshal/Unmarshal 方法
}

// article_grpc.pb.go（自動生成）
type ArticleServiceServer interface {
    GetArticle(context.Context, *GetArticleRequest) (*Article, error)
    ListArticles(context.Context, *ListArticlesRequest) (*ListArticlesResponse, error)
}

// 你只需要「實作」這個介面
type myArticleServer struct {
    UnimplementedArticleServiceServer  // 嵌入這個可以只實作需要的方法
}
func (s *myArticleServer) GetArticle(ctx context.Context, req *GetArticleRequest) (*Article, error) {
    // 你的業務邏輯
}
```

---

## 四種 RPC 通訊模式比較

| 模式 | 語法（.proto） | 請求 | 回應 | 適用場景 |
|------|---------------|------|------|---------|
| **Unary** | `rpc Get(Req) returns (Resp)` | 1 個 | 1 個 | 一般 CRUD、查詢 |
| **Server Streaming** | `rpc Watch(Req) returns (stream Resp)` | 1 個 | N 個 | 即時推播、大量資料下載 |
| **Client Streaming** | `rpc Upload(stream Req) returns (Resp)` | N 個 | 1 個 | 批量上傳、IoT 資料串流 |
| **Bidirectional** | `rpc Chat(stream Req) returns (stream Resp)` | N 個 | N 個 | 即時聊天、多人遊戲 |

```
Unary（一問一答）:
  Client ──req──▶ Server
  Client ◀──resp── Server

Server Streaming（伺服器推流）:
  Client ──req──▶ Server
  Client ◀──resp1── Server
  Client ◀──resp2── Server
  Client ◀──resp3── Server

Client Streaming（客戶端推流）:
  Client ──req1──▶ Server
  Client ──req2──▶ Server
  Client ──req3──▶ Server
  Client ◀──resp── Server

Bidirectional（雙向串流）:
  Client ──req1──▶ Server
  Client ◀──resp1── Server
  Client ──req2──▶ Server
  Client ◀──resp2── Server
```

---

## gRPC 錯誤處理與狀態碼

gRPC 有自己的一套狀態碼（不同於 HTTP），用 `status` 套件建立和解析。

### 狀態碼對照表

| gRPC Code | HTTP 對應 | 使用場景 | 範例 |
|-----------|----------|---------|------|
| `OK` (0) | 200 | 成功 | — |
| `Canceled` (1) | 499 | 客戶端取消 | 使用者取消操作 |
| `InvalidArgument` (3) | 400 | 參數錯誤 | ID 為空、格式不對 |
| `NotFound` (5) | 404 | 資源不存在 | 找不到文章 |
| `AlreadyExists` (6) | 409 | 重複建立 | 帳號已存在 |
| `PermissionDenied` (7) | 403 | 無權限 | 非管理員 |
| `Unauthenticated` (16) | 401 | 未登入 | Token 無效 |
| `ResourceExhausted` (8) | 429 | 超出限制 | Rate limit |
| `FailedPrecondition` (9) | 400 | 前置條件不滿足 | 餘額不足 |
| `Internal` (13) | 500 | 伺服器錯誤 | 未預期的 panic |
| `Unavailable` (14) | 503 | 服務不可用 | 暫時無法連線 |
| `DeadlineExceeded` (4) | 504 | 超時 | 請求太慢 |

### 發送錯誤（伺服器端）

```go
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func (s *server) GetArticle(ctx context.Context, req *GetArticleRequest) (*Article, error) {
    if req.Id == 0 {
        return nil, status.Error(codes.InvalidArgument, "ID 不能為 0")
    }
    article, err := s.repo.Find(req.Id)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "文章 %d 不存在", req.Id)
    }
    return article, nil
}
```

### 接收錯誤（客戶端）

```go
resp, err := client.GetArticle(ctx, &GetArticleRequest{Id: 999})
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        switch st.Code() {
        case codes.NotFound:
            fmt.Println("文章不存在")
        case codes.InvalidArgument:
            fmt.Println("參數錯誤:", st.Message())
        default:
            fmt.Println("未知錯誤:", st.Code(), st.Message())
        }
    }
}
```

---

## Interceptor（攔截器）— gRPC 的 Middleware

Interceptor 和 HTTP 的 Middleware 概念相同：每次 RPC 呼叫前後都會經過，可用來做日誌、認證、監控。

```
Request → [Interceptor 1] → [Interceptor 2] → Handler
                                                   │
Response ← [Interceptor 1] ← [Interceptor 2] ←────┘
```

### Unary Interceptor（一問一答）

```go
// 簽名：每個 interceptor 都接收 context、request、info、handler
func loggingInterceptor(
    ctx context.Context,
    req any,
    info *grpc.UnaryServerInfo,    // 包含 FullMethod（如 "/article.ArticleService/GetArticle"）
    handler grpc.UnaryHandler,      // 下一個 handler（可能是下一個 interceptor 或實際方法）
) (any, error) {
    start := time.Now()
    resp, err := handler(ctx, req)    // 呼叫實際 handler
    log.Printf("[gRPC] %s took %v", info.FullMethod, time.Since(start))
    return resp, err
}
```

### Stream Interceptor（串流）

```go
func streamLoggingInterceptor(
    srv any,
    ss grpc.ServerStream,
    info *grpc.StreamServerInfo,
    handler grpc.StreamHandler,
) error {
    log.Printf("[gRPC Stream] %s started", info.FullMethod)
    err := handler(srv, ss)
    log.Printf("[gRPC Stream] %s ended", info.FullMethod)
    return err
}
```

### 常見 Interceptor 用途

| 用途 | 說明 |
|------|------|
| **Logging** | 記錄每次 RPC 的方法名稱、耗時、狀態碼 |
| **Authentication** | 從 metadata 讀取 token，驗證身份 |
| **Rate Limiting** | 限制每秒請求數 |
| **Recovery** | 捕捉 panic，轉為 `codes.Internal` 錯誤 |
| **Tracing** | 注入 OpenTelemetry trace ID |
| **Metrics** | 記錄 Prometheus 指標 |

---

## gRPC vs REST 完整比較

| 比較項目 | REST API | gRPC |
|---------|---------|------|
| **傳輸格式** | JSON（文字，人類可讀） | Protobuf（二進位，小 3-10 倍） |
| **協定** | HTTP/1.1（一問一答） | HTTP/2（多路複用、雙向串流） |
| **型別安全** | 弱（靠文件、Swagger） | 強（.proto 定義，編譯時檢查） |
| **程式碼生成** | 手動或用 openapi-generator | 自動（protoc） |
| **串流** | 不原生支援（需 WebSocket） | 原生支援四種模式 |
| **瀏覽器支援** | 原生支援 | 需要 gRPC-Web proxy |
| **偵錯便利性** | curl 直接測 | 需要 grpcurl 或 Postman |
| **適合場景** | 對外公開 API、第三方整合 | 微服務內部通訊、高效能場景 |
| **學習曲線** | 低 | 中（需學 protobuf、protoc） |
| **延遲** | 較高（JSON 序列化開銷） | 較低（二進位序列化，HTTP/2） |

### 什麼時候選 gRPC？

```
選 REST：
  ✅ 對外公開 API（瀏覽器、第三方）
  ✅ 團隊不熟悉 protobuf
  ✅ 需要快速原型開發

選 gRPC：
  ✅ 微服務之間的內部通訊
  ✅ 需要串流（即時推播、大量資料）
  ✅ 對延遲和吞吐量有嚴格要求
  ✅ 多語言團隊（.proto 跨語言共用）
```

---

## 與搶票系統的連結

搶票系統中的**支付服務**就是用 gRPC 實作的（`ticket-system/internal/grpc/payment/`）：

```
搶票流程：
  使用者搶票 → HTTP API → TicketUsecase → 建立訂單
                                              │
                                              ▼
                                    MQ（訊息佇列）
                                              │
                                              ▼
                                     PaymentWorker
                                              │
                                    gRPC 呼叫 ▼
                                ┌──────────────────────┐
                                │  Payment gRPC Server │
                                │  ProcessPayment()    │
                                └──────────────────────┘
```

在 `ticket-system/internal/grpc/payment/server.go` 中：

```go
// 支付服務介面
type PaymentServiceServer interface {
    ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
}

// server 端實作：模擬支付（含隨機失敗和延遲）
func (s *paymentServer) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
    // 模擬網路延遲 + 隨機失敗
    // 用 status.Error(codes.Canceled, ...) 處理 context 取消
}
```

注意：搶票系統用 **JSON codec** 取代 protobuf，讓教學可以直接執行，不需安裝 protoc。

---

## 常見問題 FAQ

### Q: gRPC 和 REST 能共存嗎？

可以。常見做法是對外用 REST（瀏覽器友好），服務間用 gRPC（效率高）。也可以用 gRPC-Gateway 自動把 gRPC 轉成 REST API。

### Q: 一定要用 protobuf 嗎？

不一定。gRPC 預設用 protobuf，但也支援 JSON codec（搶票系統就是用 JSON codec 來簡化教學）。不過生產環境建議用 protobuf，序列化速度快 5-10 倍。

### Q: gRPC 能用在瀏覽器嗎？

原生 gRPC 不行（需要 HTTP/2 + 二進位框架）。但可以用 gRPC-Web 或 Connect 框架讓瀏覽器透過 HTTP/1.1 呼叫 gRPC 服務。

### Q: Streaming 什麼時候該用？

- **Server Streaming**：伺服器要回傳大量資料（如日誌串流、即時報價）
- **Client Streaming**：客戶端要上傳大量資料（如檔案上傳）
- **Bidirectional**：即時雙向通訊（如聊天、遊戲）
- 大多數情況 **Unary 就夠了**，不要為了用 Streaming 而用

### Q: gRPC 的錯誤處理和 HTTP 狀態碼有什麼關係？

gRPC 有自己的 status code 體系（如 `codes.NotFound`、`codes.Internal`），和 HTTP 狀態碼是不同的。gRPC-Gateway 會自動把 gRPC status code 對應到 HTTP 狀態碼（如 `NotFound` → 404）。

---

## 練習題

### 練習 1：新增一個 UpdateArticle RPC
在本課的 `articleService` 中，新增 `UpdateArticle` 方法：
- 接收 `ID`、`Title`、`Content`
- 如果文章不存在，回傳 `codes.NotFound`
- 如果 `Title` 為空，回傳 `codes.InvalidArgument`
- 更新成功回傳更新後的 `Article`

### 練習 2：實作 Recovery Interceptor
寫一個 `recoveryInterceptor`，在 handler panic 時：
- 捕捉 panic
- 記錄錯誤日誌
- 回傳 `status.Error(codes.Internal, "internal server error")`
- 不讓整個 server 崩潰

### 練習 3：Metadata 傳遞 Request ID
修改客戶端和 loggingInterceptor：
- 客戶端每次呼叫自動加上 `request-id`（UUID）
- 伺服器端 interceptor 讀取並記錄 `request-id`
- 如果缺少 `request-id`，自動生成一個

### 練習 4：模擬 Server Streaming
不使用 protobuf，手動實作一個 Server Streaming 的效果：
- 建立一個 `WatchArticles` 方法
- 使用 Go channel 模擬串流
- 每秒推送一篇新文章給客戶端
- 客戶端用 `context.WithTimeout` 在 5 秒後停止

### 練習 5：gRPC 健康檢查
實作 gRPC Health Checking Protocol：
- 新增 `HealthService`，提供 `Check` 和 `Watch` 方法
- `Check` 回傳目前服務狀態（SERVING / NOT_SERVING）
- 思考：為什麼 Kubernetes 需要 gRPC 健康檢查，而不是用 HTTP？
