package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ride4Low/contracts/types"
)

func TestCreateTrip(t *testing.T) {
	// Test case placeholder
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
		svc := NewService(mockServer.URL)

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

		svc := NewService(mockServer.URL)
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

		svc := NewService(mockServer.URL)
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
