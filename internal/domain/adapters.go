package domain

import (
	"context"

	"github.com/ride4Low/contracts/types"
)

// Service interface
type Service interface {
	CreateTrip(ctx context.Context) error
	GetRoute(ctx context.Context, pickup, dropoff types.Coordinate) (*OsrmApiResponse, error)
	EstimatePackagesPriceWithRoute(route *OsrmApiResponse) []*RideFare
	GenerateTripFares(ctx context.Context, fares []*RideFare, userID string, route *OsrmApiResponse) ([]*RideFare, error)
}

// Repository interface
type Repository interface {
	SaveTrip(ctx context.Context) error
	SaveRideFare(ctx context.Context, rideFare *RideFare) error
}
