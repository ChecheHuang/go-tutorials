// ==========================================================================
// 第十課：架構設計（Clean Architecture）— 整個教學系列最重要的一課
// ==========================================================================
//
// 執行方式：go run ./tutorials/10-clean-architecture
//
// ╔══════════════════════════════════════════════════════════════════════╗
// ║                                                                    ║
// ║   用「餐廳」來理解 Clean Architecture                                ║
// ║                                                                    ║
// ║   想像你走進一家餐廳：                                               ║
// ║                                                                    ║
// ║   🧑‍🍳 廚師（Chef）= Usecase（業務邏輯層）                            ║
// ║      → 知道怎麼炒菜、調味、控制火候（業務規則）                        ║
// ║      → 不需要知道食材從哪個市場買的（不依賴資料庫）                     ║
// ║      → 不需要知道客人長什麼樣子（不依賴 HTTP）                         ║
// ║                                                                    ║
// ║   🍽️ 服務生（Waiter）= Handler（展示層）                              ║
// ║      → 接收客人的點餐（接收 HTTP 請求）                               ║
// ║      → 把菜單交給廚師（呼叫 Usecase）                                ║
// ║      → 把做好的菜端給客人（回傳 HTTP 回應）                           ║
// ║      → 不需要知道怎麼炒菜（不包含業務邏輯）                           ║
// ║                                                                    ║
// ║   📖 食譜（Recipe）= Domain（領域層）                                 ║
// ║      → 定義「什麼是紅燒肉」（業務實體 Entity）                        ║
// ║      → 定義「廚房需要有冰箱」（Repository 介面）                      ║
// ║      → 不管冰箱是什麼牌子（不依賴具體實作）                           ║
// ║                                                                    ║
// ║   🧊 冰箱（Fridge）= Repository（資料存取層）                         ║
// ║      → 實際存放和取出食材（資料庫的具體操作）                          ║
// ║      → 可以換成不同品牌的冰箱（可以換資料庫）                         ║
// ║      → 廚師只要求「給我食材」，不管冰箱怎麼運作                       ║
// ║                                                                    ║
// ╠══════════════════════════════════════════════════════════════════════╣
// ║                                                                    ║
// ║   Clean Architecture（整潔架構）                                     ║
// ║   由 Robert C. Martin（Uncle Bob）提出                               ║
// ║   核心原則：依賴方向只能從外向內，內層不知道外層的存在                  ║
// ║                                                                    ║
// ║   ┌───────────────────────────────────────┐                        ║
// ║   │  外層：Handler（框架、HTTP、UI）        │                        ║
// ║   │  ┌───────────────────────────────┐    │                        ║
// ║   │  │  中層：Usecase（業務邏輯）      │    │                        ║
// ║   │  │  ┌───────────────────────┐    │    │                        ║
// ║   │  │  │  內層：Domain（實體）   │    │    │                        ║
// ║   │  │  └───────────────────────┘    │    │                        ║
// ║   │  └───────────────────────────────┘    │                        ║
// ║   └───────────────────────────────────────┘                        ║
// ║                                                                    ║
// ║   依賴方向：Handler → Usecase → Domain ← Repository                ║
// ║                                                                    ║
// ║   「依賴注入」（Dependency Injection）：                              ║
// ║   不是讓廚師自己去買冰箱，而是餐廳老闆（main 函式）                   ║
// ║   買好冰箱後交給廚師使用。這樣換冰箱時，廚師完全不受影響。             ║
// ║                                                                    ║
// ╚══════════════════════════════════════════════════════════════════════╝
package main // 宣告這是主程式套件，Go 程式的進入點

import ( // 匯入需要的標準函式庫
	"errors"  // errors 套件：用來建立和處理錯誤
	"fmt"     // fmt 套件：用來格式化輸出文字到終端機
	"strings" // strings 套件：用來處理字串操作（如去空白）
	"time"    // time 套件：用來處理時間相關操作
)

// ====================================================================
//
//  第 1 層：Domain（領域層）— 最內層，最純粹
//
//  🍳 餐廳比喻：這是「食譜」
//  食譜定義了「紅燒肉需要哪些食材」（Entity）
//  也定義了「廚房必須有冰箱」（Repository 介面）
//  但食譜不管你用哪個牌子的冰箱
//
//  定義：業務實體（Entity）和資料操作介面（Repository Interface）
//  規則：不依賴任何外部套件，不知道資料庫、HTTP、框架的存在
//  作用：描述「業務世界長什麼樣子」
//
//  對應部落格專案：internal/domain/
//
//  ★ 為什麼介面要定義在 Domain 層？
//  這叫「依賴反轉原則」（Dependency Inversion Principle）
//  傳統做法：Usecase 直接依賴具體的資料庫 → 換資料庫就要改 Usecase
//  正確做法：Domain 定義介面，Usecase 依賴介面 → 換資料庫只要換實作
//  就像食譜寫「需要一台冰箱」，而不是寫「需要一台 LG 冰箱」
//
// ====================================================================

// Todo 是業務實體（Entity）— 描述「一個待辦事項是什麼」
// 就像食譜中定義「紅燒肉有哪些成分」
// 它只描述資料的結構，不關心怎麼儲存、怎麼傳輸
type Todo struct { // 定義一個名為 Todo 的結構體（struct）
	ID        int       // ID 欄位：待辦事項的唯一識別碼，型別是整數
	Title     string    // Title 欄位：待辦事項的標題，型別是字串
	Completed bool      // Completed 欄位：是否已完成，型別是布林值（true/false）
	UserID    int       // UserID 欄位：這個待辦事項屬於哪個使用者
	CreatedAt time.Time // CreatedAt 欄位：建立時間，使用 time.Time 型別
}

// TodoRepository 定義「我需要什麼資料操作能力」
// ★★★ 關鍵概念：這是「介面」（Interface），不是「實作」 ★★★
//
// 介面就像食譜寫著「需要一台冰箱，這台冰箱必須能：存放食材、取出食材」
// 不管你用哪個牌子的冰箱，只要能做到這些事就行
//
// Domain 層說：「我不管你用什麼資料庫（MySQL、SQLite、記憶體...），
//
//	但你必須提供以下這五個方法」
//
// 這就是「依賴反轉」— 不是高層去依賴低層的具體實作，
// 而是高層定義介面，低層來滿足這個介面
type TodoRepository interface { // 定義 TodoRepository 介面
	Create(todo *Todo) error                 // 建立一筆待辦事項，失敗回傳 error
	FindByID(id int) (*Todo, error)          // 根據 ID 查找，回傳指標和可能的錯誤
	FindByUserID(userID int) ([]Todo, error) // 根據使用者 ID 查找所有待辦事項
	Update(todo *Todo) error                 // 更新一筆待辦事項
	Delete(id int) error                     // 根據 ID 刪除一筆待辦事項
}

// ====================================================================
//
//  第 2 層：Usecase（業務邏輯層）— 中間層
//
//  🍳 餐廳比喻：這是「廚師」
//  廚師知道怎麼炒菜（業務規則），例如：
//  - 紅燒肉要先焯水（標題不能為空）
//  - 不能用過期食材（權限檢查）
//  - 火候要適中（長度限制）
//  但廚師不需要知道食材從哪裡來（不依賴資料庫）
//  也不需要知道客人是誰（不依賴 HTTP）
//
//  定義：具體的業務規則和流程
//  規則：只依賴 Domain 層的介面，不知道資料庫的具體實作
//  作用：回答「業務邏輯是什麼」
//
//  例如：
//  - 建立 Todo 時，標題不能為空
//  - 只有擁有者可以修改自己的 Todo
//  - 完成 Todo 時要記錄完成時間
//
//  對應部落格專案：internal/usecase/
//
// ====================================================================

// TodoUsecase 定義業務邏輯的介面
// 外層（Handler）會依賴這個介面，而不是具體的 struct
type TodoUsecase interface { // 定義 Usecase 的介面
	CreateTodo(userID int, title string) (*Todo, error) // 建立待辦事項的業務邏輯
	GetUserTodos(userID int) ([]Todo, error)            // 取得使用者所有待辦事項
	CompleteTodo(id, userID int) (*Todo, error)         // 完成一個待辦事項
	DeleteTodo(id, userID int) error                    // 刪除一個待辦事項
}

// todoUsecase 是 TodoUsecase 介面的「具體實作」
// 注意：名稱小寫開頭（todoUsecase），表示這是私有的，外部套件看不到
// 外部只能透過 TodoUsecase 介面來使用它
type todoUsecase struct { // 定義 Usecase 的結構體
	repo TodoRepository // ★ 依賴的是「介面」，不是具體實作
	// 這就像廚師說「我需要一台冰箱」，而不是「我需要一台 LG 冰箱」
	// 所以不管你給他什麼牌子的冰箱（什麼資料庫），廚師都能工作
}

// NewTodoUsecase 建立 Usecase 實例 — 這是一個「建構函式」
// ★★★ 依賴注入（Dependency Injection）的核心 ★★★
//
// 「依賴注入」的意思是：不是讓廚師自己去買冰箱，
// 而是餐廳老闆（main 函式）買好冰箱後「注入」給廚師
//
// 參數是 TodoRepository「介面」型別
// → 你可以注入任何實作了這個介面的東西：
//   - 記憶體 map（開發/測試用）
//   - GORM + SQLite（本地開發）
//   - GORM + PostgreSQL（正式環境）
//   - Mock 物件（單元測試用）
//
// 回傳的是 TodoUsecase「介面」型別，而不是 *todoUsecase
// → 外部只能用介面定義的方法，看不到內部實作細節
func NewTodoUsecase(repo TodoRepository) TodoUsecase { // 接收介面，回傳介面
	return &todoUsecase{repo: repo} // 把注入的 repo 儲存在結構體中
}

// CreateTodo 業務邏輯：建立待辦事項
// 這裡只處理「業務規則」，不碰 HTTP、不碰資料庫細節
func (u *todoUsecase) CreateTodo(userID int, title string) (*Todo, error) {
	// （u *todoUsecase）是「方法接收器」，表示這個方法屬於 todoUsecase

	// === 業務規則 ===
	// 規則 1：標題不能為空（就像廚師說「沒有食材我無法料理」）
	title = strings.TrimSpace(title) // 去除字串前後的空白字元
	if title == "" {                 // 如果去除空白後是空字串
		return nil, errors.New("標題不能為空") // 回傳 nil 和錯誤訊息
	}

	// 規則 2：標題長度限制（就像廚師說「食材太多放不下鍋子」）
	if len(title) > 200 { // 如果標題長度超過 200 個字元
		return nil, errors.New("標題不能超過 200 字") // 回傳錯誤
	}

	// 建立業務實體（準備食材）
	todo := &Todo{ // 建立一個 Todo 指標
		Title:     title,      // 設定標題
		Completed: false,      // 新建的待辦事項預設為未完成
		UserID:    userID,     // 記錄是哪個使用者建立的
		CreatedAt: time.Now(), // 記錄建立時間為「現在」
	}

	// 呼叫 Repository 儲存（叫冰箱幫忙存食材）
	// 注意：這裡呼叫的是「介面方法」，不關心具體是用什麼資料庫
	if err := u.repo.Create(todo); err != nil { // 如果儲存失敗
		return nil, fmt.Errorf("建立失敗: %w", err) // 用 %w 包裝原始錯誤
	}

	return todo, nil // 成功：回傳建立好的 Todo，錯誤為 nil
}

// GetUserTodos 業務邏輯：取得使用者的所有待辦事項
// 這個方法很簡單，直接委託給 Repository
// 但如果將來需要加業務規則（例如排序、過濾），就在這裡加
func (u *todoUsecase) GetUserTodos(userID int) ([]Todo, error) {
	return u.repo.FindByUserID(userID) // 直接呼叫 Repository 的方法
}

// CompleteTodo 業務邏輯：完成待辦事項
// 這個方法展示了多個業務規則的組合
func (u *todoUsecase) CompleteTodo(id, userID int) (*Todo, error) {
	// 第一步：從 Repository 取得這個待辦事項
	todo, err := u.repo.FindByID(id) // 根據 ID 查找
	if err != nil {                  // 如果查找失敗（不存在）
		return nil, errors.New("待辦事項不存在") // 回傳友善的錯誤訊息
	}

	// === 業務規則 ===
	// 規則 1：只有擁有者可以操作（權限檢查）
	// 就像餐廳規定「只有點餐的客人可以退餐」
	if todo.UserID != userID { // 如果操作者不是待辦事項的擁有者
		return nil, errors.New("無權限操作此待辦事項") // 拒絕操作
	}

	// 規則 2：已完成的不能重複完成
	// 就像已經端上桌的菜不能再做一次
	if todo.Completed { // 如果已經完成了
		return nil, errors.New("此待辦事項已經完成") // 回傳錯誤
	}

	// 通過所有規則檢查後，標記為完成
	todo.Completed = true // 將 Completed 設為 true

	// 儲存更新到 Repository
	if err := u.repo.Update(todo); err != nil { // 如果更新失敗
		return nil, fmt.Errorf("更新失敗: %w", err) // 包裝錯誤並回傳
	}

	return todo, nil // 成功：回傳更新後的 Todo
}

// DeleteTodo 業務邏輯：刪除待辦事項
func (u *todoUsecase) DeleteTodo(id, userID int) error {
	// 先查找待辦事項是否存在
	todo, err := u.repo.FindByID(id) // 根據 ID 查找
	if err != nil {                  // 如果不存在
		return errors.New("待辦事項不存在") // 回傳錯誤
	}

	// === 業務規則 ===
	// 規則：只有擁有者可以刪除（就像只有自己能取消自己的訂單）
	if todo.UserID != userID { // 如果操作者不是擁有者
		return errors.New("無權限刪除此待辦事項") // 拒絕操作
	}

	return u.repo.Delete(id) // 委託 Repository 執行刪除
}

// ====================================================================
//
//  第 3 層：Repository 實作（基礎設施層）— 外層
//
//  🍳 餐廳比喻：這是「冰箱」的具體品牌
//  Domain 層說「我需要一台冰箱」（定義介面）
//  Repository 實作說「我是一台 LG 冰箱，我能做到」（實作介面）
//
//  定義：Domain 層 Repository 介面的具體實作
//  規則：實作 Domain 定義的介面，依賴具體的儲存技術
//  作用：回答「資料怎麼存取」
//
//  這裡用 map 模擬記憶體資料庫，實際專案用 GORM + SQLite
//  重點是：不管用什麼技術，只要滿足 TodoRepository 介面就行
//
//  對應部落格專案：internal/repository/
//
// ====================================================================

// memoryTodoRepository 用記憶體（map）實作 TodoRepository 介面
// 這就是一台「記憶體牌冰箱」— 食材存在記憶體裡，程式結束就消失
// 實際專案會換成「GORM 牌冰箱」— 食材存在 SQLite 資料庫裡
type memoryTodoRepository struct { // 定義記憶體版的 Repository 結構體
	todos  map[int]*Todo // 用 map 儲存所有 Todo，key 是 ID，value 是 Todo 指標
	nextID int           // 下一個可用的 ID（模擬資料庫的自動遞增）
}

// NewMemoryTodoRepository 建立記憶體版的 Repository
// 回傳的是 TodoRepository「介面」型別，不是 *memoryTodoRepository
// → 外部不需要知道具體是用記憶體實作的
func NewMemoryTodoRepository() TodoRepository { // 回傳介面型別
	return &memoryTodoRepository{ // 建立並回傳結構體的指標
		todos:  make(map[int]*Todo), // 初始化一個空的 map
		nextID: 1,                   // ID 從 1 開始
	}
}

// Create 實作 TodoRepository 介面的 Create 方法
// 把食材放進冰箱（把 Todo 存進 map）
func (r *memoryTodoRepository) Create(todo *Todo) error {
	todo.ID = r.nextID      // 給 Todo 一個唯一的 ID
	r.nextID++              // 讓下一個 ID 加 1，確保不重複
	r.todos[todo.ID] = todo // 把 Todo 存進 map 裡
	return nil              // 回傳 nil 表示沒有錯誤
}

// FindByID 實作 TodoRepository 介面的 FindByID 方法
// 從冰箱裡找特定的食材
func (r *memoryTodoRepository) FindByID(id int) (*Todo, error) {
	todo, exists := r.todos[id] // 從 map 中查找，exists 表示是否找到
	if !exists {                // 如果沒找到
		return nil, errors.New("not found") // 回傳 nil 和錯誤
	}
	return todo, nil // 找到了，回傳 Todo 指標
}

// FindByUserID 實作 TodoRepository 介面的 FindByUserID 方法
// 從冰箱裡找出某個人的所有食材
func (r *memoryTodoRepository) FindByUserID(userID int) ([]Todo, error) {
	var result []Todo              // 宣告一個空的 Todo 切片來存結果
	for _, todo := range r.todos { // 遍歷 map 中所有的 Todo
		if todo.UserID == userID { // 如果這個 Todo 屬於指定的使用者
			result = append(result, *todo) // 加入結果切片（*todo 取值，不是指標）
		}
	}
	return result, nil // 回傳結果（可能是空切片，但不是錯誤）
}

// Update 實作 TodoRepository 介面的 Update 方法
// 替換冰箱裡的食材
func (r *memoryTodoRepository) Update(todo *Todo) error {
	r.todos[todo.ID] = todo // 直接用新的 Todo 覆蓋 map 中的舊值
	return nil              // 回傳 nil 表示成功
}

// Delete 實作 TodoRepository 介面的 Delete 方法
// 從冰箱裡丟掉食材
func (r *memoryTodoRepository) Delete(id int) error {
	delete(r.todos, id) // 使用 Go 內建的 delete 函式從 map 中刪除
	return nil          // 回傳 nil 表示成功
}

// ====================================================================
//
//  第 4 層：Handler（展示層 / 控制層）— 最外層
//
//  🍳 餐廳比喻：這是「服務生」
//  服務生做三件事：
//  1. 接收客人的點餐（接收請求）
//  2. 把菜單交給廚師（呼叫 Usecase）
//  3. 把做好的菜端給客人（回傳回應）
//  服務生不需要會炒菜，廚師不需要會端盤子
//
//  定義：接收外部輸入，呼叫 Usecase，回傳輸出
//  規則：只依賴 Usecase 介面
//  作用：回答「使用者怎麼跟系統互動」
//
//  這裡用 CLI（命令列）模擬，實際專案用 Gin HTTP Handler
//  就像服務生可以是現場服務生、外送員、電話接單員...
//  不管是哪種方式，廚師的炒菜方法都不需要改
//
//  對應部落格專案：internal/handler/
//
// ====================================================================

// CLIHandler 用命令列介面模擬 HTTP Handler
// 在真實專案中，這會是 Gin 的 HTTP Handler
type CLIHandler struct { // 定義 CLI Handler 結構體
	usecase TodoUsecase // ★ 依賴的是 Usecase「介面」，不是具體實作
	// 服務生只需要知道「廚師會做菜」，不需要知道廚師是誰
}

// NewCLIHandler 建立 CLIHandler — 同樣使用依賴注入
func NewCLIHandler(usecase TodoUsecase) *CLIHandler { // 接收 Usecase 介面
	return &CLIHandler{usecase: usecase} // 把注入的 usecase 儲存起來
}

// HandleCreateTodo 處理「建立待辦事項」的請求
// 在真實專案中，這裡會解析 HTTP POST 請求的 JSON body
func (h *CLIHandler) HandleCreateTodo(userID int, title string) {
	// 步驟 1：顯示收到的請求（模擬接收 HTTP 請求）
	fmt.Printf("\n[操作] 使用者 %d 建立 Todo: \"%s\"\n", userID, title)

	// 步驟 2：呼叫 Usecase 處理業務邏輯（把點餐交給廚師）
	todo, err := h.usecase.CreateTodo(userID, title) // 委託給 Usecase
	if err != nil {                                  // 如果業務邏輯回傳錯誤
		fmt.Printf("[錯誤] %s\n", err) // 顯示錯誤訊息（模擬回傳 HTTP 400）
		return                       // 提早返回，不繼續執行
	}

	// 步驟 3：顯示成功結果（模擬回傳 HTTP 201 Created）
	fmt.Printf("[成功] 建立了 Todo #%d: %s\n", todo.ID, todo.Title)
}

// HandleCompleteTodo 處理「完成待辦事項」的請求
func (h *CLIHandler) HandleCompleteTodo(id, userID int) {
	fmt.Printf("\n[操作] 使用者 %d 完成 Todo #%d\n", userID, id) // 顯示操作資訊

	todo, err := h.usecase.CompleteTodo(id, userID) // 委託給 Usecase 處理
	if err != nil {                                 // 如果有錯誤
		fmt.Printf("[錯誤] %s\n", err) // 顯示錯誤
		return                       // 提早返回
	}

	fmt.Printf("[成功] Todo #%d \"%s\" 已完成 ✓\n", todo.ID, todo.Title) // 顯示成功
}

// HandleDeleteTodo 處理「刪除待辦事項」的請求
func (h *CLIHandler) HandleDeleteTodo(id, userID int) {
	fmt.Printf("\n[操作] 使用者 %d 刪除 Todo #%d\n", userID, id) // 顯示操作資訊

	err := h.usecase.DeleteTodo(id, userID) // 委託給 Usecase 處理
	if err != nil {                         // 如果有錯誤
		fmt.Printf("[錯誤] %s\n", err) // 顯示錯誤
		return                       // 提早返回
	}

	fmt.Printf("[成功] Todo #%d 已刪除\n", id) // 顯示成功
}

// HandleListTodos 處理「查看所有待辦事項」的請求
func (h *CLIHandler) HandleListTodos(userID int) {
	fmt.Printf("\n[操作] 查看使用者 %d 的所有 Todo\n", userID) // 顯示操作資訊

	todos, err := h.usecase.GetUserTodos(userID) // 委託給 Usecase 取得資料
	if err != nil {                              // 如果有錯誤
		fmt.Printf("[錯誤] %s\n", err) // 顯示錯誤
		return                       // 提早返回
	}

	if len(todos) == 0 { // 如果沒有任何待辦事項
		fmt.Println("[資訊] 沒有待辦事項") // 顯示提示訊息
		return                     // 提早返回
	}

	for _, todo := range todos { // 遍歷所有待辦事項
		status := "[ ]"     // 預設顯示未完成的符號
		if todo.Completed { // 如果已完成
			status = "[v]" // 改為完成的符號
		}
		// 印出每一個待辦事項的狀態、ID 和標題
		fmt.Printf("  %s #%d: %s\n", status, todo.ID, todo.Title)
	}
}

// ====================================================================
//
//  主程式：依賴注入（Dependency Injection）— 餐廳老闆的工作
//
//  🍳 餐廳比喻：main 函式就是「餐廳老闆」
//  老闆的工作是：
//  1. 買冰箱（建立 Repository）
//  2. 請廚師，把冰箱交給他（建立 Usecase，注入 Repository）
//  3. 請服務生，告訴他廚師是誰（建立 Handler，注入 Usecase）
//  4. 開門營業（開始接收請求）
//
//  這是「唯一知道所有具體實作」的地方
//  其他每一層都只知道自己需要的介面
//
//  對應部落格專案：cmd/server/main.go
//
// ====================================================================

func main() { // Go 程式的進入點，程式從這裡開始執行
	// 印出標題
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║  Clean Architecture 完整示範                   ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")

	// ==========================================
	// 依賴注入：由內而外組裝（老闆開店的步驟）
	// ==========================================

	// 第 1 步：建立最內層 — Repository（資料存取）
	// 🧊 老闆買冰箱：這裡選用「記憶體牌冰箱」
	// 如果想換成「SQLite 牌冰箱」，只要改這一行
	// 後面的 Usecase 和 Handler 完全不需要改動！
	repo := NewMemoryTodoRepository() // 建立記憶體版的 Repository

	// 第 2 步：建立中間層 — Usecase（業務邏輯）
	// 🧑‍🍳 老闆請廚師，把冰箱交給他（注入 Repository）
	// 廚師只知道「我有一台冰箱可以用」，不知道是什麼牌子
	usecase := NewTodoUsecase(repo) // 建立 Usecase，注入 Repository

	// 第 3 步：建立最外層 — Handler（展示層）
	// 🍽️ 老闆請服務生，告訴他廚師是誰（注入 Usecase）
	// 服務生只知道「有廚師可以接單」，不知道廚師用什麼冰箱
	handler := NewCLIHandler(usecase) // 建立 Handler，注入 Usecase

	// ==========================================
	// 模擬使用者操作（餐廳開門營業！）
	// ==========================================

	fmt.Println("\n========== 情境 1：正常的 CRUD 操作 ==========")

	// 使用者 1 建立幾個 Todo（客人點了三道菜）
	handler.HandleCreateTodo(1, "學習 Go 語言基礎") // 建立第一個待辦事項
	handler.HandleCreateTodo(1, "學習 Gin 框架")  // 建立第二個待辦事項
	handler.HandleCreateTodo(1, "完成部落格 API")  // 建立第三個待辦事項

	// 查看使用者 1 的 Todo（客人問「我點了什麼？」）
	handler.HandleListTodos(1) // 列出所有待辦事項

	// 完成一個 Todo（客人說「第一道菜吃完了」）
	handler.HandleCompleteTodo(1, 1) // 使用者 1 完成 Todo #1

	// 查看更新後的列表（確認狀態已更新）
	handler.HandleListTodos(1) // 再次列出，可以看到 Todo #1 變成 [v]

	fmt.Println("\n========== 情境 2：業務規則驗證 ==========")

	// 空標題 → 業務規則拒絕（廚師說「你沒告訴我要什麼菜」）
	handler.HandleCreateTodo(1, "") // 標題為空，會被 Usecase 的規則擋下

	// 重複完成 → 業務規則拒絕（廚師說「這道菜已經做完了」）
	handler.HandleCompleteTodo(1, 1) // Todo #1 已經完成過了

	fmt.Println("\n========== 情境 3：權限控制 ==========")

	// 使用者 2 嘗試完成使用者 1 的 Todo（別桌的客人想吃你的菜）
	handler.HandleCompleteTodo(2, 2) // 使用者 2 沒有權限操作 Todo #2

	// 使用者 2 嘗試刪除使用者 1 的 Todo（別桌的客人想退你的菜）
	handler.HandleDeleteTodo(3, 2) // 使用者 2 沒有權限刪除 Todo #3

	// 使用者 1 自己刪除（客人自己取消訂單）
	handler.HandleDeleteTodo(3, 1) // 使用者 1 可以刪除自己的 Todo

	// 確認刪除結果
	handler.HandleListTodos(1) // 確認 Todo #3 已被刪除

	// ==========================================
	// 展示架構的可替換性
	// ==========================================

	fmt.Println("\n========== 架構的好處 ==========")
	fmt.Println()
	fmt.Println("以上程式完全沒有使用資料庫和 HTTP，但結構和部落格專案一模一樣：")
	fmt.Println()
	fmt.Println("  本課範例（餐廳模擬）          部落格專案（真正的餐廳）")
	fmt.Println("  ──────────────────────────  ──────────────────────────")
	fmt.Println("  Todo (struct)           →   domain.Article (struct)")
	fmt.Println("  TodoRepository (介面)   →   domain.ArticleRepository (介面)")
	fmt.Println("  memoryTodoRepository    →   repository.articleRepository (GORM)")
	fmt.Println("  todoUsecase             →   usecase.articleUsecase")
	fmt.Println("  CLIHandler              →   handler.ArticleHandler (Gin)")
	fmt.Println("  main() 組裝             →   cmd/server/main.go 組裝")
	fmt.Println()
	fmt.Println("想換資料庫？只要寫一個新的 Repository 實作（換一台冰箱）")
	fmt.Println("想換框架？  只要寫一個新的 Handler（換一個服務生）")
	fmt.Println("業務邏輯？  永遠不需要改（廚師的手藝不變）")
}
