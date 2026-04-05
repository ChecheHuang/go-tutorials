// Package usecase 實作業務邏輯層
// 這一層負責協調 domain 層的實體與 repository，處理具體的業務規則
package usecase

import (
	"time"

	"blog-api/internal/domain"
	"blog-api/pkg/apperror"
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
func NewUserUsecase(userRepo domain.UserRepository, cfg *config.Config) UserUsecase {
	return &userUsecase{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// Register 處理使用者註冊邏輯
func (u *userUsecase) Register(req domain.RegisterRequest) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrInternal, "密碼加密失敗")
	}

	user := &domain.User{
		Username: req.Username,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := u.userRepo.Create(user); err != nil {
		return nil, apperror.Wrap(apperror.ErrConflict, "使用者名稱或信箱已被使用")
	}

	return user, nil
}

// Login 處理使用者登入邏輯
func (u *userUsecase) Login(req domain.LoginRequest) (*domain.LoginResponse, error) {
	user, err := u.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrUnauthorized, "信箱或密碼錯誤")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, apperror.Wrap(apperror.ErrUnauthorized, "信箱或密碼錯誤")
	}

	token, err := u.generateToken(user.ID)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrInternal, "產生 Token 失敗")
	}

	return &domain.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

// GetByID 根據 ID 取得使用者資料
func (u *userUsecase) GetByID(id uint) (*domain.User, error) {
	user, err := u.userRepo.FindByID(id)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrNotFound, "使用者 ID=%d", id)
	}
	return user, nil
}

// generateToken 產生 JWT Token
func (u *userUsecase) generateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(u.cfg.JWT.Expiration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.cfg.JWT.Secret))
}
