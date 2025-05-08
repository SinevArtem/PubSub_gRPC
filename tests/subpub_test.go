package tests

import (
	"VK_task/pkg/subpub"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSubPub(t *testing.T) {
	t.Run("Subscribe/Publish", func(t *testing.T) {
		eventBus := subpub.NewSubPub()
		defer eventBus.Close(context.Background())

		received := make(chan interface{}, 1)
		sub, err := eventBus.Subscribe("test", func(msg interface{}) {
			received <- msg
		})
		assert.NoError(t, err)
		defer sub.Unsubscribe()

		err = eventBus.Publish("test", "msg")
		assert.NoError(t, err)

		select {
		case msg := <-received:
			assert.Equal(t, "msg", msg)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("message not received")
		}

	})

	t.Run("unsubscribe", func(t *testing.T) {
		eventBus := subpub.NewSubPub()
		received := make(chan interface{}, 1)

		sub, err := eventBus.Subscribe("test", func(msg interface{}) {
			received <- msg
		})
		assert.NoError(t, err)

		sub.Unsubscribe()
		err = eventBus.Publish("test", "msg")
		assert.NoError(t, err)

		select {
		case <-received:
			t.Fatal("received message after unsubscribe")
		case <-time.After(100 * time.Millisecond):

		}
	})

	t.Run("close with active subscriptions", func(t *testing.T) {
		eventBus := subpub.NewSubPub()

		_, err := eventBus.Subscribe("test", func(msg interface{}) {})
		assert.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err = eventBus.Close(ctx)
		assert.NoError(t, err)
	})
}
