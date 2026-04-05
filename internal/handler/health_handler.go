package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler 處理健康檢查請求
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler 建立健康檢查 Handler
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Healthz 存活探針（Liveness Probe）
// 只要程式還在跑就回傳 200，Kubernetes 用來判斷是否需要重啟
func (h *HealthHandler) Healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Readyz 就緒探針（Readiness Probe）
// 檢查資料庫連線是否正常，Kubernetes 用來判斷是否可以接收流量
func (h *HealthHandler) Readyz(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"reason": "無法取得資料庫連線",
		})
		return
	}

	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "error",
			"reason": "資料庫連線失敗",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
