package service

import (
	"context"
	"fmt"

	"github.com/ride4Low/contracts/proto/driver"
	"github.com/ride4Low/contracts/proto/trip"
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

func (s *service) CreateTrip(ctx context.Context, fare *types.RideFare) (*types.Trip, error) {
	t := &types.Trip{
		ID:       primitive.NewObjectID(),
		UserID:   fare.UserID,
		Status:   "pending",
		RideFare: fare,
		Driver:   &trip.TripDriver{},
	}

	return s.repo.CreateTrip(ctx, t)
}

func (s *service) GetRoute(ctx context.Context, pickup, dropoff types.Coordinate) (*types.OsrmApiResponse, error) {
	return s.routeProvider.GetRoute(ctx, pickup, dropoff)
}

func (s *service) CreateTripFares(ctx context.Context, rideFares []*types.RideFare, userID string, route *types.OsrmApiResponse) ([]*types.RideFare, error) {
	fares := make([]*types.RideFare, len(rideFares))

	for i, f := range rideFares {
		id := primitive.NewObjectID()

		fare := &types.RideFare{
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

func (s *service) GetAndValidateFare(ctx context.Context, fareID, userID string) (*types.RideFare, error) {
	fare, err := s.repo.GetRideFareByID(ctx, fareID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trip fare: %w", err)
	}

	if fare == nil {
		return nil, fmt.Errorf("fare does not exist")
	}

	// User fare validation (user is owner of this fare?)
	if userID != fare.UserID {
		return nil, fmt.Errorf("fare does not belong to the user")
	}

	return fare, nil
}

func (s *service) GetTripByID(ctx context.Context, id string) (*types.Trip, error) {
	return s.repo.GetTripByID(ctx, id)
}

func (s *service) UpdateTrip(ctx context.Context, tripID string, status string, driver *driver.Driver) error {
	return s.repo.UpdateTrip(ctx, tripID, status, driver)
}
