package pubsub

import "context"

type Publisher interface {
	Publish(ctx context.Context, topic string, msg string) error
}
type Subscriber interface {
	Subscribe(ctx context.Context, topic string) (<-chan string, error)
}
