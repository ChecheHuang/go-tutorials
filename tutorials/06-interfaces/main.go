// 第六課：介面（Interfaces）
// 這是 Go 最強大的特性之一，也是 Clean Architecture 的基礎
// 執行方式：go run main.go
package main

import (
	"fmt"
	"math"
)

// ========================================
// 1. 定義介面
// ========================================

// Shape 定義了「形狀」的介面
// 任何擁有 Area() 和 Perimeter() 方法的型別都自動滿足這個介面
type Shape interface {
	Area() float64
	Perimeter() float64
}

// ========================================
// 2. 實作介面（隱式實作）
// ========================================

// Circle 圓形
type Circle struct {
	Radius float64
}

// Circle 實作了 Shape 介面（不需要寫 implements！）
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// Rectangle 矩形
type Rectangle struct {
	Width, Height float64
}

// Rectangle 也實作了 Shape 介面
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// ========================================
// 3. 使用介面（多型）
// ========================================

// printShapeInfo 接收任何 Shape 介面的實作
// 不管是圓形、矩形還是三角形，只要有 Area() 和 Perimeter() 就能用
func printShapeInfo(s Shape) {
	fmt.Printf("  面積: %.2f\n", s.Area())
	fmt.Printf("  周長: %.2f\n", s.Perimeter())
}

// ========================================
// 4. 實際範例：模擬 Clean Architecture 的 Repository 模式
// ========================================

// User 實體
type User struct {
	ID    int
	Name  string
	Email string
}

// UserRepository 介面（對應專案的 domain.UserRepository）
type UserRepository interface {
	Create(user *User) error
	FindByID(id int) (*User, error)
}

// --- 實作 1：記憶體儲存（用於測試）---

type MemoryUserRepository struct {
	users  map[int]*User
	nextID int
}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users:  make(map[int]*User),
		nextID: 1,
	}
}

func (r *MemoryUserRepository) Create(user *User) error {
	user.ID = r.nextID
	r.nextID++
	r.users[user.ID] = user
	fmt.Printf("  [Memory] 已儲存使用者：%s (ID=%d)\n", user.Name, user.ID)
	return nil
}

func (r *MemoryUserRepository) FindByID(id int) (*User, error) {
	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("使用者 ID=%d 不存在", id)
	}
	fmt.Printf("  [Memory] 找到使用者：%s\n", user.Name)
	return user, nil
}

// --- 實作 2：模擬資料庫儲存（用於生產）---

type DatabaseUserRepository struct {
	// 實際上會有 *gorm.DB，這裡簡化
}

func (r *DatabaseUserRepository) Create(user *User) error {
	fmt.Printf("  [Database] INSERT INTO users ... (%s)\n", user.Name)
	user.ID = 1
	return nil
}

func (r *DatabaseUserRepository) FindByID(id int) (*User, error) {
	fmt.Printf("  [Database] SELECT * FROM users WHERE id = %d\n", id)
	return &User{ID: id, Name: "來自資料庫的使用者"}, nil
}

// ========================================
// 5. UserService 依賴介面而非具體實作
// ========================================

// UserService 代表業務邏輯層（對應專案的 Usecase）
type UserService struct {
	repo UserRepository // 依賴介面！不是具體型別
}

// NewUserService 注入任何滿足 UserRepository 介面的實作
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) RegisterUser(name, email string) (*User, error) {
	user := &User{Name: name, Email: email}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetUser(id int) (*User, error) {
	return s.repo.FindByID(id)
}

// ========================================
// 6. 空介面 interface{} 和 any
// ========================================

// printAnything 接收任何型別的值
// interface{} 是 Go 1.18 之前的寫法
// any 是 Go 1.18 之後的別名（兩者完全等價）
func printAnything(value any) {
	fmt.Printf("  型別: %-10T 值: %v\n", value, value)
}

func main() {
	// ========================================
	// 示範 1：多型
	// ========================================
	fmt.Println("=== 介面與多型 ===")

	shapes := []Shape{
		Circle{Radius: 5},
		Rectangle{Width: 4, Height: 6},
	}

	for _, shape := range shapes {
		fmt.Printf("\n%T:\n", shape)
		printShapeInfo(shape) // 同一個函式，處理不同型別
	}

	// ========================================
	// 示範 2：依賴注入 —— 使用記憶體儲存
	// ========================================
	fmt.Println("\n=== 使用 MemoryRepository ===")

	memoryRepo := NewMemoryUserRepository()
	service := NewUserService(memoryRepo) // 注入記憶體實作

	service.RegisterUser("Alice", "alice@example.com")
	service.GetUser(1)

	// ========================================
	// 示範 3：依賴注入 —— 切換為資料庫儲存
	// ========================================
	fmt.Println("\n=== 使用 DatabaseRepository ===")

	dbRepo := &DatabaseUserRepository{}
	service2 := NewUserService(dbRepo) // 注入資料庫實作

	service2.RegisterUser("Bob", "bob@example.com")
	service2.GetUser(1)

	// UserService 的程式碼完全沒有改變！
	// 只是注入了不同的 Repository 實作

	// ========================================
	// 示範 4：空介面
	// ========================================
	fmt.Println("\n=== 空介面 (any) ===")

	printAnything(42)
	printAnything("hello")
	printAnything(true)
	printAnything(Circle{Radius: 3})

	// ========================================
	// 示範 5：型別斷言（Type Assertion）
	// ========================================
	fmt.Println("\n=== 型別斷言 ===")

	var something any = "Hello, Go!"

	// 型別斷言：從 any 中取出具體型別
	str, ok := something.(string)
	if ok {
		fmt.Println("是字串:", str)
	}

	// 型別 switch
	checkType := func(v any) {
		switch val := v.(type) {
		case int:
			fmt.Printf("  整數: %d\n", val)
		case string:
			fmt.Printf("  字串: %s\n", val)
		case bool:
			fmt.Printf("  布林: %t\n", val)
		default:
			fmt.Printf("  未知型別: %T\n", val)
		}
	}

	checkType(42)
	checkType("test")
	checkType(true)
	checkType(3.14)
}
