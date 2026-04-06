// ============================================================
// 第五課：指標（Pointers）
// ============================================================
// 指標是什麼？想像你有一棟房子（變數），指標就是那棟房子的「地址」。
// 你可以透過地址找到房子，也可以透過地址去修改房子裡的東西。
//
// 你會學到：
//   1. & 取址運算子 和 * 解引用運算子
//   2. 傳值（pass by value）vs 傳指標（pass by pointer）
//   3. 結構體指標和 Go 的語法糖
//   4. new() 函式
//   5. nil 指標和安全檢查
//   6. 為什麼 Repository 回傳 *User 而不是 User
//   7. 【重點】切片（Slice）和 Map 的內部指標機制
//   8. 什麼時候「不要」用指標
//
// 執行方式：go run ./tutorials/05-pointers
// ============================================================

package main // 可執行程式的套件名稱

import "fmt" // 印東西到螢幕上

// ========================================
// 定義結構體（放在 package 層級，整個檔案都能用）
// ========================================

// User 代表一個使用者，模擬部落格系統的 User
type User struct {
	ID   int    // 使用者 ID
	Name string // 使用者名稱
	Age  int    // 使用者年齡
}

// ========================================
// 2. 傳值 vs 傳指標的輔助函式
// ========================================

// doubleByValue 接收「值的副本」—— 修改不會影響外面的變數
func doubleByValue(n int) { // n 是原始值的副本
	n = n * 2 // 只修改了副本，外面的變數不受影響
}

// doubleByPointer 接收「指標」—— 透過指標能修改原始值
func doubleByPointer(n *int) { // n 是指向原始變數的指標，型別是 *int
	*n = *n * 2 // *n 解引用：取出指標指向的值，乘以 2 後寫回去
}

// ========================================
// 3. 結構體指標的輔助函式
// ========================================

// celebrateBirthday 用指標接收者讓函式能修改原始的 User
func celebrateBirthday(u *User) { // u 是指向 User 的指標
	u.Age++ // Go 的語法糖：不需要寫 (*u).Age++，Go 會自動解引用
}

// ========================================
// 6. Repository 模式的輔助函式
// ========================================

// findUserByID 模擬資料庫查詢
// 回傳 *User（指標）而不是 User（值），原因：
//   - 找到了 → 回傳 &user（指向實際資料的指標）
//   - 找不到 → 回傳 nil（「不存在」的意思）
//   - 如果回傳的是 User（值），找不到時只能回傳空的 User{}，
//     無法區分「找到了但欄位都是空的」和「根本沒找到」
func findUserByID(id int) *User { // 回傳型別是 *User（指向 User 的指標）
	// 模擬資料庫中的資料
	users := map[int]User{ // 用 map 模擬資料庫，key 是 ID
		1: {ID: 1, Name: "Alice", Age: 25}, // ID 為 1 的使用者
		2: {ID: 2, Name: "Bob", Age: 30},   // ID 為 2 的使用者
	}

	user, exists := users[id] // 從 map 查詢，exists 是布林值
	if !exists {              // 如果 ID 不存在
		return nil // 回傳 nil 代表「找不到」
	}
	return &user // 回傳指向 user 的指標（& 取得位址）
}

// findUserByIDValue 用「回傳值」的方式——不推薦
// 回傳 (User, bool) 雖然也能表示「找不到」，但不如 *User 直覺
func findUserByIDValue(id int) (User, bool) { // 回傳值和是否找到
	users := map[int]User{ // 用 map 模擬資料庫
		1: {ID: 1, Name: "Alice", Age: 25}, // 測試資料
	}

	user, exists := users[id] // 從 map 查詢
	return user, exists       // 回傳查詢結果和是否存在
}

// ========================================
// 7. 切片和 Map 內部指標的輔助函式
// ========================================

// modifySliceElements 修改切片「內部」的元素
// 注意：切片傳入函式時，底層陣列的指標會被複製
// 所以函式內修改元素，外面也會看到變化
func modifySliceElements(nums []int) { // nums 是切片（不是指標，但內部有指標）
	for i := range nums { // 用 range 遍歷所有索引
		nums[i] *= 2 // 直接修改元素，外面也會被改變！
	}
}

// tryAppendToSlice 嘗試在函式內 append
// 重點：append 可能建立新的底層陣列，所以外面的切片不會看到新元素！
func tryAppendToSlice(nums []int) { // 接收切片（複製了 header，不是原始 header）
	nums = append(nums, 999) // append 可能建立新的底層陣列
	// 即使沒有建立新陣列，外面的 length 也沒更新
	fmt.Println("  函式內 append 後:", nums) // 函式內看得到 999
}

// appendToSlice 正確的做法：回傳新的切片
func appendToSlice(nums []int) []int { // 接收切片，回傳新切片
	nums = append(nums, 999) // append 可能建立新底層陣列，也可能用原來的
	return nums              // 把新的切片 header 回傳出去
}

// modifyMap 修改 map 的值
// map 和切片類似，內部已經包含指標
// 傳入函式時，函式能直接修改原始的 map
func modifyMap(m map[string]int) { // m 是 map（內部已包含指標）
	m["new_key"] = 42 // 直接修改，外面也看得到
}

func main() { // 程式進入點

	// ========================================
	// 1. 指標的基本概念：& 和 *
	// ========================================
	fmt.Println("=== 1. 指標基礎：& 和 * ===") // 章節標題

	x := 42 // 宣告一個 int 變數 x，值是 42（這是「房子」）

	p := &x // & 取址運算子：取得 x 的記憶體位址（這是「房子的地址」）
	// p 的型別是 *int（指向 int 的指標）

	fmt.Println("x 的值:", x)      // 印出 42（直接看房子裡的東西）
	fmt.Println("x 的位址:", p)     // 印出類似 0xc0000b4008 的位址
	fmt.Printf("p 的型別: %T\n", p) // *int（指向 int 的指標型別）

	fmt.Println("*p 的值:", *p) // * 解引用：透過地址找到房子，看裡面的值（42）

	*p = 100                    // 透過指標修改原始值（透過地址改了房子裡的東西）
	fmt.Println("修改後 x 的值:", x) // 100（x 真的被改變了！）

	// 再看一個例子：兩個指標指向同一個變數
	p2 := &x                         // p2 也指向 x（兩張名片都寫著同一個地址）
	fmt.Println("p == p2:", p == p2) // true（兩個指標指向同一個位址）
	fmt.Println("*p2:", *p2)         // 100（透過 p2 也能看到 x 的值）

	// ========================================
	// 2. 傳值（Pass by Value）vs 傳指標（Pass by Pointer）
	// ========================================
	fmt.Println("\n=== 2. 傳值 vs 傳指標 ===") // 章節標題

	value := 10 // 宣告一個變數 value，值是 10

	// Go 預設「傳值」：函式收到的是副本
	doubleByValue(value)             // 傳入 value 的「副本」
	fmt.Println("傳值後 value:", value) // 仍然是 10（副本被修改，原始值不受影響）

	// 傳入指標：函式可以修改原始值
	doubleByPointer(&value)           // & 取址：傳入 value 的「地址」
	fmt.Println("傳指標後 value:", value) // 變成 20（透過指標修改了原始值）

	// ========================================
	// 3. 結構體指標（Struct Pointers）
	// ========================================
	fmt.Println("\n=== 3. 結構體指標 ===") // 章節標題

	// 建立一個 User 結構體（用的是 package 層級定義的 User）
	user := User{ID: 1, Name: "Alice", Age: 25} // 建立 User 值

	userPtr := &user // & 取得 user 的指標

	// Go 的語法糖：指標存取結構體欄位不需要寫 (*userPtr).Name
	fmt.Println("指標存取:", userPtr.Name) // Go 自動把 userPtr.Name 轉成 (*userPtr).Name
	fmt.Println("直接存取:", user.Name)    // 透過值直接存取

	// 用函式修改結構體
	fmt.Println("生日前年齡:", user.Age) // 25
	celebrateBirthday(&user)        // 傳入指標，函式能修改原始的 user
	fmt.Println("生日後年齡:", user.Age) // 26（被修改了！）

	// ========================================
	// 4. new() 函式
	// ========================================
	fmt.Println("\n=== 4. new() 函式 ===") // 章節標題

	// new(T) 做兩件事：
	//   1. 分配一塊 T 型別的記憶體
	//   2. 把這塊記憶體初始化為零值
	//   3. 回傳指向這塊記憶體的指標 *T

	intPtr := new(int)                   // 分配一個 int，初始化為 0，回傳 *int
	fmt.Println("new(int) 的值:", *intPtr) // 0（int 的零值）

	*intPtr = 42                 // 透過指標賦值
	fmt.Println("賦值後:", *intPtr) // 42

	userPtr2 := new(User)                // 分配一個 User，所有欄位都是零值
	fmt.Println("new(User):", *userPtr2) // {0  0}（ID=0, Name="", Age=0）

	// new() 等同於這種寫法：
	// var u User       // 宣告零值的 User
	// userPtr2 := &u   // 取得它的指標

	// ========================================
	// 5. nil 指標——一定要檢查！
	// ========================================
	fmt.Println("\n=== 5. nil 指標 ===") // 章節標題

	var nilPtr *User               // 宣告一個指標但沒有初始化，值是 nil
	fmt.Println("nil 指標:", nilPtr) // <nil>（不指向任何東西）

	// ⚠️ 使用指標前一定要檢查 nil！
	if nilPtr != nil { // 檢查是否為 nil
		fmt.Println("使用者:", nilPtr.Name) // 安全：只有不是 nil 才存取
	} else { // 如果是 nil
		fmt.Println("指標是 nil，不能使用！") // 提醒開發者
	}

	// 如果不檢查直接用：
	// fmt.Println(nilPtr.Name)  // ← panic: runtime error: invalid memory address
	// 這是 Go 最常見的 runtime 錯誤之一！

	// ========================================
	// 6. 為什麼 Repository 回傳 *User 而不是 User
	// ========================================
	fmt.Println("\n=== 6. Repository 為什麼回傳 *User ===") // 章節標題

	// 方式 A：回傳 *User（推薦）—— 找不到時回傳 nil
	foundUser := findUserByID(1) // 查詢 ID 為 1 的使用者
	if foundUser != nil {        // 檢查是否找到
		fmt.Println("找到使用者:", foundUser.Name) // Alice
	}

	notFound := findUserByID(999) // 查詢不存在的 ID
	if notFound == nil {          // nil 代表「不存在」，語意非常清楚
		fmt.Println("使用者 999 不存在") // 清楚知道是「沒找到」
	}

	// 方式 B：回傳 (User, bool) —— 也可以，但不如 *User 直覺
	userVal, exists := findUserByIDValue(1) // 回傳值和布林值
	if exists {                             // 用 bool 判斷是否存在
		fmt.Println("方式 B 找到:", userVal.Name) // Alice
	}

	// 為什麼 *User 比較好？
	// 1. nil 的語意比 bool 更直覺：nil = 不存在
	// 2. 避免複製整個結構體（大結構體效能更好）
	// 3. 和 GORM 等 ORM 的慣例一致
	// 4. 可以讓 GORM 透過指標回寫 ID 等欄位（db.Create(&user)）

	// ========================================
	// 7.【重點】切片（Slice）和 Map 的內部指標機制
	// ========================================
	fmt.Println("\n=== 7. 切片和 Map 的內部指標 ===") // 章節標題

	// ----------------------------------------
	// 7a. 切片（Slice）的真實結構
	// ----------------------------------------
	fmt.Println("\n--- 7a. 切片的真實結構 ---") // 小節標題

	// 切片不是陣列！切片其實是一個「結構體」，包含三個欄位：
	//
	//   type slice struct {
	//       ptr *array  // 指向底層陣列的指標
	//       len int     // 目前的長度（有多少元素）
	//       cap int     // 容量（底層陣列能放多少元素）
	//   }
	//
	// 所以當你把切片傳給函式時：
	//   - 切片 header（ptr, len, cap）會被「複製」
	//   - 但 ptr 指向的底層陣列是「共享的」
	//   - 這就是為什麼函式內修改元素，外面也看得到

	nums := []int{1, 2, 3, 4, 5} // 建立一個切片（底層是一個長度 5 的陣列）

	fmt.Println("原始切片:", nums)        // [1 2 3 4 5]
	fmt.Println("長度 len:", len(nums)) // 5（目前有 5 個元素）
	fmt.Println("容量 cap:", cap(nums)) // 5（底層陣列能放 5 個元素）

	// ----------------------------------------
	// 7b. 切片傳入函式：修改元素 ✅ 有效
	// ----------------------------------------
	fmt.Println("\n--- 7b. 切片傳入函式 ---") // 小節標題

	data := []int{10, 20, 30} // 建立切片
	fmt.Println("修改前:", data) // [10 20 30]

	modifySliceElements(data) // 傳入切片（複製 header，但底層陣列共享）
	fmt.Println("修改後:", data) // [20 40 60]（元素真的被改了！）

	// 為什麼會被改？因為函式內的切片和外面的切片
	// 共享同一個底層陣列（ptr 指向同一塊記憶體）

	// ----------------------------------------
	// 7c. append() 的陷阱：可能建立新的底層陣列！
	// ----------------------------------------
	fmt.Println("\n--- 7c. append() 的陷阱 ---") // 小節標題

	original := []int{1, 2, 3}   // 建立切片（len=3, cap=3）
	fmt.Println("原始:", original) // [1 2 3]

	tryAppendToSlice(original)    // 在函式內 append
	fmt.Println("函式外:", original) // [1 2 3]（沒有 999！）

	// 為什麼函式外看不到 999？
	// 因為 append 發現容量不夠（len=3, cap=3），
	// 就建立了一個新的、更大的底層陣列，
	// 函式內的切片 header 指向了新陣列，
	// 但外面的切片 header 還是指向舊陣列！

	// 正確的做法：用回傳值接收 append 的結果
	result := appendToSlice(original) // 接收回傳的新切片
	fmt.Println("正確做法:", result)      // [1 2 3 999]（有 999 了！）

	// 再看一個例子：當容量足夠時 append 不會建立新陣列
	withCap := make([]int, 3, 10)                        // 建立切片（len=3, cap=10，有額外空間）
	withCap[0] = 1                                       // 設定第一個元素
	withCap[1] = 2                                       // 設定第二個元素
	withCap[2] = 3                                       // 設定第三個元素
	fmt.Println("有額外容量:", withCap, "cap:", cap(withCap)) // [1 2 3] cap: 10

	// 即使容量足夠，函式內的 append 也不會更新外面的 len
	tryAppendToSlice(withCap)      // 函式內 append 了，底層陣列確實被寫入了
	fmt.Println("外面看不到:", withCap) // [1 2 3]（len 沒變，所以看不到新元素）
	// 結論：append 後「一定要」用 = 接收回傳值！

	// ----------------------------------------
	// 7d. Map 也包含內部指標
	// ----------------------------------------
	fmt.Println("\n--- 7d. Map 的內部指標 ---") // 小節標題

	// Map 的底層是一個指向雜湊表的指標
	// 所以傳入函式時，函式能直接修改原始的 map
	// （和切片類似，但 map 不需要擔心 append 的問題）

	scores := map[string]int{ // 建立一個 map
		"Alice": 90, // 鍵 "Alice"，值 90
		"Bob":   85, // 鍵 "Bob"，值 85
	}

	fmt.Println("修改前:", scores) // map[Alice:90 Bob:85]
	modifyMap(scores)           // 傳入 map（內部指標被複製，指向同一個雜湊表）
	fmt.Println("修改後:", scores) // map[Alice:90 Bob:85 new_key:42]（新的鍵值對出現了！）

	// 總結：為什麼切片和 map 不需要 & 就能在函式內修改？
	// 因為它們的「值」本身就包含指標：
	//   - 切片 = {指標, 長度, 容量}
	//   - map  = 指向雜湊表的指標
	// 傳入函式時，雖然是「傳值」，但複製的值裡面有指標，
	// 所以透過這個指標能修改底層資料

	// ========================================
	// 8. 什麼時候「不要」用指標
	// ========================================
	fmt.Println("\n=== 8. 什麼時候不要用指標 ===") // 章節標題

	// 不是所有情況都該用指標！以下情況用「值」比較好：

	// 情況 1：小型結構體（幾個欄位而已）
	// 複製的成本很低，用值更簡單、更安全
	type Point struct { // 只有兩個欄位的小結構體
		X int // X 座標
		Y int // Y 座標
	}
	p1 := Point{X: 1, Y: 2} // 值語意：複製很便宜
	p3 := p1                // 複製一份，p3 和 p1 完全獨立
	p3.X = 100              // 修改 p3 不影響 p1
	fmt.Println("p1:", p1)  // {1 2}（不受影響）
	fmt.Println("p3:", p3)  // {100 2}（只有 p3 被改了）

	// 情況 2：只需要讀取、不需要修改
	// 傳值可以保證函式不會意外修改你的資料
	fmt.Println("使用者名稱:", user.Name) // 只是讀取，不需要指標

	// 情況 3：基本型別（int, string, bool 等）
	// 這些型別本身就很小，用指標反而增加複雜度
	age := 25                            // int 很小，不需要指標
	name := "Alice"                      // string 在 Go 內部已經有指標機制
	fmt.Println("年齡:", age, "名稱:", name) // 直接用值就好

	// ========================================
	// 總結
	// ========================================
	fmt.Println("\n=== 總結 ===") // 章節標題

	fmt.Println("1. & 取址、* 解引用")                        // 基本語法
	fmt.Println("2. Go 預設傳值，需要修改時傳指標")                  // 傳值 vs 傳指標
	fmt.Println("3. 結構體指標有語法糖：ptr.Field")               // 語法糖
	fmt.Println("4. nil 指標一定要檢查才能用")                    // 安全性
	fmt.Println("5. Repository 回傳 *User 可以用 nil 表示不存在") // 實際應用
	fmt.Println("6. 切片 = {指標, 長度, 容量}，內部已有指標")          // 切片真相
	fmt.Println("7. Map 內部也有指標，傳入函式能直接修改")              // Map 真相
	fmt.Println("8. append() 一定要用 = 接收回傳值！")            // append 陷阱
}
