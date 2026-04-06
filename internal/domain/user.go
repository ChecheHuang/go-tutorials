// Package domain 定義核心業務實體與 Repository 介面
// 這是 Clean Architecture 的最內層，不依賴任何外部套件
package domain

import "time"

// User 定義使用者實體
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Username  string    `json:"username" gorm:"uniqueIndex;size:50;not null"`
	Email     string    `json:"email" gorm:"uniqueIndex;size:100;not null"`
	Password  string    `json:"-" gorm:"not null"` // json:"-" 確保密碼不會出現在 API 回應中
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RegisterRequest 定義使用者註冊的請求結構
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"newuser"`
	Email    string `json:"email" binding:"required,email" example:"newuser@example.com"`
	Password string `json:"password" binding:"required,min=6,max=100" example:"password123"`
}

// LoginRequest 定義使用者登入的請求結構
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"alice@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse 定義登入成功的回應結構
type LoginResponse struct {
	Token string `json:"token"` // JWT Token
	User  User   `json:"user"`  // 使用者資訊
}

// UserRepository 定義使用者的資料存取介面
// Clean Architecture 中，此介面定義在 domain 層，由 repository 層實作
type UserRepository interface {
	Create(user *User) error                 // 建立使用者
	FindByID(id uint) (*User, error)         // 根據 ID 查詢使用者
	FindByEmail(email string) (*User, error) // 根據 Email 查詢使用者
}
