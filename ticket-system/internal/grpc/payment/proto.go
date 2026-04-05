// Package payment 定義支付服務的 gRPC 訊息和服務介面
// 用純 Go 結構模擬 protobuf 生成的程式碼（不需要 protoc）
// 學習者可以之後加入真正的 .proto 檔案替換此檔案
package payment

import "context"

// PaymentRequest 對應 protobuf message PaymentRequest
type PaymentRequest struct {
	OrderID uint    `json:"order_id"`
	UserID  string  `json:"user_id"`
	Amount  float64 `json:"amount"`
}

// PaymentResponse 對應 protobuf message PaymentResponse
type PaymentResponse struct {
	Success       bool   `json:"success"`
	TransactionID string `json:"transaction_id"`
	Message       string `json:"message"`
}

// PaymentServiceServer 對應 protobuf service 定義（server 端介面）
type PaymentServiceServer interface {
	ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
}

// PaymentServiceClient 對應 gRPC 生成的 client 介面
type PaymentServiceClient interface {
	ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error)
}
