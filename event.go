package eventbus

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/go-redis/redis/v8"
	memorystore "github.com/stephenafamo/eventbus/store/memory"
	redisstore "github.com/stephenafamo/eventbus/store/redis"
)

type Event[Payload any] interface {
	RegisterHandler(id string, f EventHandler[Payload])
	UnregisterHandler(id string)
	Publish(ctx context.Context, payload Payload) error
}

type EventHandler[Payload any] interface {
	Handle(payload Payload)
}

type EventHandlerFunc[Payload any] func(payload Payload)

func (e EventHandlerFunc[Payload]) Handle(payload Payload) {
	e(payload)
}

func NewMemoryEvent[Payload any](ctx context.Context, buffer int) (Event[Payload], error) {
	var store = memorystore.New[Payload](buffer)
	return NewEvent[Payload](ctx, store)
}

func NewRedisEvent[Payload any](ctx context.Context, client *redis.Client, channel string) (Event[Payload], error) {
	var store = redisstore.New[Payload](client, channel)
	return NewEvent[Payload](ctx, store)
}

func NewEvent[Payload any](ctx context.Context, store Store[Payload]) (Event[Payload], error) {
	channel, err := store.Subscribe(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not subscribe: %w", err)
	}

	e := &event[Payload]{
		store:   store,
		channel: channel,
	}

	// Will exit when the subscription channel closes
	go e.subscribe(ctx)

	return e, nil
}

type event[Payload any] struct {
	store    Store[Payload]
	handlers map[string]EventHandler[Payload]
	mu       sync.RWMutex
	channel  <-chan Payload
}

func (e *event[Payload]) subscribe(ctx context.Context) {
	for payload := range e.channel {
		// so that the defer is scoped
		func() {
			e.mu.RLock()
			defer e.mu.RUnlock()
			for _, handler := range e.handlers {
				go handler.Handle(payload)
			}
		}()
	}

	log.Println("exiting")
}

// Register registers an event handler
func (e *event[Payload]) RegisterHandler(id string, f EventHandler[Payload]) {
	log.Printf("registering handler %q", id)
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.handlers == nil {
		e.handlers = map[string]EventHandler[Payload]{}
	}

	// Maybe return an error if the id is duplicated?
	// or panic?
	e.handlers[id] = f
	log.Printf("registered handler %q", id)
}

// Register registers an event handler
func (c *event[Payload]) UnregisterHandler(id string) {
	log.Printf("unregistering handler %q", id)
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.handlers, id)
}

func (e *event[Payload]) Publish(ctx context.Context, payload Payload) error {
	return e.store.Publish(ctx, payload)
}
