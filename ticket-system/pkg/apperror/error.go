package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrNotFound     = errors.New("資源不存在")
	ErrUnauthorized = errors.New("未授權")
	ErrForbidden    = errors.New("無權限")
	ErrConflict     = errors.New("資源衝突")
	ErrInternal     = errors.New("內部錯誤")
	ErrBadRequest   = errors.New("請求參數錯誤")
	ErrSoldOut      = errors.New("已售罄")
	ErrPayment      = errors.New("支付失敗")
)

func Wrap(sentinel error, format string, args ...any) error {
	msg := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", msg, sentinel)
}

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
	case errors.Is(err, ErrSoldOut):
		return http.StatusConflict
	case errors.Is(err, ErrPayment):
		return http.StatusPaymentRequired
	default:
		return http.StatusInternalServerError
	}
}
