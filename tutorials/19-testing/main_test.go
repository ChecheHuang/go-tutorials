// ==========================================================================
// 第十七課：單元測試 — 測試檔案
// ==========================================================================
//
// 測試檔案的命名規則：
//   - 檔名必須以 _test.go 結尾（例如 main_test.go、user_test.go）
//   - 與被測試的檔案必須在同一個 package（這裡是 package main）
//
// 測試函式的命名規則：
//   - 必須以 Test 開頭（大寫 T）
//   - 參數固定是 (t *testing.T)
//   - 例如：TestAdd、TestDivide、TestUserValidate
//
// 執行方式：
//   go test -v ./tutorials/17-testing/        # 詳細輸出所有測試結果
//   go test -cover ./tutorials/17-testing/    # 顯示測試覆蓋率（幾 % 的程式碼被測試到）
//   go test -run TestAdd ./tutorials/17-testing/  # 只跑名稱包含 TestAdd 的測試
// ==========================================================================

package main // 必須和 main.go 同一個 package

import ( // 匯入需要的套件
	"errors"          // 標準庫：建立錯誤
	"io"              // 標準庫：讀取回應內容
	"net/http"        // 標準庫：HTTP 功能
	"net/http/httptest" // 標準庫：HTTP 測試工具（模擬請求，不需要真正啟動伺服器）
	"strings"         // 標準庫：字串處理
	"testing"         // 標準庫：Go 的測試框架
)

// ==========================================================================
// 1. 最基本的測試
// ==========================================================================

// TestAdd 測試 Add 函式
// 測試函式名稱 = Test + 被測試函式名稱
func TestAdd(t *testing.T) { // t 是測試控制器，用來記錄錯誤和控制測試
	result := Add(2, 3) // 呼叫被測試的函式
	if result != 5 {    // 如果結果不符合預期
		// t.Errorf：記錄錯誤訊息，但測試繼續執行（不會立即停止）
		t.Errorf("Add(2, 3) = %d，期望 5", result) // %d 是整數佔位符
	}
}

// ==========================================================================
// 2. 表格驅動測試（Table-Driven Tests）——Go 社群最推薦的測試方式
// ==========================================================================
//
// 為什麼用表格驅動測試？
//   不用表格：每個測試案例都要寫一個測試函式（重複程式碼很多）
//   用表格：把所有測試案例集中在一個表格裡，程式碼更簡潔、更好維護

// TestAddTableDriven 用表格驅動方式測試 Add 函式
func TestAddTableDriven(t *testing.T) { // 表格驅動測試函式
	// 定義測試案例表格（匿名結構體切片）
	tests := []struct { // 定義每個測試案例的格式
		name     string // 測試案例的名稱（出錯時方便識別）
		a, b     int    // 輸入：兩個整數
		expected int    // 期望的輸出
	}{ // 開始定義具體的測試案例
		{"正數相加", 2, 3, 5},      // 最常見的情況
		{"負數相加", -1, -2, -3},   // 負數
		{"零相加", 0, 0, 0},        // 邊界條件：零
		{"正負相加", 5, -3, 2},     // 正負混合
		{"大數相加", 1000, 2000, 3000}, // 較大的數字
	}

	for _, tt := range tests { // 走訪每個測試案例（tt 是 "table test" 的縮寫）
		// t.Run 建立子測試（Sub-test）
		// 這樣每個測試案例可以獨立執行、獨立報告
		t.Run(tt.name, func(t *testing.T) { // 子測試的名稱是 tt.name
			result := Add(tt.a, tt.b) // 執行被測試的函式
			if result != tt.expected { // 如果結果不符合期望
				t.Errorf("Add(%d, %d) = %d，期望 %d", tt.a, tt.b, result, tt.expected) // 報告錯誤
			}
		})
	}
}

// ==========================================================================
// 3. 測試有錯誤回傳的函式
// ==========================================================================

// TestDivide 測試 Divide 函式（包括正常情況和錯誤情況）
func TestDivide(t *testing.T) { // 測試除法函式
	// t.Run 讓你在一個測試函式裡分組測試不同情況
	t.Run("正常除法", func(t *testing.T) { // 測試正常除法
		result, err := Divide(10, 3)  // 10 ÷ 3 ≈ 3.333...
		if err != nil {               // 如果回傳了錯誤（不應該）
			// t.Fatalf：記錄錯誤並立即停止當前子測試（比 t.Errorf 更嚴格）
			t.Fatalf("不應該有錯誤，但得到: %v", err) // 致命錯誤，立即停止
		}
		// 浮點數比較不能用 ==（因為有精度問題），要用「範圍比較」
		if result < 3.33 || result > 3.34 { // 結果應該在 3.33 和 3.34 之間
			t.Errorf("Divide(10, 3) = %f，期望約 3.33", result) // 報告錯誤
		}
	})

	t.Run("除以零", func(t *testing.T) { // 測試除以零的錯誤情況
		_, err := Divide(10, 0) // _ 忽略第一個回傳值（我們只想測試 error）
		if err == nil {         // 如果沒有回傳錯誤（不符合預期）
			t.Fatal("除以零應該回傳錯誤") // 這是嚴重問題，立即停止子測試
		}
		// 補充：可以用 t.Log 輸出額外資訊（只有加 -v 才會顯示）
		t.Logf("正確地回傳了錯誤：%v", err) // 記錄錯誤訊息（-v 模式下可見）
	})
}

// ==========================================================================
// 4. 測試結構體方法
// ==========================================================================

// TestUserValidate 用表格驅動方式測試 User.Validate 方法
func TestUserValidate(t *testing.T) { // 測試使用者驗證
	// 定義測試案例表格
	tests := []struct { // 匿名結構體切片
		name    string // 測試案例名稱
		user    User   // 輸入：User 物件
		wantErr bool   // 期望是否有錯誤（true = 應該有錯誤，false = 不應該有錯誤）
	}{ // 具體的測試案例
		{
			name:    "合法使用者",                                    // 正常情況
			user:    User{Name: "Alice", Email: "alice@test.com", Age: 25}, // 合法的使用者
			wantErr: false, // 不應該有錯誤
		},
		{
			name:    "空名稱",                                        // 錯誤情況：名稱為空
			user:    User{Name: "", Email: "test@test.com", Age: 25}, // 名稱是空字串
			wantErr: true, // 應該有錯誤
		},
		{
			name:    "負數年齡",                                      // 錯誤情況：年齡是負數
			user:    User{Name: "Bob", Email: "bob@test.com", Age: -1}, // 年齡 -1 不合理
			wantErr: true, // 應該有錯誤
		},
		{
			name:    "年齡過大",                                       // 錯誤情況：年齡超過 150
			user:    User{Name: "Old", Email: "old@test.com", Age: 200}, // 200 歲不合理
			wantErr: true, // 應該有錯誤
		},
	}

	for _, tt := range tests { // 走訪每個測試案例
		t.Run(tt.name, func(t *testing.T) { // 建立子測試
			err := tt.user.Validate()           // 執行驗證方法
			if (err != nil) != tt.wantErr {     // 如果「有沒有錯誤」與期望不符
				t.Errorf("Validate() error = %v，wantErr = %v", err, tt.wantErr) // 報告錯誤
			}
		})
	}
}

// TestUserIsAdult 測試 User.IsAdult 方法
func TestUserIsAdult(t *testing.T) { // 測試是否成年
	tests := []struct { // 測試案例表格
		age      int  // 輸入：年齡
		expected bool // 期望輸出：是否成年
	}{ // 具體案例
		{17, false}, // 17 歲：未成年
		{18, true},  // 18 歲：成年（邊界值）
		{25, true},  // 25 歲：成年
		{0, false},  // 0 歲：未成年（邊界值）
	}

	for _, tt := range tests { // 走訪每個測試案例
		user := User{Name: "Test", Age: tt.age} // 建立測試用的 User
		if result := user.IsAdult(); result != tt.expected { // 如果結果不符
			t.Errorf("Age=%d, IsAdult()=%v，期望 %v", tt.age, result, tt.expected) // 報告錯誤
		}
	}
}

// ==========================================================================
// 5. Mock 測試（模擬資料庫依賴）
// ==========================================================================
//
// Mock 的實作方式：
//   1. 定義一個 struct，實作 UserRepository 介面的所有方法
//   2. 內部用 map 模擬資料庫（記憶體中的資料）
//   3. 把 Mock 注入到 UserService 中進行測試

// mockUserRepository 是 UserRepository 介面的「假實作」
type mockUserRepository struct { // 模擬的資料庫
	users map[int]*User // 用 map 模擬資料表（key=ID, value=User）
}

// newMockRepo 建立一個預設有資料的 Mock Repository
func newMockRepo() *mockUserRepository { // 建構函式
	return &mockUserRepository{ // 建立 mockUserRepository
		users: map[int]*User{ // 預先放入測試資料
			1: {Name: "Alice", Email: "alice@test.com", Age: 25}, // ID=1 的使用者
			2: {Name: "Bob", Email: "bob@test.com", Age: 30},     // ID=2 的使用者
		},
	}
}

// FindByID 實作 UserRepository 介面的 FindByID 方法
func (m *mockUserRepository) FindByID(id int) (*User, error) { // Mock 的查詢方法
	user, exists := m.users[id] // 從 map 中查詢
	if !exists {                 // 如果不存在
		return nil, errors.New("使用者不存在") // 回傳錯誤（模擬資料庫找不到的情況）
	}
	return user, nil // 回傳找到的使用者
}

// Create 實作 UserRepository 介面的 Create 方法
func (m *mockUserRepository) Create(user *User) error { // Mock 的建立方法
	return nil // 模擬成功（不需要真的存入任何地方）
}

// TestUserServiceGetUser 測試 UserService.GetUser 方法（使用 Mock）
func TestUserServiceGetUser(t *testing.T) { // 測試取得使用者
	repo := newMockRepo()          // 建立 Mock Repository
	service := NewUserService(repo) // 把 Mock 注入到 Service（依賴注入！）

	t.Run("存在的使用者", func(t *testing.T) { // 測試查詢存在的使用者
		user, err := service.GetUser(1) // 查詢 ID=1（Alice）
		if err != nil {                 // 如果有錯誤（不應該）
			t.Fatalf("不應該有錯誤: %v", err) // 致命錯誤
		}
		if user.Name != "Alice" { // 如果名稱不是 Alice
			t.Errorf("名稱不符: 得到 %s，期望 Alice", user.Name) // 報告錯誤
		}
	})

	t.Run("不存在的使用者", func(t *testing.T) { // 測試查詢不存在的使用者
		_, err := service.GetUser(999) // 查詢 ID=999（不存在）
		if err == nil {                // 如果沒有回傳錯誤（不符合預期）
			t.Fatal("應該回傳錯誤") // 致命錯誤
		}
	})
}

// TestUserServiceCreateUser 測試 UserService.CreateUser 方法
func TestUserServiceCreateUser(t *testing.T) { // 測試建立使用者
	repo := newMockRepo()           // 建立 Mock Repository
	service := NewUserService(repo) // 注入依賴

	t.Run("正常建立", func(t *testing.T) { // 測試正常建立
		user, err := service.CreateUser("Carol", "carol@test.com", 28) // 建立 Carol
		if err != nil { // 如果有錯誤
			t.Fatalf("不應該有錯誤: %v", err) // 致命錯誤
		}
		if user.Name != "Carol" { // 如果名稱不對
			t.Errorf("名稱不符: %s", user.Name) // 報告錯誤
		}
	})

	t.Run("空名稱應失敗", func(t *testing.T) { // 測試空名稱會被驗證攔下
		_, err := service.CreateUser("", "test@test.com", 25) // 名稱是空字串
		if err == nil { // 如果沒有回傳錯誤（應該要有）
			t.Fatal("空名稱應該回傳錯誤") // 致命錯誤
		}
	})
}

// ==========================================================================
// 6. HTTP Handler 測試（使用 httptest 套件）
// ==========================================================================
//
// 什麼是 httptest？
//   httptest 是 Go 標準庫中的 HTTP 測試工具
//   它讓你「模擬 HTTP 請求」，不需要真正啟動伺服器
//
//   就像用「假的客戶」測試服務員——客戶說一句話，看服務員怎麼回應
//
// httptest 的兩個核心工具：
//   - httptest.NewRecorder()：建立一個「假的回應寫入器」，記錄 Handler 的回應
//   - httptest.NewRequest()：建立一個「假的 HTTP 請求」

// TestGetUserHandler 測試 GetUserHandler 函式
func TestGetUserHandler(t *testing.T) { // 測試取得使用者的 HTTP Handler
	// 建立 Handler（在真實專案中，這些依賴會在 main.go 中組裝）
	repo := newMockRepo()           // 建立 Mock Repository
	service := NewUserService(repo) // 建立 Service
	handler := NewUserHandler(service) // 建立 Handler

	t.Run("正確的 ID 回傳使用者資料", func(t *testing.T) { // 測試 ID=1 的情況
		// 1. 建立假的 HTTP 請求
		// httptest.NewRequest(方法, URL, 請求 Body)
		req := httptest.NewRequest(http.MethodGet, "/users?id=1", nil) // GET /users?id=1
		// nil 表示沒有請求 Body（GET 請求通常沒有 Body）

		// 2. 建立假的回應記錄器（用來接收 Handler 的回應）
		w := httptest.NewRecorder() // 建立 ResponseRecorder

		// 3. 直接呼叫 Handler（不需要啟動 HTTP 伺服器！）
		handler.GetUserHandler(w, req) // 執行 Handler

		// 4. 從記錄器取出回應結果
		resp := w.Result() // 取得 http.Response 物件

		// 5. 驗證 HTTP 狀態碼
		if resp.StatusCode != http.StatusOK { // 期望 200 OK
			t.Errorf("狀態碼 = %d，期望 %d", resp.StatusCode, http.StatusOK) // 報告錯誤
		}

		// 6. 驗證回應內容
		body, _ := io.ReadAll(resp.Body) // 讀取回應 Body（_ 忽略 error）
		defer resp.Body.Close()          // 記得關閉 Body（避免資源洩漏）

		if !strings.Contains(string(body), "Alice") { // 如果回應中沒有 "Alice"
			t.Errorf("回應 Body 應該包含 Alice，實際得到: %s", string(body)) // 報告錯誤
		}
	})

	t.Run("缺少 id 參數回傳 400", func(t *testing.T) { // 測試沒有 id 參數的情況
		req := httptest.NewRequest(http.MethodGet, "/users", nil) // 沒有 ?id= 參數
		w := httptest.NewRecorder()                                // 建立記錄器

		handler.GetUserHandler(w, req) // 執行 Handler

		if w.Code != http.StatusBadRequest { // w.Code 是記錄器記錄的狀態碼，期望 400
			t.Errorf("狀態碼 = %d，期望 %d", w.Code, http.StatusBadRequest) // 報告錯誤
		}
	})

	t.Run("不存在的 ID 回傳 404", func(t *testing.T) { // 測試 ID=999 不存在的情況
		req := httptest.NewRequest(http.MethodGet, "/users?id=999", nil) // ID=999
		w := httptest.NewRecorder()                                       // 建立記錄器

		handler.GetUserHandler(w, req) // 執行 Handler

		if w.Code != http.StatusNotFound { // 期望 404 Not Found
			t.Errorf("狀態碼 = %d，期望 %d", w.Code, http.StatusNotFound) // 報告錯誤
		}
	})
}

// TestCreateUserHandler 測試 CreateUserHandler 函式
func TestCreateUserHandler(t *testing.T) { // 測試建立使用者的 HTTP Handler
	repo := newMockRepo()             // 建立 Mock Repository
	service := NewUserService(repo)   // 建立 Service
	handler := NewUserHandler(service) // 建立 Handler

	t.Run("POST 成功建立", func(t *testing.T) { // 測試正確的 POST 請求
		// 模擬 POST 請求（真實情況下 Body 會有 JSON 資料）
		req := httptest.NewRequest(http.MethodPost, "/users", nil) // POST /users
		w := httptest.NewRecorder()                                 // 建立記錄器

		handler.CreateUserHandler(w, req) // 執行 Handler

		if w.Code != http.StatusCreated { // 期望 201 Created
			t.Errorf("狀態碼 = %d，期望 %d", w.Code, http.StatusCreated) // 報告錯誤
		}
	})

	t.Run("GET 方法回傳 405", func(t *testing.T) { // 測試使用錯誤的 HTTP 方法
		req := httptest.NewRequest(http.MethodGet, "/users", nil) // 用 GET 打 POST 路由
		w := httptest.NewRecorder()                                // 建立記錄器

		handler.CreateUserHandler(w, req) // 執行 Handler

		if w.Code != http.StatusMethodNotAllowed { // 期望 405 Method Not Allowed
			t.Errorf("狀態碼 = %d，期望 %d", w.Code, http.StatusMethodNotAllowed) // 報告錯誤
		}
	})
}

// ==========================================================================
// 補充：t.Errorf vs t.Fatalf vs t.Log 的差別
// ==========================================================================
//
// t.Errorf(...)  → 記錄錯誤，繼續執行後面的測試（非致命）
// t.Fatalf(...)  → 記錄錯誤，立即停止當前測試函式（致命）
// t.Fatal(...)   → 同 Fatalf，但不格式化
// t.Log(...)     → 記錄訊息（只有加 -v 才顯示，測試通過也會記錄）
// t.Logf(...)    → 同 Log，但支援格式化字串
//
// 使用原則：
//   如果後面的測試依賴前面的結果 → 用 Fatalf（避免 nil pointer 等問題）
//   如果後面的測試是獨立的 → 用 Errorf（收集更多錯誤資訊）
