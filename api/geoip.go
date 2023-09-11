package api

import (
	"context"
	"math"
	"strconv"

	"github.com/gnasnik/titan-explorer/core/generated/model"
	"github.com/gnasnik/titan-explorer/core/statistics"
	"github.com/golang/geo/s2"
)

type IPCoordinate interface {
	GetLatLng(ctx context.Context, ip string) (float64, float64, error)
}

type ipCoordinate struct {
	// *geoip2.Reader
}

func NewIPCoordinate() IPCoordinate {
	return &ipCoordinate{}
}

func (coordinate *ipCoordinate) GetLatLng(ctx context.Context, ip string) (float64, float64, error) {
	var loc model.Location
	err := statistics.GetIpLocation(ctx, ip, &loc, "en")
	if err != nil {
		return 0, 0, err
	}

	longitude, err := strconv.ParseFloat(loc.Longitude, 64)
	if err != nil {
		return 0, 0, err
	}

	latitude, err := strconv.ParseFloat(loc.Latitude, 64)
	if err != nil {
		return 0, 0, err
	}

	return latitude, longitude, nil
}

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	p1 := s2.PointFromLatLng(s2.LatLngFromDegrees(lat1, lon1))
	p2 := s2.PointFromLatLng(s2.LatLngFromDegrees(lat2, lon2))

	distance := s2.ChordAngleBetweenPoints(p1, p2).Angle().Radians()

	distanceKm := distance * 6371.0

	return distanceKm
}

func calculateTwoIPDistance(ctx context.Context, ip1, ip2 string, coordinate IPCoordinate) (float64, error) {
	lat1, lon1, err := coordinate.GetLatLng(ctx, ip1)
	if err != nil {
		return 0, err
	}

	lat2, lon2, err := coordinate.GetLatLng(ctx, ip2)
	if err != nil {
		return 0, err
	}

	distance := calculateDistance(lat1, lon1, lat2, lon2)
	return distance, nil
}

func GetUserNearestIP(ctx context.Context, userIP string, ipList []string, coordinate IPCoordinate) string {
	ipDistanceMap := make(map[string]float64)
	for _, ip := range ipList {
		distance, err := calculateTwoIPDistance(ctx, userIP, ip, coordinate)
		if err != nil {
			log.Errorf("calculate tow ip distance error %s", err.Error())
			continue
		}
		ipDistanceMap[ip] = distance
	}

	minDistance := math.MaxFloat64
	var nearestIP string
	for ip, distance := range ipDistanceMap {
		if distance < minDistance {
			minDistance = distance
			nearestIP = ip
		}
	}

	return nearestIP
}
