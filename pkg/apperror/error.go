// Package apperror 定義應用程式層級的錯誤類型
// 支援 Error Wrapping（errors.Is / errors.As）與 HTTP 狀態碼對應
package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// 預定義的 Sentinel Error（哨兵錯誤）
// 使用 errors.Is(err, apperror.ErrNotFound) 來判斷錯誤類型
var (
	ErrNotFound     = errors.New("資源不存在")
	ErrUnauthorized = errors.New("未授權")
	ErrForbidden    = errors.New("無權限")
	ErrConflict     = errors.New("資源衝突")
	ErrInternal     = errors.New("內部錯誤")
	ErrBadRequest   = errors.New("請求參數錯誤")
)

// Wrap 用 fmt.Errorf %w 包裝錯誤，保留錯誤鏈
// 範例：apperror.Wrap(apperror.ErrNotFound, "找不到文章 ID=%d", id)
func Wrap(sentinel error, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", msg, sentinel)
}

// HTTPStatus 根據錯誤類型回傳對應的 HTTP 狀態碼
func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
