package pubsub

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisBroker struct {
	client *redis.Client
}

func NewRedisBroker(ctx context.Context, address string, password string) (*RedisBroker, error) {
	// Initial connection pool and client
	rdb := redis.NewClient(&redis.Options{
		Addr: address, Password: password, DB: 0,
	})

	// NewClient() is lazy initialization, need Ping to check connection
	err := rdb.Ping(ctx).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to ping redis at %s: %w", address, err)
	}

	return &RedisBroker{client: rdb}, nil
}

func (rdb *RedisBroker) Publish(ctx context.Context, topic string, msg string) error {
	client := rdb.client
	err := client.Publish(ctx, topic, msg).Err() // publish message then return the connection error

	if err != nil {
		return fmt.Errorf("Failed to publish message to topic '%s': %w", topic, err)
	}
	return nil
}

func (rdb *RedisBroker) Subscribe(ctx context.Context, topic string) (<-chan string, error) {
	client := rdb.client
	sub := client.Subscribe(ctx, topic) // Subscribe to topic

	_, err := sub.Receive(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to subscribe to topic %s: %w", topic, err)
	}

	msgChannel := make(chan string) //create message channel that only return string

	go func() {
		defer sub.Close()
		defer close(msgChannel)

		// channel from redis (channel type is *redis.Message - include topic,pattern,payload)
		redisChannel := sub.Channel()
		for {
			select {
			case <-ctx.Done(): // time out or user Ctrl C
				return
			case msg, ok := <-redisChannel:
				if !ok {
					return
				}
				msgChannel <- msg.Payload
			}
		}
	}()
	return msgChannel, nil
}
