// 第七課：錯誤處理（Error Handling）
// Go 不使用 try/catch，而是用回傳值來處理錯誤
// 執行方式：go run main.go
package main

import (
	"errors"
	"fmt"
	"strconv"
)

// ========================================
// 1. error 介面
// ========================================

// Go 的 error 是一個內建介面：
//
// type error interface {
//     Error() string
// }
//
// 任何有 Error() string 方法的型別都是 error

// ========================================
// 2. 建立錯誤的幾種方式
// ========================================

// 方式 1：errors.New（最簡單）
var ErrNotFound = errors.New("找不到資源")

// 方式 2：fmt.Errorf（可以格式化訊息）
func findUser(id int) (*User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("無效的使用者 ID: %d", id)
	}
	if id > 100 {
		return nil, ErrNotFound
	}
	return &User{ID: id, Name: "Alice"}, nil
}

// ========================================
// 3. 自訂錯誤型別
// ========================================

// ValidationError 是自訂的錯誤型別
type ValidationError struct {
	Field   string // 哪個欄位出錯
	Message string // 錯誤訊息
}

// 實作 error 介面
func (e *ValidationError) Error() string {
	return fmt.Sprintf("驗證錯誤 [%s]: %s", e.Field, e.Message)
}

func validateAge(age int) error {
	if age < 0 {
		return &ValidationError{Field: "age", Message: "年齡不能為負數"}
	}
	if age > 150 {
		return &ValidationError{Field: "age", Message: "年齡不合理"}
	}
	return nil // nil 代表沒有錯誤
}

// ========================================
// 4. 錯誤包裝（Error Wrapping）
// ========================================

func getUserFromDB(id int) (*User, error) {
	// 模擬底層錯誤
	if id == 0 {
		return nil, errors.New("資料庫連線失敗")
	}
	return &User{ID: id, Name: "Bob"}, nil
}

func getUserProfile(id int) (*User, error) {
	user, err := getUserFromDB(id)
	if err != nil {
		// %w 動詞可以「包裝」原始錯誤，保留錯誤鏈
		return nil, fmt.Errorf("取得使用者資料失敗: %w", err)
	}
	return user, nil
}

// ========================================
// 5. panic 和 recover（很少直接使用）
// ========================================

func riskyOperation() {
	defer func() {
		// recover 只能在 defer 中使用
		if r := recover(); r != nil {
			fmt.Println("  已恢復 panic:", r)
		}
	}()

	fmt.Println("  執行危險操作...")
	panic("出大事了！") // 通常只在真的不可恢復的情況下使用
	// panic 之後的程式碼不會執行
}

// ========================================
// 輔助結構體
// ========================================

type User struct {
	ID   int
	Name string
}

func main() {
	// ========================================
	// 1. 基本錯誤處理模式
	// ========================================
	fmt.Println("=== 基本錯誤處理 ===")

	// Go 的標準模式：呼叫函式 → 檢查 err → 處理或繼續
	user, err := findUser(1)
	if err != nil {
		fmt.Println("錯誤:", err)
	} else {
		fmt.Println("找到使用者:", user.Name)
	}

	// 錯誤的情況
	_, err = findUser(-1)
	if err != nil {
		fmt.Println("錯誤:", err)
	}

	_, err = findUser(999)
	if err != nil {
		fmt.Println("錯誤:", err)
	}

	// ========================================
	// 2. errors.Is：判斷錯誤類型
	// ========================================
	fmt.Println("\n=== errors.Is ===")

	_, err = findUser(999)
	if errors.Is(err, ErrNotFound) {
		fmt.Println("確認是 NotFound 錯誤")
	}

	// ========================================
	// 3. 自訂錯誤型別
	// ========================================
	fmt.Println("\n=== 自訂錯誤型別 ===")

	err = validateAge(-5)
	if err != nil {
		fmt.Println("錯誤:", err)

		// errors.As：取出特定型別的錯誤資訊
		var valErr *ValidationError
		if errors.As(err, &valErr) {
			fmt.Println("  欄位:", valErr.Field)
			fmt.Println("  訊息:", valErr.Message)
		}
	}

	// ========================================
	// 4. 錯誤包裝與解包
	// ========================================
	fmt.Println("\n=== 錯誤包裝 ===")

	_, err = getUserProfile(0)
	if err != nil {
		fmt.Println("外層錯誤:", err)

		// Unwrap：取得被包裝的原始錯誤
		inner := errors.Unwrap(err)
		fmt.Println("內層錯誤:", inner)
	}

	// ========================================
	// 5. 實際範例：串聯多個可能出錯的操作
	// ========================================
	fmt.Println("\n=== 串聯錯誤處理 ===")

	result, err := processInput("42")
	if err != nil {
		fmt.Println("錯誤:", err)
	} else {
		fmt.Println("結果:", result)
	}

	result, err = processInput("abc")
	if err != nil {
		fmt.Println("錯誤:", err)
	}

	result, err = processInput("-5")
	if err != nil {
		fmt.Println("錯誤:", err)
	}

	// ========================================
	// 6. panic 和 recover
	// ========================================
	fmt.Println("\n=== panic 和 recover ===")

	riskyOperation()
	fmt.Println("  程式繼續執行（panic 已被 recover）")

	// ========================================
	// 7. 錯誤處理的最佳實踐
	// ========================================
	fmt.Println("\n=== 最佳實踐 ===")
	fmt.Println("1. 永遠檢查 error，不要用 _ 忽略")
	fmt.Println("2. 用 fmt.Errorf + %w 包裝錯誤，保留上下文")
	fmt.Println("3. 在呼叫端處理錯誤，不要在底層隱藏錯誤")
	fmt.Println("4. 只在真正不可恢復的情況下使用 panic")
	fmt.Println("5. 用 errors.Is/errors.As 判斷錯誤類型")
}

// processInput 示範串聯多個錯誤檢查
func processInput(input string) (int, error) {
	// 步驟 1：把字串轉為數字
	num, err := strconv.Atoi(input)
	if err != nil {
		return 0, fmt.Errorf("無法解析輸入 '%s': %w", input, err)
	}

	// 步驟 2：驗證數字範圍
	if num < 0 {
		return 0, fmt.Errorf("數字必須為正數，收到: %d", num)
	}

	// 步驟 3：處理
	return num * 2, nil
}
