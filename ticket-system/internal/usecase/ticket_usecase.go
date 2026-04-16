// Package usecase 搶票核心業務邏輯
package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"ticket-system/internal/domain"
	"ticket-system/internal/mq"
	"ticket-system/pkg/apperror"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TicketUsecase 搶票業務邏輯
type TicketUsecase struct {
	stock      domain.StockStore
	eventRepo  domain.EventRepository
	orderWrite domain.OrderWriteRepository
	orderRead  domain.OrderReadRepository
	broker     *mq.Broker
	tracer     trace.Tracer
}

// NewTicketUsecase 建立搶票 Usecase
func NewTicketUsecase(
	stock domain.StockStore,
	eventRepo domain.EventRepository,
	orderWrite domain.OrderWriteRepository,
	orderRead domain.OrderReadRepository,
	broker *mq.Broker,
	tracer trace.Tracer,
) *TicketUsecase {
	return &TicketUsecase{
		stock:      stock,
		eventRepo:  eventRepo,
		orderWrite: orderWrite,
		orderRead:  orderRead,
		broker:     broker,
		tracer:     tracer,
	}
}

// GrabTicket 搶票（核心流程）
func (u *TicketUsecase) GrabTicket(ctx context.Context, cmd domain.OrderCommand) (*domain.Order, error) {
	ctx, span := u.tracer.Start(ctx, "usecase.GrabTicket",
		trace.WithAttributes(
			attribute.Int("event_id", int(cmd.EventID)),
			attribute.String("user_id", cmd.UserID),
			attribute.Int("quantity", cmd.Quantity),
		),
	)
	defer span.End()

	// 1. 檢查活動存在
	event, err := u.eventRepo.FindByID(cmd.EventID)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrNotFound, "活動 ID=%d", cmd.EventID)
	}

	// 2. 原子扣庫存（Redis Lua script 或 Memory CAS，不會超賣）
	stockKey := fmt.Sprintf("stock:event:%d", cmd.EventID)
	remaining, err := u.stock.DecrIfSufficient(ctx, stockKey, int64(cmd.Quantity))
	if errors.Is(err, domain.ErrInsufficientStock) {
		slog.Info("庫存不足", "event_id", cmd.EventID, "remaining", remaining)
		return nil, apperror.Wrap(apperror.ErrSoldOut, "活動 %s 已售罄", event.Name)
	}
	if err != nil {
		span.RecordError(err)
		return nil, apperror.Wrap(apperror.ErrInternal, "庫存扣減失敗")
	}

	span.AddEvent("庫存扣減成功", trace.WithAttributes(
		attribute.Int64("remaining", remaining),
	))

	// 3. 建立 pending 訂單
	order := &domain.Order{
		EventID:  cmd.EventID,
		UserID:   cmd.UserID,
		Quantity: cmd.Quantity,
		Amount:   event.Price * float64(cmd.Quantity),
		Status:   domain.OrderPending,
	}

	if err := u.orderWrite.Create(order); err != nil {
		u.stock.IncrBy(ctx, stockKey, int64(cmd.Quantity))
		span.RecordError(err)
		return nil, apperror.Wrap(apperror.ErrInternal, "建立訂單失敗")
	}

	// 4. 發送到 Message Queue
	u.broker.Publish("order.created", order)
	span.AddEvent("訂單已發送到佇列")

	// 5. 廣播剩餘票數
	u.broker.Publish("stock.updated", domain.TicketStock{
		EventID:   cmd.EventID,
		Total:     event.TotalTickets,
		Remaining: int(remaining),
	})

	slog.Info("搶票成功",
		"order_id", order.ID,
		"event_id", cmd.EventID,
		"user_id", cmd.UserID,
		"remaining", remaining,
	)

	return order, nil
}

// GetStock 取得即時庫存
func (u *TicketUsecase) GetStock(ctx context.Context, eventID uint) (*domain.TicketStock, error) {
	event, err := u.eventRepo.FindByID(eventID)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrNotFound, "活動 ID=%d", eventID)
	}

	stockKey := fmt.Sprintf("stock:event:%d", eventID)
	remaining, err := u.stock.Get(ctx, stockKey)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrInternal, "讀取庫存失敗")
	}

	return &domain.TicketStock{
		EventID:   eventID,
		Total:     event.TotalTickets,
		Remaining: remaining,
	}, nil
}

// InitStock 初始化活動庫存
func (u *TicketUsecase) InitStock(ctx context.Context, eventID uint, total int) error {
	stockKey := fmt.Sprintf("stock:event:%d", eventID)
	return u.stock.Set(ctx, stockKey, total)
}

// GetOrder 查詢訂單（CQRS Read）
func (u *TicketUsecase) GetOrder(ctx context.Context, orderID uint) (*domain.OrderView, error) {
	view, err := u.orderRead.FindByID(orderID)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrNotFound, "訂單 ID=%d", orderID)
	}
	return view, nil
}

// GetUserOrders 查詢使用者的所有訂單（CQRS Read）
func (u *TicketUsecase) GetUserOrders(ctx context.Context, userID string) ([]domain.OrderView, error) {
	return u.orderRead.FindByUserID(userID)
}

// GetAllEvents 取得所有活動
func (u *TicketUsecase) GetAllEvents(ctx context.Context) ([]domain.Event, error) {
	return u.eventRepo.FindAll()
}
