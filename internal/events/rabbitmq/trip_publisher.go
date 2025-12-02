package rabbitmq

import (
	"context"

	"github.com/ride4Low/contracts/events"
	"github.com/ride4Low/contracts/pkg/rabbitmq"
	"github.com/ride4Low/trip-service/internal/domain"
)

type TripEventPublisher struct {
	rabbitMQ *rabbitmq.RabbitMQ
}

func NewTripEventPublisher(rmq *rabbitmq.RabbitMQ) TripEventPublisher {
	return TripEventPublisher{
		rabbitMQ: rmq,
	}
}

func (p *TripEventPublisher) PublishTripCreated(ctx context.Context, trip domain.Trip) error {
	payload := events.TripEventData{
		Trip: trip.ToProto(),
	}

	amqpMsg := events.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    payload,
	}

	return p.rabbitMQ.PublishMessage(ctx, events.TripEventCreated, amqpMsg)
}
