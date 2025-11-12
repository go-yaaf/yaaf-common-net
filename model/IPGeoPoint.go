package model

import "fmt"

// IPGeoPoint represents a WGS85 geographic location of an IP address
type IPGeoPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func NewIPGeoPoint(lon, lat float64) *IPGeoPoint {
	return &IPGeoPoint{
		Latitude:  lat,
		Longitude: lon,
	}
}

// WKT return string representation of the point in WKT format
func (p *IPGeoPoint) WKT() string {
	return fmt.Sprintf("POINT(%f %f)", p.Longitude, p.Latitude)
}
