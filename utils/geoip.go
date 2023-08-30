package utils

import (
	"math"
	"net"

	"github.com/golang/geo/s2"
	"github.com/oschwald/geoip2-golang"
)

type IPCoordinate interface {
	GetLatLng(ip string) (float64, float64, error)
}

type ipCoordinate struct {
	*geoip2.Reader
}

func NewIPCoordinate(geoDB string) (IPCoordinate, error) {
	db, err := geoip2.Open(geoDB)
	if err != nil {
		return nil, err
	}
	return &ipCoordinate{db}, nil
}

func (coordinate *ipCoordinate) GetLatLng(ip string) (float64, float64, error) {
	city, err := coordinate.City(net.ParseIP(ip))
	if err != nil {
		return 0, 0, err
	}

	return city.Location.Latitude, city.Location.Longitude, nil
}

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	p1 := s2.PointFromLatLng(s2.LatLngFromDegrees(lat1, lon1))
	p2 := s2.PointFromLatLng(s2.LatLngFromDegrees(lat2, lon2))

	distance := s2.ChordAngleBetweenPoints(p1, p2).Angle().Radians()

	distanceKm := distance * 6371.0

	return distanceKm
}

func calculateTwoIPDistance(ip1, ip2 string, coordinate IPCoordinate) (float64, error) {
	lat1, lon1, err := coordinate.GetLatLng(ip1)
	if err != nil {
		return 0, err
	}

	lat2, lon2, err := coordinate.GetLatLng(ip2)
	if err != nil {
		return 0, err
	}

	distance := calculateDistance(lat1, lon1, lat2, lon2)
	return distance, nil
}

func GetUserNearestIP(userIP string, ipList []string, coordinate IPCoordinate) string {
	ipDistanceMap := make(map[string]float64)
	for _, ip := range ipList {
		distance, err := calculateTwoIPDistance(userIP, ip, coordinate)
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
