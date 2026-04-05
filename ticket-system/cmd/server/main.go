// 搶票系統 — 高併發教學專案
//
// 展示技術：Goroutine、WebSocket、gRPC、Wire DI、Generics、
//           Message Queue、Circuit Breaker、OpenTelemetry、CQRS
//
// 啟動方式：
//   1. 先啟動 Redis：docker run -p 6379:6379 redis:7-alpine
//   2. cd ticket-system && go run ./cmd/server/
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ticket-system/internal/di"
	"ticket-system/internal/domain"
	"ticket-system/internal/usecase"
	"ticket-system/pkg/config"
	"ticket-system/pkg/logger"
	"ticket-system/pkg/telemetry"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	// === 1. 設定 ===
	cfg := config.Load()
	logger.Init(cfg.Log.Level, cfg.Log.Format)
	slog.Info("搶票系統啟動中", "port", cfg.Server.Port)

	// === 2. OpenTelemetry 追蹤（第 35 課）===
	tracer, tracerShutdown := telemetry.InitTracer("ticket-system")
	defer tracerShutdown()

	// === 3. Wire DI：一行建立所有依賴（第 28 課）===
	app, cleanup, err := di.InitializeApp(cfg, tracer)
	if err != nil {
		slog.Error("應用程式初始化失敗", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	// === 4. 啟動背景 Worker（第 19 課 Goroutine）===
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app.PaymentWorker.Start(ctx)
	app.StockBroadcaster.Start(ctx)

	// === 5. 填入示範活動 ===
	seedEvents(app.DB, app.TicketUsecase)

	// === 6. 路由 ===
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/ws", app.WSHandler.HandleWebSocket)
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	v1 := r.Group("/api/v1")
	{
		v1.GET("/events", app.TicketHandler.GetEvents)
		v1.GET("/events/:id/stock", app.TicketHandler.GetStock)
		v1.POST("/tickets/grab", app.TicketHandler.GrabTicket)
		v1.GET("/orders/:id", app.TicketHandler.GetOrder)
		v1.GET("/users/:user_id/orders", app.TicketHandler.GetUserOrders)
	}

	// === 7. 啟動伺服器 + Graceful Shutdown ===
	printBanner(cfg.Server.Port)

	srv := &http.Server{Addr: ":" + cfg.Server.Port, Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("伺服器啟動失敗", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("正在優雅關閉...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)

	slog.Info("搶票系統已關閉")
}

// seedEvents 填入示範活動並初始化 Redis 庫存
func seedEvents(db *gorm.DB, uc *usecase.TicketUsecase) {
	var count int64
	db.Model(&domain.Event{}).Count(&count)
	if count > 0 {
		var events []domain.Event
		db.Find(&events)
		for _, e := range events {
			var sold int64
			db.Model(&domain.Order{}).Where("event_id = ? AND status != ?", e.ID, domain.OrderCancelled).Count(&sold)
			remaining := e.TotalTickets - int(sold)
			uc.InitStock(context.Background(), e.ID, remaining)
		}
		return
	}

	events := []domain.Event{
		{Name: "五月天 2026 巡迴演唱會", Venue: "台北小巨蛋", TotalTickets: 100, Price: 2800, StartTime: time.Now().Add(30 * 24 * time.Hour)},
		{Name: "NBA 季後賽 Game 7", Venue: "Crypto.com Arena", TotalTickets: 50, Price: 5000, StartTime: time.Now().Add(60 * 24 * time.Hour)},
		{Name: "程式設計大會 GopherCon 2026", Venue: "台北國際會議中心", TotalTickets: 200, Price: 1500, StartTime: time.Now().Add(90 * 24 * time.Hour)},
	}

	for i := range events {
		db.Create(&events[i])
		uc.InitStock(context.Background(), events[i].ID, events[i].TotalTickets)
	}
	slog.Info("示範活動已建立", "count", len(events))
}

func printBanner(port string) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              🎫 搶票系統 Ticket System                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("  API:       http://localhost:%s/api/v1\n", port)
	fmt.Printf("  WebSocket: ws://localhost:%s/ws\n", port)
	fmt.Printf("  Health:    http://localhost:%s/healthz\n", port)
	fmt.Println()
	fmt.Println("  路由表：")
	fmt.Println("    GET  /api/v1/events            活動列表")
	fmt.Println("    GET  /api/v1/events/:id/stock   即時庫存")
	fmt.Println("    POST /api/v1/tickets/grab       搶票")
	fmt.Println("    GET  /api/v1/orders/:id         查詢訂單")
	fmt.Println("    GET  /api/v1/users/:id/orders   使用者訂單")
	fmt.Println("    GET  /ws                        WebSocket 即時推送")
	fmt.Println()
}
