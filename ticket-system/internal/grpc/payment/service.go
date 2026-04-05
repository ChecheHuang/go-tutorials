// Package payment 用純 Go 模擬 gRPC 支付服務（第 27 課）
// 不需要 protoc 編譯，用 Go struct 模擬 protobuf message
// 學習者可以之後加入真正的 .proto 檔案
package payment

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"time"
)

// PaymentRequest 模擬 protobuf 的請求結構
type PaymentRequest struct {
	OrderID uint
	UserID  string
	Amount  float64
}

// PaymentResponse 模擬 protobuf 的回應結構
type PaymentResponse struct {
	Success       bool
	TransactionID string
	Message       string
}

// PaymentService 支付服務介面（模擬 gRPC service 定義）
type PaymentService interface {
	ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
}

// paymentServer 支付服務實作（模擬外部支付閘道）
type paymentServer struct {
	failRate float64 // 模擬失敗率（0.0 ~ 1.0）
	latency  time.Duration
}

// NewPaymentServer 建立支付服務
// failRate: 模擬失敗率（例如 0.2 = 20% 失敗）
// latency: 模擬網路延遲
func NewPaymentServer(failRate float64, latency time.Duration) PaymentService {
	return &paymentServer{
		failRate: failRate,
		latency:  latency,
	}
}

func (s *paymentServer) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	slog.Info("[支付服務] 處理支付請求",
		"order_id", req.OrderID,
		"user_id", req.UserID,
		"amount", req.Amount,
	)

	// 模擬網路延遲
	select {
	case <-time.After(s.latency):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 模擬隨機失敗
	if rand.Float64() < s.failRate {
		slog.Warn("[支付服務] 支付失敗", "order_id", req.OrderID)
		return &PaymentResponse{
			Success: false,
			Message: "支付閘道暫時不可用",
		}, nil
	}

	txID := fmt.Sprintf("TX-%d-%d", req.OrderID, time.Now().UnixNano())
	slog.Info("[支付服務] 支付成功", "order_id", req.OrderID, "tx_id", txID)

	return &PaymentResponse{
		Success:       true,
		TransactionID: txID,
		Message:       "支付成功",
	}, nil
}
