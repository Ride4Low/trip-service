package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/ride4Low/contracts/proto/trip"
	"github.com/ride4Low/contracts/types"
	"github.com/ride4Low/trip-service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockService is a mock implementation of service.Service for testing
type mockService struct {
	getRouteFunc                       func(ctx context.Context, pickup, dropoff types.Coordinate) (*domain.OsrmApiResponse, error)
	estimatePackagesPriceWithRouteFunc func(route *domain.OsrmApiResponse) []*domain.RideFare
	createTripFaresFunc                func(ctx context.Context, rideFares []*domain.RideFare, userID string, route *domain.OsrmApiResponse) ([]*domain.RideFare, error)
	getAndValidateFareFunc             func(ctx context.Context, fareID, userID string) (*domain.RideFare, error)
	createTripFunc                     func(ctx context.Context, fare *domain.RideFare) (*domain.Trip, error)
}

func (m *mockService) GetRoute(ctx context.Context, pickup, dropoff types.Coordinate) (*domain.OsrmApiResponse, error) {
	if m.getRouteFunc != nil {
		return m.getRouteFunc(ctx, pickup, dropoff)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) CreateTrip(ctx context.Context, fare *domain.RideFare) (*domain.Trip, error) {
	if m.createTripFunc != nil {
		return m.createTripFunc(ctx, fare)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) EstimatePackagesPriceWithRoute(route *domain.OsrmApiResponse) []*domain.RideFare {
	if m.estimatePackagesPriceWithRouteFunc != nil {
		return m.estimatePackagesPriceWithRouteFunc(route)
	}
	return nil
}

func (m *mockService) GetAndValidateFare(ctx context.Context, fareID, userID string) (*domain.RideFare, error) {
	if m.getAndValidateFareFunc != nil {
		return m.getAndValidateFareFunc(ctx, fareID, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) CreateTripFares(ctx context.Context, rideFares []*domain.RideFare, userID string, route *domain.OsrmApiResponse) ([]*domain.RideFare, error) {
	if m.createTripFaresFunc != nil {
		return m.createTripFaresFunc(ctx, rideFares, userID, route)
	}
	return nil, errors.New("not implemented")
}

func TestPreviewTrip(t *testing.T) {
	t.Run("successful route retrieval", func(t *testing.T) {
		// Create mock service that returns a successful response
		mockSvc := &mockService{
			getRouteFunc: func(ctx context.Context, pickup, dropoff types.Coordinate) (*domain.OsrmApiResponse, error) {
				// Verify the coordinates are passed correctly
				if pickup.Latitude != 13.736717 || pickup.Longitude != 100.523186 {
					t.Errorf("unexpected pickup coordinates: %+v", pickup)
				}
				if dropoff.Latitude != 13.746717 || dropoff.Longitude != 100.533186 {
					t.Errorf("unexpected dropoff coordinates: %+v", dropoff)
				}

				return &domain.OsrmApiResponse{
					Routes: []struct {
						Distance float64 `json:"distance"`
						Duration float64 `json:"duration"`
						Geometry struct {
							Coordinates [][]float64 `json:"coordinates"`
						} `json:"geometry"`
					}{
						{
							Distance: 1234.5,
							Duration: 567.8,
							Geometry: struct {
								Coordinates [][]float64 `json:"coordinates"`
							}{
								Coordinates: [][]float64{
									{100.523186, 13.736717},
									{100.533186, 13.746717},
								},
							},
						},
					},
				}, nil
			},
			estimatePackagesPriceWithRouteFunc: func(route *domain.OsrmApiResponse) []*domain.RideFare {
				return []*domain.RideFare{}
			},
			createTripFaresFunc: func(ctx context.Context, rideFares []*domain.RideFare, userID string, route *domain.OsrmApiResponse) ([]*domain.RideFare, error) {
				return []*domain.RideFare{}, nil
			},
			createTripFunc: func(ctx context.Context, fare *domain.RideFare) (*domain.Trip, error) {
				return &domain.Trip{}, nil
			},
		}

		h := &handler{svc: mockSvc}

		req := &trip.PreviewTripRequest{
			UserID: "user123",
			PickupLocation: &trip.Coordinate{
				Latitude:  13.736717,
				Longitude: 100.523186,
			},
			DropoffLocation: &trip.Coordinate{
				Latitude:  13.746717,
				Longitude: 100.533186,
			},
		}

		resp, err := h.PreviewTrip(context.Background(), req)

		// Assertions
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp == nil {
			t.Fatal("expected response, got nil")
		}
		if resp.Route == nil {
			t.Fatal("expected route, got nil")
		}
		if resp.Route.Distance != 1234.5 {
			t.Errorf("expected distance 1234.5, got %f", resp.Route.Distance)
		}
		if resp.Route.Duration != 567.8 {
			t.Errorf("expected duration 567.8, got %f", resp.Route.Duration)
		}
		if len(resp.Route.Geometry) != 1 {
			t.Fatalf("expected 1 geometry, got %d", len(resp.Route.Geometry))
		}
		if len(resp.Route.Geometry[0].Coordinates) != 2 {
			t.Errorf("expected 2 coordinates, got %d", len(resp.Route.Geometry[0].Coordinates))
		}
	})

	t.Run("missing pickup location", func(t *testing.T) {
		mockSvc := &mockService{}
		h := &handler{svc: mockSvc}

		req := &trip.PreviewTripRequest{
			UserID: "user123",
			DropoffLocation: &trip.Coordinate{
				Latitude:  13.746717,
				Longitude: 100.533186,
			},
		}

		resp, err := h.PreviewTrip(context.Background(), req)

		// Should return InvalidArgument error
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatal("expected gRPC status error")
		}
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument code, got %v", st.Code())
		}
		if resp != nil {
			t.Errorf("expected nil response, got %v", resp)
		}
	})

	t.Run("missing dropoff location", func(t *testing.T) {
		mockSvc := &mockService{}
		h := &handler{svc: mockSvc}

		req := &trip.PreviewTripRequest{
			UserID: "user123",
			PickupLocation: &trip.Coordinate{
				Latitude:  13.736717,
				Longitude: 100.523186,
			},
		}

		resp, err := h.PreviewTrip(context.Background(), req)

		// Should return InvalidArgument error
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatal("expected gRPC status error")
		}
		if st.Code() != codes.InvalidArgument {
			t.Errorf("expected InvalidArgument code, got %v", st.Code())
		}
		if resp != nil {
			t.Errorf("expected nil response, got %v", resp)
		}
	})

	t.Run("service returns error", func(t *testing.T) {
		// Create mock service that returns an error
		mockSvc := &mockService{
			getRouteFunc: func(ctx context.Context, pickup, dropoff types.Coordinate) (*domain.OsrmApiResponse, error) {
				return nil, errors.New("OSRM service unavailable")
			},
		}

		h := &handler{svc: mockSvc}

		req := &trip.PreviewTripRequest{
			UserID: "user123",
			PickupLocation: &trip.Coordinate{
				Latitude:  13.736717,
				Longitude: 100.523186,
			},
			DropoffLocation: &trip.Coordinate{
				Latitude:  13.746717,
				Longitude: 100.533186,
			},
		}

		resp, err := h.PreviewTrip(context.Background(), req)

		// Should return Internal error
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatal("expected gRPC status error")
		}
		if st.Code() != codes.Internal {
			t.Errorf("expected Internal code, got %v", st.Code())
		}
		if resp != nil {
			t.Errorf("expected nil response, got %v", resp)
		}
	})
}
