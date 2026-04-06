// Package circuitbreaker 封裝 gobreaker v2（第 39 課）
package circuitbreaker

import (
	"log/slog"
	"time"

	"github.com/sony/gobreaker/v2"
)

// NewBreaker 建立熔斷器
// name: 熔斷器名稱（用於日誌）
// maxFailures: 最大連續失敗次數後開啟熔斷
// timeout: 熔斷開啟後多久嘗試恢復
func NewBreaker[T any](name string, maxFailures uint32, timeout time.Duration) *gobreaker.CircuitBreaker[T] {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 3, // Half-Open 狀態下最多允許幾個請求通過
		Interval:    30 * time.Second,
		Timeout:     timeout, // Open → Half-Open 的等待時間
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= maxFailures
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			slog.Warn("熔斷器狀態變更",
				"name", name,
				"from", from.String(),
				"to", to.String(),
			)
		},
	}

	return gobreaker.NewCircuitBreaker[T](settings)
}
