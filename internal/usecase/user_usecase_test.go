package usecase

import (
	"errors"
	"testing"

	"blog-api/internal/domain"
	"blog-api/pkg/config"

	"golang.org/x/crypto/bcrypt"
)

// === Mock Repository ===
// mockUserRepository 是 UserRepository 的模擬實作，用於單元測試
type mockUserRepository struct {
	users map[string]*domain.User // 以 email 為 key 的使用者儲存
	nextID uint
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:  make(map[string]*domain.User),
		nextID: 1,
	}
}

func (m *mockUserRepository) Create(user *domain.User) error {
	// 模擬唯一性檢查
	if _, exists := m.users[user.Email]; exists {
		return errors.New("duplicate email")
	}
	user.ID = m.nextID
	m.nextID++
	m.users[user.Email] = user
	return nil
}

func (m *mockUserRepository) FindByID(id uint) (*domain.User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) FindByEmail(email string) (*domain.User, error) {
	user, exists := m.users[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// === 測試用的設定 ===
func testConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 3600,
		},
	}
}

// TestRegister_Success 測試正常的使用者註冊流程
func TestRegister_Success(t *testing.T) {
	repo := newMockUserRepository()
	uc := NewUserUsecase(repo, testConfig())

	req := domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	user, err := uc.Register(req)
	if err != nil {
		t.Fatalf("預期註冊成功，但得到錯誤：%v", err)
	}

	if user.Username != req.Username {
		t.Errorf("使用者名稱不符：預期 %s，得到 %s", req.Username, user.Username)
	}

	if user.Email != req.Email {
		t.Errorf("Email 不符：預期 %s，得到 %s", req.Email, user.Email)
	}

	// 確認密碼已被雜湊
	if user.Password == req.Password {
		t.Error("密碼應該被雜湊處理，但儲存了明文密碼")
	}
}

// TestRegister_DuplicateEmail 測試重複 Email 註冊
func TestRegister_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepository()
	uc := NewUserUsecase(repo, testConfig())

	req := domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	// 第一次註冊應該成功
	_, err := uc.Register(req)
	if err != nil {
		t.Fatalf("第一次註冊應該成功：%v", err)
	}

	// 第二次註冊相同 Email 應該失敗
	_, err = uc.Register(req)
	if err == nil {
		t.Error("重複 Email 註冊應該失敗")
	}
}

// TestLogin_Success 測試正常的登入流程
func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepository()
	uc := NewUserUsecase(repo, testConfig())

	// 先註冊一個使用者
	password := "password123"
	_, err := uc.Register(domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: password,
	})
	if err != nil {
		t.Fatalf("註冊失敗：%v", err)
	}

	// 嘗試登入
	loginResp, err := uc.Login(domain.LoginRequest{
		Email:    "test@example.com",
		Password: password,
	})
	if err != nil {
		t.Fatalf("預期登入成功，但得到錯誤：%v", err)
	}

	// 確認回傳了 Token
	if loginResp.Token == "" {
		t.Error("登入成功應該回傳 Token")
	}

	// 確認回傳了使用者資訊
	if loginResp.User.Email != "test@example.com" {
		t.Errorf("使用者 Email 不符：預期 test@example.com，得到 %s", loginResp.User.Email)
	}
}

// TestLogin_WrongPassword 測試錯誤密碼登入
func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockUserRepository()
	uc := NewUserUsecase(repo, testConfig())

	// 先註冊一個使用者
	_, err := uc.Register(domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "correct-password",
	})
	if err != nil {
		t.Fatalf("註冊失敗：%v", err)
	}

	// 使用錯誤密碼登入
	_, err = uc.Login(domain.LoginRequest{
		Email:    "test@example.com",
		Password: "wrong-password",
	})
	if err == nil {
		t.Error("使用錯誤密碼登入應該失敗")
	}
}

// TestLogin_NonExistentUser 測試不存在的使用者登入
func TestLogin_NonExistentUser(t *testing.T) {
	repo := newMockUserRepository()
	uc := NewUserUsecase(repo, testConfig())

	_, err := uc.Login(domain.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password",
	})
	if err == nil {
		t.Error("不存在的使用者登入應該失敗")
	}
}

// TestGetByID_Success 測試根據 ID 取得使用者
func TestGetByID_Success(t *testing.T) {
	repo := newMockUserRepository()
	uc := NewUserUsecase(repo, testConfig())

	// 先註冊一個使用者
	registered, err := uc.Register(domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("註冊失敗：%v", err)
	}

	// 根據 ID 查詢
	user, err := uc.GetByID(registered.ID)
	if err != nil {
		t.Fatalf("預期查詢成功，但得到錯誤：%v", err)
	}

	if user.ID != registered.ID {
		t.Errorf("使用者 ID 不符：預期 %d，得到 %d", registered.ID, user.ID)
	}
}

// 避免 import 未使用警告
var _ = bcrypt.DefaultCost
