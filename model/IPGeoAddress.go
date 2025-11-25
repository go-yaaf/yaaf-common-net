package model

import (
	"fmt"
	"strings"
)

// IPGeoAddress represents a WGS85 geographic location
type IPGeoAddress struct {
	CountryCode string `json:"country_code"` // Country code (2 letters)
	CountryName string `json:"country_name"` // Country name
	RegionName  string `json:"region_name"`  // Region name
	CityName    string `json:"city_name"`    // City name
	ZipCode     string `json:"zip_code"`     // Zip code
	TimeZone    string `json:"time_zone"`    // Time zone
	ASName      string `json:"as_name"`      // Autonomous System name
	ASNumber    string `json:"as_number"`    // Autonomous System number
}

// NewIPGeoAddress creates a new IPGeoAddress instance
func NewIPGeoAddress() *IPGeoAddress {
	return &IPGeoAddress{}
}

// WithCountryCode sets the country code
func (p *IPGeoAddress) WithCountryCode(value string) *IPGeoAddress {
	p.CountryCode = value
	return p
}

// WithCountryName sets the country name
func (p *IPGeoAddress) WithCountryName(value string) *IPGeoAddress {
	p.CountryName = value
	return p
}

// WithRegionName sets the region name
func (p *IPGeoAddress) WithRegionName(value string) *IPGeoAddress {
	p.RegionName = value
	return p
}

// WithCityName sets the city name
func (p *IPGeoAddress) WithCityName(value string) *IPGeoAddress {
	p.CityName = value
	return p
}

// WithZipCode sets the zip code
func (p *IPGeoAddress) WithZipCode(value string) *IPGeoAddress {
	p.ZipCode = value
	return p
}

// WithTimeZone sets the time zone
func (p *IPGeoAddress) WithTimeZone(value string) *IPGeoAddress {
	p.TimeZone = value
	return p
}

// WithASName sets the AS name
func (p *IPGeoAddress) WithASName(value string) *IPGeoAddress {
	p.ASName = value
	return p
}

// WithASNumber sets the AS number
func (p *IPGeoAddress) WithASNumber(value string) *IPGeoAddress {
	p.ASNumber = value
	return p
}

// String returns string representation of the point in the provided format
// If format is empty, a default format is used
func (p *IPGeoAddress) String(format string) string {
	if len(format) == 0 {
		return fmt.Sprintf("[%s], %s, %s, %s %s)", p.ASName, p.CityName, p.RegionName, p.CountryName, p.ZipCode)
	} else {
		format = strings.ReplaceAll(format, "{country_code}", p.CountryCode)
		format = strings.ReplaceAll(format, "{country_name}", p.CountryName)
		format = strings.ReplaceAll(format, "{region_name}", p.RegionName)
		format = strings.ReplaceAll(format, "{city_name}", p.CityName)
		format = strings.ReplaceAll(format, "{zip_code}", p.ZipCode)
		format = strings.ReplaceAll(format, "{time_zone}", p.ZipCode)
		format = strings.ReplaceAll(format, "{as_name}", p.ASName)
		format = strings.ReplaceAll(format, "{as_number}", p.ASNumber)
		return format
	}
}
