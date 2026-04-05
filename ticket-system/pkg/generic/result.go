// Package generic 提供泛型工具函式（第 31 課）
package generic

// Result 是 Rust 風格的泛型結果容器
type Result[T any] struct {
	Value T
	Err   error
}

// NewResult 建立成功的 Result
func NewResult[T any](value T) Result[T] {
	return Result[T]{Value: value}
}

// NewError 建立失敗的 Result
func NewError[T any](err error) Result[T] {
	return Result[T]{Err: err}
}

// IsOk 檢查是否成功
func (r Result[T]) IsOk() bool {
	return r.Err == nil
}

// Map 對成功的值做轉換
func Map[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

// Filter 過濾符合條件的元素
func Filter[T any](slice []T, fn func(T) bool) []T {
	var result []T
	for _, v := range slice {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Contains 檢查元素是否存在
func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Keys 取得 map 的所有 key
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
