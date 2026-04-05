// Package domain 定義搶票系統的核心實體
package domain

import "time"

// Event 定義活動（演唱會、球賽等）
type Event struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:200;not null"`
	Venue       string    `json:"venue" gorm:"size:200"`
	TotalTickets int      `json:"total_tickets" gorm:"not null"`
	Price       float64   `json:"price" gorm:"not null"`
	StartTime   time.Time `json:"start_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EventRepository 定義活動的資料存取介面
type EventRepository interface {
	Create(event *Event) error
	FindByID(id uint) (*Event, error)
	FindAll() ([]Event, error)
}
