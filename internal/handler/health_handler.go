package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// HealthHandler 處理健康檢查請求
type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client // 可為 nil（Redis 未啟用時）
}

// NewHealthHandler 建立健康檢查 Handler
func NewHealthHandler(db *gorm.DB, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redisClient}
}

// Healthz 存活探針（Liveness Probe）
func (h *HealthHandler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readyz 就緒探針（Readiness Probe）
func (h *HealthHandler) Readyz(c *gin.Context) {
	checks := gin.H{}

	// 檢查資料庫
	sqlDB, err := h.db.DB()
	if err != nil {
		checks["database"] = "error: " + err.Error()
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "checks": checks})
		return
	}
	if err := sqlDB.Ping(); err != nil {
		checks["database"] = "error: " + err.Error()
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "checks": checks})
		return
	}
	checks["database"] = "ok"

	// 檢查 Redis（如果有啟用）
	if h.redis != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := h.redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = "error: " + err.Error()
			// Redis 掛了不影響整體就緒狀態（graceful degradation）
		} else {
			checks["redis"] = "ok"
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "checks": checks})
}
