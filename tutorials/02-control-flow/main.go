// ============================================================
// 第二課：控制流程（Control Flow）
// ============================================================
// 上一課學了「變數」——怎麼存資料
// 這一課要學「控制流程」——怎麼讓程式做判斷和重複
//
// 你會學到三個核心概念：
//   1. if / else    → 做判斷（「如果⋯就⋯否則⋯」）
//   2. for          → 重複做事（迴圈）
//   3. switch       → 多重選擇（像選擇題）
//
// 執行方式：go run ./tutorials/02-control-flow
// ============================================================

package main // 宣告這個檔案屬於 main 套件（可執行程式）

import "fmt" // 引入 fmt 套件，用來印東西到螢幕上

func main() {

	// ========================================
	// 1. if / else if / else —— 條件判斷
	// ========================================
	// 「如果成績 >= 90，就印 A；否則如果 >= 80，就印 B⋯⋯」
	//
	// 語法：
	//   if 條件 {
	//       // 條件成立時執行這裡
	//   } else if 另一個條件 {
	//       // 第一個不成立、第二個成立時執行這裡
	//   } else {
	//       // 全部都不成立時執行這裡
	//   }
	//
	// 注意！Go 跟其他語言的不同：
	//   ✅ if score >= 90 {       ← 不需要小括號
	//   ❌ if (score >= 90) {     ← 這樣也能跑，但不是 Go 風格
	//   ❌ if score >= 90         ← 大括號 {} 是必須的，不能省略

	fmt.Println("=== 1. if 條件判斷 ===")

	score := 85 // 假設考了 85 分

	if score >= 90 { // 條件：分數 >= 90？
		fmt.Println("成績：A") // 條件成立才會執行
	} else if score >= 80 { // 上面不成立，再判斷：>= 80？
		fmt.Println("成績：B") // ← 85 >= 80 成立，會印這行
	} else if score >= 70 { // 上面不成立，再判斷：>= 70？
		fmt.Println("成績：C")
	} else { // 以上全部都不成立
		fmt.Println("成績：不及格")
	}

	// ========================================
	// 2. if + 初始化語句 —— Go 的獨特語法
	// ========================================
	// Go 允許在 if 裡面「先做一件事，再判斷」
	// 語法：if 初始化; 條件 { ... }
	// 分號 ; 前面是初始化，後面是條件

	fmt.Println("\n=== 2. if + 初始化語句 ===")

	// 先算出 10 除以 3 的餘數，再判斷餘數是不是 0
	if remainder := 10 % 3; remainder == 0 { // % 是取餘數運算符
		fmt.Println("10 可以被 3 整除") // 餘數是 0 才會執行
	} else {
		fmt.Printf("10 除以 3 的餘數是 %d\n", remainder) // ← 餘數是 1，執行這行
	}
	// 重要：remainder 這個變數只在 if/else 的 {} 裡面有效
	// 出了大括號就不能用了（這叫「作用域 scope」）

	// ========================================
	// 3. for 迴圈 —— Go 只有 for，沒有 while！
	// ========================================
	// 很多語言有 for、while、do-while 三種迴圈
	// Go 只有 for，但 for 可以寫成各種形式

	fmt.Println("\n=== 3. for 迴圈 ===")

	// --- 形式 1：標準 for（最常見）---
	// for 初始化; 條件; 每次結束後做的事 { ... }
	// 「從 i=0 開始，只要 i<5，每次 i 加 1」
	fmt.Print("標準 for: ")
	for i := 0; i < 5; i++ { // i++ 等於 i = i + 1
		fmt.Print(i, " ") // 印出：0 1 2 3 4
	}
	fmt.Println() // 換行

	// --- 形式 2：類似 while（只有條件）---
	// 「只要 count < 5，就繼續」
	fmt.Print("while 風格: ")
	count := 0      // 初始化放在外面
	for count < 5 { // 只寫條件
		fmt.Print(count, " ") // 印出：0 1 2 3 4
		count++               // 自己手動加 1
	}
	fmt.Println()

	// --- 形式 3：無限迴圈 + break ---
	// for 後面什麼都不寫 = 無限迴圈（永遠不會自己停）
	// 必須用 break 來跳出
	fmt.Print("無限迴圈 + break: ")
	n := 0
	for { // 無限迴圈開始
		if n >= 5 { // 當 n >= 5 時⋯
			break // ← break = 「立刻跳出整個迴圈」
		}
		fmt.Print(n, " ") // 印出：0 1 2 3 4
		n++
	}
	fmt.Println()

	// --- continue：跳過這一次，直接進入下一次 ---
	fmt.Print("跳過偶數（只印奇數）: ")
	for i := 0; i < 10; i++ {
		if i%2 == 0 { // 如果 i 是偶數（除以 2 餘數為 0）
			continue // ← continue = 「跳過下面的程式碼，直接進入下一次迴圈」
		}
		fmt.Print(i, " ") // 只有奇數會執行到這裡：1 3 5 7 9
	}
	fmt.Println()

	// ========================================
	// 4. 標籤式 break / continue（進階）
	// ========================================
	// 當你有「迴圈裡面套迴圈」時，break 只會跳出最內層的迴圈
	// 如果想跳出外層迴圈，要用「標籤（label）」

	fmt.Println("\n=== 4. 標籤式 break ===")

	fmt.Println("在二維表格中找到 (1,2) 就停止：")
outer: // ← 這是一個標籤，名字叫 outer（可以取任何名字）
	for i := 0; i < 3; i++ { // 外層迴圈
		for j := 0; j < 3; j++ { // 內層迴圈
			fmt.Printf("  檢查 (%d,%d)\n", i, j)
			if i == 1 && j == 2 { // 找到目標座標
				fmt.Println("  → 找到了！跳出所有迴圈")
				break outer // ← 跳出「外層」迴圈（不是只跳出內層）
			}
		}
	}
	// 如果只寫 break（沒有 outer），只會跳出內層迴圈，外層會繼續

	// ========================================
	// 5. for range —— 遍歷（走訪每一個元素）
	// ========================================
	// for range 用來「一個一個拜訪」字串、陣列、切片、map 裡的元素
	// 語法：for 索引, 值 := range 要遍歷的東西 { ... }

	fmt.Println("\n=== 5. for range ===")

	// 遍歷字串：每個字元（character）
	message := "Hello Go"
	fmt.Println("遍歷字串 \"Hello Go\"：")
	for index, char := range message {
		// index = 這個字元在字串中的位置（從 0 開始）
		// char  = 這個字元的 Unicode 碼（rune 型別）
		// %c    = 把 Unicode 碼印成字元
		fmt.Printf("  位置 %d → '%c'\n", index, char)
	}

	// --- 用 _ 忽略不需要的值 ---
	// Go 規定：宣告了變數就一定要用到，否則編譯錯誤
	// 如果你不需要 index，用 _（底線）來忽略它

	fmt.Print("只取值（忽略索引）: ")
	for _, char := range "ABC" { // _ 忽略索引
		fmt.Printf("%c ", char) // 印出：A B C
	}
	fmt.Println()

	// 只需要索引，不需要值：直接省略第二個變數
	fmt.Print("只取索引: ")
	for i := range "ABC" { // 只有一個變數 = 索引
		fmt.Print(i, " ") // 印出：0 1 2
	}
	fmt.Println()

	// ========================================
	// 6. switch —— 多重選擇
	// ========================================
	// 像選擇題：根據值來選擇要執行哪一段
	//
	// Go 的 switch 跟其他語言最大的不同：
	//   ✅ 不需要 break！每個 case 執行完會自動跳出
	//   ✅ 一個 case 可以寫多個值（用逗號分隔）

	fmt.Println("\n=== 6. switch ===")

	day := "星期三"

	switch day { // 根據 day 的值來選擇
	case "星期一", "星期二", "星期三", "星期四", "星期五":
		// 一個 case 可以匹配多個值（用逗號分隔）
		fmt.Println(day, "是工作日")
	case "星期六", "星期日":
		fmt.Println(day, "是假日")
	default: // 以上都不匹配時執行（像 else）
		fmt.Println("未知的日期")
	}
	// 注意：不需要寫 break！Go 的 switch 自動 break

	// --- switch 不帶表達式（像 if-else 鏈）---
	// 不寫 switch 後面的值，直接在 case 裡寫條件
	temperature := 32

	switch { // 注意：switch 後面沒有值
	case temperature >= 35:
		fmt.Println("天氣：酷熱 🔥")
	case temperature >= 28:
		fmt.Println("天氣：炎熱 ☀️") // ← 32 >= 28 成立
	case temperature >= 20:
		fmt.Println("天氣：舒適 🌤️")
	default:
		fmt.Println("天氣：寒冷 ❄️")
	}

	// ========================================
	// 7. 實際應用：判斷 HTTP 狀態碼
	// ========================================
	// 在部落格專案中，API 會回傳不同的狀態碼
	// 我們可以用 switch 來分類處理

	fmt.Println("\n=== 7. 實際應用：HTTP 狀態碼 ===")

	statusCode := 404 // HTTP 狀態碼

	switch {
	case statusCode >= 200 && statusCode < 300:
		// 200-299 = 成功（例如 200 OK、201 Created）
		fmt.Printf("%d: ✅ 成功\n", statusCode)
	case statusCode >= 400 && statusCode < 500:
		// 400-499 = 客戶端錯誤（例如 404 Not Found、401 Unauthorized）
		fmt.Printf("%d: ❌ 客戶端錯誤\n", statusCode)
	case statusCode >= 500:
		// 500+ = 伺服器錯誤（例如 500 Internal Server Error）
		fmt.Printf("%d: 💥 伺服器錯誤\n", statusCode)
	}

	// 在部落格專案的 pkg/response/response.go 中
	// 就是用類似的邏輯來回傳不同的 HTTP 狀態碼：
	//   response.Success(c, data)            → 回傳 200
	//   response.Created(c, data)            → 回傳 201
	//   response.BadRequest(c, "錯誤訊息")    → 回傳 400
	//   response.NotFound(c, "找不到")        → 回傳 404
}
