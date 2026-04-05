// ============================================================
// 第七課：錯誤處理（Error Handling）
// ============================================================
// Go 的錯誤處理哲學：「錯誤是值」（Errors are values）
//
// 大多數語言用 try/catch 來處理錯誤（像是在程式碼上裝了一個安全網）。
// Go 不一樣 —— Go 把錯誤當作「普通的值」，用多重回傳值來傳遞。
//
// 想像你在餐廳點餐：
//   - 「抱歉，牛排賣完了」→ 這是 error（正常的錯誤，可以處理）
//   - 「廚房著火了！」→ 這是 panic（嚴重的意外，無法正常處理）
//
// Go 的做法：
//   result, err := doSomething()  // 函式回傳「結果」和「錯誤」兩個值
//   if err != nil {                // 檢查是否有錯誤
//       // 處理錯誤
//   }
//
// 你會學到：
//   1. error 介面的本質
//   2. 建立錯誤的三種方式
//   3. 哨兵錯誤（Sentinel Error）模式：var ErrXxx = errors.New(...)
//   4. 自訂錯誤型別：攜帶更多資訊
//   5. 錯誤包裝（%w）與解包（Unwrap）
//   6. errors.Is vs errors.As：判斷錯誤的兩種方式
//   7. panic 和 recover：最後的手段
//   8. 錯誤邊界：在哪裡處理 vs 在哪裡傳遞
//
// 執行方式：go run ./tutorials/07-error-handling
// ============================================================

package main // 可執行程式的套件名稱

import (         // 匯入需要的標準函式庫
	"errors"  // errors 套件：提供錯誤相關的工具函式
	"fmt"     // fmt 套件：格式化輸出
	"strconv" // strconv 套件：字串和其他型別的轉換
)

// ========================================
// 1. error 介面的本質
// ========================================
// Go 的 error 是一個內建的介面（第六課學過介面）：
//
//   type error interface {
//       Error() string    // 只有一個方法：回傳錯誤訊息字串
//   }
//
// 任何型別只要有 Error() string 方法，就是一個 error！
// 這就是介面的威力 —— 你可以自己定義各種錯誤型別。

// ========================================
// 2. 建立錯誤的三種方式
// ========================================

// --- 方式 1：哨兵錯誤（Sentinel Error）---
// 用 errors.New() 建立一個「固定的」錯誤值
// 命名慣例：以 Err 開頭（ErrNotFound, ErrUnauthorized, ErrInvalidInput...）
// 用途：當你想讓呼叫方用 errors.Is() 來判斷「是不是這個特定的錯誤」
var ErrNotFound = errors.New("找不到資源")       // 哨兵錯誤：資源不存在
var ErrInvalidID = errors.New("無效的 ID")       // 哨兵錯誤：ID 不合法

// --- 方式 2：fmt.Errorf() — 帶有動態訊息的錯誤 ---
// 當錯誤訊息需要包含變數（例如 ID、名稱）時使用

// findUser 根據 ID 查找使用者，示範不同的錯誤回傳方式
func findUser(id int) (*User, error) { // 回傳使用者指標和錯誤（Go 的標準模式）
	if id <= 0 { // 如果 ID 不合法
		return nil, ErrInvalidID // 回傳哨兵錯誤（呼叫方可以用 errors.Is 判斷）
	}
	if id > 100 { // 如果 ID 超出範圍（模擬找不到）
		return nil, ErrNotFound // 回傳哨兵錯誤
	}
	return &User{ID: id, Name: "Alice"}, nil // 找到了！回傳使用者，錯誤為 nil
}

// --- 方式 3：自訂錯誤型別 — 攜帶更多結構化資訊 ---
// 當你需要在錯誤中攜帶「額外資訊」（哪個欄位出錯、錯誤代碼等）時使用

// ========================================
// 3. 哨兵錯誤（Sentinel Error）模式
// ========================================
// 什麼是「哨兵」？就像站崗的衛兵，是一個「已知的、固定的」錯誤值。
//
// 命名慣例：
//   var ErrNotFound    = errors.New("...")  // Err + 描述
//   var ErrUnauthorized = errors.New("...")
//   var ErrTimeout     = errors.New("...")
//
// 什麼時候用哨兵錯誤？
//   - 錯誤是「可預期的」且「全域統一的」（例如「找不到」、「未授權」）
//   - 呼叫方需要判斷「是不是這個特定的錯誤」來做不同處理
//
// 什麼時候不用？
//   - 錯誤需要攜帶動態資訊（用 fmt.Errorf 或自訂錯誤型別）
//   - 錯誤只在一個函式內使用（直接回傳即可）

// ========================================
// 4. 自訂錯誤型別
// ========================================
// 當哨兵錯誤的資訊不夠用時，可以定義自己的錯誤型別。
// 只要實作 Error() string 方法，就是一個合法的 error。

// ValidationError 是自訂的驗證錯誤型別
type ValidationError struct { // 自訂錯誤結構體
	Field   string // 哪個欄位出錯（例如 "age"、"email"）
	Message string // 具體的錯誤訊息
}

// Error 實作 error 介面（讓 ValidationError 成為一個合法的 error）
func (e *ValidationError) Error() string { // 指標接收者，實作 error 介面
	return fmt.Sprintf("驗證錯誤 [%s]: %s", e.Field, e.Message) // 格式化錯誤訊息
}

// validateAge 驗證年齡是否合理，回傳自訂錯誤型別
func validateAge(age int) error { // 回傳值型別是 error 介面
	if age < 0 { // 年齡不能是負數
		return &ValidationError{Field: "age", Message: "年齡不能為負數"} // 回傳自訂錯誤
	}
	if age > 150 { // 年齡不合理
		return &ValidationError{Field: "age", Message: "年齡不合理"} // 回傳自訂錯誤
	}
	return nil // nil 代表「沒有錯誤」，一切正常
}

// validateEmail 驗證 email 是否有效
func validateEmail(email string) error { // 回傳 error 介面
	if len(email) == 0 { // 檢查是否為空字串
		return &ValidationError{Field: "email", Message: "電子郵件不能為空"} // 回傳自訂錯誤
	}
	return nil // 驗證通過
}

// ========================================
// 5. 錯誤包裝（Error Wrapping）— 保留錯誤鏈
// ========================================
// 當函式 A 呼叫函式 B，B 出錯時，A 應該：
// 1. 加上自己的上下文（「我在做什麼時出錯了」）
// 2. 保留原始錯誤（讓上層還能用 errors.Is 判斷）
//
// 用 fmt.Errorf 搭配 %w 動詞來包裝錯誤：
//   return fmt.Errorf("取得使用者失敗: %w", originalErr)
//                                       ^^
//                                   %w = wrap（包裝）

// getUserFromDB 模擬從資料庫取得使用者（底層函式）
func getUserFromDB(id int) (*User, error) { // 模擬資料庫操作
	if id == 0 { // 模擬資料庫連線失敗
		return nil, errors.New("資料庫連線失敗") // 回傳底層錯誤
	}
	if id == 999 { // 模擬找不到使用者
		return nil, ErrNotFound // 回傳哨兵錯誤
	}
	return &User{ID: id, Name: "Bob"}, nil // 成功回傳使用者
}

// getUserProfile 是中間層函式，呼叫底層並包裝錯誤
func getUserProfile(id int) (*User, error) { // 中間層函式
	user, err := getUserFromDB(id) // 呼叫底層資料庫函式
	if err != nil {                // 如果底層出錯
		// 用 %w 包裝原始錯誤：加上上下文，但保留原始錯誤供 errors.Is 使用
		return nil, fmt.Errorf("取得使用者資料失敗 (id=%d): %w", id, err)
	}
	return user, nil // 成功，直接回傳
}

// ========================================
// 6. 錯誤邊界：在哪裡處理 vs 在哪裡傳遞
// ========================================
// 原則：
//   - 底層函式：回傳錯誤，讓上層決定怎麼處理
//   - 中間層：包裝錯誤（加上下文），繼續向上傳遞
//   - 最上層（main / handler）：真正處理錯誤（印出、回傳 HTTP 狀態碼等）
//
// 類比：
//   員工發現問題 → 回報主管（加上「我在做 X 時發現的」）→ 主管回報經理 → 經理做決定

// ========================================
// 7. panic 和 recover（最後的手段）
// ========================================
// panic = 程式崩潰（類似 Java 的 RuntimeException）
// recover = 攔截 panic，防止程式崩潰（類似 catch）
//
// 重要：99% 的情況應該用 error，不是 panic！
// panic 只在「真的不可恢復」的情況下使用：
//   - 程式初始化失敗（例如設定檔讀不到）
//   - 程式邏輯錯誤（應該永遠不會發生的情況）

// riskyOperation 示範 panic 和 recover 的用法
func riskyOperation() { // 模擬一個可能 panic 的操作
	defer func() { // defer 確保這個匿名函式在 riskyOperation 結束前執行
		if r := recover(); r != nil { // recover() 攔截 panic，r 是 panic 的值
			fmt.Println("  已恢復 panic:", r) // 印出被攔截的 panic 訊息
		}
	}() // 立即呼叫這個匿名函式（注意結尾的 ()）

	fmt.Println("  執行危險操作...") // 這行會正常執行
	panic("出大事了！")             // 觸發 panic —— 通常只在不可恢復的情況下使用
	// panic 之後的程式碼不會執行    // 這行永遠不會執行到
}

// ========================================
// 輔助結構體
// ========================================

// User 是使用者結構體（模擬部落格專案的 domain.User）
type User struct { // 簡化的使用者結構體
	ID   int    // 使用者 ID
	Name string // 使用者名稱
}

// ========================================
// 主程式入口
// ========================================

func main() { // 程式的入口點

	// ========================================
	// 1. 基本錯誤處理模式
	// ========================================
	fmt.Println("=== 1. 基本錯誤處理模式 ===") // 印出區塊標題
	fmt.Println("Go 的核心模式：呼叫函式 → 檢查 err → 處理或繼續") // 說明

	// 成功的情況
	user, err := findUser(1) // 呼叫 findUser，取得兩個回傳值
	if err != nil {          // 永遠先檢查 err 是否為 nil
		fmt.Println("  錯誤:", err) // 有錯誤就處理
	} else {                       // 沒錯誤就使用結果
		fmt.Println("  找到使用者:", user.Name) // 印出使用者名稱
	}

	// 失敗的情況 1：無效 ID
	_, err = findUser(-1)    // 用 _ 忽略不需要的回傳值（因為出錯時 user 是 nil）
	if err != nil {          // 檢查錯誤
		fmt.Println("  錯誤:", err) // 印出：無效的 ID
	}

	// 失敗的情況 2：找不到
	_, err = findUser(999)   // ID 超出範圍
	if err != nil {          // 檢查錯誤
		fmt.Println("  錯誤:", err) // 印出：找不到資源
	}

	// ========================================
	// 2. errors.Is — 判斷是不是「某個特定的錯誤值」
	// ========================================
	fmt.Println("\n=== 2. errors.Is（比較錯誤值）===") // 印出區塊標題

	_, err = findUser(999) // 這會回傳 ErrNotFound
	if errors.Is(err, ErrNotFound) { // errors.Is 檢查 err 是否為（或包裝了）ErrNotFound
		fmt.Println("  確認是 NotFound 錯誤 → 可以回傳 HTTP 404") // 根據錯誤類型做不同處理
	}

	_, err = findUser(-1) // 這會回傳 ErrInvalidID
	if errors.Is(err, ErrInvalidID) { // 檢查是否為 ErrInvalidID
		fmt.Println("  確認是 InvalidID 錯誤 → 可以回傳 HTTP 400") // 做對應的處理
	}

	// errors.Is 的超能力：即使錯誤被包裝過，也能找到原始錯誤！
	_, err = getUserProfile(999) // 這會回傳「包裝過的」ErrNotFound
	if errors.Is(err, ErrNotFound) { // 即使被包裝了，errors.Is 也能穿透找到 ErrNotFound
		fmt.Println("  包裝過的錯誤，errors.Is 也能找到原始的 ErrNotFound！") // 成功！
	}

	// ========================================
	// 3. 自訂錯誤型別 + errors.As
	// ========================================
	fmt.Println("\n=== 3. 自訂錯誤型別 + errors.As（比較錯誤型別）===") // 印出區塊標題

	err = validateAge(-5) // 驗證年齡，會回傳 *ValidationError
	if err != nil {       // 有錯誤
		fmt.Println("  錯誤:", err) // 印出錯誤訊息（會呼叫 Error() 方法）

		// errors.As：從錯誤鏈中取出特定「型別」的錯誤
		// errors.Is 比較「值」，errors.As 比較「型別」
		var valErr *ValidationError                // 先宣告目標型別的變數
		if errors.As(err, &valErr) {               // 嘗試從 err 中取出 *ValidationError
			fmt.Println("  出錯欄位:", valErr.Field)   // 取出額外資訊：哪個欄位
			fmt.Println("  錯誤訊息:", valErr.Message) // 取出額外資訊：具體訊息
		}
	}

	// 再試一個
	err = validateEmail("") // 驗證空的 email
	if err != nil {         // 有錯誤
		var valErr *ValidationError            // 宣告目標型別變數
		if errors.As(err, &valErr) {           // 從 err 中取出 *ValidationError
			fmt.Println("  出錯欄位:", valErr.Field, "→", valErr.Message) // 印出詳細資訊
		}
	}

	// ========================================
	// 4. 錯誤包裝（%w）與解包（Unwrap）
	// ========================================
	fmt.Println("\n=== 4. 錯誤包裝與解包 ===") // 印出區塊標題

	_, err = getUserProfile(0) // 傳入 0 會觸發「資料庫連線失敗」
	if err != nil {            // 有錯誤
		fmt.Println("  外層錯誤:", err) // 印出包裝後的完整錯誤訊息

		// errors.Unwrap：剝開一層包裝，取得被包裝的原始錯誤
		inner := errors.Unwrap(err)    // 取得內層錯誤
		fmt.Println("  內層錯誤:", inner) // 印出原始的「資料庫連線失敗」
	}

	// 示範 errors.Is 穿透包裝的能力
	fmt.Println("\n  errors.Is 穿透包裝:") // 小標題
	_, err = getUserProfile(999)           // 會回傳包裝過的 ErrNotFound
	if err != nil {                        // 有錯誤
		fmt.Println("  完整錯誤:", err)                                       // 印出包裝後的錯誤
		fmt.Println("  errors.Is(err, ErrNotFound):", errors.Is(err, ErrNotFound)) // true！穿透包裝
	}

	// ========================================
	// 5. 串聯多個可能出錯的操作
	// ========================================
	fmt.Println("\n=== 5. 串聯錯誤處理 ===") // 印出區塊標題
	fmt.Println("模式：每一步都檢查 err，出錯就提早返回") // 說明

	// 成功案例
	result, err := processInput("42") // 傳入合法的數字字串
	if err != nil {                   // 檢查錯誤
		fmt.Println("  錯誤:", err)  // 有錯誤就印出
	} else {
		fmt.Println("  結果:", result) // 成功印出結果
	}

	// 失敗案例 1：無法解析的字串
	_, err = processInput("abc") // "abc" 無法轉成數字
	if err != nil {              // 一定有錯誤
		fmt.Println("  錯誤:", err) // 印出：無法解析輸入
	}

	// 失敗案例 2：數字不合法
	_, err = processInput("-5") // 負數不合法
	if err != nil {             // 有錯誤
		fmt.Println("  錯誤:", err) // 印出：數字必須為正數
	}

	// ========================================
	// 6. panic 和 recover
	// ========================================
	fmt.Println("\n=== 6. panic 和 recover ===") // 印出區塊標題
	fmt.Println("panic = 嚴重錯誤（廚房著火），recover = 消防員") // 類比說明

	riskyOperation()                                       // 這個函式內部會 panic，但被 recover 攔截
	fmt.Println("  程式繼續執行（panic 已被 recover 攔截）") // 程式沒有崩潰！

	// ========================================
	// 7. panic vs error 決策指南
	// ========================================
	fmt.Println("\n=== 7. 何時用 error，何時用 panic？===") // 印出區塊標題
	fmt.Println("  ┌──────────────┬─────────────────────────────────┐") // 表格頂部
	fmt.Println("  │ 情境         │ 使用方式                         │") // 表頭
	fmt.Println("  ├──────────────┼─────────────────────────────────┤") // 分隔線
	fmt.Println("  │ 檔案不存在   │ error（可預期，呼叫方能處理）     │") // 可預期錯誤
	fmt.Println("  │ 網路逾時     │ error（常見情況，需要重試邏輯）   │") // 可預期錯誤
	fmt.Println("  │ 使用者輸入錯 │ error（驗證失敗是正常的）         │") // 可預期錯誤
	fmt.Println("  │ 設定檔壞了   │ panic（程式根本無法啟動）         │") // 不可恢復
	fmt.Println("  │ 邏輯 bug     │ panic（不應該發生的情況）         │") // 不可恢復
	fmt.Println("  └──────────────┴─────────────────────────────────┘") // 表格底部

	// ========================================
	// 8. 最佳實踐總結
	// ========================================
	fmt.Println("\n=== 8. 最佳實踐 ===")                               // 印出區塊標題
	fmt.Println("  1. 永遠檢查 error，不要用 _ 忽略它")                  // 規則 1
	fmt.Println("  2. 用 fmt.Errorf + %w 包裝錯誤，保留上下文和錯誤鏈") // 規則 2
	fmt.Println("  3. 底層回傳錯誤，上層處理錯誤（錯誤邊界）")            // 規則 3
	fmt.Println("  4. 只在真正不可恢復的情況下使用 panic")                // 規則 4
	fmt.Println("  5. 用 errors.Is 比較值，errors.As 比較型別")          // 規則 5
	fmt.Println("  6. 哨兵錯誤用 ErrXxx 命名，方便呼叫方判斷")           // 規則 6

	// ========================================
	// 結語
	// ========================================
	fmt.Println("\n=== 學習完成！ ===")                                                // 結束訊息
	fmt.Println("Go 的錯誤處理看起來很囉唆（到處都是 if err != nil），")                  // 感想
	fmt.Println("但這種「顯式處理」讓你永遠知道程式在哪裡可能出錯、出了什麼錯。")          // 優點
	fmt.Println("下一課：套件與模組（Packages & Modules）— 如何組織和分享你的程式碼。")   // 預告
}

// ========================================
// 輔助函式
// ========================================

// processInput 示範串聯多個可能出錯的步驟
// 每一步失敗都提早返回，不繼續往下走
func processInput(input string) (int, error) { // 接收字串，回傳處理後的數字和錯誤
	// 步驟 1：把字串轉為數字
	num, err := strconv.Atoi(input) // Atoi = ASCII to Integer
	if err != nil {                 // 如果轉換失敗（例如 "abc" 無法轉成數字）
		return 0, fmt.Errorf("無法解析輸入 '%s': %w", input, err) // 包裝錯誤並回傳
	}

	// 步驟 2：驗證數字範圍（必須是正數）
	if num < 0 { // 如果是負數
		return 0, fmt.Errorf("數字必須為正數，收到: %d", num) // 回傳錯誤
	}

	// 步驟 3：處理（所有檢查都通過了！）
	return num * 2, nil // 回傳結果，錯誤為 nil
}
