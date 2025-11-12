package model

import (
	"fmt"
	"strings"
)

// IPGeoAddress represents a WGS85 geographic location
type IPGeoAddress struct {
	CountryCode string `json:"country_code"`
	CountryName string `json:"country_name"`
	RegionName  string `json:"region_name"`
	CityName    string `json:"city_name"`
	ZipCode     string `json:"zip_code"`
	TimeZone    string `json:"time_zone"`
	ASName      string `json:"as_name"`
	ASNumber    string `json:"as_number"`
}

func NewIPGeoAddress() *IPGeoAddress {
	return &IPGeoAddress{}
}

func (p *IPGeoAddress) WithCountryCode(value string) *IPGeoAddress {
	p.CountryCode = value
	return p
}

func (p *IPGeoAddress) WithCountryName(value string) *IPGeoAddress {
	p.CountryName = value
	return p
}

func (p *IPGeoAddress) WithRegionName(value string) *IPGeoAddress {
	p.RegionName = value
	return p
}

func (p *IPGeoAddress) WithCityName(value string) *IPGeoAddress {
	p.CityName = value
	return p
}

func (p *IPGeoAddress) WithZipCode(value string) *IPGeoAddress {
	p.ZipCode = value
	return p
}

func (p *IPGeoAddress) WithTimeZone(value string) *IPGeoAddress {
	p.TimeZone = value
	return p
}

func (p *IPGeoAddress) WithASName(value string) *IPGeoAddress {
	p.ASName = value
	return p
}

func (p *IPGeoAddress) WithASNumber(value string) *IPGeoAddress {
	p.ASNumber = value
	return p
}

// WKT return string representation of the point in WKT format
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
