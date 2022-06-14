package redisstore

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/go-redis/redis/v8"
)

func New[Payload any](client *redis.Client, channel string) *redisStore[Payload] {
	return &redisStore[Payload]{
		client:  client,
		channel: channel,
	}
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

	log.Printf("publishing %s", payload)
	return r.client.Publish(ctx, r.channel, buf.String()).Err()
}

func (r *redisStore[Payload]) Subscribe(ctx context.Context) (<-chan Payload, error) {
	log.Println("subscribing...")
	pub := r.client.Subscribe(ctx, r.channel)
	rChan := pub.Channel()
	c := make(chan Payload, 0)

	// close and cleanup the channel
	go func() {
		<-ctx.Done()
		pub.Close()
		close(c)
	}()

	go func() {
		for msg := range rChan {
			var payload Payload
			var buf = bytes.NewBufferString(msg.Payload)
			err := gob.NewDecoder(buf).Decode(&payload)
			if err != nil {
				fmt.Printf("ERROR: %v", err)
			}
			c <- payload
			log.Printf("sending to channel %s", payload)
		}

		// The loop exits so that means that
		// the redis channel has been closed
		close(c)
	}()

	return c, nil
}
