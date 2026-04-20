//go:build wireinject
// +build wireinject

// Wire Injector 定義（第 31 課）
// 執行 wire ./internal/di/ 會根據此檔案自動生成 wire_gen.go
package di

import (
	"ticket-system/pkg/config"

	"github.com/google/wire"
	"go.opentelemetry.io/otel/trace"
)

// InitializeApp 是 Wire 的 Injector 函式
func InitializeApp(cfg *config.Config, tracer trace.Tracer) (*App, func(), error) {
	wire.Build(
		ProvideDB,
		ProvideStockStore,
		ProvideBroker,
		ProvideHub,
		ProvideGRPC,
		ProvideBreaker,
		ProvideEventRepo,
		ProvideOrderWriteRepo,
		ProvideOrderReadRepo,
		ProvideTicketUsecase,
		ProvideTicketHandler,
		ProvideWSHandler,
		ProvidePaymentWorker,
		ProvideStockBroadcaster,
		wire.Struct(new(App), "*"),
	)
	return nil, nil, nil
}
