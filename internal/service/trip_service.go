package service

import (
	"context"
	"fmt"

	"github.com/ride4Low/contracts/types"
	"github.com/ride4Low/trip-service/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// service struct implementing Service
type service struct {
	routeProvider domain.RouteProvider
	repo          domain.Repository
}

func NewService(routeProvider domain.RouteProvider, repo domain.Repository) domain.Service {
	return &service{
		routeProvider: routeProvider,
		repo:          repo,
	}
}

func (s *service) CreateTrip(ctx context.Context) error {
	return nil
}

func (s *service) GetRoute(ctx context.Context, pickup, dropoff types.Coordinate) (*domain.OsrmApiResponse, error) {
	return s.routeProvider.GetRoute(ctx, pickup, dropoff)
}

func (s *service) CreateTripFares(ctx context.Context, rideFares []*domain.RideFare, userID string, route *domain.OsrmApiResponse) ([]*domain.RideFare, error) {
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
