package domain

import (
	"context"

	"github.com/ride4Low/contracts/types"
)

// Service interface
type Service interface {
	CreateTrip(ctx context.Context, fare *RideFare) (*Trip, error)
	GetRoute(ctx context.Context, pickup, dropoff types.Coordinate) (*OsrmApiResponse, error)
	EstimatePackagesPriceWithRoute(route *OsrmApiResponse) []*RideFare
	CreateTripFares(ctx context.Context, fares []*RideFare, userID string, route *OsrmApiResponse) ([]*RideFare, error)
	GetAndValidateFare(ctx context.Context, fareID, userID string) (*RideFare, error)
}

// Repository interface
type Repository interface {
	SaveRideFare(ctx context.Context, rideFare *RideFare) error
	GetRideFareByID(ctx context.Context, id string) (*RideFare, error)
	CreateTrip(ctx context.Context, trip *Trip) (*Trip, error)
}

// RouteProvider interface
type RouteProvider interface {
	GetRoute(ctx context.Context, pickup, dropoff types.Coordinate) (*OsrmApiResponse, error)
}
