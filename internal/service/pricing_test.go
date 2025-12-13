package service

import (
	"testing"

	"github.com/ride4Low/contracts/types"
)

func TestEstimatePackagesPriceWithRoute(t *testing.T) {
	// Setup
	svc := &service{}

	// Mock route
	// Distance: 1000 meters
	// Duration: 600 seconds (10 minutes)
	route := &types.OsrmApiResponse{
		Routes: []struct {
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
			Geometry struct {
				Coordinates [][]float64 `json:"coordinates"`
			} `json:"geometry"`
		}{
			{
				Distance: 1000.0,
				Duration: 600.0,
			},
		},
	}

	// Expected calculation:
	// PricePerUnitOfDistance = 1.5
	// PricingPerMinute = 0.25
	// DistanceFare = 1000 * 1.5 = 1500
	// TimeFare = 600 * 0.25 = 150
	// Base Additions = 1500 + 150 = 1650

	// Base Fares:
	// SUV: 200 + 1650 = 1850
	// Sedan: 350 + 1650 = 2000
	// Van: 400 + 1650 = 2050
	// Luxury: 1000 + 1650 = 2650

	expectedPrices := map[string]float64{
		"suv":    1850.0,
		"sedan":  2000.0,
		"van":    2050.0,
		"luxury": 2650.0,
	}

	// Execute
	fares := svc.EstimatePackagesPriceWithRoute(route)

	// Verify
	if len(fares) != 4 {
		t.Errorf("expected 4 fares, got %d", len(fares))
	}

	for _, fare := range fares {
		expected, ok := expectedPrices[fare.PackageSlug]
		if !ok {
			t.Errorf("unexpected package slug: %s", fare.PackageSlug)
			continue
		}

		if fare.TotalPriceInCents != expected {
			t.Errorf("expected price %f for %s, got %f", expected, fare.PackageSlug, fare.TotalPriceInCents)
		}
	}
}
