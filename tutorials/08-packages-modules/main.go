// ============================================================
// 第八課：套件與模組（Packages & Modules）
// ============================================================
// 套件就像「工具箱」——每個工具箱裡放著一組相關的工具。
// 模組就像「整間工廠」——一間工廠裡有很多不同的工具箱。
//
// 在其他語言中：
//   Python: import os          →  Go: import "os"
//   Java:   import java.util.* →  Go: import "fmt"
//   Node:   require('fs')      →  Go: import "os"
//
// 想像一下：
//   工廠（Module）= 你的專案（go.mod 定義）
//   ├── 工具箱 A（Package fmt）  = 印東西的工具
//   ├── 工具箱 B（Package strings）= 處理文字的工具
//   ├── 工具箱 C（Package time）  = 處理時間的工具
//   └── 工具箱 D（Package math）  = 數學運算的工具
//
// 重要規則：
//   大寫開頭 = 公開（Public）→ 外部套件可以使用
//   小寫開頭 = 私有（Private）→ 只有套件內部能用
//   這就像工具箱上標記「共用」或「僅限內部使用」
//
// 你會學到：
//   1. 什麼是套件（Package）
//   2. import 語句的用法
//   3. 公開 vs 私有（大寫 vs 小寫）
//   4. 標準庫常用套件（fmt, strings, strconv, time, math）
//   5. 部落格專案如何組織套件
//   6. 匯入路徑的慣例
//
// 執行方式：go run ./tutorials/08-packages-modules
// ============================================================

package main // 每個可執行程式都必須屬於 main 套件

// ============================================================
// import 語句：告訴 Go「我需要用哪些工具箱」
// ============================================================
// Go 的 import 有三種寫法：
//
// 寫法一：單行匯入
//   import "fmt"
//
// 寫法二：群組匯入（最常用，推薦！）
//   import (
//       "fmt"
//       "strings"
//   )
//
// 寫法三：別名匯入
//   import str "strings"  ← 用 str 代替 strings
//
// 在專案中，匯入的排列慣例是三組：
//   第一組：標準庫（fmt, net/http, time ...）
//   第二組：自己專案的套件（blog-api/internal/domain ...）
//   第三組：第三方套件（github.com/gin-gonic/gin ...）
// ============================================================
import ( // 群組匯入：把所有需要的套件放在括號裡
	"fmt"     // fmt = format，用來格式化輸出（印東西到螢幕上）
	"math"    // math = 數學運算工具箱（圓周率、平方根、四捨五入等）
	"strconv" // strconv = string conversion，字串和數字互相轉換
	"strings" // strings = 字串處理工具箱（大小寫轉換、分割、取代等）
	"time"    // time = 時間相關工具箱（現在幾點、經過多久、格式化日期）
)

// ============================================================
// 1. 公開 vs 私有（Exported vs Unexported）
// ============================================================
// 這是 Go 最獨特的設計之一：
//   不需要 public / private 關鍵字！
//   只看名稱的第一個字母是大寫還是小寫。
//
// 大寫開頭 = 公開（Exported）→ 任何套件都能用
// 小寫開頭 = 私有（Unexported）→ 只有同一個套件內能用
//
// 就像你的辦公室：
//   「會議室」（大寫）→ 其他部門的人也可以來開會
//   「個人抽屜」（小寫）→ 只有自己能打開
// ============================================================

// BlogPost 是一個公開的結構體（大寫開頭 B）
// 任何匯入這個套件的程式碼都能使用 BlogPost
type BlogPost struct { // 定義一個部落格文章的結構體
	Title   string // Title 大寫開頭 → 公開欄位，外部可以讀寫
	Content string // Content 大寫開頭 → 公開欄位
	Author  string // Author 大寫開頭 → 公開欄位
	views   int    // views 小寫開頭 → 私有欄位，外部無法直接存取
}

// NewBlogPost 是公開的建構函式（大寫開頭 N）
// 外部套件透過這個函式來建立 BlogPost
func NewBlogPost(title, content, author string) BlogPost { // 建構函式：建立並回傳一個新的 BlogPost
	return BlogPost{ // 回傳初始化好的 BlogPost
		Title:   title,   // 設定文章標題
		Content: content, // 設定文章內容
		Author:  author,  // 設定作者名稱
		views:   0,       // 私有欄位：初始瀏覽次數為 0
	}
}

// AddView 是公開的方法（大寫開頭 A）
// 外部可以呼叫此方法來增加瀏覽次數
func (p *BlogPost) AddView() { // 指標接收者：因為要修改 views 的值
	p.views++ // 每呼叫一次，瀏覽次數加 1
}

// GetViews 是公開的方法（大寫開頭 G）
// 外部透過這個方法「安全地」取得私有的 views 值
func (p BlogPost) GetViews() int { // 值接收者：只需要讀取，不需要修改
	return p.views // 回傳瀏覽次數
}

// formatForDisplay 是私有的函式（小寫開頭 f）
// 只有這個套件內部能呼叫（在這裡就是 main 套件）
func formatForDisplay(post BlogPost) string { // 格式化文章資訊為顯示用字串
	return fmt.Sprintf("[%s] %s（作者：%s，瀏覽：%d 次）", // 用 Sprintf 格式化字串
		post.Title,     // 文章標題
		post.Content,   // 文章內容
		post.Author,    // 作者名稱
		post.GetViews(), // 透過公開方法取得私有的瀏覽次數
	)
}

// ============================================================
// 主程式入口
// ============================================================
func main() { // main 函式：程式的起點
	fmt.Println("===== 第八課：套件與模組 =====") // 印出課程標題
	fmt.Println() // 印出空行，讓輸出更易讀

	// ========================================
	// 示範 1：公開 vs 私有
	// ========================================
	fmt.Println("--- 1. 公開 vs 私有（大寫 vs 小寫）---") // 印出區段標題

	post := NewBlogPost("Go 入門", "學習 Go 語言基礎", "小明") // 用公開的建構函式建立文章
	post.AddView()                                              // 用公開的方法增加瀏覽次數
	post.AddView()                                              // 再增加一次瀏覽次數
	post.AddView()                                              // 再增加一次，現在是 3 次

	fmt.Println("文章標題:", post.Title)       // 可以直接存取公開欄位 Title
	fmt.Println("瀏覽次數:", post.GetViews())  // 必須透過公開方法取得私有欄位 views
	fmt.Println("格式化:", formatForDisplay(post)) // 呼叫私有函式（同一個套件內可以）
	// 如果在其他套件中，formatForDisplay 會編譯錯誤！
	fmt.Println() // 空行分隔

	// ========================================
	// 示範 2：fmt 套件 — 格式化輸出
	// ========================================
	// fmt 是 Go 中最常用的套件，全名是 "format"
	// 主要功能：把東西印到螢幕上、把東西格式化成字串
	fmt.Println("--- 2. fmt 套件：格式化輸出 ---") // 印出區段標題

	name := "Go 語言"                                     // 定義一個字串變數
	version := 1.26                                       // 定義一個浮點數變數
	fmt.Println("Println:", name)                         // Println = Print Line，印完會換行
	fmt.Printf("Printf: %s 版本 %.2f\n", name, version)  // Printf = Print Format，用佔位符格式化
	result := fmt.Sprintf("Sprintf: 歡迎使用 %s", name)  // Sprintf = String Printf，格式化後回傳字串（不印出）
	fmt.Println(result)                                   // 印出 Sprintf 的結果

	// 常用的格式化佔位符：
	// %s = 字串（string）
	// %d = 整數（decimal）
	// %f = 浮點數（float），%.2f 表示小數點後 2 位
	// %v = 任意值（value），Go 自動判斷型別
	// %T = 型別名稱（Type）
	// %t = 布林值（true/false）
	// %p = 指標位址（pointer）
	fmt.Printf("型別示範: %T, %T, %T\n", name, version, true) // 印出各個值的型別
	fmt.Println() // 空行分隔

	// ========================================
	// 示範 3：strings 套件 — 字串處理
	// ========================================
	// strings 套件提供了各種字串操作工具
	// 就像一把瑞士刀，處理文字的各種需求
	fmt.Println("--- 3. strings 套件：字串處理 ---") // 印出區段標題

	title := "  Hello, Go World!  "                            // 定義一個前後有空格的字串
	fmt.Println("原始字串:    ", title)                         // 印出原始字串（注意前後空格）
	fmt.Println("轉大寫:      ", strings.ToUpper(title))       // ToUpper：全部轉成大寫字母
	fmt.Println("轉小寫:      ", strings.ToLower(title))       // ToLower：全部轉成小寫字母
	fmt.Println("去空格:      ", strings.TrimSpace(title))     // TrimSpace：去掉前後的空白字元
	fmt.Println("包含 Go?    ", strings.Contains(title, "Go")) // Contains：檢查是否包含子字串
	fmt.Println("取代:        ", strings.ReplaceAll(title, "Go", "Golang")) // ReplaceAll：取代所有符合的子字串

	// Split：用指定的分隔符把字串切成切片（slice）
	tags := "go,web,api,blog"                                  // 用逗號分隔的標籤字串
	tagList := strings.Split(tags, ",")                        // 用逗號切割，得到字串切片
	fmt.Println("分割標籤:    ", tagList)                       // 印出切割後的結果：[go web api blog]

	// Join：把切片合併成一個字串（Split 的反操作）
	joined := strings.Join(tagList, " | ")                     // 用 " | " 把切片元素合併
	fmt.Println("合併標籤:    ", joined)                        // 印出合併後的結果

	// HasPrefix / HasSuffix：檢查字串的開頭或結尾
	url := "/api/v1/articles"                                       // 一個 API 路徑
	fmt.Println("以 /api 開頭? ", strings.HasPrefix(url, "/api"))   // HasPrefix：檢查是否以指定字串開頭
	fmt.Println("以 /articles 結尾?", strings.HasSuffix(url, "/articles")) // HasSuffix：檢查是否以指定字串結尾
	fmt.Println() // 空行分隔

	// ========================================
	// 示範 4：strconv 套件 — 型別轉換
	// ========================================
	// strconv = string conversion（字串轉換）
	// 在 Web 開發中超常用！因為 HTTP 請求的參數都是字串
	// 例如：URL 裡的 ?page=2，"2" 是字串，需要轉成數字
	fmt.Println("--- 4. strconv 套件：型別轉換 ---") // 印出區段標題

	// 字串轉整數：Atoi = ASCII to Integer
	pageStr := "42"                                  // 模擬從 URL 參數取得的頁碼（字串）
	pageNum, err := strconv.Atoi(pageStr)            // Atoi 回傳兩個值：轉換結果和錯誤
	if err != nil {                                  // 如果轉換失敗（例如 "abc" 無法轉成數字）
		fmt.Println("轉換失敗:", err) // 印出錯誤訊息
	} else { // 轉換成功
		fmt.Printf("字串 %q → 整數 %d\n", pageStr, pageNum) // %q 會幫字串加上引號
	}

	// 整數轉字串：Itoa = Integer to ASCII
	statusCode := 200                                        // HTTP 狀態碼（整數）
	statusStr := strconv.Itoa(statusCode)                    // 把整數轉成字串
	fmt.Printf("整數 %d → 字串 %q\n", statusCode, statusStr) // 印出轉換結果

	// 字串轉布林值
	activeStr := "true"                                      // 模擬從設定檔讀取的布林值字串
	isActive, err := strconv.ParseBool(activeStr)            // ParseBool：把 "true"/"false" 轉成 bool
	if err != nil {                                          // 如果轉換失敗
		fmt.Println("轉換失敗:", err) // 印出錯誤
	} else { // 轉換成功
		fmt.Printf("字串 %q → 布林 %t\n", activeStr, isActive) // %t 印出 true 或 false
	}

	// 字串轉浮點數
	priceStr := "19.99"                                         // 模擬從資料庫讀取的價格字串
	price, err := strconv.ParseFloat(priceStr, 64)              // ParseFloat：字串轉浮點數，64 表示 float64 精度
	if err != nil {                                             // 如果轉換失敗
		fmt.Println("轉換失敗:", err) // 印出錯誤
	} else { // 轉換成功
		fmt.Printf("字串 %q → 浮點數 %.2f\n", priceStr, price) // %.2f 保留小數點後 2 位
	}
	fmt.Println() // 空行分隔

	// ========================================
	// 示範 5：time 套件 — 時間處理
	// ========================================
	// time 套件處理所有和時間相關的事情
	// 在部落格專案中：文章建立時間、JWT 過期時間等都會用到
	fmt.Println("--- 5. time 套件：時間處理 ---") // 印出區段標題

	// 取得現在時間
	now := time.Now()                                     // Now() 回傳當前的日期和時間
	fmt.Println("現在時間:", now)                           // 印出完整的時間資訊

	// 格式化時間（Go 的時間格式化很特別！）
	// Go 不用 YYYY-MM-DD，而是用一個「參考時間」：
	//   2006-01-02 15:04:05
	//   （這是 Go 的誕生時刻，你只需要記住這個數字）
	formatted := now.Format("2006-01-02 15:04:05")        // 用參考時間的格式來格式化
	fmt.Println("格式化:  ", formatted)                    // 印出格式化後的時間字串

	// 只取日期
	dateOnly := now.Format("2006-01-02")                  // 只取年月日
	fmt.Println("只取日期:", dateOnly)                     // 印出日期

	// 取得時間的各個部分
	fmt.Printf("年: %d, 月: %d, 日: %d\n",              // 分別取出年、月、日
		now.Year(), now.Month(), now.Day())                // Year()、Month()、Day() 方法

	// 時間運算
	oneHourLater := now.Add(1 * time.Hour)                // Add：加上一段時間，time.Hour = 1 小時
	fmt.Println("一小時後:", oneHourLater.Format("15:04:05")) // 印出一小時後的時間

	// 計算時間差
	startTime := time.Now()                               // 記錄開始時間
	// 模擬一些處理...
	for i := 0; i < 1000000; i++ {                        // 做一百萬次空迴圈
		_ = i * i // 簡單的運算，_ 表示丟棄結果
	}
	elapsed := time.Since(startTime)                      // Since：計算從 startTime 到現在經過多久
	fmt.Printf("處理耗時: %v\n", elapsed)                 // %v 自動格式化 Duration 型別
	fmt.Println() // 空行分隔

	// ========================================
	// 示範 6：math 套件 — 數學運算
	// ========================================
	// math 套件提供常見的數學函式和常數
	fmt.Println("--- 6. math 套件：數學運算 ---") // 印出區段標題

	fmt.Printf("圓周率 π = %.10f\n", math.Pi)         // math.Pi：圓周率常數
	fmt.Printf("自然對數 e = %.10f\n", math.E)         // math.E：自然對數的底數
	fmt.Printf("√16 = %.0f\n", math.Sqrt(16))         // Sqrt：平方根，√16 = 4
	fmt.Printf("2³ = %.0f\n", math.Pow(2, 3))         // Pow：次方運算，2 的 3 次方 = 8
	fmt.Printf("|-7| = %.0f\n", math.Abs(-7))         // Abs：絕對值，|-7| = 7
	fmt.Printf("ceil(3.2) = %.0f\n", math.Ceil(3.2))  // Ceil：無條件進位，3.2 → 4
	fmt.Printf("floor(3.8) = %.0f\n", math.Floor(3.8)) // Floor：無條件捨去，3.8 → 3
	fmt.Printf("round(3.5) = %.0f\n", math.Round(3.5)) // Round：四捨五入，3.5 → 4
	fmt.Println() // 空行分隔

	// ========================================
	// 示範 7：匯入路徑的慣例
	// ========================================
	// 標準庫：直接用套件名稱
	//   "fmt"         → 格式化輸出
	//   "net/http"    → HTTP 伺服器和客戶端
	//   "os"          → 作業系統相關操作
	//
	// 自己專案的套件：用模組名稱 + 路徑
	//   "blog-api/internal/domain"     → 部落格專案的領域層
	//   "blog-api/internal/handler"    → 部落格專案的處理層
	//   "blog-api/pkg/config"          → 部落格專案的設定套件
	//
	// 第三方套件：用完整的 URL 路徑
	//   "github.com/gin-gonic/gin"     → Gin 網頁框架
	//   "gorm.io/gorm"                 → GORM 資料庫 ORM
	//   "github.com/golang-jwt/jwt/v5" → JWT 驗證套件
	fmt.Println("--- 7. 部落格專案的套件結構 ---") // 印出區段標題

	// 部落格專案的目錄結構（用字串印出來說明）
	fmt.Println("blog-api/")                            // 專案根目錄
	fmt.Println("├── cmd/")                             // cmd/ = 程式進入點
	fmt.Println("│   └── server/")                      // server/ = 伺服器程式
	fmt.Println("│       └── main.go")                  // main.go = 程式起點
	fmt.Println("├── internal/")                        // internal/ = 內部套件（只有本專案能用）
	fmt.Println("│   ├── domain/")                      // domain/ = 領域模型（Article, User 等）
	fmt.Println("│   │   ├── article.go")               // 文章相關的型別定義
	fmt.Println("│   │   └── user.go")                  // 使用者相關的型別定義
	fmt.Println("│   ├── handler/")                     // handler/ = HTTP 請求處理器
	fmt.Println("│   │   ├── article_handler.go")       // 處理文章相關的 API 請求
	fmt.Println("│   │   └── user_handler.go")          // 處理使用者相關的 API 請求
	fmt.Println("│   └── repository/")                  // repository/ = 資料存取層
	fmt.Println("│       ├── article_repository.go")    // 文章的資料庫操作
	fmt.Println("│       └── user_repository.go")       // 使用者的資料庫操作
	fmt.Println("├── pkg/")                             // pkg/ = 公開套件（其他專案也能用）
	fmt.Println("│   ├── config/")                      // config/ = 設定檔讀取
	fmt.Println("│   └── response/")                    // response/ = 統一回應格式
	fmt.Println("└── go.mod")                           // go.mod = 模組定義檔
	fmt.Println()                                       // 空行分隔

	// 說明 internal/ 和 pkg/ 的差別
	fmt.Println("internal/ vs pkg/ 的差別：")           // 印出比較標題
	fmt.Println("  internal/ → 只有 blog-api 這個模組內部能匯入") // internal 的限制
	fmt.Println("             Go 編譯器會強制執行這個規則！")     // 不是慣例，是強制的
	fmt.Println("  pkg/      → 任何外部專案都可以匯入使用")       // pkg 是公開的
	fmt.Println("             這是一個慣例（convention），不是強制的") // 只是慣例
	fmt.Println()                                       // 空行分隔

	// ========================================
	// 示範 8：綜合練習 — 用多個套件處理部落格資料
	// ========================================
	fmt.Println("--- 8. 綜合練習：處理部落格資料 ---") // 印出區段標題

	// 模擬從 HTTP 請求取得的查詢參數（都是字串）
	queryPage := "3"                                    // 模擬 ?page=3
	queryKeyword := "  go language  "                   // 模擬 ?keyword=  go language  （有多餘空格）

	// 用 strconv 轉換頁碼
	page, err := strconv.Atoi(queryPage)                // 把字串 "3" 轉成整數 3
	if err != nil {                                     // 如果轉換失敗
		page = 1 // 預設第 1 頁
	}

	// 用 strings 清理搜尋關鍵字
	keyword := strings.TrimSpace(queryKeyword)          // 去掉前後空格
	keyword = strings.ToLower(keyword)                  // 統一轉成小寫，方便搜尋比對

	// 用 fmt 格式化輸出查詢資訊
	fmt.Printf("查詢第 %d 頁，關鍵字: %q\n", page, keyword) // 印出處理後的查詢參數

	// 用 time 記錄查詢時間
	queryTime := time.Now()                             // 記錄查詢發生的時間
	fmt.Printf("查詢時間: %s\n", queryTime.Format("2006-01-02 15:04:05")) // 格式化印出

	// 模擬查詢結果
	articles := []string{                               // 建立一個文章標題的切片
		"Go 語言入門教學",                              // 第一篇文章
		"Go Web 開發實戰",                              // 第二篇文章
		"Go Language 基礎觀念",                         // 第三篇文章（英文 language）
	}

	// 用 strings 套件篩選包含關鍵字的文章
	fmt.Println("搜尋結果:")                             // 印出結果標題
	matchCount := 0                                     // 計數器：符合條件的文章數
	for i, article := range articles {                  // 遍歷所有文章
		if strings.Contains(strings.ToLower(article), keyword) { // 不分大小寫地搜尋
			fmt.Printf("  [%d] %s\n", i+1, article)    // 印出符合條件的文章（編號從 1 開始）
			matchCount++                                 // 符合數加 1
		}
	}
	fmt.Printf("共找到 %d 篇符合的文章\n", matchCount)  // 印出符合的總數

	// ========================================
	// 總結
	// ========================================
	fmt.Println()                                       // 空行分隔
	fmt.Println("===== 重點回顧 =====")                 // 印出回顧標題
	fmt.Println("1. 套件（Package）= 工具箱，把相關的程式碼放在一起")              // 重點 1
	fmt.Println("2. 模組（Module）= 工廠，由 go.mod 定義，包含多個套件")           // 重點 2
	fmt.Println("3. 大寫開頭 = 公開（Public），小寫開頭 = 私有（Private）")        // 重點 3
	fmt.Println("4. import 匯入套件，按照 標準庫 / 自己的 / 第三方 分組")          // 重點 4
	fmt.Println("5. internal/ 目錄下的套件只有本模組能用，Go 編譯器強制執行")       // 重點 5
	fmt.Println("6. go mod tidy 會自動整理依賴，go get 安裝第三方套件")            // 重點 6
}
