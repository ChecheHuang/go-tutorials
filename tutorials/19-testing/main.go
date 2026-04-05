// ==========================================================================
// 第十九課：單元測試（Unit Testing）
// ==========================================================================
//
// 什麼是單元測試？
//   單元測試就是「自動化的品質檢查員」
//   你寫程式碼，同時也寫測試程式碼，讓電腦自動幫你確認程式是否正確
//
//   就像工廠生產零件時，有一台專門的「檢測機器」
//   每次生產完就跑一遍，確保品質沒有問題
//
// 為什麼需要測試？
//   - 改 A 功能時，測試能立刻告訴你有沒有壞掉 B
//   - 重構程式碼時，測試給你信心（測試過了就沒問題）
//   - 新加入的人可以放心改程式碼
//
// Go 內建測試框架：
//   - 不需要安裝任何第三方套件
//   - 只需要 import "testing"
//   - 檔案命名：xxx_test.go
//   - 函式命名：TestXxx (大寫 T 開頭)
//
// 執行方式：
//   go test -v ./tutorials/17-testing/        # 跑所有測試（詳細輸出）
//   go test -cover ./tutorials/17-testing/    # 顯示測試覆蓋率
// ==========================================================================

package main // 宣告這是 main 套件（測試檔案必須和被測試的檔案同 package）

import ( // 匯入需要的套件
	"errors"   // 標準庫：錯誤處理
	"fmt"      // 標準庫：格式化輸出
	"net/http" // 標準庫：HTTP 功能（用於 Handler）
)

// ==========================================================================
// 1. 最基本的函式（用來示範測試）
// ==========================================================================

// Add 加法函式
// 這個函式很簡單，但它示範了「被測試的程式碼」長什麼樣子
func Add(a, b int) int { // 接受兩個整數，回傳整數
	return a + b // 回傳兩數之和
}

// Divide 除法函式（帶錯誤處理）
// 示範如何測試「可能回傳錯誤」的函式
func Divide(a, b float64) (float64, error) { // 接受兩個浮點數，回傳結果和錯誤
	if b == 0 { // 如果除數為零
		return 0, errors.New("除數不能為零") // 回傳零值和錯誤（不能除以零）
	}
	return a / b, nil // 回傳計算結果和 nil（nil 表示沒有錯誤）
}

// ==========================================================================
// 2. 結構體和方法（用來示範更複雜的測試）
// ==========================================================================

// User 使用者結構體
// 示範如何測試「結構體的方法」
type User struct { // 定義 User 結構體
	Name  string // 使用者名稱
	Email string // 電子信箱
	Age   int    // 年齡
}

// Validate 驗證使用者資料是否合法
// 回傳 nil 表示合法，回傳 error 表示不合法
func (u *User) Validate() error { // 指標接收者：可以修改 User 的內容（雖然這裡沒修改）
	if u.Name == "" { // 如果名稱是空字串
		return errors.New("名稱不能為空") // 回傳錯誤
	}
	if u.Age < 0 || u.Age > 150 { // 如果年齡不合理
		return errors.New("年齡不合理") // 回傳錯誤
	}
	return nil // 所有驗證通過，回傳 nil
}

// IsAdult 判斷是否成年（18 歲以上）
func (u *User) IsAdult() bool { // 回傳布林值
	return u.Age >= 18 // 大於等於 18 歲就是成年
}

// ==========================================================================
// 3. 介面和 Mock（示範依賴注入的可測試性）
// ==========================================================================
//
// 什麼是 Mock？
//   Mock 是一個「假的」實作，用來代替真實的依賴（如資料庫）
//
//   真實環境：UserService → 真實資料庫（慢、不可控）
//   測試環境：UserService → Mock 資料庫（快、可控）
//
//   因為 UserService 依賴的是「介面」（UserRepository）而不是具體的資料庫
//   所以測試時可以輕鬆換成 Mock！

// UserRepository 介面：定義「資料庫操作」的規格
// 任何實作了這個介面的物件，都可以作為 UserService 的資料庫
type UserRepository interface { // 定義介面
	FindByID(id int) (*User, error) // 根據 ID 查詢使用者
	Create(user *User) error        // 建立新使用者
}

// UserService 業務邏輯層：依賴 UserRepository 介面（不依賴具體實作）
type UserService struct { // 定義 UserService 結構體
	repo UserRepository // 儲存 UserRepository 介面（不管是真實的還是 Mock 的）
}

// NewUserService 建構函式：建立 UserService 並注入 Repository
func NewUserService(repo UserRepository) *UserService { // 依賴注入（Dependency Injection）
	return &UserService{repo: repo} // 建立並回傳 UserService
}

// GetUser 取得使用者資料
func (s *UserService) GetUser(id int) (*User, error) { // 方法：根據 ID 取得使用者
	user, err := s.repo.FindByID(id) // 呼叫 Repository 查詢
	if err != nil {                  // 如果查詢出錯
		return nil, fmt.Errorf("取得使用者失敗: %w", err) // 包裝錯誤後回傳
	}
	return user, nil // 回傳使用者資料
}

// CreateUser 建立新使用者（先驗證，再存入）
func (s *UserService) CreateUser(name, email string, age int) (*User, error) { // 方法：建立使用者
	user := &User{Name: name, Email: email, Age: age} // 建立 User 物件
	if err := user.Validate(); err != nil {            // 先驗證資料是否合法
		return nil, err // 驗證失敗直接回傳錯誤
	}
	if err := s.repo.Create(user); err != nil { // 驗證通過，存入 Repository
		return nil, fmt.Errorf("建立使用者失敗: %w", err) // 包裝錯誤後回傳
	}
	return user, nil // 回傳建立好的使用者
}

// ==========================================================================
// 4. HTTP Handler（示範如何測試 HTTP 處理函式）
// ==========================================================================
//
// 什麼是 HTTP Handler 測試？
//   測試 HTTP Handler 時，我們不需要真正啟動伺服器
//   Go 的 net/http/httptest 套件提供了「模擬 HTTP 請求」的工具
//
//   就像用「模擬器」測試手機 App，不需要真的買一台手機

// UserHandler 處理使用者相關的 HTTP 請求
type UserHandler struct { // 定義 Handler 結構體
	service *UserService // 依賴 UserService（業務邏輯層）
}

// NewUserHandler 建構函式
func NewUserHandler(service *UserService) *UserHandler { // 依賴注入
	return &UserHandler{service: service} // 建立並回傳 UserHandler
}

// GetUserHandler 處理 GET /users/{id} 的請求
// 注意：這裡使用標準庫的 http.ResponseWriter 和 *http.Request（不是 Gin）
// 因為這個教學著重在測試概念，標準庫更容易示範
func (h *UserHandler) GetUserHandler(w http.ResponseWriter, r *http.Request) { // HTTP Handler 函式
	// 從 URL 查詢參數取得 user_id
	// 例如：GET /users?id=1 → r.URL.Query().Get("id") = "1"
	idStr := r.URL.Query().Get("id") // 取得 URL 查詢參數 id

	if idStr == "" { // 如果沒有提供 id 參數
		http.Error(w, "缺少 id 參數", http.StatusBadRequest) // 回傳 400 Bad Request
		return                                                // 提前返回
	}

	// 簡化版本：這裡只示範基本的 HTTP 測試流程
	// 實際上應該用 strconv.Atoi 轉換 idStr → int，再查詢資料庫
	if idStr == "1" { // 模擬 ID=1 存在
		w.Header().Set("Content-Type", "application/json") // 設定回應格式為 JSON
		w.WriteHeader(http.StatusOK)                        // 設定 HTTP 狀態碼 200
		fmt.Fprintf(w, `{"id":1,"name":"Alice"}`)           // 寫入 JSON 回應內容
	} else { // 其他 ID 視為不存在
		http.Error(w, "使用者不存在", http.StatusNotFound) // 回傳 404 Not Found
	}
}

// CreateUserHandler 處理 POST /users 的請求（建立使用者）
func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) { // HTTP Handler 函式
	if r.Method != http.MethodPost { // 如果不是 POST 方法
		http.Error(w, "只允許 POST 方法", http.StatusMethodNotAllowed) // 回傳 405 Method Not Allowed
		return                                                          // 提前返回
	}

	// 模擬成功建立使用者
	w.Header().Set("Content-Type", "application/json") // 設定回應格式
	w.WriteHeader(http.StatusCreated)                   // 設定狀態碼 201 Created
	fmt.Fprintf(w, `{"id":2,"name":"Bob","message":"建立成功"}`) // 寫入 JSON 回應
}

// ==========================================================================
// 主程式（這個檔案主要是被測試用的，主程式很簡單）
// ==========================================================================

func main() { // 程式進入點
	fmt.Println("這個檔案主要用於被測試") // 說明這個檔案的用途
	fmt.Println("請執行: go test -v ./tutorials/17-testing/") // 提示測試指令
	fmt.Println()                                              // 空行

	// 示範依賴注入：UserHandler 依賴 UserService，UserService 依賴 UserRepository
	// 在真實專案中，這些依賴會在 main.go 中組裝（依賴注入）
	// 在測試中，我們用 Mock 替換真實的 Repository
	fmt.Println("程式架構示意：")                              // 說明架構
	fmt.Println("  UserHandler → UserService → UserRepository") // 依賴關係
	fmt.Println("  測試時：把 UserRepository 換成 MockRepository") // 測試策略
}
