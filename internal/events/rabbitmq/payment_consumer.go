package rabbitmq

import (
	"context"
	"fmt"
	"log"

	"github.com/bytedance/sonic"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/ride4Low/contracts/events"
	"github.com/ride4Low/trip-service/internal/domain"
)

type PaymentEventHandler struct {
	service domain.Service
}

func NewPaymentEventHandler(service domain.Service) *PaymentEventHandler {
	return &PaymentEventHandler{
		service: service,
	}
}

func (h *PaymentEventHandler) Handle(ctx context.Context, msg amqp.Delivery) error {
	var message events.AmqpMessage
	if msg.Body == nil {
		return fmt.Errorf("message body is nil")
	}
	if err := sonic.Unmarshal(msg.Body, &message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %v", err)
	}

	switch msg.RoutingKey {
	case events.PaymentEventSuccess:
		return h.handlePaymentSuccess(ctx, message)
	default:
		return fmt.Errorf("unknown routing key: %s", msg.RoutingKey)
	}
}

func (h *PaymentEventHandler) handlePaymentSuccess(ctx context.Context, message events.AmqpMessage) error {
	var payload events.PaymentStatusUpdateData
	if err := sonic.Unmarshal(message.Data, &payload); err != nil {
		log.Printf("Failed to unmarshal payload: %v", err)
		return err
	}

	return h.service.UpdateTrip(
		ctx,
		payload.TripID,
		"paid",
		nil,
	)
}
