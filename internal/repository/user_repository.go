// Package repository 實作 domain 層定義的 Repository 介面
// 使用 GORM 作為 ORM 框架，SQLite 作為資料庫
// 教學對應：第 10 課（Clean Architecture Repository 層）、第 14 課（GORM CRUD）
package repository

import (
	"blog-api/internal/domain"

	"gorm.io/gorm"
)

// userRepository 實作 domain.UserRepository 介面
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 建立使用者 Repository 實例
// 接收 GORM 資料庫連線作為依賴注入
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// Create 建立新使用者
func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

// FindByID 根據 ID 查詢使用者
func (r *userRepository) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根據 Email 查詢使用者
// 常用於登入驗證時查找使用者
func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
