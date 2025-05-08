package subpub

import (
	"context"
	"sync"
)

type MessageHandler func(msg interface{})

type Subscription interface {
	Unsubscribe()
}

type SubPub interface {
	Subscribe(subject string, cb MessageHandler) (Subscription, error)

	Publish(subject string, msg interface{}) error

	Close(ctx context.Context) error
}

type subscription struct {
	subject string
	handler MessageHandler
	msgChan chan interface{}
	cancel  context.CancelFunc
	subPub  *subPubImpl
	wg      *sync.WaitGroup
}

func (s *subscription) Unsubscribe() {
	s.subPub.unsubscribe(s)
}

// EventBus
type subPubImpl struct {
	mu          sync.Mutex
	subscribers map[string][]*subscription
	wg          sync.WaitGroup
	closed      bool
}

func NewSubPub() SubPub {
	return &subPubImpl{
		subscribers: make(map[string][]*subscription),
	}
}

func (eb *subPubImpl) Subscribe(subject string, cb MessageHandler) (Subscription, error) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	sub := &subscription{
		subject: subject,
		handler: cb,
		msgChan: make(chan interface{}, 100),
		cancel:  cancel, //  функця отмены в подписке //Это позволяет внешне управлять горутиной (завершить её при необходимости)
		subPub:  eb,
		wg:      &eb.wg,
	}

	eb.subscribers[subject] = append(eb.subscribers[subject], sub)
	eb.wg.Add(1)

	go func() {
		defer eb.wg.Done()
		for {
			select {
			case msg, ok := <-sub.msgChan:
				if !ok {
					return
				}
				sub.handler(msg)
			case <-ctx.Done():
				return
			}
		}
	}()

	return sub, nil
}

func (eb *subPubImpl) Publish(subject string, msg interface{}) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subs, ok := eb.subscribers[subject]
	if !ok {
		return nil
	}

	for _, sub := range subs {
		select {
		case sub.msgChan <- msg:
		default:
			// если канал переполнен ничего не делаем
		}

	}

	return nil
}

func (eb *subPubImpl) unsubscribe(sub *subscription) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subs, ok := eb.subscribers[sub.subject]
	if !ok {
		return
	}

	for i, s := range subs {
		if s == sub {
			eb.subscribers[sub.subject] = append(subs[:i], subs[i+1:]...)

			sub.cancel()
			close(sub.msgChan)
			break
		}
	}

}

func (eb *subPubImpl) Close(ctx context.Context) error {
	eb.mu.Lock()
	if eb.closed {
		eb.mu.Unlock()
		return nil
	}
	eb.closed = true

	var allSubs []*subscription
	for _, subs := range eb.subscribers {
		allSubs = append(allSubs, subs...)
	}
	eb.mu.Unlock()

	for _, sub := range allSubs {
		sub.Unsubscribe()
	}

	done := make(chan struct{})
	go func() {
		eb.wg.Wait()
		close(done)
	}()

	select {
	case <-done: // Срабатывает при закрытии канала done
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}

}
