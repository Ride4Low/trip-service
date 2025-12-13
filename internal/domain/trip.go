package domain

import (
	"github.com/ride4Low/contracts/proto/trip"
	"github.com/ride4Low/contracts/types"
)

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

func ToRideFaresProto(fares []*types.RideFare) []*trip.RideFare {
	var protoFares []*trip.RideFare
	for _, f := range fares {
		protoFares = append(protoFares, f.ToProto())
	}
	return protoFares
}
