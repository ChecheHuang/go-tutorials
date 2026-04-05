package usecase

import (
	"errors"
	"testing"
	"time"

	"blog-api/internal/domain"
	"blog-api/pkg/apperror"
	"blog-api/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// === Mock Repository ===
type mockUserRepository struct {
	users  map[string]*domain.User
	nextID uint
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:  make(map[string]*domain.User),
		nextID: 1,
	}
}

func (m *mockUserRepository) Create(user *domain.User) error {
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

func testConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key",
			Expiration: 1 * time.Hour,
		},
	}
}

func TestUserUsecase_Register(t *testing.T) {
	t.Run("註冊成功", func(t *testing.T) {
		repo := newMockUserRepository()
		uc := NewUserUsecase(repo, testConfig())

		user, err := uc.Register(domain.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		})

		require.NoError(t, err)
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
		assert.NotEqual(t, "password123", user.Password, "密碼應該被雜湊")
	})

	t.Run("重複 Email 回傳 ErrConflict", func(t *testing.T) {
		repo := newMockUserRepository()
		uc := NewUserUsecase(repo, testConfig())

		req := domain.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		}

		_, err := uc.Register(req)
		require.NoError(t, err)

		_, err = uc.Register(req)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperror.ErrConflict), "應該是 ErrConflict，但得到：%v", err)
	})
}

func TestUserUsecase_Login(t *testing.T) {
	setup := func() UserUsecase {
		repo := newMockUserRepository()
		uc := NewUserUsecase(repo, testConfig())
		uc.Register(domain.RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password123",
		})
		return uc
	}

	t.Run("登入成功", func(t *testing.T) {
		uc := setup()

		resp, err := uc.Login(domain.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		})

		require.NoError(t, err)
		assert.NotEmpty(t, resp.Token)
		assert.Equal(t, "test@example.com", resp.User.Email)
	})

	t.Run("密碼錯誤回傳 ErrUnauthorized", func(t *testing.T) {
		uc := setup()

		_, err := uc.Login(domain.LoginRequest{
			Email:    "test@example.com",
			Password: "wrong-password",
		})

		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperror.ErrUnauthorized))
	})

	t.Run("使用者不存在回傳 ErrUnauthorized", func(t *testing.T) {
		uc := setup()

		_, err := uc.Login(domain.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password",
		})

		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperror.ErrUnauthorized))
	})
}

func TestUserUsecase_GetByID(t *testing.T) {
	repo := newMockUserRepository()
	uc := NewUserUsecase(repo, testConfig())

	registered, err := uc.Register(domain.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	})
	require.NoError(t, err)

	t.Run("找到使用者", func(t *testing.T) {
		user, err := uc.GetByID(registered.ID)
		require.NoError(t, err)
		assert.Equal(t, registered.ID, user.ID)
	})

	t.Run("使用者不存在回傳 ErrNotFound", func(t *testing.T) {
		_, err := uc.GetByID(999)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperror.ErrNotFound))
	})
}
