package eventbus

import (
	"context"
)

type Store[Payload any] interface {
	Publish(ctx context.Context, payload Payload) error

	// Free any resources when the context is done
	Subscribe(ctx context.Context) (<-chan Payload, error)
}
