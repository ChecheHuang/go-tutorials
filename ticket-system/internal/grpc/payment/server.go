package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// paymentServer 支付服務的 gRPC server 端實作
type paymentServer struct {
	failRate float64
	latency  time.Duration
}

// NewServer 建立支付服務 server 實作
func NewServer(failRate float64, latency time.Duration) PaymentServiceServer {
	return &paymentServer{failRate: failRate, latency: latency}
}

func (s *paymentServer) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	slog.Info("[gRPC 支付服務] 收到請求",
		"order_id", req.OrderID,
		"user_id", req.UserID,
		"amount", req.Amount,
	)

	// 模擬網路延遲
	select {
	case <-time.After(s.latency):
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "請求已取消")
	}

	// 模擬隨機失敗
	if rand.Float64() < s.failRate {
		slog.Warn("[gRPC 支付服務] 支付失敗", "order_id", req.OrderID)
		return &PaymentResponse{
			Success: false,
			Message: "支付閘道暫時不可用",
		}, nil
	}

	txID := fmt.Sprintf("TX-%d-%d", req.OrderID, time.Now().UnixNano())
	slog.Info("[gRPC 支付服務] 支付成功", "order_id", req.OrderID, "tx_id", txID)

	return &PaymentResponse{
		Success:       true,
		TransactionID: txID,
		Message:       "支付成功",
	}, nil
}

// grpcServiceDesc 手動定義 gRPC service descriptor（取代 protoc 生成）
var grpcServiceDesc = grpc.ServiceDesc{
	ServiceName: "payment.PaymentService",
	HandlerType: (*PaymentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ProcessPayment",
			Handler:    processPaymentHandler,
		},
	},
}

// processPaymentHandler gRPC method handler（解碼請求 → 呼叫實作 → 編碼回應）
func processPaymentHandler(srv any, ctx context.Context, dec func(any) error, _ grpc.UnaryServerInterceptor) (any, error) {
	req := &PaymentRequest{}
	if err := dec(req); err != nil {
		return nil, err
	}
	return srv.(PaymentServiceServer).ProcessPayment(ctx, req)
}

// StartGRPCServer 啟動 gRPC server（可用 net.Listener 或 bufconn）
func StartGRPCServer(lis net.Listener, svc PaymentServiceServer) *grpc.Server {
	// 註冊自訂 codec（用 JSON 取代 protobuf，免去 protoc 依賴）
	s := grpc.NewServer()
	s.RegisterService(&grpcServiceDesc, svc)

	go func() {
		slog.Info("[gRPC] 支付服務啟動", "addr", lis.Addr().String())
		if err := s.Serve(lis); err != nil {
			slog.Error("[gRPC] 服務啟動失敗", "error", err)
		}
	}()

	return s
}

// jsonCodec 用 JSON 編解碼（取代 protobuf，教學用途）
type jsonCodec struct{}

func (jsonCodec) Marshal(v any) ([]byte, error)   { return json.Marshal(v) }
func (jsonCodec) Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }
func (jsonCodec) Name() string                       { return "json" }
