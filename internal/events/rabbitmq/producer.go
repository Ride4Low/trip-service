package rabbitmq

import "context"

// Producer interface
type Producer interface {
	PublishEvent(ctx context.Context, event interface{}) error
}

// producer struct implementing Producer
type producer struct {
}

func NewProducer() Producer {
	return &producer{}
}

func (p *producer) PublishEvent(ctx context.Context, event interface{}) error {
	return nil
}
