# 第二十七課：gRPC 基礎

> **一句話總結**：gRPC 讓你像呼叫本地函式一樣呼叫遠端服務，比 REST 快 5-10 倍，是微服務標配。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解 gRPC 概念，與 REST API 的差異 |
| 🔴 資深工程師 | **重點**：微服務架構中 gRPC 是服務間通訊的主流 |

## 執行方式

```bash
go run ./tutorials/27-grpc
```

## gRPC vs REST

| | REST API | gRPC |
|--|---------|------|
| 格式 | JSON（人類可讀）| Protocol Buffers（二進位，小 3-10 倍）|
| 協定 | HTTP/1.1 | HTTP/2（多路複用）|
| 型別 | 弱（靠文件）| 強（.proto 定義）|
| 適合 | 對外公開 API | 微服務內部通訊 |

## 四種通訊模式

```
1. Unary（一問一答）← 本課示範
   client.GetArticle(req) → article

2. Server Streaming（伺服器推流）
   client.Watch(req) → stream of events

3. Client Streaming（客戶端推流）
   client.BatchCreate(stream) → result

4. Bidirectional（雙向串流）
   client ↔ server（即時通訊）
```

## 完整使用 protobuf（正式環境）

```bash
# 安裝工具
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 從 .proto 生成程式碼
protoc --go_out=. --go-grpc_out=. article.proto
```

## gRPC 狀態碼

| codes.Code | 對應 HTTP | 使用場景 |
|-----------|----------|---------|
| `OK` | 200 | 成功 |
| `NotFound` | 404 | 資源不存在 |
| `InvalidArgument` | 400 | 參數錯誤 |
| `Unauthenticated` | 401 | 未登入 |
| `PermissionDenied` | 403 | 無權限 |
| `Internal` | 500 | 伺服器錯誤 |
| `Canceled` | — | 客戶端取消 |
