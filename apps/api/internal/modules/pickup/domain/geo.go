package domain

import (
	"math"
)

const EarthRadiusMeters = 6371000.0

type LatLng struct {
	Latitude  float64
	Longitude float64
}

func NewLatLng(lat, lon float64) (LatLng, error) {
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return LatLng{}, ErrInvalidCoordinates
	}
	return LatLng{
		Latitude:  lat,
		Longitude: lon,
	}, nil
}

// DestinationPoint вычисляет координаты точки назначения по начальной точке, азимуту в градусах и расстоянию в метрах.
// Использует сферическую формулу геодезии (ADR-012).
func DestinationPoint(origin LatLng, bearingDeg, distanceMeters float64) LatLng {
	delta := distanceMeters / EarthRadiusMeters
	theta := bearingDeg * math.Pi / 180.0

	phi1 := origin.Latitude * math.Pi / 180.0
	lambda1 := origin.Longitude * math.Pi / 180.0

	sinPhi1 := math.Sin(phi1)
	cosPhi1 := math.Cos(phi1)
	sinDelta := math.Sin(delta)
	cosDelta := math.Cos(delta)

	sinPhi2 := sinPhi1*cosDelta + cosPhi1*sinDelta*math.Cos(theta)
	phi2 := math.Asin(sinPhi2)

	y := math.Sin(theta) * sinDelta * cosPhi1
	x := cosDelta - sinPhi1*sinPhi2
	lambda2 := lambda1 + math.Atan2(y, x)

	// Нормализация долготы к [-180, 180]
	lonDeg := lambda2 * 180.0 / math.Pi
	lonDeg = math.Mod(lonDeg+540.0, 360.0) - 180.0

	return LatLng{
		Latitude:  phi2 * 180.0 / math.Pi,
		Longitude: lonDeg,
	}
}

// DistanceBetween вычисляет расстояние между двумя точками по формуле гаверсинусов в метрах.
func DistanceBetween(p1, p2 LatLng) float64 {
	phi1 := p1.Latitude * math.Pi / 180.0
	phi2 := p2.Latitude * math.Pi / 180.0
	deltaPhi := (p2.Latitude - p1.Latitude) * math.Pi / 180.0
	deltaLambda := (p2.Longitude - p1.Longitude) * math.Pi / 180.0

	a := math.Sin(deltaPhi/2.0)*math.Sin(deltaPhi/2.0) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(deltaLambda/2.0)*math.Sin(deltaLambda/2.0)
	c := 2.0 * math.Atan2(math.Sqrt(a), math.Sqrt(1.0-a))

	return EarthRadiusMeters * c
}
