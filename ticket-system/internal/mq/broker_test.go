package mq

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroker_PublishSubscribe(t *testing.T) {
	b := NewBroker(10)
	defer b.Close()

	ch := b.Subscribe("orders")

	b.Publish("orders", "order-1")
	b.Publish("orders", "order-2")

	msg1 := <-ch
	msg2 := <-ch

	assert.Equal(t, "order-1", msg1.Payload)
	assert.Equal(t, "order-2", msg2.Payload)
	assert.Equal(t, "orders", msg1.Topic)
}

func TestBroker_MultipleSubscribers(t *testing.T) {
	b := NewBroker(10)
	defer b.Close()

	ch1 := b.Subscribe("events")
	ch2 := b.Subscribe("events")

	b.Publish("events", "event-1")

	msg1 := <-ch1
	msg2 := <-ch2

	assert.Equal(t, "event-1", msg1.Payload)
	assert.Equal(t, "event-1", msg2.Payload)
}

func TestBroker_DifferentTopics(t *testing.T) {
	b := NewBroker(10)
	defer b.Close()

	ordersCh := b.Subscribe("orders")
	stockCh := b.Subscribe("stock")

	b.Publish("orders", "order-data")
	b.Publish("stock", "stock-data")

	assert.Equal(t, "order-data", (<-ordersCh).Payload)
	assert.Equal(t, "stock-data", (<-stockCh).Payload)
}

func TestBroker_PublishToNonexistentTopic_NoPanic(t *testing.T) {
	b := NewBroker(10)
	defer b.Close()

	// 沒有人訂閱，不應 panic
	assert.NotPanics(t, func() {
		b.Publish("nobody-listens", "data")
	})
}

func TestBroker_CloseClosesChannels(t *testing.T) {
	b := NewBroker(10)
	ch := b.Subscribe("orders")

	b.Close()

	// channel 被關閉後，讀取應回傳 zero value + ok=false
	msg, ok := <-ch
	assert.False(t, ok)
	assert.Empty(t, msg.Payload)
}

func TestBroker_PublishAfterClose_NoPanic(t *testing.T) {
	b := NewBroker(10)
	b.Subscribe("orders")
	b.Close()

	// Close 後 Publish 不應 panic（closed flag 會攔截）
	assert.NotPanics(t, func() {
		b.Publish("orders", "should-be-dropped")
	})
}

func TestBroker_DoubleClose_NoPanic(t *testing.T) {
	b := NewBroker(10)
	b.Subscribe("orders")

	assert.NotPanics(t, func() {
		b.Close()
		b.Close() // 重複關閉不應 panic
	})
}

func TestBroker_BufferFull_DropsMessage(t *testing.T) {
	b := NewBroker(1) // buffer 只有 1
	defer b.Close()

	ch := b.Subscribe("orders")

	b.Publish("orders", "msg-1") // 進 buffer
	b.Publish("orders", "msg-2") // buffer 滿，應該被丟棄

	msg := <-ch
	assert.Equal(t, "msg-1", msg.Payload)

	// channel 應該是空的（msg-2 被丟棄了）
	select {
	case <-ch:
		t.Fatal("不應該有第二條訊息")
	default:
		// 預期行為
	}
}

func TestBroker_ConcurrentPublish_NoPanic(t *testing.T) {
	b := NewBroker(1000)
	defer b.Close()

	b.Subscribe("orders")

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			b.Publish("orders", n)
		}(i)
	}
	wg.Wait()
}

func TestBroker_ConcurrentPublishAndClose_NoPanic(t *testing.T) {
	b := NewBroker(100)
	b.Subscribe("orders")

	var wg sync.WaitGroup

	// 50 個 goroutine 不停 publish
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				b.Publish("orders", j)
			}
		}()
	}

	// 同時 close
	time.Sleep(1 * time.Millisecond)
	b.Close()

	wg.Wait()
}

func TestBroker_MessageHasTimestamp(t *testing.T) {
	b := NewBroker(10)
	defer b.Close()

	ch := b.Subscribe("orders")

	before := time.Now()
	b.Publish("orders", "data")
	after := time.Now()

	msg := <-ch
	require.False(t, msg.CreatedAt.IsZero())
	assert.True(t, msg.CreatedAt.After(before) || msg.CreatedAt.Equal(before))
	assert.True(t, msg.CreatedAt.Before(after) || msg.CreatedAt.Equal(after))
}
