# 第三十七課：OpenTelemetry 分散式追蹤

> **一句話總結**：OpenTelemetry 就像包裹追蹤系統（FedEx/DHL），讓你追蹤每一個請求在多個服務之間的完整旅程，精準找出「時間花在哪裡」。

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解 Trace/Span 概念，能在單一服務中加入追蹤 |
| 🔴 資深工程師 | **必備**：能設計跨服務追蹤方案，使用 Jaeger/Tempo 分析效能瓶頸 |
| 🏢 SRE/DevOps | 建立可觀測性（Observability）平台：Logs + Metrics + Traces |

## 你會學到什麼？

- 可觀測性三大支柱：Logs（第 21 課）、Metrics（第 35 課）、Traces（本課）
- OpenTelemetry 核心概念：Trace、Span、SpanContext、Propagation
- 如何在 Go 程式中建立和管理 Span
- Span 的屬性（Attributes）、事件（Events）、狀態（Status）
- 跨服務的 Context 傳遞（W3C Trace Context）
- Exporter 設定：stdout、Jaeger、OTLP
- Gin middleware 整合
- 取樣策略：Always、Probability、Tail-based
- 搶票系統中如何追蹤跨 gRPC 服務的請求

## 執行方式

```bash
# 啟動範例程式（輸出 Span JSON 到 stdout）
go run ./tutorials/37-opentelemetry

# 搭配 Jaeger 查看視覺化 Trace
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest

# 啟動程式（使用 OTLP exporter）
OTEL_EXPORTER=otlp go run ./tutorials/37-opentelemetry

# 開啟 Jaeger UI
# http://localhost:16686
```

## 生活比喻：包裹追蹤系統

```
寄一個包裹從台北到紐約（= 一個使用者請求經過多個微服務）：

📦 包裹追蹤（= Trace）
│
├── 台北倉庫收件          09:00（= Span: API Gateway）
│   追蹤碼: TW-2024-001         ← TraceID（全程唯一）
│   站點編號: A                   ← SpanID
│
├── 桃園機場通關          09:30（= Span: Auth Service）
│   追蹤碼: TW-2024-001         ← 同一個 TraceID
│   站點編號: B                   ← 新的 SpanID
│   上一站: A                     ← ParentSpanID
│
├── 東京轉運中心          14:00（= Span: Order Service）
│   追蹤碼: TW-2024-001
│   站點編號: C
│   備註: 包裹重新分揀            ← Span Event
│
├── 紐約海關              20:00（= Span: Payment Service）
│   追蹤碼: TW-2024-001
│   站點編號: D
│   屬性: 重量=2kg, 類別=電子產品  ← Span Attributes
│
└── 收件人簽收            22:00（= Span: Notification Service）
    追蹤碼: TW-2024-001
    站點編號: E
    狀態: 成功送達                 ← Span Status: OK

每一站 = 一個 Span
整趟旅程 = 一個 Trace
追蹤碼 = TraceID（讓你追蹤完整路徑）
```

## 可觀測性三大支柱

可觀測性（Observability）是指你能從系統的「外部輸出」來理解系統的「內部狀態」。它由三大支柱組成：

| 支柱 | 回答的問題 | 工具 | 本系列課程 |
|------|-----------|------|-----------|
| **Logs（日誌）** | 「發生了什麼事？」 | zap + Loki | 第 21 課 |
| **Metrics（指標）** | 「系統整體狀況如何？」 | Prometheus + Grafana | 第 35 課 |
| **Traces（追蹤）** | 「這個請求經過了哪些服務？每一步花了多久？」 | OTel + Jaeger | **本課** |

```
三者的關係：

  使用者回報「下單很慢」
    │
    ▼
  Metrics：發現 P99 延遲從 200ms 飆到 3s      ← 發現「有問題」
    │
    ▼
  Traces：追蹤慢請求，發現時間花在 Payment 服務   ← 定位「哪裡有問題」
    │
    ▼
  Logs：查看 Payment 服務的 log，發現連不上銀行 API ← 理解「為什麼有問題」
```

> **重要**：三個支柱互相補充，單靠任何一個都不夠。OpenTelemetry 的願景是用一套統一的標準來收集這三種訊號。

## 核心概念

### Trace 和 Span

```
一個 Trace（追蹤）= 一個請求的完整旅程
一個 Span（跨度）= 旅程中的一個步驟

Trace（TraceID: abc123）
│
├── Span: HTTP GET /order/123           duration: 350ms  ← Root Span
│   │
│   ├── Span: validateOrder             duration: 10ms   ← Child Span
│   │
│   ├── Span: chargePayment             duration: 200ms  ← Child Span
│   │   │
│   │   └── Span: POST bank-api.com     duration: 180ms  ← Grandchild Span
│   │
│   └── Span: updateInventory           duration: 50ms   ← Child Span
│       │
│       └── Span: db.UPDATE inventory   duration: 30ms   ← DB Span
```

### SpanContext（跨服務傳遞的關鍵）

```
SpanContext 包含：
  - TraceID:    abc123def456...（128-bit，全程唯一）
  - SpanID:     789abc...（64-bit，每個 Span 唯一）
  - TraceFlags: 01（是否被取樣）
  - TraceState: 廠商特定資料

當一個服務呼叫另一個服務時，SpanContext 會透過 HTTP Header 傳遞：

  Service A                              Service B
  ┌─────────────┐   HTTP Request        ┌─────────────┐
  │ Span: call  │ ──────────────────→   │ Span: handle│
  │             │   Header:              │             │
  │ TraceID: abc│   traceparent:         │ TraceID: abc│ ← 同一個！
  │ SpanID:  111│   00-abc-111-01        │ SpanID:  222│ ← 新的
  └─────────────┘                        │ Parent:  111│ ← 連接起來
                                         └─────────────┘
```

## 基本用法：建立 Span

```go
package main

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
)

// Step 1：初始化 TracerProvider
func initTracer() (*sdktrace.TracerProvider, error) {
    // 建立 exporter（先用 stdout，之後換 Jaeger/OTLP）
    exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
    if err != nil {
        return nil, err
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),     // 批次送出（效能好）
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceName("blog-api"),        // 服務名稱
            semconv.ServiceVersion("1.0.0"),        // 服務版本
        )),
    )

    // 設定全域 TracerProvider
    otel.SetTracerProvider(tp)
    // 設定全域 Propagator（用 W3C Trace Context 標準）
    otel.SetTextMapPropagator(propagation.TraceContext{})

    return tp, nil
}

// Step 2：建立 Tracer（通常每個 package 一個）
var tracer = otel.Tracer("blog-api/handler")

// Step 3：在函式中建立 Span
func handleCreateArticle(ctx context.Context, article Article) error {
    // 建立一個新的 Span（自動成為子 Span）
    ctx, span := tracer.Start(ctx, "handleCreateArticle")
    defer span.End()  // 一定要 End！不然 Span 不會被送出

    // 設定屬性（key-value 資料，可以搜尋和篩選）
    span.SetAttributes(
        attribute.String("article.title", article.Title),
        attribute.Int("article.author_id", article.AuthorID),
        attribute.String("article.category", article.Category),
    )

    // 記錄事件（帶時間戳的日誌）
    span.AddEvent("開始驗證文章")

    // 呼叫子函式（傳遞 ctx，自動建立 parent-child 關係）
    if err := validateArticle(ctx, article); err != nil {
        // 記錄錯誤
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    span.AddEvent("驗證通過，開始儲存")

    if err := saveArticle(ctx, article); err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return err
    }

    span.SetStatus(codes.Ok, "文章建立成功")
    return nil
}

// 子函式也建立自己的 Span
func validateArticle(ctx context.Context, article Article) error {
    ctx, span := tracer.Start(ctx, "validateArticle")
    defer span.End()

    if article.Title == "" {
        return fmt.Errorf("title is required")
    }

    span.AddEvent("驗證通過")
    return nil
}
```

## Span 的三種重要資料

| 資料類型 | 用途 | 範例 |
|---------|------|------|
| **Attributes** | 靜態的 key-value 資料，用於搜尋和篩選 | `user.id=123`, `http.method=GET` |
| **Events** | 帶時間戳的事件紀錄 | 「快取命中」、「重試第 2 次」 |
| **Status** | Span 的最終狀態 | `Ok`, `Error` |

```go
// Attributes：描述「這個操作的性質」
span.SetAttributes(
    attribute.String("http.method", "POST"),
    attribute.String("http.url", "/api/articles"),
    attribute.Int("http.status_code", 201),
    attribute.String("db.system", "postgresql"),
    attribute.String("db.statement", "INSERT INTO articles..."),
)

// Events：記錄「過程中發生了什麼事」
span.AddEvent("cache_miss", trace.WithAttributes(
    attribute.String("cache.key", "article:123"),
))
span.AddEvent("retry", trace.WithAttributes(
    attribute.Int("retry.count", 2),
    attribute.String("retry.reason", "connection timeout"),
))

// Status：標記「結果是成功還是失敗」
span.SetStatus(codes.Ok, "")
span.SetStatus(codes.Error, "database connection refused")
```

## 跨服務 Context 傳遞（W3C Trace Context）

當 Service A 呼叫 Service B 時，需要把 TraceID 傳遞過去，讓兩邊的 Span 串在一起：

```go
// ===== Service A：發送方（注入 Trace Context 到 HTTP Header）=====
func callServiceB(ctx context.Context) error {
    req, _ := http.NewRequestWithContext(ctx, "GET", "http://service-b/api/data", nil)

    // 把 SpanContext 注入到 HTTP Header
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
    // 注入後的 Header：
    // traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
    //              ^^─version  ^^─trace-id                ^^─span-id    ^^─flags

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return nil
}

// ===== Service B：接收方（從 HTTP Header 提取 Trace Context）=====
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // 從 HTTP Header 提取 SpanContext
    ctx := otel.GetTextMapPropagator().Extract(
        r.Context(),
        propagation.HeaderCarrier(r.Header),
    )

    // 建立新的 Span（自動成為 Service A 的 child span）
    ctx, span := tracer.Start(ctx, "ServiceB.handleRequest")
    defer span.End()

    // 處理請求...
}
```

```
W3C Trace Context Header 格式：

traceparent: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
             ──  ────────────────────────────────  ────────────────  ──
             │           │                              │            │
          version    trace-id (128-bit)            span-id (64-bit) flags
                    （整個 Trace 唯一）          （這個 Span 唯一）  （01=已取樣）
```

## Exporter 設定

Exporter 決定 Trace 資料要送到哪裡：

| Exporter | 用途 | 適用場景 |
|----------|------|---------|
| **stdout** | 直接印到終端 | 開發/除錯 |
| **Jaeger** | 送到 Jaeger 後端 | 小型團隊 |
| **OTLP** | 送到任何 OTLP 相容後端 | 生產環境（推薦）|

```go
// ===== stdout exporter（開發用）=====
import "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"

exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())

// ===== Jaeger exporter =====
import "go.opentelemetry.io/otel/exporters/jaeger"

exporter, err := jaeger.New(
    jaeger.WithCollectorEndpoint(
        jaeger.WithEndpoint("http://localhost:14268/api/traces"),
    ),
)

// ===== OTLP exporter（推薦用於生產環境）=====
import "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"

exporter, err := otlptracehttp.New(ctx,
    otlptracehttp.WithEndpoint("localhost:4318"),
    otlptracehttp.WithInsecure(),  // 開發環境不用 TLS
)
```

> **推薦架構**：應用程式 → OTLP → OpenTelemetry Collector → Jaeger/Tempo/等後端。Collector 可以做取樣、轉換、分流，是生產環境的最佳實踐。

## 與 Gin Middleware 整合

```go
import "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

func setupRouter() *gin.Engine {
    r := gin.New()

    // 加入 OpenTelemetry middleware
    // 自動為每個 HTTP 請求建立 Span，並設定標準屬性
    r.Use(otelgin.Middleware("blog-api"))

    r.GET("/articles", getArticles)
    r.POST("/articles", createArticle)

    return r
}

// Handler 中可以直接從 ctx 取得 Span
func getArticles(c *gin.Context) {
    // otelgin middleware 已經建立了 root span
    // 這裡建立的是 child span
    ctx, span := tracer.Start(c.Request.Context(), "getArticles.queryDB")
    defer span.End()

    span.SetAttributes(
        attribute.Int("page", page),
        attribute.Int("limit", limit),
    )

    articles, err := repo.FindAll(ctx, page, limit)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        c.JSON(500, gin.H{"error": "internal error"})
        return
    }

    span.SetAttributes(attribute.Int("result.count", len(articles)))
    c.JSON(200, articles)
}
```

加入 middleware 後，每個 HTTP 請求自動產生的 Span 會包含：

| 屬性 | 範例值 |
|------|--------|
| `http.method` | `GET` |
| `http.target` | `/articles?page=1` |
| `http.status_code` | `200` |
| `http.route` | `/articles` |
| `net.host.name` | `localhost` |
| `net.host.port` | `8080` |

## 取樣策略（Sampling）

不是每個請求都需要追蹤——在高流量系統中，100% 追蹤會產生大量資料。取樣策略決定哪些 Trace 要保留：

```go
// ===== Always Sample（全部追蹤）=====
// 適合：開發環境、低流量服務
tp := sdktrace.NewTracerProvider(
    sdktrace.WithSampler(sdktrace.AlwaysSample()),
)

// ===== Probability Sample（機率取樣）=====
// 適合：中等流量，需要統計代表性
tp := sdktrace.NewTracerProvider(
    sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1)), // 10% 取樣
)

// ===== Parent-Based Sample（跟隨上游決定）=====
// 適合：微服務架構，確保整個 Trace 的取樣一致
tp := sdktrace.NewTracerProvider(
    sdktrace.WithSampler(sdktrace.ParentBased(
        sdktrace.TraceIDRatioBased(0.1), // root span 10% 取樣
    )),
)

// ===== Never Sample（完全不追蹤）=====
// 適合：極高流量且不需要追蹤的服務
tp := sdktrace.NewTracerProvider(
    sdktrace.WithSampler(sdktrace.NeverSample()),
)
```

| 策略 | 取樣率 | 資料量 | 適用場景 |
|------|--------|--------|---------|
| **AlwaysSample** | 100% | 最大 | 開發、低流量 |
| **TraceIDRatioBased(0.1)** | 10% | 中等 | 中等流量 |
| **TraceIDRatioBased(0.01)** | 1% | 最小 | 高流量生產環境 |
| **ParentBased** | 跟隨上游 | 取決於上游 | 微服務（推薦）|
| **Tail-based**（在 Collector） | 動態 | 智慧選擇 | 只保留慢請求/錯誤 |

> **Tail-based Sampling**：在 OpenTelemetry Collector 中設定，等 Trace 完成後再決定是否保留。可以設定「只保留超過 500ms 的請求」或「只保留有錯誤的請求」，讓你用最少的資料量抓到最有用的 Trace。

## 搶票系統中的分散式追蹤

在我們的搶票系統中，一個「買票」請求會經過多個服務：

```
使用者點「立即搶票」
│
▼
┌─────────────┐     gRPC      ┌──────────────┐     gRPC     ┌──────────────┐
│ API Gateway │ ────────────→ │ Ticket       │ ───────────→ │ Payment      │
│ (Gin)       │               │ Service      │              │ Service      │
│             │               │              │              │              │
│ Span: HTTP  │               │ Span: Check  │              │ Span: Charge │
│ POST /buy   │               │ Inventory    │              │ Credit Card  │
│ 5ms         │               │ 20ms         │              │ 200ms        │
└─────────────┘               └──────┬───────┘              └──────────────┘
                                     │
                                     │ Redis
                                     ▼
                              ┌──────────────┐
                              │ Redis        │
                              │              │
                              │ Span: DECR   │
                              │ inventory    │
                              │ 2ms          │
                              └──────────────┘

在 Jaeger UI 中看到的 Trace：

  HTTP POST /buy          ████████████████████████████████████████  227ms
    │
    ├─ Check Inventory    ████████                                  22ms
    │   │
    │   └─ Redis DECR     ██                                        2ms
    │
    └─ Charge Credit Card ████████████████████████████████          200ms

  → 一眼就能看出：瓶頸在 Payment Service 的信用卡扣款！
```

### gRPC 中的 Context 傳遞

```go
// gRPC 自動傳遞 Trace Context（透過 metadata）
import "go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

// Client 端
conn, err := grpc.Dial(
    "ticket-service:50051",
    grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
    grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
)

// Server 端
server := grpc.NewServer(
    grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
    grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
)
```

## 什麼時候建立 Span？什麼時候用 Event？

| 情況 | 用 Span | 用 Event |
|------|---------|---------|
| 呼叫外部服務（HTTP/gRPC/DB） | 是 | |
| 重要的業務邏輯步驟 | 是 | |
| 快取命中/未命中 | | 是 |
| 重試 | | 是 |
| 輸入驗證 | 看情況（很快就用 Event） | 看情況 |
| 單一服務內的小函式 | | 是（避免太多 Span）|

> **原則**：Span 有開銷（建立、傳送、儲存），不要對每個函式都建立 Span。一般來說，跨越服務邊界（HTTP、gRPC、DB 查詢）的操作用 Span，同一服務內的小步驟用 Event。

## FAQ

### Q1：OpenTelemetry 和 Jaeger/Zipkin 是什麼關係？

OpenTelemetry 是一套「標準」和「SDK」，負責在你的程式碼中收集 Trace 資料。Jaeger 和 Zipkin 是「後端」，負責儲存和視覺化 Trace 資料。就像 OpenTelemetry 是相機（拍照），Jaeger 是相簿（存照片和看照片）。你也可以把 OpenTelemetry 的資料送到 Grafana Tempo、Datadog 等其他後端。

### Q2：加了追蹤會影響效能嗎？

會有一點點影響，但通常很小。每個 Span 大約增加幾微秒的開銷。如果擔心效能，可以用取樣策略（例如 10% 取樣）來降低開銷。在大多數情況下，追蹤帶來的除錯便利遠超過微小的效能代價。

### Q3：Trace 和 Log 有什麼不同？什麼時候該用哪個？

Log 記錄的是「事件」，適合記錄詳細的除錯資訊。Trace 記錄的是「因果關係」和「時間分布」，適合理解請求流經多個服務的路徑和耗時。最佳實踐是在 Log 中加入 TraceID，這樣你可以從 Trace 跳轉到相關的 Log，反之亦然。

### Q4：為什麼推薦 OTLP 而不是直接用 Jaeger exporter？

OTLP（OpenTelemetry Protocol）是 OpenTelemetry 的原生協議，所有後端都支援。使用 OTLP 的好處是：(1) 如果以後要換後端（從 Jaeger 換到 Tempo），不需要改程式碼；(2) 可以搭配 OpenTelemetry Collector 做取樣、轉換、分流；(3) 同時支援 Traces、Metrics、Logs。

### Q5：如何在本地開發時快速看到 Trace？

最簡單的方式是用 stdout exporter，Trace 資料會直接印在終端。想要視覺化的話，用一行指令啟動 Jaeger：`docker run -d -p 16686:16686 -p 4318:4318 jaegertracing/all-in-one`，然後用 OTLP exporter 把資料送到 `localhost:4318`，打開 `http://localhost:16686` 就能看到了。

## 練習

1. 在一個 HTTP handler 中建立 Span，加上 `user.id` 和 `article.id` 屬性
2. 在 usecase 層建立子 Span，觀察 parent-child 關係
3. 用 SpanEvent 記錄「快取命中」和「快取未命中」事件
4. 設定 Jaeger exporter（`docker run jaegertracing/all-in-one`），在 UI 中查看 trace
5. 思考並實作：在搶票系統中，為「檢查庫存 → 扣減庫存 → 建立訂單」三個步驟各建立一個 Span，觀察完整的 Trace

## 下一課預告

下一課我們會學習 **Rate Limiting（限流）**——當搶票活動開始，瞬間湧入的請求可能壓垮你的系統。限流就像高速公路的匝道管制，控制進入的流量，保護系統不被打掛。搭配本課的 Trace，你可以清楚看到被限流的請求和正常請求的差異。
