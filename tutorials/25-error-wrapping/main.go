// ==========================================================================
// 第二十五課：Error Wrapping 完整示範
// ==========================================================================
//
// 第七課學了基本的錯誤處理（errors.New、fmt.Errorf）
// 這一課學更進階的「錯誤包裝」技巧，這是正式環境必備的能力
//
// 什麼是 Error Wrapping（錯誤包裝）？
//
//   資料庫回傳了 "record not found" 錯誤
//   Repository 把它包成 "查詢使用者失敗: record not found"
//   Usecase 再包成 "建立訂單失敗: 查詢使用者失敗: record not found"
//   Handler 看到錯誤，能判斷「最根本的原因」是什麼
//
//   這就像剝洋蔥：外層是業務錯誤，一層層剝開後，最裡面是原始錯誤
//
// 為什麼要包裝錯誤？
//   1. 保留原始錯誤（方便除錯）
//   2. 加上上下文資訊（在哪裡、做什麼時出錯）
//   3. 讓呼叫者可以判斷錯誤類型（errors.Is/As）
//
// 執行方式：go run ./tutorials/25-error-wrapping
// ==========================================================================

package main // 宣告這是 main 套件

import (
	"errors"  // 標準庫：errors.New、errors.Is、errors.As、errors.Unwrap
	"fmt"     // 標準庫：fmt.Errorf（支援 %w 包裝錯誤）
)

// ==========================================================================
// 1. fmt.Errorf 和 %w（最常用的包裝方式）
// ==========================================================================
//
// %w 是 Go 1.13 加入的特殊佔位符，用來「包裝」錯誤：
//   err := fmt.Errorf("上層操作失敗: %w", originalErr)
//   → err.Error() = "上層操作失敗: 原始錯誤訊息"
//   → errors.Unwrap(err) = originalErr（可以取出原始錯誤）
//
// 注意：用 %v 只是把錯誤訊息嵌入字串，不保留錯誤類型：
//   err := fmt.Errorf("失敗: %v", originalErr)  ← 不能用 errors.Is 解包！

// 模擬的底層錯誤（資料庫層）
var ErrNotFound = errors.New("記錄不存在")        // 資料庫：找不到記錄
var ErrDuplicate = errors.New("記錄已存在")        // 資料庫：重複插入
var ErrConnectionFailed = errors.New("資料庫連線失敗") // 資料庫：連線錯誤

func demonstrateBasicWrapping() { // 示範基本包裝
	fmt.Println("=== 1. fmt.Errorf %w 基本包裝 ===\n")

	// 模擬 Repository 層：把底層錯誤包裝後回傳
	findUser := func(id int) error {
		// 模擬資料庫回傳 "記錄不存在"
		dbErr := ErrNotFound                                      // 底層錯誤
		return fmt.Errorf("findUser(id=%d): %w", id, dbErr)      // 包裝：加上上下文
	}

	// 模擬 Usecase 層：再包裝一層
	getOrder := func(userID int) error {
		if err := findUser(userID); err != nil {
			return fmt.Errorf("getOrder: 查詢使用者失敗: %w", err) // 再包一層
		}
		return nil
	}

	err := getOrder(42) // 觸發錯誤鏈

	fmt.Printf("錯誤訊息（最外層）: %v\n", err)
	// 輸出：getOrder: 查詢使用者失敗: findUser(id=42): 記錄不存在

	// errors.Is：沿著整條錯誤鏈向內找，看是否包含特定錯誤
	fmt.Printf("errors.Is(err, ErrNotFound) = %v\n", errors.Is(err, ErrNotFound))
	// 輸出：true（即使 ErrNotFound 被包了兩層，也能找到！）

	fmt.Printf("errors.Is(err, ErrDuplicate) = %v\n", errors.Is(err, ErrDuplicate))
	// 輸出：false

	// errors.Unwrap：剝一層（只剝一層）
	unwrapped := errors.Unwrap(err)
	fmt.Printf("Unwrap 一層: %v\n", unwrapped)
	// 輸出：findUser(id=42): 記錄不存在

	unwrapped2 := errors.Unwrap(unwrapped)
	fmt.Printf("Unwrap 兩層: %v\n", unwrapped2)
	// 輸出：記錄不存在（這就是 ErrNotFound 本身）
}

// ==========================================================================
// 2. 自訂錯誤類型（Custom Error Type）
// ==========================================================================
//
// 有時候只有錯誤訊息不夠，還需要帶額外資訊：
//   - HTTP 狀態碼（這個錯誤對應 404 還是 400？）
//   - 錯誤代碼（給前端顯示的 error_code）
//   - 影響的欄位名稱（表單驗證時，是哪個欄位出問題）
//
// 自訂 error type 只需要實作 Error() string 方法

// AppError 應用程式層的自訂錯誤
type AppError struct {
	Code    string // 錯誤代碼（給前端使用）
	Message string // 使用者看到的錯誤訊息
	Status  int    // 對應的 HTTP 狀態碼
	Cause   error  // 原始錯誤（可選）
}

// Error 實作 error 介面（必須實作這個方法）
func (e *AppError) Error() string {
	if e.Cause != nil { // 如果有原始錯誤
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause) // 加上原因
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message) // 只有代碼和訊息
}

// Unwrap 讓 errors.Is/As 可以穿透這個錯誤找到 Cause
func (e *AppError) Unwrap() error {
	return e.Cause // 回傳原始錯誤（讓 errors.Is 可以繼續向內找）
}

// 預先定義常用的 AppError（讓呼叫者可以用 errors.Is 判斷）
var (
	ErrUserNotFound = &AppError{
		Code:    "USER_NOT_FOUND",
		Message: "使用者不存在",
		Status:  404,
	}
	ErrUnauthorized = &AppError{
		Code:    "UNAUTHORIZED",
		Message: "未授權的操作",
		Status:  401,
	}
	ErrValidationFailed = &AppError{
		Code:    "VALIDATION_FAILED",
		Message: "資料驗證失敗",
		Status:  400,
	}
)

// newAppError 建立一個新的 AppError，包裝原始錯誤
func newAppError(template *AppError, cause error) *AppError {
	return &AppError{
		Code:    template.Code,
		Message: template.Message,
		Status:  template.Status,
		Cause:   cause, // 包裝原始錯誤
	}
}

func demonstrateCustomError() { // 示範自訂錯誤類型
	fmt.Println("\n=== 2. 自訂錯誤類型（Custom Error Type）===\n")

	// 模擬 Repository：資料庫找不到使用者
	findUserInDB := func(id int) error {
		return fmt.Errorf("SELECT FROM users WHERE id=%d: %w", id, ErrNotFound) // DB 錯誤
	}

	// 模擬 Usecase：把 DB 錯誤轉換成 AppError
	getUser := func(id int) error {
		dbErr := findUserInDB(id)
		if dbErr != nil {
			if errors.Is(dbErr, ErrNotFound) { // 如果是「找不到」
				return newAppError(ErrUserNotFound, dbErr) // 包裝成 AppError
			}
			return fmt.Errorf("getUser 發生未知錯誤: %w", dbErr)
		}
		return nil
	}

	err := getUser(99) // 觸發錯誤

	fmt.Printf("錯誤訊息: %v\n", err)

	// errors.As：在錯誤鏈中找到特定類型的錯誤（比 errors.Is 更強大）
	// errors.Is 只能判斷「是不是同一個錯誤」
	// errors.As 可以「取出錯誤物件」，然後存取它的欄位
	var appErr *AppError
	if errors.As(err, &appErr) { // 如果錯誤鏈中有 *AppError
		fmt.Printf("錯誤代碼: %s\n", appErr.Code)       // USER_NOT_FOUND
		fmt.Printf("HTTP 狀態碼: %d\n", appErr.Status) // 404
		fmt.Printf("使用者訊息: %s\n", appErr.Message) // 使用者不存在
	}

	// 也可以用 errors.Is 判斷是否是特定的預定義錯誤
	fmt.Printf("errors.Is(err, ErrUserNotFound) = %v\n", errors.Is(err, ErrUserNotFound))
	// 注意：因為 ErrUserNotFound 和 newAppError 建立的是不同實例，
	// 需要在 Is() 方法中實作比較邏輯，或改用其他方式
}

// ==========================================================================
// 3. errors.Is 和 errors.As 的完整用法
// ==========================================================================

// ValidationError 欄位驗證錯誤（帶欄位名稱）
type ValidationError struct {
	Field   string // 哪個欄位有問題
	Message string // 錯誤說明
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("欄位 '%s' 驗證失敗: %s", e.Field, e.Message)
}

// MultiError 多個錯誤的集合（一次回傳所有驗證錯誤）
type MultiError struct {
	Errors []error // 所有錯誤的列表
}

func (m *MultiError) Error() string {
	msg := fmt.Sprintf("發現 %d 個錯誤:", len(m.Errors))
	for _, e := range m.Errors {
		msg += "\n  - " + e.Error()
	}
	return msg
}

// Unwrap 讓 errors.Is/As 可以遍歷所有錯誤（Go 1.20+）
func (m *MultiError) Unwrap() []error {
	return m.Errors // 回傳所有錯誤的 slice（Go 1.20+ 支援）
}

func demonstrateErrorsIsAs() { // 示範 errors.Is 和 errors.As
	fmt.Println("\n=== 3. errors.Is vs errors.As 完整示範 ===\n")

	// ---- errors.Is：判斷「是不是這個錯誤」----
	fmt.Println("【errors.Is：沿著錯誤鏈找特定錯誤】")

	sentinel := errors.New("基礎錯誤") // Sentinel error（哨兵錯誤）
	wrapped := fmt.Errorf("中層: %w", fmt.Errorf("底層: %w", sentinel))

	fmt.Printf("errors.Is(wrapped, sentinel) = %v\n", errors.Is(wrapped, sentinel)) // true
	fmt.Printf("錯誤鏈: %v\n", wrapped)

	// ---- errors.As：取出特定類型的錯誤物件 ----
	fmt.Println("\n【errors.As：從錯誤鏈中取出特定類型的錯誤】")

	validateForm := func() error {
		errs := &MultiError{ // 收集所有驗證錯誤
			Errors: []error{
				&ValidationError{Field: "email", Message: "格式不正確"},
				&ValidationError{Field: "password", Message: "長度不足（最少 8 字元）"},
				fmt.Errorf("伺服器內部錯誤: %w", errors.New("timeout")),
			},
		}
		return fmt.Errorf("表單驗證失敗: %w", errs) // 包裝成外層錯誤
	}

	err := validateForm()
	fmt.Printf("錯誤訊息: %v\n", err)

	// 用 errors.As 取出 *MultiError
	var multiErr *MultiError
	if errors.As(err, &multiErr) {
		fmt.Printf("\n找到 MultiError，共 %d 個錯誤:\n", len(multiErr.Errors))
		for i, e := range multiErr.Errors {
			// 再用 errors.As 看每個錯誤是不是 ValidationError
			var valErr *ValidationError
			if errors.As(e, &valErr) {
				fmt.Printf("  [%d] 欄位: %-10s → %s\n", i+1, valErr.Field, valErr.Message)
			} else {
				fmt.Printf("  [%d] 系統錯誤: %v\n", i+1, e)
			}
		}
	}
}

// ==========================================================================
// 4. 實際場景：Clean Architecture 中的錯誤流向
// ==========================================================================
//
// 正式專案中，錯誤從最底層流向最頂層：
//
//   資料庫錯誤（gorm.ErrRecordNotFound）
//        ↓ Repository 包裝
//   領域錯誤（ErrArticleNotFound）
//        ↓ Usecase 加上下文
//   業務錯誤（fmt.Errorf("getArticle: %w", ErrArticleNotFound)）
//        ↓ Handler 判斷
//   HTTP 回應（404 Not Found）

// 模擬 GORM 的錯誤
var gormErrRecordNotFound = errors.New("gorm: record not found")

// 領域錯誤（Domain Layer）
var ErrArticleNotFound = errors.New("文章不存在")

func demonstrateErrorFlow() { // 示範錯誤在各層的流向
	fmt.Println("\n=== 4. Clean Architecture 中的錯誤流向 ===\n")

	// Repository 層：把 DB 錯誤轉換成領域錯誤
	findArticleRepo := func(id int) error {
		dbErr := gormErrRecordNotFound // 模擬 DB 錯誤
		if errors.Is(dbErr, gormErrRecordNotFound) {
			return fmt.Errorf("findArticle(id=%d): %w", id, ErrArticleNotFound) // 轉換成領域錯誤
		}
		return fmt.Errorf("findArticle: 資料庫錯誤: %w", dbErr)
	}

	// Usecase 層：加上業務上下文
	getArticleUsecase := func(id int) error {
		if err := findArticleRepo(id); err != nil {
			return fmt.Errorf("getArticle(userRequest): %w", err) // 加上業務上下文
		}
		return nil
	}

	// Handler 層：判斷錯誤類型，回傳對應的 HTTP 狀態碼
	handleGetArticle := func(id int) {
		err := getArticleUsecase(id)
		if err == nil {
			fmt.Printf("  GET /articles/%d → 200 OK\n", id)
			return
		}

		// 判斷錯誤類型（不管包了幾層，errors.Is 都能找到）
		switch {
		case errors.Is(err, ErrArticleNotFound):
			fmt.Printf("  GET /articles/%d → 404 Not Found: %v\n", id, err)
		default:
			fmt.Printf("  GET /articles/%d → 500 Internal Server Error: %v\n", id, err)
		}
	}

	handleGetArticle(1)  // 模擬找不到文章 → 404
	handleGetArticle(42) // 模擬找不到文章 → 404

	fmt.Println("\n💡 重點：")
	fmt.Println("  Repository 把 gorm 錯誤轉成領域錯誤（隱藏 DB 細節）")
	fmt.Println("  Usecase 再包裝加上業務上下文（方便 debug）")
	fmt.Println("  Handler 用 errors.Is 判斷，不管包了幾層都找得到")
}

// ==========================================================================
// 5. 錯誤處理最佳實踐總結
// ==========================================================================

func demonstrateBestPractices() { // 示範最佳實踐
	fmt.Println("\n=== 5. 錯誤處理最佳實踐 ===\n")

	fmt.Println("✅ 正確做法：")
	fmt.Println()
	fmt.Println("  1. 每層都包裝錯誤，加上上下文：")
	fmt.Println(`     return fmt.Errorf("createUser(email=%s): %w", email, err)`)
	fmt.Println()
	fmt.Println("  2. 用 errors.Is 判斷錯誤類型（不要比較 err.Error() 字串！）：")
	fmt.Println("     ✅ errors.Is(err, ErrNotFound)")
	fmt.Println(`     ❌ err.Error() == "not found"`)
	fmt.Println()
	fmt.Println("  3. 用 errors.As 取出錯誤物件（需要存取欄位時）：")
	fmt.Println("     var valErr *ValidationError")
	fmt.Println("     if errors.As(err, &valErr) { ... }")
	fmt.Println()
	fmt.Println("  4. Sentinel Error（哨兵錯誤）用 var 定義在 package 層級：")
	fmt.Println("     var ErrNotFound = errors.New(\"not found\")")
	fmt.Println()
	fmt.Println("  5. 自訂 Error type 時，實作 Unwrap() 讓 errors.Is/As 可以穿透：")
	fmt.Println("     func (e *MyError) Unwrap() error { return e.Cause }")
	fmt.Println()
	fmt.Println("❌ 錯誤做法：")
	fmt.Println()
	fmt.Println(`  1. 用 %v 不用 %w（無法用 errors.Is 判斷）：`)
	fmt.Println(`     ❌ return fmt.Errorf("失敗: %v", err)`)
	fmt.Println(`     ✅ return fmt.Errorf("失敗: %w", err)`)
	fmt.Println()
	fmt.Println("  2. 吞掉錯誤（不處理也不記錄）：")
	fmt.Println("     ❌ result, _ := doSomething()  // _ 丟掉錯誤！")
	fmt.Println()
	fmt.Println("  3. 同一個錯誤記錄兩次（每層都 log，會產生重複日誌）：")
	fmt.Println("     ❌ 每一層都 log.Error(err)  // 最後只在最頂層記錄一次就夠")
}

func main() { // 程式進入點
	fmt.Println("==========================================")
	fmt.Println(" 第二十五課：Error Wrapping 完整示範")
	fmt.Println("==========================================")

	demonstrateBasicWrapping()  // 1. fmt.Errorf %w
	demonstrateCustomError()    // 2. 自訂錯誤類型
	demonstrateErrorsIsAs()     // 3. errors.Is vs errors.As
	demonstrateErrorFlow()      // 4. Clean Architecture 錯誤流向
	demonstrateBestPractices()  // 5. 最佳實踐

	fmt.Println("\n==========================================")
	fmt.Println(" 教學完成！")
	fmt.Println("==========================================")
}
