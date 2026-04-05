package repository

import (
	"ticket-system/internal/domain"

	"gorm.io/gorm"
)

// orderWriteRepository CQRS 寫入端
type orderWriteRepository struct{ db *gorm.DB }

func NewOrderWriteRepository(db *gorm.DB) domain.OrderWriteRepository {
	return &orderWriteRepository{db: db}
}

func (r *orderWriteRepository) Create(order *domain.Order) error {
	return r.db.Create(order).Error
}

func (r *orderWriteRepository) UpdateStatus(id uint, status domain.OrderStatus) error {
	return r.db.Model(&domain.Order{}).Where("id = ?", id).Update("status", status).Error
}
