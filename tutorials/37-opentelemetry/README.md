# 第三十七課：OpenTelemetry 分散式追蹤

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解 Trace/Span 概念，能在單一服務中加入追蹤 |
| 🔴 資深工程師 | **必備**：能設計跨服務追蹤方案，使用 Jaeger/Tempo 分析效能瓶頸 |
| 🏢 SRE/DevOps | 建立可觀測性（Observability）平台：Logs + Metrics + Traces |

## 核心概念

```
Trace（一個請求的完整追蹤）
└── Span（handleOrder） ← 根 Span
    ├── Span（validateOrder）
    ├── Span（chargePayment）← Client Span（呼叫外部 API）
    └── Span（updateInventory）
          └── Span（db.update） ← DB Span
```

## 基本用法

```go
// 初始化
tp := sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(exporter),
    sdktrace.WithResource(res),
)
otel.SetTracerProvider(tp)

// 建立 Tracer
tracer := otel.Tracer("my-service")

// 建立 Span
func handleRequest(ctx context.Context) error {
    ctx, span := tracer.Start(ctx, "handleRequest")
    defer span.End()

    // 設定屬性
    span.SetAttributes(attribute.String("user.id", "123"))

    // 記錄事件
    span.AddEvent("開始處理")

    // 記錄錯誤
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    }
    return nil
}
```

## 跨服務傳遞 TraceID

```go
// 發送方（注入到 HTTP Header）
otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
// Header: traceparent: 00-4bf92f3577b34da6-00f067aa0ba902b7-01

// 接收方（從 HTTP Header 提取）
ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(req.Header))
```

## 三大可觀測性支柱

| 支柱 | 工具 | 課程 |
|------|------|------|
| **Logs（日誌）** | zap + Loki | 第 21 課 |
| **Metrics（指標）** | Prometheus + Grafana | 第 29 課 |
| **Traces（追蹤）** | OTel + Jaeger | 本課 |

## 執行方式

```bash
go run ./tutorials/37-opentelemetry
# 查看 stdout 輸出的 Span JSON 數據
```
