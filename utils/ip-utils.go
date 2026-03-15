package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-yaaf/yaaf-common-net/model"
	"github.com/go-yaaf/yaaf-common/utils/collections"

	"github.com/ip2location/ip2location-io-go/ip2locationio"
)

var wellKnownDNS []string

const (
	bulkLookupURL       = "https://bulk.ip2location.io/"
	bulkLookupBatchSize = 1000
	bulkLookupMaxWorker = 4
	bulkLookupFields    = "country_code,country_name,region_name,city_name,latitude,longitude,zip_code,time_zone,asn,as"
)

var defaultIPUtilsHTTPClient = func() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 16
	transport.MaxIdleConnsPerHost = 8

	return &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
	}
}()

// IPUtilsStruct is a structure for IP utilities
type IPUtilsStruct struct {
	apiKey      string
	httpClient  *http.Client
	bulkBaseURL string
}

// IPUtils is a factory method that acts as a static member
func IPUtils(apiKey string) *IPUtilsStruct {
	if len(apiKey) == 0 {
		apiKey = "A804D17F1EE16FBE269FE00610B95C97"
	}
	return &IPUtilsStruct{
		apiKey:      apiKey,
		httpClient:  defaultIPUtilsHTTPClient,
		bulkBaseURL: bulkLookupURL,
	}
}

type bulkLookupResult struct {
	CountryCode string  `json:"country_code"`
	CountryName string  `json:"country_name"`
	RegionName  string  `json:"region_name"`
	CityName    string  `json:"city_name"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ZipCode     string  `json:"zip_code"`
	TimeZone    string  `json:"time_zone"`
	ASNumber    string  `json:"asn"`
	ASName      string  `json:"as"`
}

type bulkLookupErrorResponse struct {
	Error struct {
		ErrorCode    int    `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	} `json:"error"`
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
		WithCityName(res.CityName).
		WithLatitude(res.Latitude).
		WithLongitude(res.Longitude).
		WithASName(res.AS).
		WithASNumber(res.Asn).
		WithZipCode(res.ZipCode).
		WithTimeZone(res.TimeZone)
	return ipga, nil
}

// LookupGeoAndAddress resolves a list of IP addresses through the IP2Location bulk API.
// It validates the input, removes private IPs from the lookup set, de-duplicates repeated public IPs,
// batches requests in groups of 1000, performs the remote lookups with bounded concurrency, and preserves
// the relative order of the remaining public IPs in the returned slice.
// The method returns an error if any IP is invalid, if the remote API fails, or if the API omits
// data for one of the requested public IPs.
func (t *IPUtilsStruct) LookupGeoAndAddress(ips []string) ([]model.IPGeoAddress, error) {
	if len(ips) == 0 {
		return []model.IPGeoAddress{}, nil
	}
	if strings.TrimSpace(t.apiKey) == "" {
		return nil, fmt.Errorf("bulk lookup requires a non-empty API key")
	}

	uniqueIPs, positions, publicCount, err := normalizeLookupIPs(ips)
	if err != nil {
		return nil, err
	}
	if publicCount == 0 {
		return []model.IPGeoAddress{}, nil
	}

	batches := splitIPBatches(uniqueIPs, bulkLookupBatchSize)
	results := make([]model.IPGeoAddress, publicCount)
	workerCount := minInt(len(batches), bulkLookupMaxWorker)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errOnce sync.Once
	errCh := make(chan error, 1)
	sem := make(chan struct{}, workerCount)

	reportError := func(err error) {
		errOnce.Do(func() {
			errCh <- err
			cancel()
		})
	}

	for _, batch := range batches {
		batch := batch
		wg.Add(1)

		go func() {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			lookups, lookupErr := t.lookupGeoAndAddressBatch(ctx, batch)
			if lookupErr != nil {
				reportError(lookupErr)
				return
			}

			mu.Lock()
			defer mu.Unlock()

			for _, ip := range batch {
				addr, ok := lookups[ip]
				if !ok {
					reportError(fmt.Errorf("bulk lookup returned no result for IP %s", ip))
					return
				}

				for _, idx := range positions[ip] {
					results[idx] = addr
				}
			}
		}()
	}

	wg.Wait()

	select {
	case err = <-errCh:
		return nil, err
	default:
		return results, nil
	}
}

func (t *IPUtilsStruct) lookupGeoAndAddressBatch(ctx context.Context, ips []string) (map[string]model.IPGeoAddress, error) {
	body, err := json.Marshal(ips)
	if err != nil {
		return nil, fmt.Errorf("marshal bulk lookup request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.resolveBulkLookupURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build bulk lookup request: %w", err)
	}

	query := req.URL.Query()
	query.Set("format", "json")
	query.Set("fields", bulkLookupFields)
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiKey))

	resp, err := t.resolveHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("bulk lookup request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read bulk lookup response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, parseBulkLookupError(resp.StatusCode, payload)
	}

	raw := make(map[string]bulkLookupResult, len(ips))
	if err = json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("decode bulk lookup response: %w", err)
	}

	result := make(map[string]model.IPGeoAddress, len(raw))
	for ip, item := range raw {
		result[ip] = model.IPGeoAddress{
			CountryCode: item.CountryCode,
			CountryName: item.CountryName,
			RegionName:  item.RegionName,
			CityName:    item.CityName,
			Latitude:    item.Latitude,
			Longitude:   item.Longitude,
			ZipCode:     item.ZipCode,
			TimeZone:    item.TimeZone,
			ASName:      item.ASName,
			ASNumber:    item.ASNumber,
		}
	}
	return result, nil
}

func normalizeLookupIPs(ips []string) ([]string, map[string][]int, int, error) {
	uniqueIPs := make([]string, 0, len(ips))
	positions := make(map[string][]int, len(ips))
	publicCount := 0

	for idx, rawIP := range ips {
		trimmed := strings.TrimSpace(rawIP)
		if trimmed == "" {
			return nil, nil, 0, fmt.Errorf("IP at index %d is empty", idx)
		}

		parsed := net.ParseIP(trimmed)
		if parsed == nil {
			return nil, nil, 0, fmt.Errorf("invalid IP at index %d: %q", idx, rawIP)
		}
		if parsed.IsPrivate() {
			continue
		}

		canonical := parsed.String()
		if _, exists := positions[canonical]; !exists {
			uniqueIPs = append(uniqueIPs, canonical)
		}
		positions[canonical] = append(positions[canonical], publicCount)
		publicCount++
	}

	return uniqueIPs, positions, publicCount, nil
}

func splitIPBatches(ips []string, size int) [][]string {
	if len(ips) == 0 {
		return nil
	}

	batches := make([][]string, 0, (len(ips)+size-1)/size)
	for start := 0; start < len(ips); start += size {
		end := start + size
		if end > len(ips) {
			end = len(ips)
		}
		batches = append(batches, ips[start:end])
	}
	return batches
}

func parseBulkLookupError(statusCode int, payload []byte) error {
	var apiErr bulkLookupErrorResponse
	if err := json.Unmarshal(payload, &apiErr); err == nil && apiErr.Error.ErrorMessage != "" {
		return fmt.Errorf("bulk lookup failed with status %d (code %d): %s", statusCode, apiErr.Error.ErrorCode, apiErr.Error.ErrorMessage)
	}

	message := strings.TrimSpace(string(payload))
	if message == "" {
		message = http.StatusText(statusCode)
	}
	return fmt.Errorf("bulk lookup failed with status %d: %s", statusCode, message)
}

func (t *IPUtilsStruct) resolveBulkLookupURL() string {
	if strings.TrimSpace(t.bulkBaseURL) != "" {
		return t.bulkBaseURL
	}
	return bulkLookupURL
}

func (t *IPUtilsStruct) resolveHTTPClient() *http.Client {
	if t.httpClient != nil {
		return t.httpClient
	}
	return defaultIPUtilsHTTPClient
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
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
