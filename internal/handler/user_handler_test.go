package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-api/internal/domain"
	"blog-api/pkg/apperror"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === Mock User Usecase ===
type mockUserUsecase struct{}

func (m *mockUserUsecase) Register(req domain.RegisterRequest) (*domain.User, error) {
	if req.Email == "duplicate@example.com" {
		return nil, apperror.Wrap(apperror.ErrConflict, "信箱已被使用")
	}
	return &domain.User{
		ID:       1,
		Username: req.Username,
		Email:    req.Email,
	}, nil
}

func (m *mockUserUsecase) Login(req domain.LoginRequest) (*domain.LoginResponse, error) {
	if req.Password == "wrong" {
		return nil, apperror.Wrap(apperror.ErrUnauthorized, "信箱或密碼錯誤")
	}
	return &domain.LoginResponse{
		Token: "mock-jwt-token",
		User:  domain.User{ID: 1, Email: req.Email},
	}, nil
}

func (m *mockUserUsecase) GetByID(id uint) (*domain.User, error) {
	if id == 999 {
		return nil, apperror.Wrap(apperror.ErrNotFound, "使用者不存在")
	}
	return &domain.User{
		ID:       id,
		Username: "testuser",
		Email:    "test@example.com",
	}, nil
}

func setupTestRouter() (*gin.Engine, *UserHandler) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewUserHandler(&mockUserUsecase{})
	return r, h
}

func TestRegisterHandler(t *testing.T) {
	t.Run("註冊成功回傳 201", func(t *testing.T) {
		r, h := setupTestRouter()
		r.POST("/api/v1/auth/register", h.Register)

		body, _ := json.Marshal(domain.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		})

		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("缺少欄位回傳 400", func(t *testing.T) {
		r, h := setupTestRouter()
		r.POST("/api/v1/auth/register", h.Register)

		body, _ := json.Marshal(map[string]string{"username": "ab"})

		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("重複 Email 回傳 409", func(t *testing.T) {
		r, h := setupTestRouter()
		r.POST("/api/v1/auth/register", h.Register)

		body, _ := json.Marshal(domain.RegisterRequest{
			Username: "testuser",
			Email:    "duplicate@example.com",
			Password: "password123",
		})

		req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestLoginHandler(t *testing.T) {
	t.Run("登入成功回傳 Token", func(t *testing.T) {
		r, h := setupTestRouter()
		r.POST("/api/v1/auth/login", h.Login)

		body, _ := json.Marshal(domain.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		})

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]any
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

		data, ok := resp["data"].(map[string]any)
		require.True(t, ok)
		assert.NotEmpty(t, data["token"])
	})

	t.Run("密碼錯誤回傳 401", func(t *testing.T) {
		r, h := setupTestRouter()
		r.POST("/api/v1/auth/login", h.Login)

		body, _ := json.Marshal(domain.LoginRequest{
			Email:    "test@example.com",
			Password: "wrong",
		})

		req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
