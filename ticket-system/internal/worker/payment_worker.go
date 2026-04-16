// Package worker 背景工作者（第 25 課 Goroutine Worker Pool）
package worker

import (
	"context"
	"log/slog"
	"sync"

	"ticket-system/internal/domain"
	"ticket-system/internal/grpc/payment"
	"ticket-system/internal/mq"
	"ticket-system/internal/ws"

	"github.com/sony/gobreaker/v2"
	"go.opentelemetry.io/otel/trace"
)

// PaymentWorker 支付處理工作者
// 從 Message Queue 消費訂單，透過 Circuit Breaker 呼叫支付服務
type PaymentWorker struct {
	broker      *mq.Broker
	orderWrite  domain.OrderWriteRepository
	paymentSvc  payment.PaymentServiceClient
	breaker     *gobreaker.CircuitBreaker[*payment.PaymentResponse]
	hub         *ws.Hub
	tracer      trace.Tracer
	workerCount int
	wg          sync.WaitGroup // 追蹤所有 worker goroutine 的生命週期
}

// NewPaymentWorker 建立支付工作者
func NewPaymentWorker(
	broker *mq.Broker,
	orderWrite domain.OrderWriteRepository,
	paymentSvc payment.PaymentServiceClient,
	breaker *gobreaker.CircuitBreaker[*payment.PaymentResponse],
	hub *ws.Hub,
	tracer trace.Tracer,
	workerCount int,
) *PaymentWorker {
	return &PaymentWorker{
		broker:      broker,
		orderWrite:  orderWrite,
		paymentSvc:  paymentSvc,
		breaker:     breaker,
		hub:         hub,
		tracer:      tracer,
		workerCount: workerCount,
	}
}

// Start 啟動 Worker Pool（多個 goroutine 並發消費）
//
// 呼叫後立即返回，Worker 在背景執行。
// 使用 Wait() 等待所有 Worker 完全停止（在取消 ctx 之後）。
func (w *PaymentWorker) Start(ctx context.Context) {
	orders := w.broker.Subscribe("order.created")

	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go func(id int) {
			defer w.wg.Done()
			slog.Info("支付 Worker 啟動", "worker_id", id)

			for {
				select {
				case <-ctx.Done():
					slog.Info("支付 Worker 停止", "worker_id", id)
					return
				case msg, ok := <-orders:
					if !ok {
						return
					}
					order, ok := msg.Payload.(*domain.Order)
					if !ok {
						slog.Error("訊息類型錯誤", "worker_id", id)
						continue
					}
					w.processOrder(ctx, id, order)
				}
			}
		}(i)
	}
}

// Wait 阻塞直到所有 Worker goroutine 完全停止
//
// 典型的 graceful shutdown 流程：
//
//	cancel()            // 1. 取消 context，通知 worker 停止
//	worker.Wait()       // 2. 等待所有 worker 處理完當前訂單並退出
//	broker.Close()      // 3. 關閉 broker（此時確保不會有 worker 在讀 channel）
func (w *PaymentWorker) Wait() {
	w.wg.Wait()
	slog.Info("所有支付 Worker 已停止")
}

// processOrder 處理單一訂單（透過 Circuit Breaker 呼叫支付服務）
func (w *PaymentWorker) processOrder(ctx context.Context, workerID int, order *domain.Order) {
	ctx, span := w.tracer.Start(ctx, "worker.processPayment")
	defer span.End()

	slog.Info("處理訂單", "worker_id", workerID, "order_id", order.ID)

	// 透過 Circuit Breaker 呼叫支付服務
	resp, err := w.breaker.Execute(func() (*payment.PaymentResponse, error) {
		return w.paymentSvc.ProcessPayment(ctx, &payment.PaymentRequest{
			OrderID: order.ID,
			UserID:  order.UserID,
			Amount:  order.Amount,
		})
	})

	if err != nil {
		// 熔斷器開啟或其他錯誤
		slog.Error("支付呼叫失敗（熔斷器）", "order_id", order.ID, "error", err)
		w.orderWrite.UpdateStatus(order.ID, domain.OrderFailed)
		w.notifyOrderStatus(order.ID, domain.OrderFailed)
		return
	}

	if resp.Success {
		w.orderWrite.UpdateStatus(order.ID, domain.OrderPaid)
		w.notifyOrderStatus(order.ID, domain.OrderPaid)
		slog.Info("訂單支付成功", "order_id", order.ID, "tx_id", resp.TransactionID)
	} else {
		w.orderWrite.UpdateStatus(order.ID, domain.OrderFailed)
		w.notifyOrderStatus(order.ID, domain.OrderFailed)
		slog.Warn("訂單支付失敗", "order_id", order.ID, "message", resp.Message)
	}
}

// notifyOrderStatus 透過 WebSocket 通知訂單狀態
func (w *PaymentWorker) notifyOrderStatus(orderID uint, status domain.OrderStatus) {
	w.hub.Broadcast(map[string]any{
		"type":     "order.status",
		"order_id": orderID,
		"status":   status,
	})
}
