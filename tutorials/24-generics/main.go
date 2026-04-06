// ==========================================================================
// 第二十四課：Go 泛型（Generics）
// ==========================================================================
//
// Go 1.18 加入泛型（2022年3月），解決重複程式碼問題：
//
//   // 沒有泛型之前
//   func SumInts(nums []int) int { ... }
//   func SumFloats(nums []float64) float64 { ... }
//   func SumInt64s(nums []int64) int64 { ... }
//
//   // 有泛型之後
//   func Sum[T int | float64 | int64](nums []T) T { ... }
//
// 泛型的三個核心概念：
//   1. 類型參數（Type Parameters）：func Fn[T any](v T) T
//   2. 類型約束（Type Constraints）：限制 T 必須滿足的條件
//   3. 類型推斷（Type Inference）：編譯器自動推斷 T 是什麼
//
// 執行方式：go run ./tutorials/31-generics
// ==========================================================================

package main

import (
	"cmp" // Go 1.21：cmp.Ordered（可排序的類型）
	"fmt"
	"slices" // Go 1.21：切片泛型工具
	"strings"
)

// ==========================================================================
// 1. 基本泛型函式
// ==========================================================================

// Min 回傳兩個可比較值中的較小者
// [T cmp.Ordered]：T 必須是「可排序的」（int, float64, string 等）
// cmp.Ordered = ~int | ~int8 | ... | ~float64 | ~string
func Min[T cmp.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// Max 回傳兩個可比較值中的較大者
func Max[T cmp.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Contains 檢查 slice 是否包含某個元素
// [T comparable]：T 必須支援 == 運算子
func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Map 把 []T 轉換成 []U（函式式程式設計的 map）
// [T, U any]：兩個獨立的類型參數
func Map[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// Filter 過濾 slice，只保留符合條件的元素
func Filter[T any](slice []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Reduce 把 slice 折疊成單一值
func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, v := range slice {
		result = fn(result, v)
	}
	return result
}

// ==========================================================================
// 2. 泛型資料結構
// ==========================================================================

// Stack 泛型堆疊（Last-In-First-Out）
type Stack[T any] struct {
	items []T
}

func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T // 回傳零值（int→0, string→"", 指標→nil）
		return zero, false
	}
	last := len(s.items) - 1
	item := s.items[last]
	s.items = s.items[:last]
	return item, true
}

func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Size() int { return len(s.items) }

// Pair 泛型鍵值對
type Pair[K, V any] struct {
	Key   K
	Value V
}

// NewPair 建立 Pair（類型推斷：編譯器自動推斷 K 和 V）
func NewPair[K, V any](key K, value V) Pair[K, V] {
	return Pair[K, V]{Key: key, Value: value}
}

// Result 泛型結果類型（類似 Rust 的 Result<T, E>）
type Result[T any] struct {
	value T
	err   error
}

func Ok[T any](value T) Result[T]    { return Result[T]{value: value} }
func Err[T any](err error) Result[T] { return Result[T]{err: err} }

func (r Result[T]) IsOk() bool { return r.err == nil }
func (r Result[T]) Unwrap() T {
	if r.err != nil {
		panic(fmt.Sprintf("called Unwrap on Err: %v", r.err))
	}
	return r.value
}
func (r Result[T]) UnwrapOr(defaultValue T) T {
	if r.err != nil {
		return defaultValue
	}
	return r.value
}

// ==========================================================================
// 3. 類型約束（Type Constraints）
// ==========================================================================

// Number 自訂類型約束：所有數值類型
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}

// Sum 計算數值切片的總和
// ~ 符號代表「底層類型（underlying type）」
// type MyInt int → MyInt 的底層類型是 int，~int 包含 MyInt
func Sum[T Number](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

// Average 計算平均值（回傳 float64）
func Average[T Number](nums []T) float64 {
	if len(nums) == 0 {
		return 0
	}
	total := Sum(nums)
	return float64(total) / float64(len(nums))
}

// Stringer 自訂約束：必須有 String() string 方法
type Stringer interface {
	String() string
}

// PrintAll 印出任何有 String() 方法的物件切片
func PrintAll[T Stringer](items []T) {
	for _, item := range items {
		fmt.Println(" ", item.String())
	}
}

// ==========================================================================
// 4. 泛型與介面的比較
// ==========================================================================

// 用介面實作（執行時決定類型）
type Sorter interface {
	Len() int
	Less(i, j int) bool
	Swap(i, j int)
}

// 用泛型實作（編譯時確定類型，更快、更安全）
func SortSlice[T cmp.Ordered](s []T) []T {
	result := make([]T, len(s))
	copy(result, s)
	slices.Sort(result) // Go 1.21 標準庫泛型函式
	return result
}

// Keys 取得 map 的所有 key（以前要手動寫，現在可以泛型化）
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values 取得 map 的所有 value
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// GroupBy 按照 key 函式分組
func GroupBy[T any, K comparable](slice []T, keyFn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range slice {
		key := keyFn(item)
		result[key] = append(result[key], item)
	}
	return result
}

// ==========================================================================
// 5. 實際使用範例
// ==========================================================================

// User 示範用的結構
type User struct {
	ID   int
	Name string
	Age  int
	Role string
}

func (u User) String() string {
	return fmt.Sprintf("User{ID:%d, Name:%s, Age:%d}", u.ID, u.Name, u.Age)
}

// ==========================================================================
// 主程式
// ==========================================================================

func main() {
	fmt.Println("==========================================")
	fmt.Println(" 第三十一課：Go 泛型（Generics）")
	fmt.Println("==========================================")

	// ──── 1. 基本泛型函式 ────
	fmt.Println("\n=== 1. 基本泛型函式 ===\n")

	fmt.Printf("Min(3, 5)        = %d\n", Min(3, 5))
	fmt.Printf("Min(3.14, 2.71)  = %.2f\n", Min(3.14, 2.71))
	fmt.Printf("Min(\"apple\", \"banana\") = %q\n", Min("apple", "banana"))

	fmt.Printf("Max(10, 20)      = %d\n", Max(10, 20))

	nums := []int{1, 2, 3, 4, 5}
	fmt.Printf("Contains(%v, 3)  = %v\n", nums, Contains(nums, 3))
	fmt.Printf("Contains(%v, 9)  = %v\n", nums, Contains(nums, 9))

	// ──── 2. 函式式操作 ────
	fmt.Println("\n=== 2. 函式式操作（Map/Filter/Reduce）===\n")

	doubled := Map(nums, func(n int) int { return n * 2 })
	fmt.Printf("Map(double): %v\n", doubled)

	asStrings := Map(nums, func(n int) string { return fmt.Sprintf("#%d", n) })
	fmt.Printf("Map(toStr):  %v\n", asStrings)

	evens := Filter(nums, func(n int) bool { return n%2 == 0 })
	fmt.Printf("Filter(even): %v\n", evens)

	total := Reduce(nums, 0, func(acc, n int) int { return acc + n })
	fmt.Printf("Reduce(sum): %d\n", total)

	// 鏈式操作
	result := Reduce(
		Filter(
			Map(nums, func(n int) int { return n * n }), // 平方
			func(n int) bool { return n > 5 },           // 過濾 > 5
		),
		0,
		func(acc, n int) int { return acc + n }, // 加總
	)
	fmt.Printf("Map²→Filter>5→Sum: %d\n", result) // 9+16+25 = 50

	// ──── 3. 泛型資料結構 ────
	fmt.Println("\n=== 3. 泛型資料結構 ===\n")

	// Stack
	var s Stack[string]
	s.Push("first")
	s.Push("second")
	s.Push("third")

	for s.Size() > 0 {
		if item, ok := s.Pop(); ok {
			fmt.Printf("Pop: %q\n", item)
		}
	}

	// Pair
	p := NewPair("userId", 42)
	fmt.Printf("Pair: %v = %v\n", p.Key, p.Value)

	// Result
	parseAge := func(s string) Result[int] {
		var n int
		if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
			return Err[int](fmt.Errorf("無效的年齡: %q", s))
		}
		return Ok(n)
	}

	fmt.Printf("parseAge(\"25\"): %d\n", parseAge("25").Unwrap())
	fmt.Printf("parseAge(\"abc\"): %d (預設值)\n", parseAge("abc").UnwrapOr(-1))

	// ──── 4. 類型約束 ────
	fmt.Println("\n=== 4. 類型約束（Number）===\n")

	ints := []int{1, 2, 3, 4, 5}
	floats := []float64{1.1, 2.2, 3.3}

	fmt.Printf("Sum(ints):     %d\n", Sum(ints))
	fmt.Printf("Sum(floats):   %.1f\n", Sum(floats))
	fmt.Printf("Average(ints): %.2f\n", Average(ints))

	// ──── 5. 實際應用：泛型工具函式 ────
	fmt.Println("\n=== 5. 實際應用：使用者資料處理 ===\n")

	users := []User{
		{1, "Alice", 30, "admin"},
		{2, "Bob", 25, "user"},
		{3, "Charlie", 35, "admin"},
		{4, "David", 28, "user"},
		{5, "Eve", 22, "moderator"},
	}

	// 取得所有姓名
	names := Map(users, func(u User) string { return u.Name })
	fmt.Printf("姓名: %v\n", names)

	// 過濾管理員
	admins := Filter(users, func(u User) bool { return u.Role == "admin" })
	fmt.Println("管理員:")
	PrintAll(admins)

	// 按角色分組
	byRole := GroupBy(users, func(u User) string { return u.Role })
	fmt.Println("按角色分組:")
	for role, members := range byRole {
		memberNames := Map(members, func(u User) string { return u.Name })
		fmt.Printf("  %s: %v\n", role, memberNames)
	}

	// 排序（泛型版）
	ages := Map(users, func(u User) int { return u.Age })
	sorted := SortSlice(ages)
	fmt.Printf("年齡排序: %v\n", sorted)

	// ──── 6. Go 1.21 標準庫的泛型函式 ────
	fmt.Println("\n=== 6. Go 1.21 標準庫的泛型函式（slices/maps 套件）===\n")

	words := []string{"banana", "apple", "cherry", "date"}
	fmt.Printf("原始:  %v\n", words)
	fmt.Printf("包含 apple: %v\n", slices.Contains(words, "apple"))
	fmt.Printf("最大值: %q\n", slices.Max(words))

	scores := map[string]int{"Alice": 95, "Bob": 87, "Charlie": 92}
	fmt.Printf("Map keys: %v\n", Keys(scores))

	// ──── 7. 泛型使用時機 ────
	fmt.Println("\n=== 7. 泛型使用時機 ===\n")

	fmt.Println("✅ 適合泛型：")
	fmt.Println("  - 資料結構（Stack, Queue, Tree, Set）")
	fmt.Println("  - 函式式工具（Map, Filter, Reduce）")
	fmt.Println("  - 數學運算（Sum, Min, Max, Sort）")
	fmt.Println("  - 通用包裝器（Result[T], Optional[T]）")
	fmt.Println()
	fmt.Println("❌ 不適合泛型：")
	fmt.Println("  - 業務邏輯（用介面，行為多型比類型多型好）")
	fmt.Println("  - 只有一種類型的情況（直接寫就好）")
	fmt.Println("  - 需要執行時動態決定類型（用 interface{}）")
	fmt.Println()
	fmt.Println("📝 建議：先寫具體類型，發現重複時再泛型化")
	fmt.Println("  （不要一開始就過度設計）")

	// ──── Map/Filter/Reduce 鏈的最終示範 ────
	fmt.Println("\n=== 完整示範：處理文章標題 ===\n")

	titles := []string{"  Go 泛型  ", "  Prometheus 監控  ", "  gRPC 教學  ", "  CI/CD  "}
	processed := Map(
		Filter(
			Map(titles, strings.TrimSpace),            // 去除空格
			func(s string) bool { return len(s) > 5 }, // 過濾短標題
		),
		strings.ToUpper, // 轉大寫
	)
	fmt.Println("處理後的標題:")
	for _, t := range processed {
		fmt.Printf("  %q\n", t)
	}

	fmt.Println("\n==========================================")
	fmt.Println(" 教學完成！")
	fmt.Println("==========================================")
}
