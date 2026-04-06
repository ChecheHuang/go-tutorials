// Package response 提供統一的 API 回應格式
// 所有 API 回應都使用相同的 JSON 結構，方便前端處理
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 定義統一的 API 回應結構
type Response struct {
	Code    int         `json:"code"`           // HTTP 狀態碼
	Message string      `json:"message"`        // 回應訊息
	Data    interface{} `json:"data,omitempty"` // 回應資料（可選）
}

// PaginatedData 定義分頁回應的資料結構
type PaginatedData struct {
	Items      interface{} `json:"items"`       // 資料列表
	Total      int64       `json:"total"`       // 總筆數
	Page       int         `json:"page"`        // 目前頁碼
	PageSize   int         `json:"page_size"`   // 每頁筆數
	TotalPages int         `json:"total_pages"` // 總頁數
}

// Success 回傳成功回應（200 OK）
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

// Created 回傳建立成功回應（201 Created）
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Code:    http.StatusCreated,
		Message: "created",
		Data:    data,
	})
}

// Error 回傳錯誤回應，使用指定的 HTTP 狀態碼
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Code:    statusCode,
		Message: message,
	})
}

// BadRequest 回傳 400 錯誤回應
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// Unauthorized 回傳 401 錯誤回應
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden 回傳 403 錯誤回應
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound 回傳 404 錯誤回應
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalServerError 回傳 500 錯誤回應
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// Paginated 回傳分頁資料回應
func Paginated(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	// 計算總頁數
	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	Success(c, PaginatedData{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	})
}
