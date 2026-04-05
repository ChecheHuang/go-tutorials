package domain

import "context"

// TicketStock 票券庫存
type TicketStock struct {
	EventID   uint `json:"event_id"`
	Total     int  `json:"total"`
	Remaining int  `json:"remaining"`
}

// StockStore 庫存儲存介面（Redis 或 Memory 實作）
type StockStore interface {
	DecrBy(ctx context.Context, key string, value int64) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	Get(ctx context.Context, key string) (int, error)
	Set(ctx context.Context, key string, value int) error
}
