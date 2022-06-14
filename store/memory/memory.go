package memorystore

import (
	"context"
	"sync"

	"github.com/gofrs/uuid"
)

func New[Payload any](buffer int) *memoryStore[Payload] {
	return &memoryStore[Payload]{
		buffer:   buffer,
		channels: make(map[[16]byte]chan Payload),
	}
}

type memoryStore[Payload any] struct {
	buffer   int
	mu       sync.RWMutex
	channels map[[16]byte]chan Payload
}

func (m *memoryStore[Payload]) Publish(ctx context.Context, payload Payload) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, c := range m.channels {
		theChan := c
		go func() {
			theChan <- payload
		}()
	}

	return nil
}

func (m *memoryStore[Payload]) Subscribe(ctx context.Context) (<-chan Payload, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c := make(chan Payload, m.buffer)

	var x [16]byte

	for {
		// keep generating a new one until an unused one is found
		x, _ = uuid.NewV4()
		if _, ok := m.channels[x]; ok {
			continue
		}

		m.channels[x] = c
		break
	}

	// close and cleanup the channel
	go func() {
		<-ctx.Done()

		m.mu.Lock()
		defer m.mu.Unlock()

		close(c)
		delete(m.channels, x)
	}()

	return c, nil
}
