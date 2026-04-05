// Package mq 用 Go channel 實作 In-Memory Message Queue（第 33 課）
// 生產環境應替換為 Redis Streams / RabbitMQ / Kafka
package mq

import (
	"log/slog"
	"sync"
	"time"
)

// Message 訊息結構
type Message struct {
	Topic     string
	Payload   any
	CreatedAt time.Time
}

// Broker 訊息代理（用 channel 實作的 Pub/Sub）
type Broker struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Message
	bufferSize  int
}

// NewBroker 建立訊息代理
func NewBroker(bufferSize int) *Broker {
	return &Broker{
		subscribers: make(map[string][]chan Message),
		bufferSize:  bufferSize,
	}
}

// Subscribe 訂閱主題，回傳接收 channel
func (b *Broker) Subscribe(topic string) <-chan Message {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan Message, b.bufferSize)
	b.subscribers[topic] = append(b.subscribers[topic], ch)
	slog.Debug("訂閱主題", "topic", topic, "subscribers", len(b.subscribers[topic]))
	return ch
}

// Publish 發布訊息到主題（非阻塞，滿了就丟棄並記錄警告）
func (b *Broker) Publish(topic string, payload any) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	msg := Message{
		Topic:     topic,
		Payload:   payload,
		CreatedAt: time.Now(),
	}

	for _, ch := range b.subscribers[topic] {
		select {
		case ch <- msg:
		default:
			slog.Warn("訊息佇列已滿，丟棄訊息", "topic", topic)
		}
	}
}

// Close 關閉所有訂閱 channel
func (b *Broker) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()

	for topic, subs := range b.subscribers {
		for _, ch := range subs {
			close(ch)
		}
		delete(b.subscribers, topic)
	}
}
