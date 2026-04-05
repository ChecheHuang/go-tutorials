// Package usecase 實作業務邏輯層
// 這一層負責協調 domain 層的實體與 repository，處理具體的業務規則
package usecase

import (
	"errors"
	"time"

	"blog-api/internal/domain"
	"blog-api/pkg/config"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// UserUsecase 定義使用者業務邏輯的介面
type UserUsecase interface {
	Register(req domain.RegisterRequest) (*domain.User, error)
	Login(req domain.LoginRequest) (*domain.LoginResponse, error)
	GetByID(id uint) (*domain.User, error)
}

// userUsecase 實作 UserUsecase 介面
type userUsecase struct {
	userRepo domain.UserRepository
	cfg      *config.Config
}

// NewUserUsecase 建立使用者 Usecase 實例
// 透過依賴注入接收 Repository 與設定
func NewUserUsecase(userRepo domain.UserRepository, cfg *config.Config) UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// Register 處理使用者註冊邏輯
// 1. 將密碼進行 bcrypt 雜湊處理
// 2. 建立使用者記錄
func (u *userUsecase) Register(req domain.RegisterRequest) (*domain.User, error) {
	// 使用 bcrypt 雜湊密碼，cost 為 10（安全與效能的平衡）
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密碼加密失敗")
	}

	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := u.userRepo.Create(user); err != nil {
		return nil, errors.New("使用者名稱或信箱已被使用")
	}

	return user, nil
}

// Login 處理使用者登入邏輯
// 1. 根據 Email 查找使用者
// 2. 驗證密碼
// 3. 產生 JWT Token
func (u *userUsecase) Login(req domain.LoginRequest) (*domain.LoginResponse, error) {
	// 根據 Email 查找使用者
	user, err := u.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("信箱或密碼錯誤")
	}

	// 比對密碼與雜湊值
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("信箱或密碼錯誤")
	}

	// 產生 JWT Token
	token, err := u.generateToken(user.ID)
	if err != nil {
		return nil, errors.New("產生 Token 失敗")
	}

	return &domain.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

// GetByID 根據 ID 取得使用者資料
func (u *userUsecase) GetByID(id uint) (*domain.User, error) {
	return u.userRepo.FindByID(id)
}

// generateToken 產生 JWT Token
// Token 中包含使用者 ID 與過期時間
func (u *userUsecase) generateToken(userID uint) (string, error) {
	// 建立 JWT Claims（聲明）
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(u.cfg.JWT.Expiration).Unix(), // 設定過期時間
		"iat":     time.Now().Unix(),                           // 簽發時間
	}

	// 使用 HS256 演算法建立 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密鑰簽名
	return token.SignedString([]byte(u.cfg.JWT.Secret))
}
