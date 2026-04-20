// Blog API - 教學用 Go REST API
//
// 使用 Gin 框架搭配 Clean Architecture 架構的部落格系統 API
// 包含使用者認證、文章管理、留言功能
//
// @title       Blog API
// @version     1.0
// @description 教學用的部落格系統 REST API，使用 Go + Gin + GORM + SQLite
//
// @host     localhost:8080
// @BasePath /api/v1
//
// @securityDefinitions.apikey BearerAuth
// @in                         header
// @name                       Authorization
// @description                在此輸入 Bearer Token，格式：Bearer <token>
//
// @tag.name        auth
// @tag.description 使用者認證（註冊、登入、個人資料）
// @tag.name        articles
// @tag.description 文章管理（CRUD、分頁、搜尋）
// @tag.name        comments
// @tag.description 留言管理（CRUD）
// 教學對應：第 10 課（依賴注入）、第 29 課（Graceful Shutdown）
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

	"blog-api/internal/domain"
	"blog-api/internal/handler"
	"blog-api/internal/repository"
	"blog-api/internal/seed"
	"blog-api/internal/usecase"
	"blog-api/pkg/cache"
	"blog-api/pkg/config"
	"blog-api/pkg/logger"

	_ "blog-api/docs"

	"github.com/fatih/color"
	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func main() {
	// === 1. 載入設定 ===
	cfg := config.Load()

	// === 2. 初始化結構化日誌 ===
	logger.Init(cfg.Log.Level, cfg.Log.Format)
	slog.Info("設定載入完成", "port", cfg.Server.Port, "log_level", cfg.Log.Level)

	// === 3. 初始化資料庫 ===
	db, err := gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		slog.Error("資料庫連線失敗", "error", err)
		os.Exit(1)
	}
	slog.Info("資料庫連線成功", "dsn", cfg.Database.DSN)

	if err := db.AutoMigrate(&domain.User{}, &domain.Article{}, &domain.Comment{}); err != nil {
		slog.Error("資料庫遷移失敗", "error", err)
		os.Exit(1)
	}
	slog.Info("資料庫遷移完成")

	seed.Run(db)

	// === 4. 初始化 Redis（可選）===
	var appCache cache.Cache
	var redisClient *redis.Client

	if cfg.Redis.Enabled {
		rc, err := cache.NewRedisCache(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
		if err != nil {
			slog.Warn("Redis 連線失敗，降級為無快取模式", "error", err)
			appCache = cache.NewNoOpCache()
		} else {
			redisClient = rc.Client()
			appCache = rc
		}
	} else {
		slog.Info("Redis 未啟用，使用無快取模式")
		appCache = cache.NewNoOpCache()
	}

	// === 5. 依賴注入：按照 Clean Architecture 由內而外初始化 ===
	userRepo := repository.NewUserRepository(db)
	articleRepo := repository.NewArticleRepository(db)
	cachedArticleRepo := repository.NewCachedArticleRepository(articleRepo, appCache)
	commentRepo := repository.NewCommentRepository(db)

	userUsecase := usecase.NewUserUsecase(userRepo, cfg)
	articleUsecase := usecase.NewArticleUsecase(cachedArticleRepo)
	commentUsecase := usecase.NewCommentUsecase(commentRepo, cachedArticleRepo)

	userHandler := handler.NewUserHandler(userUsecase)
	articleHandler := handler.NewArticleHandler(articleUsecase)
	commentHandler := handler.NewCommentHandler(commentUsecase)
	healthHandler := handler.NewHealthHandler(db, redisClient)

	// === 6. 設定路由 ===
	router := handler.SetupRouter(cfg, userHandler, articleHandler, commentHandler, healthHandler)

	printRoutes(cfg.Server.Port)

	// === 7. Graceful Shutdown ===
	srv := &http.Server{
		Addr:           ":" + cfg.Server.Port,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB max header size
	}

	// 在 goroutine 中啟動伺服器
	go func() {
		slog.Info("伺服器啟動", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("伺服器啟動失敗", "error", err)
			os.Exit(1)
		}
	}()

	// 等待中斷訊號（Ctrl+C 或 kill）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	slog.Info("收到關閉訊號，開始優雅關閉...", "signal", sig.String())

	// 給伺服器 30 秒完成正在處理的請求
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("伺服器關閉失敗", "error", err)
		os.Exit(1)
	}

	slog.Info("伺服器已優雅關閉")
}

// printRoutes 印出彩色的路由表與啟動資訊
func printRoutes(port string) {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	magenta := color.New(color.FgMagenta, color.Bold)
	white := color.New(color.FgWhite, color.Bold)
	dim := color.New(color.FgHiBlack)

	fmt.Println()
	cyan.Println("╔══════════════════════════════════════════════════════════╗")
	cyan.Println("║                    Blog API Server                      ║")
	cyan.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	white.Println("  🌐 伺服器")
	fmt.Printf("     API:     ")
	green.Printf("http://localhost:%s/api/v1\n", port)
	fmt.Printf("     Swagger: ")
	green.Printf("http://localhost:%s/swagger/index.html\n", port)
	fmt.Printf("     Health:  ")
	green.Printf("http://localhost:%s/healthz\n", port)
	fmt.Printf("     Metrics: ")
	green.Printf("http://localhost:%s/metrics\n", port)
	fmt.Printf("     Pprof:   ")
	green.Printf("http://localhost:%s/debug/pprof/\n", port)
	fmt.Println()

	white.Println("  📋 路由表")
	dim.Println("  ─────────────────────────────────────────────────────────")

	dim.Println("  系統 (System)")
	printRoute(green, " GET", "/healthz", "存活探針", "")
	printRoute(green, " GET", "/readyz", "就緒探針", "")
	printRoute(green, " GET", "/metrics", "Prometheus 指標", "")
	printRoute(green, " GET", "/debug/pprof/", "效能分析（debug 模式）", "")

	fmt.Println()
	dim.Println("  認證 (Auth)")
	printRoute(green, "POST", "/api/v1/auth/register", "使用者註冊", "")
	printRoute(green, "POST", "/api/v1/auth/login", "使用者登入", "")
	printRoute(yellow, " GET", "/api/v1/auth/profile", "取得個人資料", "🔒")

	fmt.Println()
	dim.Println("  文章 (Articles)")
	printRoute(cyan, " GET", "/api/v1/articles", "文章列表（分頁+搜尋）", "")
	printRoute(cyan, " GET", "/api/v1/articles/:id", "文章詳情", "")
	printRoute(green, "POST", "/api/v1/articles", "建立文章", "🔒")
	printRoute(yellow, " PUT", "/api/v1/articles/:id", "更新文章", "🔒")
	printRoute(magenta, " DEL", "/api/v1/articles/:id", "刪除文章", "🔒")

	fmt.Println()
	dim.Println("  留言 (Comments)")
	printRoute(cyan, " GET", "/api/v1/articles/:id/comments", "取得文章留言", "")
	printRoute(green, "POST", "/api/v1/articles/:id/comments", "建立留言", "🔒")
	printRoute(yellow, " PUT", "/api/v1/comments/:id", "更新留言", "🔒")
	printRoute(magenta, " DEL", "/api/v1/comments/:id", "刪除留言", "🔒")

	dim.Println("  ─────────────────────────────────────────────────────────")
	dim.Println("  🔒 = 需要 JWT 認證（Header: Authorization: Bearer <token>）")
	fmt.Println()

	white.Println("  👤 測試帳號")
	fmt.Println("     alice@example.com / password123")
	fmt.Println("     bob@example.com   / password123")
	fmt.Println("     carol@example.com / password123")
	fmt.Println()
}

// printRoute 印出單一路由
func printRoute(methodColor *color.Color, method, path, desc, auth string) {
	methodColor.Printf("    %-4s", method)
	fmt.Printf(" %-38s %s %s\n", path, desc, auth)
}
