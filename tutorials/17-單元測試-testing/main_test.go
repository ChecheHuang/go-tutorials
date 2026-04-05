// 第十六課：單元測試
// 測試檔案的命名規則：xxx_test.go
// 測試函式的命名規則：TestXxx（大寫 T 開頭）
//
// 執行方式：
//   go test -v          ← 詳細輸出
//   go test -cover      ← 顯示覆蓋率
//   go test -run TestAdd ← 只執行特定測試
package main

import (
	"errors"
	"testing"
)

// ========================================
// 1. 最基本的測試
// ========================================

// TestAdd 測試 Add 函式
func TestAdd(t *testing.T) {
	result := Add(2, 3)
	if result != 5 {
		// t.Errorf：記錄錯誤但繼續執行
		t.Errorf("Add(2, 3) = %d，期望 5", result)
	}
}

// ========================================
// 2. 表格驅動測試（Table-Driven Tests）
// ========================================
// 這是 Go 社群最推薦的測試方式：用表格列出所有測試案例

func TestAddTableDriven(t *testing.T) {
	// 定義測試案例
	tests := []struct {
		name     string // 測試案例名稱
		a, b     int    // 輸入
		expected int    // 期望輸出
	}{
		{"正數相加", 2, 3, 5},
		{"負數相加", -1, -2, -3},
		{"零相加", 0, 0, 0},
		{"正負相加", 5, -3, 2},
		{"大數相加", 1000, 2000, 3000},
	}

	for _, tt := range tests {
		// t.Run 建立子測試，讓每個案例獨立執行
		t.Run(tt.name, func(t *testing.T) {
			result := Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%d, %d) = %d，期望 %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// ========================================
// 3. 測試錯誤情況
// ========================================

func TestDivide(t *testing.T) {
	// 正常情況
	t.Run("正常除法", func(t *testing.T) {
		result, err := Divide(10, 3)
		if err != nil {
			t.Fatalf("不應該有錯誤，但得到: %v", err)
		}
		// 浮點數比較要用容差
		if result < 3.33 || result > 3.34 {
			t.Errorf("Divide(10, 3) = %f，期望約 3.33", result)
		}
	})

	// 錯誤情況：除以零
	t.Run("除以零", func(t *testing.T) {
		_, err := Divide(10, 0)
		if err == nil {
			t.Fatal("除以零應該回傳錯誤")
		}
	})
}

// ========================================
// 4. 測試結構體方法
// ========================================

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    User
		wantErr bool // 是否期望有錯誤
	}{
		{
			name:    "合法使用者",
			user:    User{Name: "Alice", Email: "alice@test.com", Age: 25},
			wantErr: false,
		},
		{
			name:    "空名稱",
			user:    User{Name: "", Email: "test@test.com", Age: 25},
			wantErr: true,
		},
		{
			name:    "負數年齡",
			user:    User{Name: "Bob", Email: "bob@test.com", Age: -1},
			wantErr: true,
		},
		{
			name:    "年齡過大",
			user:    User{Name: "Old", Email: "old@test.com", Age: 200},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v，wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserIsAdult(t *testing.T) {
	tests := []struct {
		age      int
		expected bool
	}{
		{17, false},
		{18, true},
		{25, true},
		{0, false},
	}

	for _, tt := range tests {
		user := User{Name: "Test", Age: tt.age}
		if result := user.IsAdult(); result != tt.expected {
			t.Errorf("Age=%d, IsAdult()=%v，期望 %v", tt.age, result, tt.expected)
		}
	}
}

// ========================================
// 5. Mock 測試（模擬依賴）
// ========================================

// mockUserRepository 是 UserRepository 的模擬實作
type mockUserRepository struct {
	users map[int]*User
}

func newMockRepo() *mockUserRepository {
	return &mockUserRepository{
		users: map[int]*User{
			1: {Name: "Alice", Email: "alice@test.com", Age: 25},
			2: {Name: "Bob", Email: "bob@test.com", Age: 30},
		},
	}
}

func (m *mockUserRepository) FindByID(id int) (*User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, errors.New("使用者不存在")
	}
	return user, nil
}

func (m *mockUserRepository) Create(user *User) error {
	return nil // 模擬成功
}

// 測試 UserService（使用 Mock Repository）
func TestUserServiceGetUser(t *testing.T) {
	// 建立 Service，注入 Mock
	repo := newMockRepo()
	service := NewUserService(repo)

	t.Run("存在的使用者", func(t *testing.T) {
		user, err := service.GetUser(1)
		if err != nil {
			t.Fatalf("不應該有錯誤: %v", err)
		}
		if user.Name != "Alice" {
			t.Errorf("名稱不符: 得到 %s，期望 Alice", user.Name)
		}
	})

	t.Run("不存在的使用者", func(t *testing.T) {
		_, err := service.GetUser(999)
		if err == nil {
			t.Fatal("應該回傳錯誤")
		}
	})
}

func TestUserServiceCreateUser(t *testing.T) {
	repo := newMockRepo()
	service := NewUserService(repo)

	t.Run("正常建立", func(t *testing.T) {
		user, err := service.CreateUser("Carol", "carol@test.com", 28)
		if err != nil {
			t.Fatalf("不應該有錯誤: %v", err)
		}
		if user.Name != "Carol" {
			t.Errorf("名稱不符: %s", user.Name)
		}
	})

	t.Run("空名稱應失敗", func(t *testing.T) {
		_, err := service.CreateUser("", "test@test.com", 25)
		if err == nil {
			t.Fatal("空名稱應該回傳錯誤")
		}
	})
}
