// 第五課：指標（Pointers）
// 執行方式：go run main.go
package main

import "fmt"

// ========================================
// 1. 指標的基本概念
// ========================================

// 指標是一個變數，它儲存的是另一個變數的「記憶體位址」
// 可以想像成：變數是房子，指標是房子的地址

func main() {
	fmt.Println("=== 指標基礎 ===")

	x := 42

	// & 取址運算子：取得變數的記憶體位址
	p := &x // p 是指向 x 的指標，型別是 *int

	fmt.Println("x 的值:", x)
	fmt.Println("x 的位址:", p)     // 印出像 0xc0000b4008 的位址
	fmt.Printf("p 的型別: %T\n", p) // *int

	// * 解引用運算子：透過指標存取原始值
	fmt.Println("*p 的值:", *p) // 42（透過指標讀取 x 的值）

	// 透過指標修改原始值
	*p = 100
	fmt.Println("修改後 x:", x) // 100（x 被改變了！）

	// ========================================
	// 2. 為什麼需要指標？—— 傳值 vs 傳參考
	// ========================================
	fmt.Println("\n=== 傳值 vs 傳參考 ===")

	value := 10

	// Go 預設是「傳值」（Pass by Value）
	// 函式收到的是副本，修改不影響原始值
	doubleByValue(value)
	fmt.Println("傳值後:", value) // 仍然是 10

	// 傳入指標就能修改原始值
	doubleByPointer(&value)
	fmt.Println("傳指標後:", value) // 變成 20

	// ========================================
	// 3. 結構體與指標
	// ========================================
	fmt.Println("\n=== 結構體與指標 ===")

	type User struct {
		Name string
		Age  int
	}

	// 建立結構體的值
	user := User{Name: "Alice", Age: 25}

	// 建立結構體的指標
	userPtr := &user

	// Go 的語法糖：用指標存取結構體欄位不需要寫 (*userPtr).Name
	fmt.Println("用指標存取:", userPtr.Name) // 等同於 (*userPtr).Name
	fmt.Println("直接存取:", user.Name)

	// 修改
	userPtr.Age = 26
	fmt.Println("修改後年齡:", user.Age) // 26

	// ========================================
	// 4. new() 函式
	// ========================================
	fmt.Println("\n=== new() 函式 ===")

	// new() 分配記憶體並回傳指標，值為零值
	intPtr := new(int) // *int，值為 0
	fmt.Println("new(int):", *intPtr)

	*intPtr = 42
	fmt.Println("賦值後:", *intPtr)

	// 等同於：
	// var i int
	// intPtr := &i

	// ========================================
	// 5. nil 指標
	// ========================================
	fmt.Println("\n=== nil 指標 ===")

	var ptr *int // 未初始化的指標，值為 nil
	fmt.Println("nil 指標:", ptr)

	// 使用前必須檢查是否為 nil
	if ptr != nil {
		fmt.Println("值:", *ptr)
	} else {
		fmt.Println("指標是 nil，不能解引用")
	}

	// *ptr = 10  // ← 這會造成 panic！不能對 nil 指標解引用

	// ========================================
	// 6. 實際範例：為什麼 Repository 回傳 *User 而非 User
	// ========================================
	fmt.Println("\n=== 實際應用 ===")

	// 模擬 Repository 的 FindByID
	foundUser := findUserByID(1)
	if foundUser != nil {
		fmt.Println("找到使用者:", foundUser.Name)
	}

	notFound := findUserByID(999)
	if notFound == nil {
		fmt.Println("使用者不存在")
	}

	// 如果回傳的是值（User 而非 *User），
	// 即使找不到也會回傳一個空的 User{}，
	// 很難判斷到底是「找到了但欄位都是空的」還是「沒找到」
	// 所以用指標可以回傳 nil 代表「不存在」

	// ========================================
	// 7. 切片和 Map 是參考型別
	// ========================================
	fmt.Println("\n=== 參考型別 ===")

	// 切片（Slice）和 Map 內部已經包含指標
	// 傳給函式時不需要額外用 &

	nums := []int{1, 2, 3}
	doubleSlice(nums)
	fmt.Println("切片修改後:", nums) // [2, 4, 6]（被修改了）

	// 這就是為什麼 GORM 的方法接收 *User 而非 User：
	// db.Create(&user)  ← 傳入指標，GORM 會設定 user.ID
}

// doubleByValue 接收值的副本
func doubleByValue(n int) {
	n = n * 2 // 只修改了副本
}

// doubleByPointer 接收指標，能修改原始值
func doubleByPointer(n *int) {
	*n = *n * 2 // 修改原始值
}

type User struct {
	Name string
	Age  int
}

// findUserByID 模擬資料庫查詢
// 回傳 *User（指標），找不到時回傳 nil
func findUserByID(id int) *User {
	// 模擬資料庫
	users := map[int]User{
		1: {Name: "Alice", Age: 25},
		2: {Name: "Bob", Age: 30},
	}

	user, exists := users[id]
	if !exists {
		return nil // 找不到就回傳 nil
	}
	return &user // 回傳指標
}

// doubleSlice 修改切片中的值
func doubleSlice(nums []int) {
	for i := range nums {
		nums[i] *= 2
	}
}
