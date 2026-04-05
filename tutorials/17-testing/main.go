// 第十六課：單元測試
// 這個檔案是被測試的程式碼
// 測試程式碼在 main_test.go
package main

import (
	"errors"
	"fmt"
)

// ========================================
// 被測試的函式和結構體
// ========================================

// Add 加法
func Add(a, b int) int {
	return a + b
}

// Divide 除法（帶錯誤處理）
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("除數不能為零")
	}
	return a / b, nil
}

// User 使用者
type User struct {
	Name  string
	Email string
	Age   int
}

// Validate 驗證使用者資料
func (u *User) Validate() error {
	if u.Name == "" {
		return errors.New("名稱不能為空")
	}
	if u.Age < 0 || u.Age > 150 {
		return errors.New("年齡不合理")
	}
	return nil
}

// IsAdult 判斷是否成年
func (u *User) IsAdult() bool {
	return u.Age >= 18
}

// ========================================
// 模擬 Repository 介面（用於展示 Mock 測試）
// ========================================

// UserRepository 介面
type UserRepository interface {
	FindByID(id int) (*User, error)
	Create(user *User) error
}

// UserService 依賴 UserRepository 介面
type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(id int) (*User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("取得使用者失敗: %w", err)
	}
	return user, nil
}

func (s *UserService) CreateUser(name, email string, age int) (*User, error) {
	user := &User{Name: name, Email: email, Age: age}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("建立使用者失敗: %w", err)
	}
	return user, nil
}

func main() {
	fmt.Println("這個檔案主要用於被測試")
	fmt.Println("請執行: go test -v")
}
