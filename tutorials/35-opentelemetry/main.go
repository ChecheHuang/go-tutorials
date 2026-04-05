// ==========================================================================
// 第三十五課：OpenTelemetry 分散式追蹤
// ==========================================================================
//
// 什麼是分散式追蹤（Distributed Tracing）？
//   微服務架構中，一個請求可能流經多個服務：
//   用戶端 → API Gateway → 使用者服務 → 訂單服務 → 庫存服務 → 支付服務
//
//   問題：某個請求很慢，是哪個服務的問題？
//   追蹤：給每個請求一個唯一 ID（TraceID），記錄每個步驟的耗時
//
// OpenTelemetry（OTel）是什麼？
//   CNCF 的可觀測性標準，統一了：
//   - Traces（追蹤）：請求在各服務的流向和耗時
//   - Metrics（指標）：數值型監控數據（第 29 課）
//   - Logs（日誌）：結構化日誌（第 21 課）
//
// 核心概念：
//   Trace    → 一個完整請求的追蹤記錄（由多個 Span 組成）
//   Span     → 一個操作的記錄（有開始/結束時間、屬性、事件）
//   TraceID  → 跨服務的唯一追蹤 ID（128 bit）
//   SpanID   → 每個 Span 的唯一 ID（64 bit）
//   Context  → 攜帶 TraceID 在服務間傳遞（HTTP Header、gRPC metadata）
//
// 執行方式：go run ./tutorials/35-opentelemetry
// ==========================================================================

package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// ==========================================================================
// 1. 初始化 OpenTelemetry
// ==========================================================================

// initTracer 初始化 OTel Tracer（回傳 shutdown 函式）
func initTracer() (func(context.Context) error, error) {
	// Exporter：決定 trace 數據輸出到哪裡
	// 真實環境：輸出到 Jaeger、Zipkin、Tempo 等系統
	// 本課示範：輸出到標準輸出（方便看）
	exporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint(), // 格式化 JSON 輸出
	)
	if err != nil {
		return nil, err
	}

	// Resource：描述這個服務本身的屬性
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("tutorial-35-otel"),    // 服務名稱
			semconv.ServiceVersion("1.0.0"),            // 版本
			attribute.String("environment", "development"),
		),
	)
	if err != nil {
		return nil, err
	}

	// TracerProvider：管理 Tracer 的建立和 Span 的處理
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),               // 批次發送（效能更好）
		sdktrace.WithResource(res),                   // 設定 Resource
		sdktrace.WithSampler(sdktrace.AlwaysSample()), // 採樣率：100%（生產環境通常 1-10%）
	)

	// 設定為全域 TracerProvider
	otel.SetTracerProvider(tp)

	// 設定 Propagator（跨服務傳遞 TraceID 的方式）
	// W3C TraceContext：標準格式（traceparent HTTP header）
	// W3C Baggage：傳遞自訂鍵值對
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

// ==========================================================================
// 2. 模擬微服務呼叫鏈
// ==========================================================================

// tracer 取得這個套件的 Tracer（慣例：用套件名稱）
var tracer = otel.Tracer("tutorial-35/service")

// handleOrder 模擬 API Handler（追蹤入口）
func handleOrder(ctx context.Context, orderID string) error {
	// 建立 Span（這個 Span 是整個請求的根）
	ctx, span := tracer.Start(ctx, "handleOrder",
		trace.WithSpanKind(trace.SpanKindServer), // 這是個 Server Span
	)
	defer span.End() // 函式結束時自動關閉 Span

	// 設定 Span 屬性（鍵值對，方便搜尋和過濾）
	span.SetAttributes(
		attribute.String("order.id", orderID),
		attribute.String("http.method", "POST"),
		attribute.String("http.path", "/api/orders"),
	)

	// 記錄事件（Span 中的時間點，比 Span 更細緻）
	span.AddEvent("開始處理訂單")

	// 呼叫下游服務（ctx 攜帶 TraceID）
	if err := validateOrder(ctx, orderID); err != nil {
		span.RecordError(err)                               // 記錄錯誤
		span.SetStatus(codes.Error, "訂單驗證失敗")         // 設定 Span 狀態
		return fmt.Errorf("handleOrder: %w", err)
	}

	if err := chargePayment(ctx, orderID, 299.0); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "支付失敗")
		return fmt.Errorf("handleOrder: %w", err)
	}

	if err := updateInventory(ctx, orderID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "庫存更新失敗")
		return fmt.Errorf("handleOrder: %w", err)
	}

	span.AddEvent("訂單處理完成")
	span.SetStatus(codes.Ok, "") // 成功
	return nil
}

// validateOrder 模擬訂單驗證服務
func validateOrder(ctx context.Context, orderID string) error {
	// 從父 ctx 建立子 Span（自動繼承 TraceID，形成樹狀結構）
	_, span := tracer.Start(ctx, "validateOrder",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	span.SetAttributes(attribute.String("order.id", orderID))

	time.Sleep(10 * time.Millisecond) // 模擬驗證耗時

	if orderID == "INVALID" {
		return errors.New("訂單 ID 格式無效")
	}

	span.AddEvent("驗證通過")
	return nil
}

// chargePayment 模擬支付服務（外部 API 呼叫）
func chargePayment(ctx context.Context, orderID string, amount float64) error {
	_, span := tracer.Start(ctx, "chargePayment",
		trace.WithSpanKind(trace.SpanKindClient), // 呼叫外部服務用 Client
	)
	defer span.End()

	span.SetAttributes(
		attribute.String("order.id", orderID),
		attribute.Float64("payment.amount", amount),
		attribute.String("payment.provider", "stripe"),
	)

	// 模擬呼叫支付 API
	time.Sleep(50 * time.Millisecond)

	// 記錄支付結果
	span.AddEvent("支付請求已發送",
		trace.WithAttributes(attribute.String("payment.status", "pending")),
	)
	time.Sleep(20 * time.Millisecond) // 等待支付結果

	span.AddEvent("支付成功",
		trace.WithAttributes(
			attribute.String("payment.status", "completed"),
			attribute.String("payment.transaction_id", "txn-abc123"),
		),
	)
	return nil
}

// updateInventory 模擬庫存服務（資料庫操作）
func updateInventory(ctx context.Context, orderID string) error {
	ctx, span := tracer.Start(ctx, "updateInventory",
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

	span.SetAttributes(attribute.String("order.id", orderID))

	// 模擬資料庫操作（子 Span）
	if err := dbUpdate(ctx, "inventory", orderID); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "資料庫更新失敗")
		return err
	}

	return nil
}

// dbUpdate 模擬資料庫操作
func dbUpdate(ctx context.Context, table, key string) error {
	_, span := tracer.Start(ctx, "db.update",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	// 資料庫 Span 的標準屬性（遵循 OTel 語義慣例）
	span.SetAttributes(
		semconv.DBSystemSqlite,
		attribute.String("db.operation", "UPDATE"),
		attribute.String("db.sql.table", table),
		attribute.String("db.statement", fmt.Sprintf("UPDATE %s SET quantity = quantity - 1 WHERE id = ?", table)),
	)

	time.Sleep(15 * time.Millisecond) // 模擬 DB 耗時
	return nil
}

// ==========================================================================
// 主程式
// ==========================================================================

func main() {
	fmt.Println("==========================================")
	fmt.Println(" 第三十五課：OpenTelemetry 分散式追蹤")
	fmt.Println("==========================================")
	fmt.Println()

	// 初始化 OTel
	ctx := context.Background()
	shutdown, err := initTracer()
	if err != nil {
		panic(err)
	}
	defer shutdown(ctx)

	// ──── 說明 OTel 概念 ────
	fmt.Println("=== OpenTelemetry 核心概念 ===")
	fmt.Println()
	fmt.Println("Trace（追蹤）：一個請求從開始到結束的完整記錄")
	fmt.Println("  └── Span（操作）：每個步驟")
	fmt.Println("        ├── SpanID：這個 Span 的 ID")
	fmt.Println("        ├── ParentSpanID：父 Span 的 ID（形成樹狀）")
	fmt.Println("        ├── StartTime / EndTime")
	fmt.Println("        ├── Attributes（鍵值對屬性）")
	fmt.Println("        ├── Events（Span 內的時間點事件）")
	fmt.Println("        └── Status（OK / Error）")
	fmt.Println()
	fmt.Println("追蹤樹狀結構（本課示範）：")
	fmt.Println("  handleOrder [Server]")
	fmt.Println("  ├── validateOrder [Internal]")
	fmt.Println("  ├── chargePayment [Client → Stripe API]")
	fmt.Println("  └── updateInventory [Internal]")
	fmt.Println("        └── db.update [Client → SQLite]")
	fmt.Println()
	fmt.Println("=== 執行追蹤（Trace 數據輸出到 stdout）===")
	fmt.Println()

	// ──── 執行正常請求 ────
	fmt.Println("--- 正常訂單 ---")
	if err := handleOrder(ctx, "ORD-001"); err != nil {
		fmt.Printf("錯誤: %v\n", err)
	} else {
		fmt.Println("訂單處理成功！（查看上方的 JSON Span 數據）")
	}

	fmt.Println()
	fmt.Println("=== OTel 在生產環境的使用 ===")
	fmt.Println()
	fmt.Println("1. 用 Jaeger 替換 stdout exporter：")
	fmt.Println("   import \"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp\"")
	fmt.Println("   exporter, _ := otlptracehttp.New(ctx,")
	fmt.Println("       otlptracehttp.WithEndpoint(\"jaeger:4318\"),")
	fmt.Println("   )")
	fmt.Println()
	fmt.Println("2. 啟動 Jaeger（Docker）：")
	fmt.Println("   docker run -d -p 16686:16686 -p 4318:4318 jaegertracing/all-in-one")
	fmt.Println("   然後訪問 http://localhost:16686")
	fmt.Println()
	fmt.Println("3. 在 HTTP Middleware 中自動注入 Span：")
	fmt.Println("   import \"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp\"")
	fmt.Println("   http.Handle(\"/api/\", otelhttp.NewHandler(handler, \"api\"))")
	fmt.Println()
	fmt.Println("4. 跨服務傳遞 TraceID：")
	fmt.Println("   // 發送方：注入 TraceID 到 HTTP Header")
	fmt.Println("   otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))")
	fmt.Println("   // 接收方：從 HTTP Header 提取 TraceID")
	fmt.Println("   ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(req.Header))")

	fmt.Println("\n==========================================")
	fmt.Println(" 教學完成！")
	fmt.Println("==========================================")
}
