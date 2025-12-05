package domain

import (
	"context"

	"github.com/ride4Low/contracts/proto/driver"
	"github.com/ride4Low/contracts/types"
)

// Service interface
type Service interface {
	CreateTrip(ctx context.Context, fare *RideFare) (*Trip, error)
	GetRoute(ctx context.Context, pickup, dropoff types.Coordinate) (*OsrmApiResponse, error)
	EstimatePackagesPriceWithRoute(route *OsrmApiResponse) []*RideFare
	CreateTripFares(ctx context.Context, fares []*RideFare, userID string, route *OsrmApiResponse) ([]*RideFare, error)
	GetAndValidateFare(ctx context.Context, fareID, userID string) (*RideFare, error)
	GetTripByID(ctx context.Context, id string) (*Trip, error)
	UpdateTrip(ctx context.Context, tripID string, status string, driver *driver.Driver) error
}

// Repository interface
type Repository interface {
	SaveRideFare(ctx context.Context, rideFare *RideFare) error
	GetRideFareByID(ctx context.Context, id string) (*RideFare, error)
	CreateTrip(ctx context.Context, trip *Trip) (*Trip, error)
	GetTripByID(ctx context.Context, id string) (*Trip, error)
	UpdateTrip(ctx context.Context, tripID string, status string, driver *driver.Driver) error
}

// RouteProvider interface
type RouteProvider interface {
	GetRoute(ctx context.Context, pickup, dropoff types.Coordinate) (*OsrmApiResponse, error)
}
