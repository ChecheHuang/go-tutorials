// ============================================================
// 第九課：切片與映射（Slices & Maps）
// ============================================================
// 切片（Slice）就像「可以伸縮的書架」——可以隨時加書或拿走書。
// 映射（Map）就像「電話簿」——用名字查電話號碼，快速找到對應的值。
//
// 在其他語言中：
//   JavaScript: Array / Object  →  Go: Slice / Map
//   Python:     list / dict     →  Go: Slice / Map
//   Java:       ArrayList / HashMap → Go: Slice / Map
//
// 陣列（Array）vs 切片（Slice）的核心差異：
//   陣列：長度固定，宣告時就決定了大小，像固定大小的書架
//   切片：長度可變，可以動態增減，像可以伸縮的書架
//   實際開發中 99% 使用切片，幾乎不會直接用陣列！
//
// 切片的內部結構（連結第五課指標觀念）：
//   切片本質上是一個包含三個欄位的結構體：
//   ┌──────────┬──────────┬──────────┐
//   │ pointer  │  length  │ capacity │
//   │ 指向底層  │  目前有   │  最多能   │
//   │ 陣列的   │  幾個元素  │  放幾個   │
//   │ 指標     │          │  元素     │
//   └──────────┴──────────┴──────────┘
//   這就是為什麼切片傳入函式時，修改會影響原始資料——
//   因為它們共用同一個底層陣列（pointer 指向同一個地方）！
//
// 你會學到：
//   1. 陣列 vs 切片的差異
//   2. 切片的建立與操作
//   3. make() 預先分配容量
//   4. append() 的擴容行為
//   5. 切片的陷阱（共享底層陣列）
//   6. Map 的建立與操作
//   7. Map 的零值（nil map vs 空 map）
//   8. Map 遍歷順序是隨機的
//   9. 函式式操作（filter、map）
//   10. 實際應用：部落格文章搜尋
//
// 執行方式：go run ./tutorials/09-slices-maps
// ============================================================

package main // 每個可執行程式都必須屬於 main 套件

import ( // import 區塊：引入需要的套件
	"fmt"     // fmt = format，格式化輸出
	"sort"    // sort = 排序工具
	"strings" // strings = 字串處理工具
)

// ============================================================
// 主程式入口
// ============================================================
func main() { // main 函式：程式的起點
	fmt.Println("===== 第九課：切片與映射 =====") // 印出課程標題
	fmt.Println()                        // 空行分隔

	// ========================================
	// 1. 陣列（Array）vs 切片（Slice）
	// ========================================
	fmt.Println("--- 1. 陣列 vs 切片 ---") // 印出區段標題

	// 陣列（Array）：長度固定，宣告時就決定大小
	// [3]int 表示「剛好 3 個 int 的陣列」，不能多也不能少
	var arr [3]int = [3]int{10, 20, 30} // 宣告並初始化一個長度為 3 的整數陣列
	fmt.Println("陣列:", arr)             // 印出陣列：[10 20 30]
	fmt.Printf("陣列型別: %T\n", arr)       // 印出型別：[3]int（長度是型別的一部分！）
	// arr = append(arr, 40) ← 編譯錯誤！陣列不能用 append

	// 切片（Slice）：長度可變，實際開發幾乎都用切片
	// []int 表示「int 的切片」，注意方括號裡沒有數字
	slice := []int{10, 20, 30}      // 宣告並初始化一個整數切片
	fmt.Println("切片:", slice)       // 印出切片：[10 20 30]
	fmt.Printf("切片型別: %T\n", slice) // 印出型別：[]int（沒有長度，因為長度可變）

	// 關鍵差異：[3]int 和 []int 是完全不同的型別！
	// [3]int 是陣列（固定長度），[]int 是切片（可變長度）
	fmt.Println() // 空行分隔

	// ========================================
	// 2. 切片的內部結構（pointer + length + capacity）
	// ========================================
	// 連結第五課指標觀念：
	// 切片不是直接存放資料，而是持有一個指向底層陣列的指標
	// 就像「書架的標籤」指向「實際的書架」
	fmt.Println("--- 2. 切片的內部結構 ---") // 印出區段標題

	original := []int{1, 2, 3, 4, 5}             // 建立原始切片，底層陣列是 [1,2,3,4,5]
	fmt.Printf("original: %v\n", original)       // 印出切片內容
	fmt.Printf("  長度(len): %d\n", len(original)) // len() 回傳目前有幾個元素
	fmt.Printf("  容量(cap): %d\n", cap(original)) // cap() 回傳底層陣列能容納幾個元素

	// 從切片中切出一段：[start:end]
	// 包含 start，不包含 end（和 Python 一樣）
	subSlice := original[1:3]                              // 取出索引 1 到 2 的元素（不包含 3）
	fmt.Printf("subSlice = original[1:3]: %v\n", subSlice) // 印出 [2 3]
	fmt.Printf("  長度(len): %d\n", len(subSlice))           // 長度是 2（只有 2 個元素）
	fmt.Printf("  容量(cap): %d\n", cap(subSlice))           // 容量是 4（從索引 1 到底層陣列的末尾）
	// 容量 = 底層陣列長度 - 起始索引 = 5 - 1 = 4
	fmt.Println() // 空行分隔

	// ========================================
	// 3. 切片的陷阱：共享底層陣列！
	// ========================================
	// 這是初學者最常踩到的坑：
	// 從同一個切片切出的子切片，共用同一個底層陣列
	// 修改子切片，會影響原始切片！
	fmt.Println("--- 3. 切片陷阱：共享底層陣列 ---") // 印出區段標題

	data := []int{10, 20, 30, 40, 50} // 建立原始切片
	fmt.Println("原始 data:", data)     // 印出 [10 20 30 40 50]

	part := data[1:4]                      // 取出索引 1~3 的部分
	fmt.Println("part = data[1:4]:", part) // 印出 [20 30 40]

	// 危險操作：修改 part 的元素
	part[0] = 999                      // 把 part 的第一個元素改成 999
	fmt.Println("修改 part[0] = 999 後：") // 說明操作
	fmt.Println("  part:", part)       // part 變成 [999 30 40]
	fmt.Println("  data:", data)       // data 也變了！[10 999 30 40 50]
	// 因為 part 和 data 的底層陣列是同一個！

	// 安全做法：用 copy() 建立獨立的副本
	safeCopy := make([]int, len(data))   // 用 make 建立一個新的切片
	copy(safeCopy, data)                 // copy(目標, 來源)：複製資料到新切片
	safeCopy[0] = 777                    // 修改副本的元素
	fmt.Println("用 copy 建立副本後：")         // 說明操作
	fmt.Println("  safeCopy:", safeCopy) // safeCopy: [777 999 30 40 50]
	fmt.Println("  data:    ", data)     // data 不受影響：[10 999 30 40 50]
	fmt.Println()                        // 空行分隔

	// ========================================
	// 4. 切片操作：存取、切片、長度
	// ========================================
	fmt.Println("--- 4. 切片操作 ---") // 印出區段標題

	fruits := []string{"蘋果", "香蕉", "櫻桃", "葡萄", "芒果"} // 建立水果切片
	fmt.Println("原始:", fruits)                       // 印出所有水果

	// 存取元素（索引從 0 開始）
	fmt.Println("第一個 fruits[0]:", fruits[0])                  // 索引 0 = 第一個元素
	fmt.Println("最後一個 fruits[len-1]:", fruits[len(fruits)-1]) // 最後一個元素

	// 切片操作：[start:end]
	fmt.Println("fruits[1:3]:", fruits[1:3]) // 索引 1 到 2：[香蕉 櫻桃]
	fmt.Println("fruits[:2]: ", fruits[:2])  // 從開頭到索引 1：[蘋果 香蕉]
	fmt.Println("fruits[3:]: ", fruits[3:])  // 從索引 3 到結尾：[葡萄 芒果]
	fmt.Println()                            // 空行分隔

	// ========================================
	// 5. append()：新增元素
	// ========================================
	// append 是切片最重要的操作
	// 它回傳一個新的切片（可能指向新的底層陣列）
	fmt.Println("--- 5. append：新增元素 ---") // 印出區段標題

	numbers := []int{1, 2, 3}                                    // 建立初始切片
	fmt.Println("原始:", numbers)                                  // [1 2 3]
	fmt.Printf("  len=%d, cap=%d\n", len(numbers), cap(numbers)) // 長度 3，容量 3

	// 新增一個元素
	numbers = append(numbers, 4)                                 // append 回傳新切片，一定要接回來！
	fmt.Println("新增 4:", numbers)                                // [1 2 3 4]
	fmt.Printf("  len=%d, cap=%d\n", len(numbers), cap(numbers)) // 容量可能翻倍成 6

	// 一次新增多個元素
	numbers = append(numbers, 5, 6, 7) // 可以一次加多個
	fmt.Println("新增 5,6,7:", numbers)  // [1 2 3 4 5 6 7]

	// 合併兩個切片（用 ... 展開）
	more := []int{8, 9}                // 另一個切片
	numbers = append(numbers, more...) // ... 把切片展開成一個個元素
	fmt.Println("合併:", numbers)        // [1 2 3 4 5 6 7 8 9]
	fmt.Println()                      // 空行分隔

	// ========================================
	// 6. append() 的擴容行為
	// ========================================
	// 當容量不夠時，append 會自動擴容：
	//   - 建立一個更大的底層陣列（通常是原來的 2 倍）
	//   - 把舊資料複製過去
	//   - 回傳指向新陣列的切片
	// 所以 append 可能很昂貴！如果知道大小，用 make 預先分配
	fmt.Println("--- 6. append 的擴容行為 ---") // 印出區段標題

	growth := make([]int, 0)  // 建立空切片，長度 0，容量 0
	prevCap := cap(growth)    // 記錄前一次的容量
	for i := 0; i < 20; i++ { // 迴圈新增 20 個元素
		growth = append(growth, i)  // 每次新增一個元素
		if cap(growth) != prevCap { // 如果容量改變了，表示發生了擴容
			fmt.Printf("  len=%2d → cap 從 %2d 擴容到 %2d\n", // 印出擴容資訊
				len(growth), prevCap, cap(growth)) // 新的長度和容量
			prevCap = cap(growth) // 更新前一次容量的記錄
		}
	}
	// 你會看到容量大約以 2 倍的方式成長：0→1→2→4→8→16→32
	fmt.Println() // 空行分隔

	// ========================================
	// 7. make()：預先分配容量
	// ========================================
	// make([]型別, 長度, 容量)
	// 當你事先知道大約需要多少元素時，用 make 預先分配
	// 可以避免多次擴容，提升效能
	fmt.Println("--- 7. make：預先分配容量 ---") // 印出區段標題

	// make([]int, 0, 10)：長度 0（目前沒有元素），容量 10（預留 10 個位置）
	scores := make([]int, 0, 10)                                 // 建立一個預留 10 個位置的空切片
	fmt.Printf("初始: len=%d, cap=%d\n", len(scores), cap(scores)) // len=0, cap=10

	for i := 0; i < 5; i++ { // 迴圈新增 5 個元素
		scores = append(scores, i*10) // 新增分數
	}
	fmt.Println("填充後:", scores)                                // [0 10 20 30 40]
	fmt.Printf("  len=%d, cap=%d\n", len(scores), cap(scores)) // len=5, cap=10（沒有擴容！）

	// make([]int, 5)：長度 5（有 5 個零值元素），容量也是 5
	zeros := make([]int, 5)     // 建立長度為 5 的切片，每個元素都是 0
	fmt.Println("零值切片:", zeros) // [0 0 0 0 0]

	// 常見錯誤：make([]int, 5) 然後用 append
	// 這樣會在 5 個零之後再加元素，而不是從頭開始！
	wrong := make([]int, 5)     // [0 0 0 0 0]（已經有 5 個元素了）
	wrong = append(wrong, 1)    // 在後面加 1
	fmt.Println("常見錯誤:", wrong) // [0 0 0 0 0 1]（不是 [1]！）

	// 正確做法：make([]int, 0, 5) 然後用 append
	right := make([]int, 0, 5)  // []（空的，但預留了 5 個位置）
	right = append(right, 1)    // 加入 1
	fmt.Println("正確做法:", right) // [1]
	fmt.Println()               // 空行分隔

	// ========================================
	// 8. 遍歷切片
	// ========================================
	fmt.Println("--- 8. 遍歷切片 ---") // 印出區段標題

	colors := []string{"紅", "綠", "藍"} // 建立顏色切片

	// for range 同時取得索引和值
	fmt.Println("索引和值:")           // 標題
	for i, color := range colors { // i = 索引, color = 值
		fmt.Printf("  [%d] %s\n", i, color) // 印出每個顏色的索引和值
	}

	// 只需要值（用 _ 忽略索引）
	fmt.Print("只要值: ")             // 不換行
	for _, color := range colors { // _ 表示「我不需要這個值」
		fmt.Print(color, " ") // 印出顏色
	}
	fmt.Println() // 換行

	// 只需要索引（省略第二個變數）
	fmt.Print("只要索引: ")     // 不換行
	for i := range colors { // 只寫一個變數，就只取得索引
		fmt.Print(i, " ") // 印出索引
	}
	fmt.Println() // 換行
	fmt.Println() // 空行分隔

	// ========================================
	// 9. 函式式操作：filter / map
	// ========================================
	// Go 沒有內建的 filter/map 函式，但我們可以自己寫
	// （Go 1.21+ 有 slices 套件，但理解原理很重要）
	fmt.Println("--- 9. filter / map 操作 ---") // 印出區段標題

	nums := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10} // 建立 1 到 10 的切片

	// Filter：篩選偶數
	evens := filter(nums, func(n int) bool { // 傳入匿名函式作為篩選條件
		return n%2 == 0 // 如果能被 2 整除就是偶數
	})
	fmt.Println("偶數:", evens) // [2 4 6 8 10]

	// Map：每個元素乘以 2
	doubled := mapInts(nums, func(n int) int { // 傳入匿名函式作為轉換邏輯
		return n * 2 // 每個數字乘以 2
	})
	fmt.Println("乘以 2:", doubled) // [2 4 6 8 10 12 14 16 18 20]
	fmt.Println()                 // 空行分隔

	// ========================================
	// 10. Map（映射 / 字典）基礎
	// ========================================
	// Map 就像「電話簿」：用名字（key）查電話號碼（value）
	// map[KeyType]ValueType
	fmt.Println("--- 10. Map 基礎 ---") // 印出區段標題

	// 建立 Map（字面值初始化）
	ages := map[string]int{ // key 是 string，value 是 int
		"Alice": 25, // "Alice" 對應 25
		"Bob":   30, // "Bob" 對應 30
		"Carol": 28, // "Carol" 對應 28
	}
	fmt.Println("Map:", ages) // 印出整個 Map

	// 存取值
	aliceAge := ages["Alice"]           // 用 key 取得對應的 value
	fmt.Println("Alice 的年齡:", aliceAge) // 25

	// 新增 / 修改
	ages["Dave"] = 35         // 如果 key 不存在，就是新增
	ages["Alice"] = 26        // 如果 key 已存在，就是修改
	fmt.Println("更新後:", ages) // 印出更新後的 Map
	fmt.Println()             // 空行分隔

	// ========================================
	// 11. Map 的零值與存在性檢查
	// ========================================
	// Map 的零值是 nil，nil map 不能寫入！
	// 存取不存在的 key 不會報錯，而是回傳 value 型別的零值
	fmt.Println("--- 11. Map 零值與存在性檢查 ---") // 印出區段標題

	// nil map vs 空 map
	var nilMap map[string]int       // 宣告但未初始化 → nil map
	emptyMap := map[string]int{}    // 初始化為空 → 空 map
	madeMap := make(map[string]int) // 用 make 初始化 → 空 map

	fmt.Println("nilMap == nil?  ", nilMap == nil)   // true：沒有初始化
	fmt.Println("emptyMap == nil?", emptyMap == nil) // false：已初始化，只是沒有元素
	fmt.Println("madeMap == nil? ", madeMap == nil)  // false：已初始化

	// 讀取 nil map 不會出錯（回傳零值）
	fmt.Println("nilMap[\"key\"]:", nilMap["key"]) // 0（int 的零值）

	// 但寫入 nil map 會 panic！
	// nilMap["key"] = 1  ← 這行會導致程式崩潰（panic）！
	// 所以使用 map 之前一定要初始化！

	// 檢查 key 是否存在（comma ok 模式）
	age, exists := ages["Bob"] // 回傳兩個值：值和是否存在
	if exists {                // 如果 key 存在
		fmt.Printf("Bob 的年齡: %d\n", age) // 印出年齡
	} else { // 如果 key 不存在
		fmt.Println("Bob 不存在") // 印出不存在的訊息
	}

	// 存取不存在的 key
	unknownAge := ages["Unknown"]        // key 不存在，回傳 int 的零值 0
	fmt.Println("不存在的 key:", unknownAge) // 0（不是錯誤！）
	// 所以你沒辦法分辨「值是 0」和「key 不存在」，除非用 comma ok 模式
	fmt.Println() // 空行分隔

	// ========================================
	// 12. 從 Map 中刪除元素
	// ========================================
	fmt.Println("--- 12. 刪除 Map 元素 ---") // 印出區段標題

	fmt.Println("刪除前:", ages)      // 印出刪除前的 Map
	delete(ages, "Bob")            // delete(map, key)：從 Map 中刪除指定的 key
	fmt.Println("刪除 Bob 後:", ages) // Bob 已被移除

	// 刪除不存在的 key 不會報錯（什麼都不做）
	delete(ages, "NotExist")        // 刪除不存在的 key，不會 panic
	fmt.Println("刪除不存在的 key: 沒有報錯") // 確認沒有問題
	fmt.Println()                   // 空行分隔

	// ========================================
	// 13. 遍歷 Map（順序是隨機的！）
	// ========================================
	// 重要：Map 的遍歷順序是隨機的
	// 每次執行程式，順序可能不同
	// 如果需要固定順序，必須先把 key 取出來排序
	fmt.Println("--- 13. 遍歷 Map ---") // 印出區段標題

	studentScores := map[string]int{ // 學生分數的 Map
		"小明": 85, // 小明的分數
		"小華": 92, // 小華的分數
		"小美": 78, // 小美的分數
		"小強": 95, // 小強的分數
	}

	// 直接遍歷（順序隨機）
	fmt.Println("直接遍歷（順序可能每次不同）:")           // 提醒順序隨機
	for name, score := range studentScores { // 遍歷 Map 的每一對 key-value
		fmt.Printf("  %s: %d 分\n", name, score) // 印出學生姓名和分數
	}

	// 按 key 排序後遍歷（固定順序）
	fmt.Println("按姓名排序後遍歷:")                             // 固定順序的做法
	sortedNames := make([]string, 0, len(studentScores)) // 預先分配容量
	for name := range studentScores {                    // 只取 key（學生姓名）
		sortedNames = append(sortedNames, name) // 把所有 key 收集到切片
	}
	sort.Strings(sortedNames)          // 對切片排序
	for _, name := range sortedNames { // 按排序後的順序遍歷
		fmt.Printf("  %s: %d 分\n", name, studentScores[name]) // 用排序後的 key 查 Map
	}
	fmt.Println() // 空行分隔

	// ========================================
	// 14. 實際應用：單字計數
	// ========================================
	fmt.Println("--- 14. 實際應用：單字計數 ---") // 印出區段標題

	text := "the quick brown fox jumps over the lazy dog the fox" // 一段英文文字
	wordCount := countWords(text)                                 // 呼叫單字計數函式

	for word, count := range wordCount { // 遍歷計數結果
		if count > 1 { // 只印出出現超過一次的單字
			fmt.Printf("  '%s' 出現了 %d 次\n", word, count) // 印出單字和次數
		}
	}
	fmt.Println() // 空行分隔

	// ========================================
	// 15. 排序
	// ========================================
	fmt.Println("--- 15. 排序 ---") // 印出區段標題

	// 注意：sort 套件會直接修改原始切片（不是不可變的！）
	// 如果需要保留原始順序，先用 copy 建立副本
	unsorted := []int{5, 3, 8, 1, 9, 2} // 未排序的整數切片
	fmt.Println("排序前:", unsorted)       // 印出排序前的狀態
	sort.Ints(unsorted)                 // sort.Ints：把 int 切片從小到大排序
	fmt.Println("排序後:", unsorted)       // 印出排序後的狀態

	names := []string{"Charlie", "Alice", "Bob"} // 未排序的字串切片
	sort.Strings(names)                          // sort.Strings：按字母順序排序
	fmt.Println("字串排序:", names)                  // [Alice Bob Charlie]

	// 自訂排序（用 sort.Slice）
	type Student struct { // 定義學生結構體
		Name  string // 學生姓名
		Score int    // 學生分數
	}
	students := []Student{ // 建立學生切片
		{"小明", 85}, // 小明 85 分
		{"小華", 92}, // 小華 92 分
		{"小美", 78}, // 小美 78 分
	}
	sort.Slice(students, func(i, j int) bool { // sort.Slice：自訂排序規則
		return students[i].Score > students[j].Score // 分數高的排前面（降序）
	})
	fmt.Println("按分數降序:")        // 印出標題
	for _, s := range students { // 遍歷排序後的學生
		fmt.Printf("  %s: %d 分\n", s.Name, s.Score) // 印出每個學生的分數
	}
	fmt.Println() // 空行分隔

	// ========================================
	// 16. 部落格應用：文章搜尋
	// ========================================
	fmt.Println("--- 16. 部落格應用：文章搜尋 ---") // 印出區段標題

	// 模擬部落格文章資料
	type Article struct { // 定義文章結構體
		ID     int      // 文章 ID
		Title  string   // 文章標題
		Author string   // 作者
		Tags   []string // 標籤（用切片儲存多個標籤）
	}

	// 種子資料（用切片儲存所有文章）
	articles := []Article{ // 建立文章切片
		{1, "Go 入門教學", "小明", []string{"go", "beginner"}},      // 文章 1
		{2, "Go Web 開發", "小華", []string{"go", "web"}},         // 文章 2
		{3, "Python 機器學習", "小美", []string{"python", "ml"}},    // 文章 3
		{4, "Go 並發程式設計", "小明", []string{"go", "concurrency"}}, // 文章 4
	}

	// 用 Map 建立標籤索引（快速查詢）
	tagIndex := make(map[string][]int) // key=標籤, value=文章ID列表
	for _, article := range articles { // 遍歷所有文章
		for _, tag := range article.Tags { // 遍歷每篇文章的所有標籤
			tagIndex[tag] = append(tagIndex[tag], article.ID) // 把文章 ID 加入對應標籤的列表
		}
	}

	fmt.Println("標籤索引:")             // 印出標題
	for tag, ids := range tagIndex { // 遍歷標籤索引
		fmt.Printf("  #%s → 文章 ID: %v\n", tag, ids) // 印出每個標籤對應的文章 ID
	}

	// 搜尋含有 "go" 標籤的文章
	fmt.Println("搜尋 #go 標籤的文章:")      // 印出搜尋目標
	goArticleIDs := tagIndex["go"]    // 從索引中取得 "go" 標籤的文章 ID 列表
	for _, id := range goArticleIDs { // 遍歷文章 ID
		for _, article := range articles { // 在文章列表中找到對應的文章
			if article.ID == id { // 如果 ID 相符
				fmt.Printf("  [%d] %s（作者: %s）\n", // 印出文章資訊
					article.ID, article.Title, article.Author)
			}
		}
	}

	// 用 Map 統計每個作者的文章數
	authorCount := make(map[string]int) // key=作者, value=文章數
	for _, article := range articles {  // 遍歷所有文章
		authorCount[article.Author]++ // 作者的文章數加 1（map 零值是 0，直接 ++ 不會報錯）
	}
	fmt.Println("作者文章統計:")                   // 印出標題
	for author, count := range authorCount { // 遍歷統計結果
		fmt.Printf("  %s: %d 篇文章\n", author, count) // 印出每個作者的文章數
	}

	// ========================================
	// 總結
	// ========================================
	fmt.Println()                                                 // 空行分隔
	fmt.Println("===== 重點回顧 =====")                               // 印出回顧標題
	fmt.Println("1. 陣列固定長度 [3]int，切片可變長度 []int，實際開發用切片")          // 重點 1
	fmt.Println("2. 切片 = 指標 + 長度 + 容量，子切片共享底層陣列")                 // 重點 2
	fmt.Println("3. make([]T, len, cap) 預先分配容量，避免多次擴容")           // 重點 3
	fmt.Println("4. append() 可能觸發擴容，一定要用 s = append(s, ...) 接回來") // 重點 4
	fmt.Println("5. Map 使用前必須初始化（make 或字面值），nil map 不能寫入")        // 重點 5
	fmt.Println("6. Map 用 val, ok := m[key] 檢查 key 是否存在")         // 重點 6
	fmt.Println("7. Map 遍歷順序是隨機的，需要排序要自己處理")                      // 重點 7
}

// ============================================================
// 輔助函式
// ============================================================

// filter 篩選切片中符合條件的元素
// 參數：nums 是要篩選的切片，predicate 是判斷條件的函式
// 回傳：只包含符合條件元素的新切片
func filter(nums []int, predicate func(int) bool) []int { // 接受一個判斷函式
	result := make([]int, 0) // 建立空切片存放結果（不預估容量，因為不知道有多少符合）
	for _, n := range nums { // 遍歷每個元素
		if predicate(n) { // 如果這個元素符合條件
			result = append(result, n) // 加入結果切片
		}
	}
	return result // 回傳篩選後的新切片
}

// mapInts 將切片中的每個元素進行轉換
// 參數：nums 是原始切片，transform 是轉換函式
// 回傳：轉換後的新切片（長度和原始一樣）
func mapInts(nums []int, transform func(int) int) []int { // 接受一個轉換函式
	result := make([]int, len(nums)) // 預先分配和原始一樣長的切片（因為長度不會變）
	for i, n := range nums {         // 遍歷每個元素
		result[i] = transform(n) // 轉換後直接放到對應位置
	}
	return result // 回傳轉換後的新切片
}

// countWords 計算每個單字出現的次數
// 參數：text 是要分析的文字
// 回傳：key=單字, value=出現次數 的 Map
func countWords(text string) map[string]int { // 回傳 Map
	counts := make(map[string]int) // 建立空的 Map（一定要初始化！）
	words := strings.Fields(text)  // Fields 用空白字元分割字串，得到單字切片
	for _, word := range words {   // 遍歷每個單字
		counts[word]++ // 把這個單字的計數加 1
		// map 的零值是 0，所以第一次遇到新單字時，counts[word] 是 0，++ 後變成 1
	}
	return counts // 回傳計數結果
}
