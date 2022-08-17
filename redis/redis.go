package redisevents

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
	"github.com/stephenafamo/eventbus"
)

func newStore[Payload any](client *redis.Client, channel string) *redisStore[Payload] {
	return &redisStore[Payload]{
		client:  client,
		channel: channel,
	}
}

func New[Payload any](ctx context.Context, client *redis.Client, channelName string) (eventbus.Event[Payload], error) {
	store := newStore[Payload](client, channelName)
	return eventbus.NewEvent[Payload](ctx, store)
}

// An implemnetation of KVStore based on Redis
type redisStore[Payload any] struct {
	client  *redis.Client
	channel string
}

func (r *redisStore[Payload]) Publish(ctx context.Context, payload Payload) error {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(payload)
	if err != nil {
		return fmt.Errorf("gob encoding: %w", err)
	}

	return r.client.Publish(ctx, r.channel, buf.String()).Err()
}

func (r *redisStore[Payload]) Subscribe(ctx context.Context, f func(Payload)) error {
	sub := r.client.Subscribe(ctx, r.channel)

	go func() {
		for {
			select {
			case msg := <-sub.Channel():
				var payload Payload
				buf := bytes.NewBufferString(msg.Payload)
				err := gob.NewDecoder(buf).Decode(&payload)
				if err != nil {
					fmt.Printf("ERROR: %v", err)
				}
				f(payload)
				log.Printf("sending to channel %#v", payload)
			case <-ctx.Done():
				// close and cleanup the channels
				sub.Close()
				return
			}
		}
	}()

	return nil
}
