package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"blog-api/internal/domain"
	"blog-api/pkg/response"

	"github.com/gin-gonic/gin"
)

// === Mock User Usecase ===
type mockUserUsecase struct{}

func (m *mockUserUsecase) Register(req domain.RegisterRequest) (*domain.User, error) {
	return &domain.User{
		ID:       1,
		Username: req.Username,
		Email:    req.Email,
	}, nil
}

func (m *mockUserUsecase) Login(req domain.LoginRequest) (*domain.LoginResponse, error) {
	return &domain.LoginResponse{
		Token: "mock-jwt-token",
		User: domain.User{
			ID:    1,
			Email: req.Email,
		},
	}, nil
}

func (m *mockUserUsecase) GetByID(id uint) (*domain.User, error) {
	return &domain.User{
		ID:       id,
		Username: "testuser",
		Email:    "test@example.com",
	}, nil
}

// setupTestRouter 建立測試用的 Gin 路由
func setupTestRouter() (*gin.Engine, *UserHandler) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewUserHandler(&mockUserUsecase{})
	return r, h
}

// TestRegisterHandler_Success 測試註冊 API 端點
func TestRegisterHandler_Success(t *testing.T) {
	r, h := setupTestRouter()
	r.POST("/api/v1/auth/register", h.Register)

	// 準備請求 body
	body := domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	// 建立 HTTP 請求
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// 執行請求並記錄回應
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 驗證狀態碼
	if w.Code != http.StatusCreated {
		t.Errorf("預期狀態碼 %d，得到 %d", http.StatusCreated, w.Code)
	}

	// 解析回應
	var resp response.Response
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Code != http.StatusCreated {
		t.Errorf("回應碼不符：預期 %d，得到 %d", http.StatusCreated, resp.Code)
	}
}

// TestRegisterHandler_InvalidInput 測試無效輸入的註冊請求
func TestRegisterHandler_InvalidInput(t *testing.T) {
	r, h := setupTestRouter()
	r.POST("/api/v1/auth/register", h.Register)

	// 缺少必要欄位的請求
	body := map[string]string{
		"username": "ab", // 太短（最少 3 字元）
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 應該回傳 400 Bad Request
	if w.Code != http.StatusBadRequest {
		t.Errorf("預期狀態碼 %d，得到 %d", http.StatusBadRequest, w.Code)
	}
}

// TestLoginHandler_Success 測試登入 API 端點
func TestLoginHandler_Success(t *testing.T) {
	r, h := setupTestRouter()
	r.POST("/api/v1/auth/login", h.Login)

	body := domain.LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("預期狀態碼 %d，得到 %d", http.StatusOK, w.Code)
	}

	// 確認回應中包含 Token
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("回應中缺少 data 欄位")
	}

	if _, exists := data["token"]; !exists {
		t.Error("回應中缺少 token")
	}
}
