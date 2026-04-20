package saga

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaga_AllStepsSucceed(t *testing.T) {
	var executionOrder []string

	s := New("test-saga")
	s.AddStep("步驟一",
		func(ctx context.Context) error { executionOrder = append(executionOrder, "exec:1"); return nil },
		func(ctx context.Context) error { executionOrder = append(executionOrder, "comp:1"); return nil },
	)
	s.AddStep("步驟二",
		func(ctx context.Context) error { executionOrder = append(executionOrder, "exec:2"); return nil },
		func(ctx context.Context) error { executionOrder = append(executionOrder, "comp:2"); return nil },
	)
	s.AddStep("步驟三",
		func(ctx context.Context) error { executionOrder = append(executionOrder, "exec:3"); return nil },
		func(ctx context.Context) error { executionOrder = append(executionOrder, "comp:3"); return nil },
	)

	result, err := s.Run(context.Background())

	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Empty(t, result.FailedStep)
	assert.Empty(t, result.CompErrors)
	// 只執行，不補償
	assert.Equal(t, []string{"exec:1", "exec:2", "exec:3"}, executionOrder)
}

func TestSaga_MiddleStepFails_CompensatesInReverse(t *testing.T) {
	var executionOrder []string

	s := New("test-saga")
	s.AddStep("扣庫存",
		func(ctx context.Context) error { executionOrder = append(executionOrder, "exec:deduct"); return nil },
		func(ctx context.Context) error { executionOrder = append(executionOrder, "comp:deduct"); return nil },
	)
	s.AddStep("建訂單",
		func(ctx context.Context) error {
			executionOrder = append(executionOrder, "exec:order")
			return errors.New("DB connection refused")
		},
		func(ctx context.Context) error { executionOrder = append(executionOrder, "comp:order"); return nil },
	)
	s.AddStep("發送 MQ",
		func(ctx context.Context) error { executionOrder = append(executionOrder, "exec:mq"); return nil },
		func(ctx context.Context) error { executionOrder = append(executionOrder, "comp:mq"); return nil },
	)

	result, err := s.Run(context.Background())

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "建訂單", result.FailedStep)
	assert.Empty(t, result.CompErrors)

	// 步驟三不應被執行，步驟一應被補償（步驟二失敗但未完成，不需要補償）
	assert.Equal(t, []string{"exec:deduct", "exec:order", "comp:deduct"}, executionOrder)
}

func TestSaga_CompensationStepPanics_ContinuesRemainingCompensations(t *testing.T) {
	compensated := make(map[string]bool)

	s := New("test-saga")
	s.AddStep("步驟一",
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { compensated["步驟一"] = true; return nil },
	)
	s.AddStep("步驟二",
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error {
			panic("補償爆炸了！")
		},
	)
	s.AddStep("步驟三",
		func(ctx context.Context) error {
			return errors.New("步驟三失敗")
		},
		func(ctx context.Context) error { compensated["步驟三"] = true; return nil },
	)

	result, err := s.Run(context.Background())

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Equal(t, "步驟三", result.FailedStep)

	// 步驟二的補償 panic 了，但步驟一的補償仍然要執行
	assert.True(t, compensated["步驟一"], "步驟一應該被補償（即使步驟二的補償 panic）")

	// 補償錯誤中應包含 panic 的資訊
	require.Len(t, result.CompErrors, 1)
	assert.Contains(t, result.CompErrors[0].Error(), "panic")
}

func TestSaga_ExecuteStepPanics_TriggersCompensation(t *testing.T) {
	compensated := false

	s := New("test-saga")
	s.AddStep("正常步驟",
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { compensated = true; return nil },
	)
	s.AddStep("會 panic 的步驟",
		func(ctx context.Context) error { panic("execute panic!") },
		func(ctx context.Context) error { return nil },
	)

	result, err := s.Run(context.Background())

	require.Error(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, err.Error(), "panic")
	assert.True(t, compensated, "即使 Execute panic，前面的步驟也要被補償")
}

func TestSaga_NilCompensation_Skipped(t *testing.T) {
	s := New("test-saga")
	s.AddStep("無補償步驟",
		func(ctx context.Context) error { return nil },
		nil, // 沒有補償函式
	)
	s.AddStep("會失敗的步驟",
		func(ctx context.Context) error { return errors.New("fail") },
		func(ctx context.Context) error { return nil },
	)

	result, err := s.Run(context.Background())

	require.Error(t, err)
	assert.False(t, result.Success)
	// nil 補償不應 panic
	assert.Empty(t, result.CompErrors)
}

func TestSaga_CompensationFails_CollectsError(t *testing.T) {
	s := New("test-saga")
	s.AddStep("步驟一",
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return errors.New("補償失敗：DB 斷線") },
	)
	s.AddStep("步驟二",
		func(ctx context.Context) error { return errors.New("步驟二失敗") },
		func(ctx context.Context) error { return nil },
	)

	result, err := s.Run(context.Background())

	require.Error(t, err)
	require.Len(t, result.CompErrors, 1)
	assert.Contains(t, result.CompErrors[0].Error(), "補償失敗：DB 斷線")
}
