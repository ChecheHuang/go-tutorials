package payment

import (
	"context"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// grpcClient 透過 gRPC 連線呼叫支付服務
type grpcClient struct {
	conn *grpc.ClientConn
}

// NewClient 建立 gRPC client
// target: gRPC server 地址（例如 "localhost:9090"）或 bufconn 用的 "bufnet"
// opts: 額外的 DialOption（例如 bufconn 的 WithContextDialer）
func NewClient(target string, opts ...grpc.DialOption) (PaymentServiceClient, func(), error) {
	// 預設使用不安全連線（教學用途）
	defaultOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.ForceCodec(jsonCodec{})),
	}
	allOpts := append(defaultOpts, opts...)

	conn, err := grpc.NewClient(target, allOpts...)
	if err != nil {
		return nil, nil, err
	}

	slog.Info("[gRPC Client] 已連線", "target", target)

	cleanup := func() {
		conn.Close()
	}

	return &grpcClient{conn: conn}, cleanup, nil
}

// ProcessPayment 透過 gRPC 呼叫遠端支付服務
func (c *grpcClient) ProcessPayment(ctx context.Context, req *PaymentRequest) (*PaymentResponse, error) {
	resp := &PaymentResponse{}
	err := c.conn.Invoke(ctx, "/payment.PaymentService/ProcessPayment", req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
