package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ride4Low/contracts/types"
	"github.com/ride4Low/trip-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// service struct implementing Service
type service struct {
	osrmURL string
	repo    domain.Repository
}

func NewService(osrmURL string, repo domain.Repository) domain.Service {
	return &service{
		osrmURL: osrmURL,
		repo:    repo,
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

func (s *service) GenerateTripFares(ctx context.Context, rideFares []*domain.RideFare, userID string, route *domain.OsrmApiResponse) ([]*domain.RideFare, error) {
	fares := make([]*domain.RideFare, len(rideFares))

	for i, f := range rideFares {
		id := primitive.NewObjectID()

		fare := &domain.RideFare{
			UserID:            userID,
			ID:                id,
			TotalPriceInCents: f.TotalPriceInCents,
			PackageSlug:       f.PackageSlug,
			Route:             route,
		}

		if err := s.repo.SaveRideFare(ctx, fare); err != nil {
			return nil, fmt.Errorf("failed to save trip fare: %w", err)
		}

		fares[i] = fare
	}

	return fares, nil
}
