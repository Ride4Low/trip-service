package service

import (
	"log"

	"github.com/ride4Low/contracts/types"
	"github.com/ride4Low/trip-service/internal/domain"
)

func (s *service) EstimatePackagesPriceWithRoute(route *types.OsrmApiResponse) []*types.RideFare {
	baseFares := getBaseFares()
	estimatedFares := make([]*types.RideFare, len(baseFares))

	for i, f := range baseFares {
		estimatedFares[i] = estimateFareRoute(f, route)
	}

	return estimatedFares
}

func getBaseFares() []*types.RideFare {
	return []*types.RideFare{
		{
			PackageSlug:       "suv",
			TotalPriceInCents: 200,
		},
		{
			PackageSlug:       "sedan",
			TotalPriceInCents: 350,
		},
		{
			PackageSlug:       "van",
			TotalPriceInCents: 400,
		},
		{
			PackageSlug:       "luxury",
			TotalPriceInCents: 1000,
		},
	}
}

func estimateFareRoute(f *types.RideFare, route *types.OsrmApiResponse) *types.RideFare {
	pricingCfg := domain.DefaultPricingConfig()
	carPackagePrice := f.TotalPriceInCents

	distanceM := route.Routes[0].Distance
	durationInSeconds := route.Routes[0].Duration

	log.Printf("Distance in meters: %f, Duration in Seconds: %f\n", distanceM, durationInSeconds)

	distanceFare := distanceM * pricingCfg.PricePerUnitOfDistance
	timeFare := durationInSeconds * pricingCfg.PricingPerMinute
	log.Printf("CarPackagePrice: %f, DistanceFare: %f, TimeFare: %f\n", carPackagePrice, distanceFare, timeFare)
	totalPrice := carPackagePrice + distanceFare + timeFare
	log.Printf("Total: %f\n", totalPrice)

	return &types.RideFare{
		TotalPriceInCents: totalPrice,
		PackageSlug:       f.PackageSlug,
	}
}
