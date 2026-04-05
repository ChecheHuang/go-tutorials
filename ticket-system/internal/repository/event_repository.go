package repository

import (
	"ticket-system/internal/domain"

	"gorm.io/gorm"
)

type eventRepository struct{ db *gorm.DB }

func NewEventRepository(db *gorm.DB) domain.EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(event *domain.Event) error {
	return r.db.Create(event).Error
}

func (r *eventRepository) FindByID(id uint) (*domain.Event, error) {
	var event domain.Event
	if err := r.db.First(&event, id).Error; err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *eventRepository) FindAll() ([]domain.Event, error) {
	var events []domain.Event
	err := r.db.Order("start_time ASC").Find(&events).Error
	return events, err
}
