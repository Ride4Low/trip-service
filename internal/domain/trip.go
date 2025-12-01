package domain

import (
	"time"

	"github.com/ride4Low/contracts/proto/trip"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Trip struct {
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	UserID   string             `bson:"userID"`
	Status   string             `bson:"status"`
	RideFare *RideFare          `bson:"rideFare"`
	Driver   *trip.TripDriver   `bson:"driver"`
}

type RideFare struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	UserID            string             `bson:"userID"`
	PackageSlug       string             `bson:"packageSlug"` // ex: van, luxury, sedan
	TotalPriceInCents float64            `bson:"totalPriceInCents"`
	Route             *OsrmApiResponse   `bson:"route"`
	CreatedAt         time.Time          `bson:"created_at"`
}

func (r *RideFare) ToProto() *trip.RideFare {
	return &trip.RideFare{
		Id:                r.ID.Hex(),
		UserID:            r.UserID,
		PackageSlug:       r.PackageSlug,
		TotalPriceInCents: r.TotalPriceInCents,
	}
}

type OsrmApiResponse struct {
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"routes"`
}

func (o *OsrmApiResponse) ToProto() *trip.Route {
	route := o.Routes[0]
	geometry := route.Geometry.Coordinates
	coordinates := make([]*trip.Coordinate, len(geometry))
	for i, coord := range geometry {
		coordinates[i] = &trip.Coordinate{
			Latitude:  coord[0],
			Longitude: coord[1],
		}
	}

	return &trip.Route{
		Geometry: []*trip.Geometry{
			{
				Coordinates: coordinates,
			},
		},
		Distance: route.Distance,
		Duration: route.Duration,
	}
}

type PricingConfig struct {
	PricePerUnitOfDistance float64
	PricingPerMinute       float64
}

func DefaultPricingConfig() *PricingConfig {
	return &PricingConfig{
		PricePerUnitOfDistance: 1.5,
		PricingPerMinute:       0.25,
	}
}

func ToRideFaresProto(fares []*RideFare) []*trip.RideFare {
	var protoFares []*trip.RideFare
	for _, f := range fares {
		protoFares = append(protoFares, f.ToProto())
	}
	return protoFares
}
