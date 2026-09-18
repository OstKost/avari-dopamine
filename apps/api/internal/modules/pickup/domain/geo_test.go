package domain_test

import (
	"math"
	"testing"

	"github.com/ostkost/dopamine-market/api/internal/modules/pickup/domain"
)

func TestGeo_DestinationPointAndDistance(t *testing.T) {
	origin := domain.LatLng{Latitude: 47.2357, Longitude: 39.7015} // Rostov-on-Don

	bearings := []float64{0, 45, 90, 135, 180, 225, 270, 315}
	distances := []float64{100, 250, 500, 1000, 5000}

	for _, bearing := range bearings {
		for _, dist := range distances {
			dest := domain.DestinationPoint(origin, bearing, dist)

			measuredDist := domain.DistanceBetween(origin, dest)

			// Допуск погрешности 0.1% из-за округления сферической модели
			diff := math.Abs(measuredDist - dist)
			if diff > dist*0.001 {
				t.Errorf("bearing %.1f, target dist %.1f, measured %.2f (diff %.4f)",
					bearing, dist, measuredDist, diff)
			}
		}
	}
}

func TestGeo_EquatorTest(t *testing.T) {
	origin := domain.LatLng{Latitude: 0, Longitude: 0}
	dist := 10000.0 // 10 km north
	dest := domain.DestinationPoint(origin, 0, dist)

	if dest.Longitude != 0 {
		t.Errorf("expected longitude 0, got %f", dest.Longitude)
	}
	if dest.Latitude <= 0 {
		t.Errorf("expected positive latitude, got %f", dest.Latitude)
	}

	measured := domain.DistanceBetween(origin, dest)
	if math.Abs(measured-dist) > 1.0 {
		t.Errorf("expected dist %f, got %f", dist, measured)
	}
}
