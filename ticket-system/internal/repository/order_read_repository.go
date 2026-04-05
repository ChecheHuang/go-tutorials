package repository

import (
	"ticket-system/internal/domain"

	"gorm.io/gorm"
)

// orderReadRepository CQRS 讀取端（JOIN 查詢，讀最佳化）
type orderReadRepository struct{ db *gorm.DB }

func NewOrderReadRepository(db *gorm.DB) domain.OrderReadRepository {
	return &orderReadRepository{db: db}
}

func (r *orderReadRepository) FindByID(id uint) (*domain.OrderView, error) {
	var view domain.OrderView
	err := r.db.Table("orders").
		Select("orders.id, orders.event_id, events.name as event_name, orders.user_id, orders.quantity, orders.amount, orders.status, orders.created_at").
		Joins("LEFT JOIN events ON events.id = orders.event_id").
		Where("orders.id = ?", id).
		Scan(&view).Error
	if err != nil {
		return nil, err
	}
	if view.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &view, nil
}

func (r *orderReadRepository) FindByUserID(userID string) ([]domain.OrderView, error) {
	var views []domain.OrderView
	err := r.db.Table("orders").
		Select("orders.id, orders.event_id, events.name as event_name, orders.user_id, orders.quantity, orders.amount, orders.status, orders.created_at").
		Joins("LEFT JOIN events ON events.id = orders.event_id").
		Where("orders.user_id = ?", userID).
		Order("orders.created_at DESC").
		Scan(&views).Error
	return views, err
}

func (r *orderReadRepository) FindByEventID(eventID uint) ([]domain.OrderView, error) {
	var views []domain.OrderView
	err := r.db.Table("orders").
		Select("orders.id, orders.event_id, events.name as event_name, orders.user_id, orders.quantity, orders.amount, orders.status, orders.created_at").
		Joins("LEFT JOIN events ON events.id = orders.event_id").
		Where("orders.event_id = ?", eventID).
		Order("orders.created_at DESC").
		Scan(&views).Error
	return views, err
}
