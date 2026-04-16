package domain

import (
	"context"
	"errors"
)

// ErrInsufficientStock 庫存不足（原子檢查失敗時回傳，不會實際扣減）
var ErrInsufficientStock = errors.New("庫存不足")

// TicketStock 票券庫存
type TicketStock struct {
	EventID   uint `json:"event_id"`
	Total     int  `json:"total"`
	Remaining int  `json:"remaining"`
}

// StockStore 庫存儲存介面（Redis 或 Memory 實作）
type StockStore interface {
	// DecrIfSufficient 原子性扣減庫存：僅在庫存 >= value 時扣減，否則回傳 ErrInsufficientStock
	// 成功時回傳扣減後的剩餘數量；失敗時不修改庫存
	DecrIfSufficient(ctx context.Context, key string, value int64) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	Get(ctx context.Context, key string) (int, error)
	Set(ctx context.Context, key string, value int) error
}
