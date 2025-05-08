package tests

import (
	s "VK_task/pkg/subpub"
	"context"
	"sync"
	"testing"
)

func TestSubPub(t *testing.T) {
	t.Run("Subscribe/Publish", func(t *testing.T) {
		bus := s.NewSubPub()
		defer bus.Close(context.Background())

		var wg sync.WaitGroup
		wg.Add(1)

		sub, err := bus.Subscribe("test", func(msg interface{}) {
			defer wg.Done()
			if msg != "hello" {
				t.Error()
			}
		})
		if err != nil {
			t.Error()
		}
		defer sub.Unsubscribe()

		if err := bus.Publish("test", "hello"); err != nil {
			t.Error()
		}

		wg.Wait()
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		EventBus := s.NewSubPub()
		defer EventBus.Close(context.Background())

		called := false
		sub, err := EventBus.Subscribe("test", func(msg interface{}) {
			called = true
		})
		if err != nil {
			t.Error()
		}

		sub.Unsubscribe()
		EventBus.Publish("test", "hello")

		if called {
			t.Error()
		}
	})

}
