package eventbus

import (
	"context"
	"errors"
	"sync"
)

var ErrDuplicateID = errors.New("Duplicate handler ID")

type Event[Payload any] interface {
	RegisterHandler(id string, f Handler[Payload]) error
	UnregisterHandler(id string)
	Publish(ctx context.Context, payload Payload) error
}

func NewEvent[Payload any](ctx context.Context, store Store[Payload]) (Event[Payload], error) {
	e := &event[Payload]{
		store: store,
	}

	// Will exit when the subscription channel closes
	if err := store.Subscribe(ctx, e.subscribe); err != nil {
		return nil, err
	}

	return e, nil
}

type event[Payload any] struct {
	store    Store[Payload]
	handlers map[string]Handler[Payload]
	mu       sync.RWMutex
}

func (e *event[Payload]) subscribe(payload Payload) {
	for _, handler := range e.handlers {
		go handler.Handle(payload)
	}
}

// Register registers an event handler
func (e *event[Payload]) RegisterHandler(id string, f Handler[Payload]) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.handlers == nil {
		e.handlers = map[string]Handler[Payload]{}
	}

	if _, ok := e.handlers[id]; ok {
		return ErrDuplicateID
	}

	e.handlers[id] = f

	return nil
}

// Register registers an event handler
func (c *event[Payload]) UnregisterHandler(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.handlers, id)
}

func (e *event[Payload]) Publish(ctx context.Context, payload Payload) error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.store.Publish(ctx, payload)
}
