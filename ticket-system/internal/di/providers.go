// Package di 提供 Wire 依賴注入的 Provider 函式（第 28 課）
// 每個 Provider 負責建立一個依賴，Wire 自動串接依賴關係
package di

import (
	"context"
	"log/slog"
	"net"
	"time"

	"ticket-system/internal/domain"
	"ticket-system/internal/grpc/payment"
	"ticket-system/internal/handler"
	"ticket-system/internal/mq"
	"ticket-system/internal/repository"
	"ticket-system/internal/usecase"
	"ticket-system/internal/worker"
	"ticket-system/internal/ws"
	"ticket-system/pkg/circuitbreaker"
	"ticket-system/pkg/config"

	"github.com/glebarez/sqlite"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"gorm.io/gorm"

	"go.opentelemetry.io/otel/trace"
)

// ProvideDB 建立資料庫連線
func ProvideDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&domain.Event{}, &domain.Order{})
	slog.Info("資料庫就緒", "dsn", cfg.Database.DSN)
	return db, nil
}

// ProvideRedis 建立 Redis 連線
func ProvideRedis(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	slog.Info("Redis 連線成功")
	return rdb, nil
}

// ProvideBroker 建立 Message Queue Broker
func ProvideBroker() *mq.Broker {
	return mq.NewBroker(1000)
}

// ProvideHub 建立 WebSocket Hub
func ProvideHub() *ws.Hub {
	return ws.NewHub()
}

// GRPCComponents gRPC 相關元件（server + client + cleanup）
type GRPCComponents struct {
	Client  payment.PaymentServiceClient
	Server  *grpc.Server
	Cleanup func()
}

// ProvideGRPC 建立 gRPC 支付服務（in-memory bufconn）
func ProvideGRPC() (*GRPCComponents, error) {
	const bufSize = 1024 * 1024
	bufLis := bufconn.Listen(bufSize)

	svc := payment.NewServer(0.2, 100*time.Millisecond)
	grpcServer := payment.StartGRPCServer(bufLis, svc)

	client, cleanupClient, err := payment.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return bufLis.DialContext(ctx)
		}),
	)
	if err != nil {
		grpcServer.GracefulStop()
		return nil, err
	}

	cleanup := func() {
		cleanupClient()
		grpcServer.GracefulStop()
	}

	slog.Info("gRPC 支付服務就緒（in-memory bufconn）")
	return &GRPCComponents{Client: client, Server: grpcServer, Cleanup: cleanup}, nil
}

// ProvideBreaker 建立 Circuit Breaker
func ProvideBreaker() *gobreaker.CircuitBreaker[*payment.PaymentResponse] {
	return circuitbreaker.NewBreaker[*payment.PaymentResponse]("payment-service", 5, 10*time.Second)
}

// ProvideEventRepo 建立活動 Repository
func ProvideEventRepo(db *gorm.DB) domain.EventRepository {
	return repository.NewEventRepository(db)
}

// ProvideOrderWriteRepo 建立訂單寫入 Repository
func ProvideOrderWriteRepo(db *gorm.DB) domain.OrderWriteRepository {
	return repository.NewOrderWriteRepository(db)
}

// ProvideOrderReadRepo 建立訂單讀取 Repository
func ProvideOrderReadRepo(db *gorm.DB) domain.OrderReadRepository {
	return repository.NewOrderReadRepository(db)
}

// ProvideTicketUsecase 建立搶票 Usecase
func ProvideTicketUsecase(
	rdb *redis.Client,
	eventRepo domain.EventRepository,
	orderWrite domain.OrderWriteRepository,
	orderRead domain.OrderReadRepository,
	broker *mq.Broker,
	tracer trace.Tracer,
) *usecase.TicketUsecase {
	return usecase.NewTicketUsecase(rdb, eventRepo, orderWrite, orderRead, broker, tracer)
}

// ProvideTicketHandler 建立 HTTP Handler
func ProvideTicketHandler(uc *usecase.TicketUsecase) *handler.TicketHandler {
	return handler.NewTicketHandler(uc)
}

// ProvideWSHandler 建立 WebSocket Handler
func ProvideWSHandler(hub *ws.Hub) *handler.WSHandler {
	return handler.NewWSHandler(hub)
}

// ProvidePaymentWorker 建立支付 Worker
func ProvidePaymentWorker(
	broker *mq.Broker,
	orderWrite domain.OrderWriteRepository,
	grpcComps *GRPCComponents,
	breaker *gobreaker.CircuitBreaker[*payment.PaymentResponse],
	hub *ws.Hub,
	tracer trace.Tracer,
) *worker.PaymentWorker {
	return worker.NewPaymentWorker(broker, orderWrite, grpcComps.Client, breaker, hub, tracer, 3)
}

// ProvideStockBroadcaster 建立庫存廣播器
func ProvideStockBroadcaster(broker *mq.Broker, hub *ws.Hub) *worker.StockBroadcaster {
	return worker.NewStockBroadcaster(broker, hub)
}
