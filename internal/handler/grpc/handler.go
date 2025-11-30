package grpc

import (
	"context"

	"github.com/ride4Low/contracts/proto/trip"
	"github.com/ride4Low/contracts/types"
	"github.com/ride4Low/trip-service/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler struct placeholder
type handler struct {
	trip.UnimplementedTripServiceServer
	svc service.Service
}

func NewHandler(server *grpc.Server, svc service.Service) *handler {
	h := &handler{
		svc: svc,
	}
	trip.RegisterTripServiceServer(server, h)
	return h
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

	return &trip.PreviewTripResponse{
		Route: osrmResponse.ToProto(),
	}, nil
}
