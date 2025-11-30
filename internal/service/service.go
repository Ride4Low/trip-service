package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ride4Low/contracts/types"
	"github.com/ride4Low/trip-service/internal/domain"
)

// Service interface
type Service interface {
	CreateTrip(ctx context.Context) error
	GetRoute(ctx context.Context, pickup, dropoff types.Coordinate) (*domain.OsrmApiResponse, error)
}

// service struct implementing Service
type service struct {
	osrmURL string
}

func NewService(osrmURL string) Service {
	return &service{
		osrmURL: osrmURL,
	}
}

func (s *service) CreateTrip(ctx context.Context) error {
	return nil
}

func (s *service) GetRoute(ctx context.Context, pickup, dropoff types.Coordinate) (*domain.OsrmApiResponse, error) {
	url := fmt.Sprintf(
		"%s/route/v1/driving/%f,%f;%f,%f?overview=full&geometries=geojson",
		s.osrmURL,
		pickup.Longitude, pickup.Latitude,
		dropoff.Longitude, dropoff.Latitude,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OSRM request failed with status code: %d", resp.StatusCode)
	}

	var osrmResponse domain.OsrmApiResponse
	if err := json.Unmarshal(body, &osrmResponse); err != nil {
		return nil, err
	}

	return &osrmResponse, nil
}
