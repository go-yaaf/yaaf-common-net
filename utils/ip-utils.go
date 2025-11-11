package utils

import (
	"context"
	"fmt"
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
	return fmt.Sprintf("POINT(%f %f)", res.Longitude, res.Latitude), nil
}

// AddressLookup invoke Geo IP and return address as formatted string
func (t *IPUtilsStruct) AddressLookup(ip string) (string, error) {
	config, err := ip2locationio.OpenConfiguration(t.apiKey)
	if err != nil {
		return "", err
	}
	ipl, err := ip2locationio.OpenIPGeolocation(config)
	if err != nil {
		return "", err
	}

	res, err := ipl.LookUp(ip, "")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("[%s], %s, %s, %s %s)", res.AS, res.CityName, res.RegionName, res.CountryName, res.ZipCode), nil
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
