package utils

import (
	"context"
	"github.com/go-yaaf/yaaf-common-net/model"
	"net"
	"strings"
	"time"

	"github.com/ip2location/ip2location-io-go/ip2locationio"
)

type IPUtilsStruct struct {
	apiKey string
}

// IPUtils is a factory method that acts as a static member
func IPUtils(apiKey string) *IPUtilsStruct {
	if len(apiKey) > 0 {
		apiKey = "A804D17F1EE16FBE269FE00610B95C97"
	}
	return &IPUtilsStruct{
		apiKey: apiKey,
	}
}

// GeoLookupWKT invoke Geo IP and return location as WTK string
func (t *IPUtilsStruct) GeoLookupWKT(ip string) (string, error) {
	config, err := ip2locationio.OpenConfiguration(t.apiKey)
	if err != nil {
		return "", err
	}
	ipl, err := ip2locationio.OpenIPGeolocation(config)
	if err != nil {
		return "", err
	}

	res, err := ipl.LookUp(ip, "") // language parameter only available with Plus and Security plans
	if err != nil {
		return "", err
	}

	return model.NewIPGeoPoint(res.Longitude, res.Latitude).WKT(), nil
}

// AddressLookup invoke Geo IP and return address as formatted string
func (t *IPUtilsStruct) AddressLookup(ip string, format string) (string, error) {

	if ipga, err := t.FullAddressLookup(ip); err != nil {
		return "", err
	} else {
		return ipga.String(format), nil
	}
}

// FullAddressLookup invoke Geo IP and return address as object
func (t *IPUtilsStruct) FullAddressLookup(ip string) (*model.IPGeoAddress, error) {
	config, err := ip2locationio.OpenConfiguration(t.apiKey)
	if err != nil {
		return nil, err
	}
	ipl, err := ip2locationio.OpenIPGeolocation(config)
	if err != nil {
		return nil, err
	}

	res, err := ipl.LookUp(ip, "")
	if err != nil {
		return nil, err
	}

	ipga := model.NewIPGeoAddress().
		WithCountryCode(res.CountryCode).
		WithCountryName(res.CountryName).
		WithRegionName(res.RegionName).
		WithCityName(res.CityName).WithASName(res.AS).
		WithASNumber(res.Asn).
		WithZipCode(res.ZipCode).
		WithTimeZone(res.TimeZone)
	return ipga, nil
}

// DnsLookup invoke DNS resolver and return comma-separated list of DNS names
func (t *IPUtilsStruct) DnsLookup(ip string) (string, error) {
	if ip == "" {
		return "", nil
	}
	r := net.Resolver{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if names, err := r.LookupAddr(ctx, ip); err != nil || len(names) == 0 {
		return "", nil
	} else {
		return strings.Join(names, ", "), nil
	}
}
