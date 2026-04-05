# 第三十九課：Circuit Breaker（熔斷器）

## 適合哪個階段的工程師？

| 階段 | 說明 |
|------|------|
| 🟡 中級工程師 | 了解熔斷器概念，知道為什麼需要它 |
| 🔴 資深工程師 | **必備**：能設計容錯策略，包含熔斷器、重試、超時、Fallback |
| 🏢 SRE/架構師 | 設計整體系統韌性（Resilience）策略 |

## 三個狀態

```
Closed（正常）→ [失敗率 ≥ 60% 且樣本 ≥ 5] → Open（熔斷）
Open（熔斷）  → [等待 5 秒]                → Half-Open（試探）
Half-Open    → [成功]                      → Closed（恢復）
Half-Open    → [失敗]                      → Open（繼續熔斷）
```

## 核心用法

```go
settings := gobreaker.Settings{
    Name:        "payment-service",
    MaxRequests: 3,              // Half-Open 允許的測試請求數
    Interval:    10 * time.Second, // 統計週期
    Timeout:     5 * time.Second,  // Open → Half-Open 的等待時間
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        return counts.Requests >= 5 &&
               float64(counts.TotalFailures)/float64(counts.Requests) >= 0.6
    },
}

cb := gobreaker.NewCircuitBreaker[string](settings)

// 包裝外部呼叫
result, err := cb.Execute(func() (string, error) {
    return callExternalAPI(req)
})

// 判斷是熔斷還是服務錯誤
if errors.Is(err, gobreaker.ErrOpenState) {
    // 熔斷器開啟，快速失敗
}
```

## 搭配其他模式

| 模式 | 用途 | 搭配方式 |
|------|------|----------|
| **Retry（重試）** | 短暫錯誤（網路抖動）| 先重試 N 次，仍失敗算熔斷器失敗 |
| **Timeout（超時）** | 避免無限等待 | 設定 context 超時 |
| **Fallback（降級）** | 熔斷時返回預設值 | 捕獲 ErrOpenState |
| **Bulkhead（隔艙）** | 隔離不同服務的失敗 | 每個服務一個熔斷器 |

## 執行方式

```bash
go run ./tutorials/39-circuit-breaker
```
