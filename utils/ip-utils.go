package utils

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/go-yaaf/yaaf-common-net/model"
	"github.com/go-yaaf/yaaf-common/utils/collections"

	"github.com/ip2location/ip2location-io-go/ip2locationio"
)

var wellKnownDNS []string

// IPUtilsStruct is a structure for IP utilities
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

// GetKnownDnsIPs This method return list of well known DNS IPs
func (t *IPUtilsStruct) GetKnownDnsIPs() []string {
	if len(wellKnownDNS) == 0 {
		wellKnownDNS = make([]string, 0)
		wellKnownDNS = append(wellKnownDNS, "8.8.8.8", "8.8.4.4")               // Google Public DNS
		wellKnownDNS = append(wellKnownDNS, "1.1.1.1", "1.0.0.1")               // Cloudflare DNS
		wellKnownDNS = append(wellKnownDNS, "9.9.9.9", "149.112.112.112")       // Quad9 DNS
		wellKnownDNS = append(wellKnownDNS, "208.67.222.222", "208.67.220.220") // OpenDNS
		wellKnownDNS = append(wellKnownDNS, "94.140.14.14", "94.140.15.15")     // AdGuard DNS
		wellKnownDNS = append(wellKnownDNS, "77.88.8.8", "77.88.8.1")           // Yandex DNS
		wellKnownDNS = append(wellKnownDNS, "76.76.19.19", "76.223.122.150")    // Alternate DNS
		wellKnownDNS = append(wellKnownDNS, "185.228.168.9", "185.228.168.9")   // CleanBrowsing DNS
	}
	return wellKnownDNS
}

// IsKnownDnsIP check if the provided IP is in the list of well-known public DNS
func (t *IPUtilsStruct) IsKnownDnsIP(ip string) bool {
	return collections.Include(t.GetKnownDnsIPs(), ip)
}
