package service

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ride4Low/contracts/types"
	"github.com/ride4Low/trip-service/internal/adapter/osrm"
	"github.com/ride4Low/trip-service/internal/domain"
)

// mockRepository is a mock implementation of domain.Repository for testing
type mockRepository struct {
	saveTripFunc     func(ctx context.Context) error
	saveRideFareFunc func(ctx context.Context, rideFare *domain.RideFare) error
	getRideFareFunc  func(ctx context.Context, fareID string) (*domain.RideFare, error)
	createTripFunc   func(ctx context.Context, fare *domain.Trip) (*domain.Trip, error)
}

func (m *mockRepository) SaveTrip(ctx context.Context) error {
	if m.saveTripFunc != nil {
		return m.saveTripFunc(ctx)
	}
	return nil
}

func (m *mockRepository) SaveRideFare(ctx context.Context, rideFare *domain.RideFare) error {
	if m.saveRideFareFunc != nil {
		return m.saveRideFareFunc(ctx, rideFare)
	}
	return nil
}

func (m *mockRepository) GetRideFareByID(ctx context.Context, fareID string) (*domain.RideFare, error) {
	if m.getRideFareFunc != nil {
		return m.getRideFareFunc(ctx, fareID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) CreateTrip(ctx context.Context, trip *domain.Trip) (*domain.Trip, error) {
	if m.createTripFunc != nil {
		return m.createTripFunc(ctx, trip)
	}
	return nil, errors.New("not implemented")
}

func TestCreateTrip(t *testing.T) {
	t.Run("successful trip creation", func(t *testing.T) {
		// Setup
		mockRepo := &mockRepository{
			createTripFunc: func(ctx context.Context, trip *domain.Trip) (*domain.Trip, error) {
				return trip, nil
			},
		}
		svc := NewService(nil, mockRepo)

		// Execute
		_, err := svc.CreateTrip(context.Background(), nil)

		// Verify
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestGetRoute(t *testing.T) {
	t.Run("successful route retrieval", func(t *testing.T) {
		// Create a mock OSRM server
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify the request path
			expectedPath := "/route/v1/driving/100.523186,13.736717;100.523186,13.736717"
			if !strings.HasPrefix(r.URL.Path, expectedPath) {
				t.Errorf("unexpected path: got %v", r.URL.Path)
			}

			// Return a mock OSRM response
			response := `{
				"routes": [{
					"distance": 1234.5,
					"duration": 567.8,
					"geometry": {
						"coordinates": [[100.523186, 13.736717], [100.523200, 13.736800]]
					}
				}]
			}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		}))
		defer mockServer.Close()

		// Create service with mock server URL
		svc := NewService(osrm.NewClient(mockServer.URL), nil)

		// Test GetRoute
		pickup := types.Coordinate{Latitude: 13.736717, Longitude: 100.523186}
		dropoff := types.Coordinate{Latitude: 13.736717, Longitude: 100.523186}

		result, err := svc.GetRoute(context.Background(), pickup, dropoff)

		// Assertions
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("expected result, got nil")
		}
		if len(result.Routes) != 1 {
			t.Fatalf("expected 1 route, got %d", len(result.Routes))
		}
		if result.Routes[0].Distance != 1234.5 {
			t.Errorf("expected distance 1234.5, got %f", result.Routes[0].Distance)
		}
		if result.Routes[0].Duration != 567.8 {
			t.Errorf("expected duration 567.8, got %f", result.Routes[0].Duration)
		}
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		// Create a mock server that returns invalid JSON
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`invalid json`))
		}))
		defer mockServer.Close()

		svc := NewService(osrm.NewClient(mockServer.URL), nil)
		pickup := types.Coordinate{Latitude: 13.736717, Longitude: 100.523186}
		dropoff := types.Coordinate{Latitude: 13.736717, Longitude: 100.523186}

		result, err := svc.GetRoute(context.Background(), pickup, dropoff)

		if err == nil {
			t.Fatal("expected error for invalid JSON, got nil")
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("HTTP error", func(t *testing.T) {
		// Create a mock server that returns an error
		mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer mockServer.Close()

		svc := NewService(osrm.NewClient(mockServer.URL), nil)
		pickup := types.Coordinate{Latitude: 13.736717, Longitude: 100.523186}
		dropoff := types.Coordinate{Latitude: 13.736717, Longitude: 100.523186}

		result, err := svc.GetRoute(context.Background(), pickup, dropoff)

		// Should return an error about the status code
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "status code") {
			t.Errorf("expected error to mention status code, got: %v", err)
		}
		if !strings.Contains(err.Error(), "500") {
			t.Errorf("expected error to mention status code 500, got: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})
}

func TestCreateTripFares(t *testing.T) {
	t.Run("successful creation of multiple fares", func(t *testing.T) {
		// Setup
		savedFares := []*domain.RideFare{}
		mockRepo := &mockRepository{
			saveRideFareFunc: func(ctx context.Context, rideFare *domain.RideFare) error {
				savedFares = append(savedFares, rideFare)
				return nil
			},
		}
		svc := NewService(nil, mockRepo)

		// Create mock route
		route := &domain.OsrmApiResponse{
			Routes: []struct {
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
				Geometry struct {
					Coordinates [][]float64 `json:"coordinates"`
				} `json:"geometry"`
			}{
				{
					Distance: 1000.0,
					Duration: 600.0,
					Geometry: struct {
						Coordinates [][]float64 `json:"coordinates"`
					}{
						Coordinates: [][]float64{{100.0, 13.0}, {100.1, 13.1}},
					},
				},
			},
		}

		// Input fares
		inputFares := []*domain.RideFare{
			{PackageSlug: "suv", TotalPriceInCents: 1850.0},
			{PackageSlug: "sedan", TotalPriceInCents: 2000.0},
			{PackageSlug: "van", TotalPriceInCents: 2050.0},
		}
		userID := "test-user-123"

		// Execute
		result, err := svc.CreateTripFares(context.Background(), inputFares, userID, route)

		// Verify
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("expected result, got nil")
		}
		if len(result) != 3 {
			t.Fatalf("expected 3 fares, got %d", len(result))
		}
		if len(savedFares) != 3 {
			t.Fatalf("expected 3 fares saved to repository, got %d", len(savedFares))
		}

		// Verify each fare
		for i, fare := range result {
			if fare.UserID != userID {
				t.Errorf("fare %d: expected userID %s, got %s", i, userID, fare.UserID)
			}
			if fare.PackageSlug != inputFares[i].PackageSlug {
				t.Errorf("fare %d: expected package %s, got %s", i, inputFares[i].PackageSlug, fare.PackageSlug)
			}
			if fare.TotalPriceInCents != inputFares[i].TotalPriceInCents {
				t.Errorf("fare %d: expected price %f, got %f", i, inputFares[i].TotalPriceInCents, fare.TotalPriceInCents)
			}
			if fare.Route != route {
				t.Errorf("fare %d: route not set correctly", i)
			}
			if fare.ID.IsZero() {
				t.Errorf("fare %d: ID should be generated", i)
			}
		}
	})

	t.Run("repository error on save", func(t *testing.T) {
		// Setup
		expectedErr := errors.New("database connection failed")
		mockRepo := &mockRepository{
			saveRideFareFunc: func(ctx context.Context, rideFare *domain.RideFare) error {
				return expectedErr
			},
		}
		svc := NewService(nil, mockRepo)

		route := &domain.OsrmApiResponse{
			Routes: []struct {
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
				Geometry struct {
					Coordinates [][]float64 `json:"coordinates"`
				} `json:"geometry"`
			}{
				{Distance: 1000.0, Duration: 600.0},
			},
		}

		inputFares := []*domain.RideFare{
			{PackageSlug: "suv", TotalPriceInCents: 1850.0},
		}

		// Execute
		result, err := svc.CreateTripFares(context.Background(), inputFares, "test-user", route)

		// Verify
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to save trip fare") {
			t.Errorf("expected error to mention 'failed to save trip fare', got: %v", err)
		}
		if !errors.Is(err, expectedErr) {
			t.Errorf("expected error to wrap original error")
		}
		if result != nil {
			t.Errorf("expected nil result on error, got %v", result)
		}
	})

	t.Run("empty fares array", func(t *testing.T) {
		// Setup
		mockRepo := &mockRepository{}
		svc := NewService(nil, mockRepo)

		route := &domain.OsrmApiResponse{
			Routes: []struct {
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
				Geometry struct {
					Coordinates [][]float64 `json:"coordinates"`
				} `json:"geometry"`
			}{
				{Distance: 1000.0, Duration: 600.0},
			},
		}

		// Execute with empty array
		result, err := svc.CreateTripFares(context.Background(), []*domain.RideFare{}, "test-user", route)

		// Verify
		if err != nil {
			t.Errorf("expected no error for empty array, got %v", err)
		}
		if result == nil {
			t.Fatal("expected empty array, got nil")
		}
		if len(result) != 0 {
			t.Errorf("expected empty array, got %d items", len(result))
		}
	})
}
