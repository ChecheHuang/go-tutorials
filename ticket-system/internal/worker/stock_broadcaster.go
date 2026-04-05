package worker

import (
	"context"
	"log/slog"

	"ticket-system/internal/domain"
	"ticket-system/internal/mq"
	"ticket-system/internal/ws"
)

// StockBroadcaster 監聽庫存變更事件，透過 WebSocket 廣播
type StockBroadcaster struct {
	broker *mq.Broker
	hub    *ws.Hub
}

func NewStockBroadcaster(broker *mq.Broker, hub *ws.Hub) *StockBroadcaster {
	return &StockBroadcaster{broker: broker, hub: hub}
}

func (b *StockBroadcaster) Start(ctx context.Context) {
	stockUpdates := b.broker.Subscribe("stock.updated")

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-stockUpdates:
				if !ok {
					return
				}
				stock, ok := msg.Payload.(domain.TicketStock)
				if !ok {
					continue
				}
				b.hub.Broadcast(map[string]any{
					"type":      "stock.updated",
					"event_id":  stock.EventID,
					"total":     stock.Total,
					"remaining": stock.Remaining,
				})
				slog.Debug("庫存更新已廣播", "event_id", stock.EventID, "remaining", stock.Remaining)
			}
		}
	}()
}
