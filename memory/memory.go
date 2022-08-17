package memoryevents

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/stephenafamo/eventbus"
)

func newStore[Payload any]() *memoryStore[Payload] {
	return &memoryStore[Payload]{
		callbacks: make(map[[16]byte]func(Payload)),
	}
}

func New[Payload any](ctx context.Context) (eventbus.Event[Payload], error) {
	store := newStore[Payload]()
	return eventbus.NewEvent[Payload](ctx, store)
}

type memoryStore[Payload any] struct {
	mu        sync.RWMutex
	callbacks map[[16]byte]func(Payload)
}

func (m *memoryStore[Payload]) Publish(ctx context.Context, payload Payload) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, c := range m.callbacks {
		go c(payload)
	}

	return nil
}

func (m *memoryStore[Payload]) Subscribe(ctx context.Context, f func(Payload)) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var x [16]byte

	for {
		// keep generating a new one until an unused one is found
		x, _ = uuid.NewV4()
		if _, ok := m.callbacks[x]; ok {
			continue
		}

		m.callbacks[x] = f
		break
	}

	// close and cleanup the channel
	go func() {
		<-ctx.Done()

		m.mu.Lock()
		defer m.mu.Unlock()

		delete(m.callbacks, x)
	}()

	return nil
}
