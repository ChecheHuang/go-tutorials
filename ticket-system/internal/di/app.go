package di

import (
	"ticket-system/internal/domain"
	"ticket-system/internal/handler"
	"ticket-system/internal/mq"
	"ticket-system/internal/usecase"
	"ticket-system/internal/worker"
	"ticket-system/internal/ws"

	"gorm.io/gorm"
)

// App 封裝整個應用程式的所有依賴
type App struct {
	DB               *gorm.DB
	Stock            domain.StockStore
	Broker           *mq.Broker
	Hub              *ws.Hub
	GRPC             *GRPCComponents
	TicketUsecase    *usecase.TicketUsecase
	TicketHandler    *handler.TicketHandler
	WSHandler        *handler.WSHandler
	PaymentWorker    *worker.PaymentWorker
	StockBroadcaster *worker.StockBroadcaster
}
