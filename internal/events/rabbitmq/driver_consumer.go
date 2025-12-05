package rabbitmq

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ride4Low/contracts/events"
	"github.com/ride4Low/contracts/pkg/rabbitmq"
	"github.com/ride4Low/trip-service/internal/domain"
)

type DriverEventHandler struct {
	publisher *rabbitmq.Publisher
	service   domain.Service
}

func NewDriverEventHandler(publisher *rabbitmq.Publisher, service domain.Service) *DriverEventHandler {
	return &DriverEventHandler{
		publisher: publisher,
		service:   service,
	}
}

func (h *DriverEventHandler) Handle(ctx context.Context, msg amqp.Delivery) error {
	var message events.AmqpMessage

	if msg.Body == nil {
		return fmt.Errorf("message body is nil")
	}

	if err := sonic.Unmarshal(msg.Body, &message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	switch msg.RoutingKey {
	case events.DriverCmdTripAccept:
		return h.handleTripAccept(ctx, message)
	case events.DriverCmdTripDecline:
		return h.handleTripDecline(ctx, message)
	default:
		return fmt.Errorf("unknown routing key: %s", msg.RoutingKey)
	}
}

func (h *DriverEventHandler) handleTripAccept(ctx context.Context, message events.AmqpMessage) error {
	return nil
}

func (h *DriverEventHandler) handleTripDecline(ctx context.Context, message events.AmqpMessage) error {
	// When a driver declines, we try to find another driver
	var payload events.DriverTripResponseData
	if err := sonic.Unmarshal(message.Data, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	trip, err := h.service.GetTripByID(ctx, payload.TripID)
	if err != nil {
		return err
	}

	newPayload := events.TripEventData{
		Trip: trip.ToProto(),
	}

	marshalledPayload, err := sonic.Marshal(newPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	if err := h.publisher.PublishMessage(ctx, events.TripEventDriverNotInterested,
		events.AmqpMessage{
			OwnerID: payload.RiderID,
			Data:    marshalledPayload,
		},
	); err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	return nil
}
