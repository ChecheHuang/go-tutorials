// ==========================================================================
// 第三十九課：Circuit Breaker（熔斷器）
// ==========================================================================
//
// 什麼是熔斷器？
//   從電路保險絲借來的概念：當電流過大，保險絲自動斷開，保護電器。
//   在軟體中：當外部服務頻繁失敗，熔斷器自動「斷開」請求，
//   讓系統有時間恢復，同時保護呼叫方不被拖垮。
//
// 熔斷器的三個狀態：
//   Closed（關閉）→ 正常狀態，請求可以通過，記錄失敗率
//   Open（開啟）  → 熔斷！請求全部被拒絕（快速失敗），不打向下游
//   Half-Open（半開）→ 嘗試恢復，允許少量請求測試下游是否恢復
//
//   Closed → [失敗率超過閾值] → Open
//   Open   → [等待超時後]     → Half-Open
//   Half-Open → [測試成功]    → Closed
//   Half-Open → [測試失敗]    → Open
//
// 為什麼需要熔斷器？
//   微服務架構中，服務 A 呼叫服務 B，服務 B 掛了：
//   ❌ 沒有熔斷器：A 等待 B 回應（超時），佔用 A 的連線，
//                   A 也跟著崩潰（級聯失敗 Cascading Failure）
//   ✅ 有熔斷器：A 立即收到錯誤（快速失敗），繼續服務其他請求
//
// 使用 gobreaker 套件（Sony 開源）
// 執行方式：go run ./tutorials/34-circuit-breaker
// ==========================================================================

package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/sony/gobreaker/v2"
)

// ==========================================================================
// 1. 模擬不穩定的外部服務
// ==========================================================================

// ExternalService 模擬不穩定的外部服務（例如第三方 API）
type ExternalService struct {
	failRate   float64 // 失敗率（0.0 ~ 1.0）
	latency    time.Duration
	callCount  int
}

var ErrServiceUnavailable = errors.New("服務不可用")

// Call 呼叫外部服務
func (s *ExternalService) Call(request string) (string, error) {
	s.callCount++
	time.Sleep(s.latency) // 模擬網路延遲

	if rand.Float64() < s.failRate {
		return "", ErrServiceUnavailable
	}
	return fmt.Sprintf("回應: %s (第 %d 次呼叫)", request, s.callCount), nil
}

// ==========================================================================
// 2. 設定熔斷器
// ==========================================================================

// newCircuitBreaker 建立設定好的熔斷器
func newCircuitBreaker(name string) *gobreaker.CircuitBreaker[string] {
	settings := gobreaker.Settings{
		Name: name, // 熔斷器名稱（用於日誌和監控）

		// 觸發 Open 的條件（Closed 狀態下）
		MaxRequests: 3, // Half-Open 時允許通過的最大請求數
		Interval:    10 * time.Second, // 統計週期（這段時間內的失敗率）
		Timeout:     5 * time.Second,  // Open → Half-Open 的等待時間

		// ReadyToTrip：決定什麼時候要從 Closed 轉到 Open
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// 條件：總請求 >= 5 且失敗率 >= 60%
			totalRequests := counts.Requests
			if totalRequests < 5 {
				return false // 樣本太少，不熔斷
			}
			failureRatio := float64(counts.TotalFailures) / float64(totalRequests)
			return failureRatio >= 0.6
		},

		// OnStateChange：狀態改變時的 callback（用於日誌、監控）
		OnStateChange: func(name string, from, to gobreaker.State) {
			fmt.Printf("\n🔌 熔斷器 [%s] 狀態變更: %s → %s\n", name, from, to)
			switch to {
			case gobreaker.StateOpen:
				fmt.Println("   ⚠️  熔斷器開啟！請求將被快速拒絕，等待恢復...")
			case gobreaker.StateHalfOpen:
				fmt.Println("   🔍 嘗試恢復，允許少量請求測試...")
			case gobreaker.StateClosed:
				fmt.Println("   ✅ 服務恢復正常！")
			}
		},
	}

	return gobreaker.NewCircuitBreaker[string](settings)
}

// ==========================================================================
// 3. 使用熔斷器包裝外部服務呼叫
// ==========================================================================

// ServiceClient 帶熔斷器的服務客戶端
type ServiceClient struct {
	service *ExternalService
	breaker *gobreaker.CircuitBreaker[string]
}

// Call 透過熔斷器呼叫服務
func (c *ServiceClient) Call(request string) (string, error) {
	// gobreaker 的核心用法：把實際呼叫包在 Execute 裡
	result, err := c.breaker.Execute(func() (string, error) {
		return c.service.Call(request) // 實際的服務呼叫
	})

	if err != nil {
		// 判斷是熔斷器拒絕（ErrOpenState）還是服務本身的錯誤
		if errors.Is(err, gobreaker.ErrOpenState) {
			return "", fmt.Errorf("熔斷器開啟，服務暫時不可用: %w", err)
		}
		if errors.Is(err, gobreaker.ErrTooManyRequests) {
			return "", fmt.Errorf("Half-Open 階段請求數超限: %w", err)
		}
		return "", fmt.Errorf("服務錯誤: %w", err)
	}
	return result, nil
}

// printBreakerState 印出熔斷器當前狀態
func printBreakerState(cb *gobreaker.CircuitBreaker[string]) {
	counts := cb.Counts()
	fmt.Printf("  [狀態: %-10s] 請求數: %d, 成功: %d, 失敗: %d\n",
		cb.State(),
		counts.Requests,
		counts.TotalSuccesses,
		counts.TotalFailures,
	)
}

// ==========================================================================
// 示範：無熔斷器 vs 有熔斷器
// ==========================================================================

func demonstrateWithoutBreaker() {
	fmt.Println("=== 1. 沒有熔斷器（外部服務 80% 失敗率）===")
	fmt.Println()

	service := &ExternalService{failRate: 0.8, latency: 50 * time.Millisecond}
	successCount, failCount := 0, 0
	start := time.Now()

	for i := range 10 {
		_, err := service.Call(fmt.Sprintf("請求 #%d", i+1))
		if err != nil {
			failCount++
			fmt.Printf("  請求 #%d ❌ 失敗（等了 50ms）\n", i+1)
		} else {
			successCount++
			fmt.Printf("  請求 #%d ✅ 成功\n", i+1)
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("\n結果: 成功 %d, 失敗 %d, 耗時: %v\n", successCount, failCount, elapsed)
	fmt.Println("→ 每個失敗都等了 50ms，總共浪費很多時間在等待失敗的服務")
}

func demonstrateWithBreaker() {
	fmt.Println("\n=== 2. 有熔斷器（外部服務 80% 失敗率）===")
	fmt.Println()

	service := &ExternalService{failRate: 0.8, latency: 50 * time.Millisecond}
	cb := newCircuitBreaker("payment-service")
	client := &ServiceClient{service: service, breaker: cb}

	successCount, failCount, rejectedCount := 0, 0, 0
	start := time.Now()

	for i := range 20 {
		_, err := client.Call(fmt.Sprintf("請求 #%d", i+1))
		if err != nil {
			if errors.Is(err, gobreaker.ErrOpenState) {
				rejectedCount++
				fmt.Printf("  請求 #%d ⚡ 被熔斷器拒絕（立即回傳，不等待）\n", i+1)
			} else {
				failCount++
				fmt.Printf("  請求 #%d ❌ 服務失敗\n", i+1)
			}
		} else {
			successCount++
			fmt.Printf("  請求 #%d ✅ 成功\n", i+1)
		}
		printBreakerState(cb)
	}

	elapsed := time.Since(start)
	fmt.Printf("\n結果: 成功 %d, 服務失敗 %d, 熔斷拒絕 %d, 耗時: %v\n",
		successCount, failCount, rejectedCount, elapsed)
	fmt.Println("→ 熔斷後的請求立即被拒絕，不浪費時間等待")
}

func demonstrateRecovery() {
	fmt.Println("\n=== 3. 服務恢復流程（Open → Half-Open → Closed）===")
	fmt.Println()

	// 先建立一個會失敗的服務，讓熔斷器開啟
	service := &ExternalService{failRate: 1.0, latency: 10 * time.Millisecond} // 100% 失敗
	cb := newCircuitBreaker("recovery-demo")
	client := &ServiceClient{service: service, breaker: cb}

	fmt.Println("階段 1：服務 100% 失敗，觸發熔斷")
	for i := range 6 {
		_, err := client.Call(fmt.Sprintf("請求 #%d", i+1))
		if err != nil {
			fmt.Printf("  請求 #%d: %v\n", i+1, err)
		}
	}

	fmt.Println("\n階段 2：模擬等待 Timeout（5 秒後進入 Half-Open）")
	fmt.Println("  （在真實系統中這裡會等 5 秒，示範跳過）")
	fmt.Println("  熔斷器目前狀態:", cb.State())

	fmt.Println("\n階段 3：修復服務（失敗率降到 0%）")
	service.failRate = 0.0 // 「修復」服務

	fmt.Println("\n說明：真實場景中，等待 Timeout 後熔斷器進入 Half-Open，")
	fmt.Println("允許少量請求通過測試服務是否恢復。")
	fmt.Println("如果測試成功 → Closed（正常）")
	fmt.Println("如果測試失敗 → Open（繼續熔斷）")
}

// ==========================================================================
// 主程式
// ==========================================================================

func main() {
	fmt.Println("==========================================")
	fmt.Println(" 第三十四課：Circuit Breaker（熔斷器）")
	fmt.Println("==========================================")
	fmt.Println()

	demonstrateWithoutBreaker()
	demonstrateWithBreaker()
	demonstrateRecovery()

	// ──── 最佳實踐總結 ────
	fmt.Println("\n=== 熔斷器設計原則 ===")
	fmt.Println()
	fmt.Println("✅ 適合使用熔斷器的場景：")
	fmt.Println("  - 呼叫外部 HTTP API（第三方服務）")
	fmt.Println("  - 資料庫查詢（資料庫掛掉時）")
	fmt.Println("  - 微服務之間的呼叫")
	fmt.Println("  - 任何有網路延遲的操作")
	fmt.Println()
	fmt.Println("⚙️  重要參數調整：")
	fmt.Println("  - MaxRequests：Half-Open 允許通過的請求數（通常 1-5）")
	fmt.Println("  - Timeout：Open → Half-Open 的等待時間（通常 30s-5min）")
	fmt.Println("  - Interval：統計週期（通常 10-60 秒）")
	fmt.Println("  - ReadyToTrip：閾值（失敗率 + 最小樣本數）")
	fmt.Println()
	fmt.Println("📊 搭配監控：")
	fmt.Println("  - 用 Prometheus Counter 記錄熔斷器狀態變更")
	fmt.Println("  - 用 Grafana 建立熔斷器狀態看板")
	fmt.Println("  - 設定 AlertManager：熔斷器開啟超過 5 分鐘就告警")
	fmt.Println()
	fmt.Println("🔧 與 Retry 的配合：")
	fmt.Println("  Retry（重試）→ 短暫錯誤（網路抖動）")
	fmt.Println("  CircuitBreaker → 持續錯誤（服務掛掉）")
	fmt.Println("  兩者配合：先重試 N 次，仍失敗則熔斷器記錄為失敗")

	fmt.Println("\n==========================================")
	fmt.Println(" 教學完成！")
	fmt.Println("==========================================")
}
