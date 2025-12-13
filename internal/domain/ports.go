package domain

import (
	"context"

	"github.com/ride4Low/contracts/proto/driver"
	"github.com/ride4Low/contracts/types"
)

// Service interface
type Service interface {
	CreateTrip(ctx context.Context, fare *types.RideFare) (*types.Trip, error)
	GetRoute(ctx context.Context, pickup, dropoff types.Coordinate) (*types.OsrmApiResponse, error)
	EstimatePackagesPriceWithRoute(route *types.OsrmApiResponse) []*types.RideFare
	CreateTripFares(ctx context.Context, fares []*types.RideFare, userID string, route *types.OsrmApiResponse) ([]*types.RideFare, error)
	GetAndValidateFare(ctx context.Context, fareID, userID string) (*types.RideFare, error)
	GetTripByID(ctx context.Context, id string) (*types.Trip, error)
	UpdateTrip(ctx context.Context, tripID string, status string, driver *driver.Driver) error
}

// Repository interface
type Repository interface {
	SaveRideFare(ctx context.Context, rideFare *types.RideFare) error
	GetRideFareByID(ctx context.Context, id string) (*types.RideFare, error)
	CreateTrip(ctx context.Context, trip *types.Trip) (*types.Trip, error)
	GetTripByID(ctx context.Context, id string) (*types.Trip, error)
	UpdateTrip(ctx context.Context, tripID string, status string, driver *driver.Driver) error
}

// RouteProvider interface
type RouteProvider interface {
	GetRoute(ctx context.Context, pickup, dropoff types.Coordinate) (*types.OsrmApiResponse, error)
}
