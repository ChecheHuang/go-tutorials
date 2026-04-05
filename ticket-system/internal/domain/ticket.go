package domain

// TicketStock 票券庫存（Redis 中管理）
type TicketStock struct {
	EventID   uint `json:"event_id"`
	Total     int  `json:"total"`
	Remaining int  `json:"remaining"`
}
