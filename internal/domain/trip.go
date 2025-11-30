package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Trip struct {
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	UserID   string             `bson:"userID"`
	Status   string             `bson:"status"`
	RideFare *RideFare          `bson:"rideFare"`
}

type RideFare struct {
	ID                primitive.ObjectID `bson:"_id,omitempty"`
	UserID            string             `bson:"userID"`
	PackageSlug       string             `bson:"packageSlug"` // ex: van, luxury, sedan
	TotalPriceInCents float64            `bson:"totalPriceInCents"`
	Route             *OsrmApiResponse   `bson:"route"`
	CreatedAt         time.Time          `bson:"created_at"`
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

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
