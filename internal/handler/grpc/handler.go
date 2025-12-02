package grpc

import (
	"context"

	"github.com/ride4Low/contracts/proto/trip"
	"github.com/ride4Low/contracts/types"
	"github.com/ride4Low/trip-service/internal/domain"
	"github.com/ride4Low/trip-service/internal/events/rabbitmq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler struct placeholder
type handler struct {
	trip.UnimplementedTripServiceServer
	svc       domain.Service
	publisher rabbitmq.TripEventPublisher
}

func NewHandler(server *grpc.Server, svc domain.Service, publisher rabbitmq.TripEventPublisher) *handler {
	h := &handler{
		svc:       svc,
		publisher: publisher,
	}
	trip.RegisterTripServiceServer(server, h)
	return h
}

func (h *handler) CreateTrip(ctx context.Context, req *trip.CreateTripRequest) (*trip.CreateTripResponse, error) {
	fareID := req.GetRideFareID()
	userID := req.GetUserID()

	rideFare, err := h.svc.GetAndValidateFare(ctx, fareID, userID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "failed to get and validate the fare: %v", err)
	}

	t, err := h.svc.CreateTrip(ctx, rideFare)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create the trip: %v", err)
	}

	if err := h.publisher.PublishTripCreated(ctx, t); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish the trip created event: %v", err)
	}

	return &trip.CreateTripResponse{
		TripID: t.ID.Hex(),
	}, nil
}

func (h *handler) PreviewTrip(ctx context.Context, req *trip.PreviewTripRequest) (*trip.PreviewTripResponse, error) {
	pickup := req.GetPickupLocation()
	dropoff := req.GetDropoffLocation()

	if pickup == nil || dropoff == nil {
		return nil, status.Error(codes.InvalidArgument, "pickup and dropoff locations are required")
	}

	pickupCoordinates := types.Coordinate{
		Latitude:  pickup.GetLatitude(),
		Longitude: pickup.GetLongitude(),
	}

	dropoffCoordinates := types.Coordinate{
		Latitude:  dropoff.GetLatitude(),
		Longitude: dropoff.GetLongitude(),
	}

	osrmResponse, err := h.svc.GetRoute(ctx, pickupCoordinates, dropoffCoordinates)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get route")
	}

	estimatedFares := h.svc.EstimatePackagesPriceWithRoute(osrmResponse)

	fares, err := h.svc.CreateTripFares(ctx, estimatedFares, req.GetUserID(), osrmResponse)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate the ride fares: %v", err)
	}

	return &trip.PreviewTripResponse{
		Route:     osrmResponse.ToProto(),
		RideFares: domain.ToRideFaresProto(fares),
	}, nil
}
