package rabbitmq

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/ride4Low/contracts/events"
	"github.com/ride4Low/contracts/pkg/rabbitmq"
	"github.com/ride4Low/trip-service/internal/domain"
)

type TripEventPublisher struct {
	publisher *rabbitmq.Publisher
}

func NewTripEventPublisher(publisher *rabbitmq.Publisher) TripEventPublisher {
	return TripEventPublisher{
		publisher: publisher,
	}
}

// will be consumed by driver service to find available drivers
func (p *TripEventPublisher) PublishTripCreated(ctx context.Context, trip *domain.Trip) error {
	payload := events.TripEventData{
		Trip: trip.ToProto(),
	}

	data, err := sonic.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %v", err)
	}

	amqpMsg := events.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    data,
	}

	return p.publisher.PublishMessage(ctx, events.TripEventCreated, amqpMsg)
}
