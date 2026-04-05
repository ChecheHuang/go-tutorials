package domain

import "time"

// OrderStatus 訂單狀態
type OrderStatus string

const (
	OrderPending   OrderStatus = "pending"   // 待支付
	OrderPaid      OrderStatus = "paid"      // 已支付
	OrderFailed    OrderStatus = "failed"    // 支付失敗
	OrderCancelled OrderStatus = "cancelled" // 已取消
)

// Order 定義訂單實體（CQRS Write Model）
type Order struct {
	ID        uint        `json:"id" gorm:"primaryKey"`
	EventID   uint        `json:"event_id" gorm:"index;not null"`
	UserID    string      `json:"user_id" gorm:"index;not null"` // 簡化為字串，不做真實認證
	Quantity  int         `json:"quantity" gorm:"not null;default:1"`
	Amount    float64     `json:"amount" gorm:"not null"`
	Status    OrderStatus `json:"status" gorm:"size:20;not null;default:pending"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// OrderView 訂單查詢視圖（CQRS Read Model）
type OrderView struct {
	ID        uint        `json:"id"`
	EventID   uint        `json:"event_id"`
	EventName string      `json:"event_name"`
	UserID    string      `json:"user_id"`
	Quantity  int         `json:"quantity"`
	Amount    float64     `json:"amount"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

// OrderCommand 搶票指令
type OrderCommand struct {
	EventID  uint   `json:"event_id" binding:"required"`
	UserID   string `json:"user_id" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1,max=4"`
}

// OrderWriteRepository 寫入用 Repository（CQRS Write）
type OrderWriteRepository interface {
	Create(order *Order) error
	UpdateStatus(id uint, status OrderStatus) error
}

// OrderReadRepository 查詢用 Repository（CQRS Read）
type OrderReadRepository interface {
	FindByID(id uint) (*OrderView, error)
	FindByUserID(userID string) ([]OrderView, error)
	FindByEventID(eventID uint) ([]OrderView, error)
}
