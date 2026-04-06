// Package saga 實作 Saga 補償模式，處理分散式交易的自動回滾（第 40 課）
//
// Saga 模式的核心概念：
//   - 將一個長交易拆成多個步驟（Step），每個步驟有「執行」和「補償」兩個動作
//   - 依序執行每個步驟，如果某個步驟失敗，就反向執行已完成步驟的補償動作
//   - 這樣可以在不需要分散式鎖的情況下，保持最終一致性（Eventual Consistency）
//
// 搭配搶票系統的使用情境：
//   - Step 1: 扣 Redis 庫存       ← 補償：回補庫存
//   - Step 2: 建立訂單到 DB        ← 補償：刪除訂單
//   - Step 3: 發送付款請求到 MQ    ← 補償：發送取消訊息
//   - 如果 Step 3 失敗 → 自動回滾 Step 2（刪訂單）→ 回滾 Step 1（回補庫存）
package saga

import (
	"context"
	"fmt"
	"log/slog"
)

// Step 代表 Saga 中的一個步驟
//
// 每個步驟包含：
//   - Name:       步驟名稱，用於日誌和錯誤訊息
//   - Execute:    正向操作（例如扣庫存）
//   - Compensate: 補償操作（例如回補庫存），在後續步驟失敗時被呼叫
type Step struct {
	Name       string
	Execute    func(ctx context.Context) error
	Compensate func(ctx context.Context) error
}

// Result 代表 Saga 執行的結果
type Result struct {
	Success     bool   // 整個 Saga 是否成功
	FailedStep  string // 失敗的步驟名稱（如果有的話）
	CompErrors  []error // 補償過程中發生的錯誤
}

// Saga 是 Saga 模式的編排器（Orchestrator）
//
// 負責依序執行步驟，並在失敗時反向補償
type Saga struct {
	name  string
	steps []Step
}

// New 建立一個新的 Saga 實例
//
// name 用於日誌識別，例如 "搶票交易"
func New(name string) *Saga {
	return &Saga{
		name:  name,
		steps: make([]Step, 0),
	}
}

// AddStep 新增一個步驟到 Saga
//
// 步驟會依照加入順序執行，補償時則反向執行
//
// 範例：
//
//	saga.AddStep("扣除庫存", deductStock, restoreStock)
//	saga.AddStep("建立訂單", createOrder, deleteOrder)
func (s *Saga) AddStep(name string, execute, compensate func(ctx context.Context) error) {
	s.steps = append(s.steps, Step{
		Name:       name,
		Execute:    execute,
		Compensate: compensate,
	})
}

// Run 執行 Saga 中的所有步驟
//
// 執行邏輯：
//  1. 依序執行每個步驟的 Execute 函式
//  2. 如果某個步驟失敗，停止執行後續步驟
//  3. 反向執行已完成步驟的 Compensate 函式（最後完成的先補償）
//  4. 回傳 Result，包含失敗資訊和補償過程中的錯誤
//
// 範例：
//
//	result, err := saga.Run(ctx)
//	if err != nil {
//	    // Saga 失敗，已自動補償
//	    log.Printf("交易失敗於步驟 %s: %v", result.FailedStep, err)
//	}
func (s *Saga) Run(ctx context.Context) (*Result, error) {
	slog.Info("Saga 開始執行", "saga", s.name, "steps", len(s.steps))

	// 記錄已成功執行的步驟索引，用於失敗時的反向補償
	completedSteps := make([]int, 0, len(s.steps))

	for i, step := range s.steps {
		slog.Info("Saga 執行步驟",
			"saga", s.name,
			"step", step.Name,
			"index", i+1,
			"total", len(s.steps),
		)

		if err := step.Execute(ctx); err != nil {
			slog.Error("Saga 步驟失敗，開始補償",
				"saga", s.name,
				"failed_step", step.Name,
				"error", err,
			)

			// 反向補償所有已完成的步驟
			compErrors := s.compensate(ctx, completedSteps)

			result := &Result{
				Success:    false,
				FailedStep: step.Name,
				CompErrors: compErrors,
			}

			return result, fmt.Errorf("saga [%s] 在步驟 [%s] 失敗: %w", s.name, step.Name, err)
		}

		completedSteps = append(completedSteps, i)
		slog.Info("Saga 步驟完成", "saga", s.name, "step", step.Name)
	}

	slog.Info("Saga 全部完成", "saga", s.name)
	return &Result{Success: true}, nil
}

// compensate 反向執行已完成步驟的補償動作
//
// 從最後一個已完成的步驟開始，往前逐一補償
// 即使某個補償動作失敗，也會繼續補償其他步驟（收集所有錯誤）
func (s *Saga) compensate(ctx context.Context, completedSteps []int) []error {
	compErrors := make([]error, 0)

	// 反向遍歷已完成的步驟
	for i := len(completedSteps) - 1; i >= 0; i-- {
		stepIdx := completedSteps[i]
		step := s.steps[stepIdx]

		slog.Info("Saga 補償步驟", "saga", s.name, "step", step.Name)

		if step.Compensate == nil {
			slog.Warn("Saga 步驟無補償函式，跳過", "step", step.Name)
			continue
		}

		if err := step.Compensate(ctx); err != nil {
			slog.Error("Saga 補償失敗",
				"saga", s.name,
				"step", step.Name,
				"error", err,
			)
			compErrors = append(compErrors, fmt.Errorf("補償 [%s] 失敗: %w", step.Name, err))
			// 不中斷，繼續補償其他步驟
		} else {
			slog.Info("Saga 補償完成", "saga", s.name, "step", step.Name)
		}
	}

	return compErrors
}
