// 第十課：架構設計（Clean Architecture）
// 這是整個部落格專案最核心的設計思想
// 本課用一個完整的迷你專案展示 Clean Architecture 的每一層
//
// 執行方式：go run main.go
package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ╔══════════════════════════════════════════════════════════════╗
// ║                                                            ║
// ║   Clean Architecture（整潔架構）                             ║
// ║                                                            ║
// ║   由 Robert C. Martin（Uncle Bob）提出                       ║
// ║   核心原則：依賴方向只能從外向內，內層不知道外層的存在            ║
// ║                                                            ║
// ║   ┌───────────────────────────────────────┐                ║
// ║   │  外層：Handler（框架、HTTP、UI）        │                ║
// ║   │  ┌───────────────────────────────┐    │                ║
// ║   │  │  中層：Usecase（業務邏輯）      │    │                ║
// ║   │  │  ┌───────────────────────┐    │    │                ║
// ║   │  │  │  內層：Domain（實體）   │    │    │                ║
// ║   │  │  └───────────────────────┘    │    │                ║
// ║   │  └───────────────────────────────┘    │                ║
// ║   └───────────────────────────────────────┘                ║
// ║                                                            ║
// ║   依賴方向：Handler → Usecase → Domain ← Repository        ║
// ║                                                            ║
// ╚══════════════════════════════════════════════════════════════╝

// ====================================================================
//
//  第 1 層：Domain（領域層）— 最內層，最純粹
//
//  定義：業務實體（Entity）和資料操作介面（Repository Interface）
//  規則：不依賴任何外部套件，不知道資料庫、HTTP、框架的存在
//  作用：描述「業務世界長什麼樣子」
//
//  對應部落格專案：internal/domain/
//
// ====================================================================

// Todo 是業務實體（Entity）
// 它只描述「一個待辦事項是什麼」，不關心怎麼儲存、怎麼傳輸
type Todo struct {
	ID        int
	Title     string
	Completed bool
	UserID    int
	CreatedAt time.Time
}

// TodoRepository 定義「我需要什麼資料操作能力」
// 注意：這是介面（Interface），不是實作
// Domain 層說：「我不管你用什麼資料庫，但你必須提供這些方法」
type TodoRepository interface {
	Create(todo *Todo) error
	FindByID(id int) (*Todo, error)
	FindByUserID(userID int) ([]Todo, error)
	Update(todo *Todo) error
	Delete(id int) error
}

// ====================================================================
//
//  第 2 層：Usecase（業務邏輯層）— 中間層
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
type TodoUsecase interface {
	CreateTodo(userID int, title string) (*Todo, error)
	GetUserTodos(userID int) ([]Todo, error)
	CompleteTodo(id, userID int) (*Todo, error)
	DeleteTodo(id, userID int) error
}

// todoUsecase 實作 TodoUsecase
type todoUsecase struct {
	repo TodoRepository // 依賴的是「介面」，不是具體實作
}

// NewTodoUsecase 建立 Usecase 實例
// 參數是介面型別 → 可以注入任何實作（真實 DB、Mock、記憶體...）
func NewTodoUsecase(repo TodoRepository) TodoUsecase {
	return &todoUsecase{repo: repo}
}

// CreateTodo 業務邏輯：建立待辦事項
func (u *todoUsecase) CreateTodo(userID int, title string) (*Todo, error) {
	// === 業務規則 ===
	// 規則 1：標題不能為空
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("標題不能為空")
	}

	// 規則 2：標題長度限制
	if len(title) > 200 {
		return nil, errors.New("標題不能超過 200 字")
	}

	// 建立實體
	todo := &Todo{
		Title:     title,
		Completed: false,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	// 呼叫 Repository 儲存（不關心用什麼資料庫）
	if err := u.repo.Create(todo); err != nil {
		return nil, fmt.Errorf("建立失敗: %w", err)
	}

	return todo, nil
}

// GetUserTodos 業務邏輯：取得使用者的所有待辦事項
func (u *todoUsecase) GetUserTodos(userID int) ([]Todo, error) {
	return u.repo.FindByUserID(userID)
}

// CompleteTodo 業務邏輯：完成待辦事項
func (u *todoUsecase) CompleteTodo(id, userID int) (*Todo, error) {
	// 取得 Todo
	todo, err := u.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("待辦事項不存在")
	}

	// === 業務規則 ===
	// 規則：只有擁有者可以操作
	if todo.UserID != userID {
		return nil, errors.New("無權限操作此待辦事項")
	}

	// 規則：已完成的不能重複完成
	if todo.Completed {
		return nil, errors.New("此待辦事項已經完成")
	}

	// 標記為完成
	todo.Completed = true

	if err := u.repo.Update(todo); err != nil {
		return nil, fmt.Errorf("更新失敗: %w", err)
	}

	return todo, nil
}

// DeleteTodo 業務邏輯：刪除待辦事項
func (u *todoUsecase) DeleteTodo(id, userID int) error {
	todo, err := u.repo.FindByID(id)
	if err != nil {
		return errors.New("待辦事項不存在")
	}

	// === 業務規則 ===
	// 規則：只有擁有者可以刪除
	if todo.UserID != userID {
		return errors.New("無權限刪除此待辦事項")
	}

	return u.repo.Delete(id)
}

// ====================================================================
//
//  第 3 層：Repository 實作 — 外層
//
//  定義：Domain 層 Repository 介面的具體實作
//  規則：實作 Domain 定義的介面，依賴具體的儲存技術
//  作用：回答「資料怎麼存取」
//
//  這裡用 map 模擬，實際專案用 GORM + SQLite
//
//  對應部落格專案：internal/repository/
//
// ====================================================================

// memoryTodoRepository 用記憶體（map）實作 TodoRepository
type memoryTodoRepository struct {
	todos  map[int]*Todo
	nextID int
}

// NewMemoryTodoRepository 建立記憶體 Repository
func NewMemoryTodoRepository() TodoRepository {
	return &memoryTodoRepository{
		todos:  make(map[int]*Todo),
		nextID: 1,
	}
}

func (r *memoryTodoRepository) Create(todo *Todo) error {
	todo.ID = r.nextID
	r.nextID++
	r.todos[todo.ID] = todo
	return nil
}

func (r *memoryTodoRepository) FindByID(id int) (*Todo, error) {
	todo, exists := r.todos[id]
	if !exists {
		return nil, errors.New("not found")
	}
	return todo, nil
}

func (r *memoryTodoRepository) FindByUserID(userID int) ([]Todo, error) {
	var result []Todo
	for _, todo := range r.todos {
		if todo.UserID == userID {
			result = append(result, *todo)
		}
	}
	return result, nil
}

func (r *memoryTodoRepository) Update(todo *Todo) error {
	r.todos[todo.ID] = todo
	return nil
}

func (r *memoryTodoRepository) Delete(id int) error {
	delete(r.todos, id)
	return nil
}

// ====================================================================
//
//  第 4 層：Handler（展示層 / 控制層）— 最外層
//
//  定義：接收外部輸入，呼叫 Usecase，回傳輸出
//  規則：只依賴 Usecase 介面
//  作用：回答「使用者怎麼跟系統互動」
//
//  這裡用 CLI 模擬，實際專案用 Gin HTTP Handler
//
//  對應部落格專案：internal/handler/
//
// ====================================================================

// CLIHandler 用命令列介面模擬 HTTP Handler
type CLIHandler struct {
	usecase TodoUsecase // 依賴 Usecase 介面
}

func NewCLIHandler(usecase TodoUsecase) *CLIHandler {
	return &CLIHandler{usecase: usecase}
}

func (h *CLIHandler) HandleCreateTodo(userID int, title string) {
	fmt.Printf("\n[操作] 使用者 %d 建立 Todo: \"%s\"\n", userID, title)

	todo, err := h.usecase.CreateTodo(userID, title)
	if err != nil {
		fmt.Printf("[錯誤] %s\n", err)
		return
	}

	fmt.Printf("[成功] 建立了 Todo #%d: %s\n", todo.ID, todo.Title)
}

func (h *CLIHandler) HandleCompleteTodo(id, userID int) {
	fmt.Printf("\n[操作] 使用者 %d 完成 Todo #%d\n", userID, id)

	todo, err := h.usecase.CompleteTodo(id, userID)
	if err != nil {
		fmt.Printf("[錯誤] %s\n", err)
		return
	}

	fmt.Printf("[成功] Todo #%d \"%s\" 已完成 ✓\n", todo.ID, todo.Title)
}

func (h *CLIHandler) HandleDeleteTodo(id, userID int) {
	fmt.Printf("\n[操作] 使用者 %d 刪除 Todo #%d\n", userID, id)

	err := h.usecase.DeleteTodo(id, userID)
	if err != nil {
		fmt.Printf("[錯誤] %s\n", err)
		return
	}

	fmt.Printf("[成功] Todo #%d 已刪除\n", id)
}

func (h *CLIHandler) HandleListTodos(userID int) {
	fmt.Printf("\n[操作] 查看使用者 %d 的所有 Todo\n", userID)

	todos, err := h.usecase.GetUserTodos(userID)
	if err != nil {
		fmt.Printf("[錯誤] %s\n", err)
		return
	}

	if len(todos) == 0 {
		fmt.Println("[資訊] 沒有待辦事項")
		return
	}

	for _, todo := range todos {
		status := "[ ]"
		if todo.Completed {
			status = "[v]"
		}
		fmt.Printf("  %s #%d: %s\n", status, todo.ID, todo.Title)
	}
}

// ====================================================================
//
//  主程式：依賴注入（Dependency Injection）
//
//  在 main 中把所有層組裝起來
//  這是唯一知道「所有具體實作」的地方
//
//  對應部落格專案：cmd/server/main.go
//
// ====================================================================

func main() {
	fmt.Println("╔═══════════════════════════════════════════════╗")
	fmt.Println("║  Clean Architecture 完整示範                   ║")
	fmt.Println("╚═══════════════════════════════════════════════╝")

	// ==========================================
	// 依賴注入：由內而外組裝
	// ==========================================

	// 第 1 步：建立最內層 — Repository（資料存取）
	repo := NewMemoryTodoRepository()

	// 第 2 步：建立中間層 — Usecase（業務邏輯），注入 Repository
	usecase := NewTodoUsecase(repo)

	// 第 3 步：建立最外層 — Handler（展示層），注入 Usecase
	handler := NewCLIHandler(usecase)

	// ==========================================
	// 模擬使用者操作
	// ==========================================

	fmt.Println("\n========== 情境 1：正常的 CRUD 操作 ==========")

	// 使用者 1 建立幾個 Todo
	handler.HandleCreateTodo(1, "學習 Go 語言基礎")
	handler.HandleCreateTodo(1, "學習 Gin 框架")
	handler.HandleCreateTodo(1, "完成部落格 API")

	// 查看使用者 1 的 Todo
	handler.HandleListTodos(1)

	// 完成一個 Todo
	handler.HandleCompleteTodo(1, 1) // 使用者 1 完成 Todo #1

	// 查看更新後的列表
	handler.HandleListTodos(1)

	fmt.Println("\n========== 情境 2：業務規則驗證 ==========")

	// 空標題
	handler.HandleCreateTodo(1, "")

	// 重複完成
	handler.HandleCompleteTodo(1, 1) // Todo #1 已經完成過了

	fmt.Println("\n========== 情境 3：權限控制 ==========")

	// 使用者 2 嘗試完成使用者 1 的 Todo
	handler.HandleCompleteTodo(2, 2) // 使用者 2 沒有權限操作 Todo #2

	// 使用者 2 嘗試刪除使用者 1 的 Todo
	handler.HandleDeleteTodo(3, 2) // 使用者 2 沒有權限刪除 Todo #3

	// 使用者 1 自己刪除
	handler.HandleDeleteTodo(3, 1) // 使用者 1 可以刪除自己的 Todo

	handler.HandleListTodos(1) // 確認 Todo #3 已被刪除

	// ==========================================
	// 展示架構的可替換性
	// ==========================================

	fmt.Println("\n========== 架構的好處 ==========")
	fmt.Println()
	fmt.Println("以上程式完全沒有使用資料庫和 HTTP，但結構和部落格專案一模一樣：")
	fmt.Println()
	fmt.Println("  本課範例                    部落格專案")
	fmt.Println("  ──────────────────────────  ──────────────────────────")
	fmt.Println("  Todo (struct)           →   domain.Article (struct)")
	fmt.Println("  TodoRepository (介面)   →   domain.ArticleRepository (介面)")
	fmt.Println("  memoryTodoRepository    →   repository.articleRepository (GORM)")
	fmt.Println("  todoUsecase             →   usecase.articleUsecase")
	fmt.Println("  CLIHandler              →   handler.ArticleHandler (Gin)")
	fmt.Println("  main() 組裝             →   cmd/server/main.go 組裝")
	fmt.Println()
	fmt.Println("想換資料庫？只要寫一個新的 Repository 實作")
	fmt.Println("想換框架？  只要寫一個新的 Handler")
	fmt.Println("業務邏輯？  永遠不需要改")
}
