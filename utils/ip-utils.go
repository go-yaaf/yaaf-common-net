package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
	singleLookupURL     = "https://api.ip2location.io/"
	bulkLookupBatchSize = 1000
	bulkLookupMaxWorker = 4
	singleLookupWorkers = 8
	bulkLookupFields    = "country_code,country_name,region_name,city_name,latitude,longitude,zip_code,time_zone,asn,as"
)

var defaultIPUtilsHTTPClient = func() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 16
	transport.MaxIdleConnsPerHost = 8

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}()

// IPUtilsStruct is a structure for IP utilities
type IPUtilsStruct struct {
	apiKey        string
	httpClient    *http.Client
	bulkBaseURL   string
	singleBaseURL string
}

// IPUtils is a factory method that acts as a static member.
// The IP2Location API key must be supplied by the caller; lookups return an
// error when it is empty (no shared/hardcoded key is used).
func IPUtils(apiKey string) *IPUtilsStruct {
	return &IPUtilsStruct{
		apiKey:        strings.TrimSpace(apiKey),
		httpClient:    defaultIPUtilsHTTPClient,
		bulkBaseURL:   bulkLookupURL,
		singleBaseURL: singleLookupURL,
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

type bulkLookupRequestError struct {
	statusCode int
	errorCode  int
	message    string
}

func (e *bulkLookupRequestError) Error() string {
	if e.errorCode > 0 {
		return fmt.Sprintf("bulk lookup failed with status %d (code %d): %s", e.statusCode, e.errorCode, e.message)
	}
	return fmt.Sprintf("bulk lookup failed with status %d: %s", e.statusCode, e.message)
}

// GeoLookupWKT invoke Geo IP and return location as WTK string
func (t *IPUtilsStruct) GeoLookupWKT(ip string) (string, error) {
	if strings.TrimSpace(t.apiKey) == "" {
		return "", errors.New("IP2Location API key is not set")
	}
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
	if strings.TrimSpace(t.apiKey) == "" {
		return nil, errors.New("IP2Location API key is not set")
	}
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
		WithIP(ip).
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

// LookupGeoAndAddress resolves a list of IP addresses through the IP2Location APIs.
// It validates the input, removes private IPs, de-duplicates repeated public IPs, and attempts
// the bulk API first. When the bulk endpoint is unavailable for the current plan, it falls back
// to concurrent single-IP lookups while keeping the same method signature and returning a unique
// result set ordered by the first public occurrence of each IP in the input slice.
// The method returns an error if any IP is invalid, if both lookup strategies fail, or if the API
// omits data for one of the requested public IPs.
func (t *IPUtilsStruct) LookupGeoAndAddress(ips []string) ([]model.IPGeoAddress, error) {
	if len(ips) == 0 {
		return []model.IPGeoAddress{}, nil
	}
	if strings.TrimSpace(t.apiKey) == "" {
		return nil, errors.New("IP2Location API key is not set")
	}

	uniqueIPs, err := normalizeLookupIPs(ips)
	if err != nil {
		return nil, err
	}
	if len(uniqueIPs) == 0 {
		return []model.IPGeoAddress{}, nil
	}

	resultsByIP, err := t.lookupGeoAndAddressWithBulk(uniqueIPs)
	if err != nil {
		if !shouldFallbackToSingleLookup(err) {
			return nil, err
		}

		resultsByIP, err = t.lookupGeoAndAddressForSingleIp(uniqueIPs)
		if err != nil {
			return nil, err
		}
	}

	return buildOrderedAddresses(uniqueIPs, resultsByIP)
}

func (t *IPUtilsStruct) lookupGeoAndAddressWithBulk(uniqueIPs []string) (map[string]model.IPGeoAddress, error) {
	batches := splitIPBatches(uniqueIPs, bulkLookupBatchSize)
	workerCount := minInt(len(batches), bulkLookupMaxWorker)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errOnce sync.Once
	errCh := make(chan error, 1)
	sem := make(chan struct{}, workerCount)
	resultsByIP := make(map[string]model.IPGeoAddress, len(uniqueIPs))

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

			batchLookups, lookupErr := t.lookupGeoAndAddressBatch(ctx, batch)
			if lookupErr != nil {
				reportError(lookupErr)
				return
			}

			mu.Lock()
			defer mu.Unlock()

			for _, ip := range batch {
				addr, ok := batchLookups[ip]
				if !ok {
					reportError(fmt.Errorf("bulk lookup returned no result for IP %s", ip))
					return
				}
				resultsByIP[ip] = addr
			}
		}()
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
		return resultsByIP, nil
	}
}

func (t *IPUtilsStruct) lookupGeoAndAddressForSingleIp(uniqueIPs []string) (map[string]model.IPGeoAddress, error) {
	workerCount := minInt(len(uniqueIPs), singleLookupWorkers)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errOnce sync.Once
	errCh := make(chan error, 1)
	sem := make(chan struct{}, workerCount)
	resultsByIP := make(map[string]model.IPGeoAddress, len(uniqueIPs))

	reportError := func(err error) {
		errOnce.Do(func() {
			errCh <- err
			cancel()
		})
	}

	for _, ip := range uniqueIPs {
		ip := ip
		wg.Add(1)

		go func() {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			addr, lookupErr := t.lookupGeoAndAddressSingle(ctx, ip)
			if lookupErr != nil {
				reportError(lookupErr)
				return
			}

			mu.Lock()
			resultsByIP[ip] = addr
			mu.Unlock()
		}()
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
		return resultsByIP, nil
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
		result[ip] = mapLookupResult(ip, item)
	}
	return result, nil
}

func (t *IPUtilsStruct) lookupGeoAndAddressSingle(ctx context.Context, ip string) (model.IPGeoAddress, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.resolveSingleLookupURL(), nil)
	if err != nil {
		return model.IPGeoAddress{}, fmt.Errorf("build single lookup request: %w", err)
	}

	query := req.URL.Query()
	query.Set("ip", ip)
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiKey))

	resp, err := t.resolveHTTPClient().Do(req)
	if err != nil {
		return model.IPGeoAddress{}, fmt.Errorf("single lookup request failed for IP %s: %w", ip, err)
	}
	defer func() { _ = resp.Body.Close() }()

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.IPGeoAddress{}, fmt.Errorf("read single lookup response for IP %s: %w", ip, err)
	}
	if resp.StatusCode != http.StatusOK {
		return model.IPGeoAddress{}, parseSingleLookupError(resp.StatusCode, payload)
	}

	var raw bulkLookupResult
	if err = json.Unmarshal(payload, &raw); err != nil {
		return model.IPGeoAddress{}, fmt.Errorf("decode single lookup response for IP %s: %w", ip, err)
	}

	return mapLookupResult(ip, raw), nil
}

func normalizeLookupIPs(ips []string) ([]string, error) {
	uniqueIPs := make([]string, 0, len(ips))
	seen := make(map[string]struct{}, len(ips))

	for idx, rawIP := range ips {
		trimmed := strings.TrimSpace(rawIP)
		if trimmed == "" {
			return nil, fmt.Errorf("IP at index %d is empty", idx)
		}

		parsed := net.ParseIP(trimmed)
		if parsed == nil {
			return nil, fmt.Errorf("invalid IP at index %d: %q", idx, rawIP)
		}
		if parsed.IsPrivate() {
			continue
		}

		canonical := parsed.String()
		if _, exists := seen[canonical]; !exists {
			uniqueIPs = append(uniqueIPs, canonical)
			seen[canonical] = struct{}{}
		}
	}

	return uniqueIPs, nil
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
	errorCode, message := parseLookupErrorPayload(statusCode, payload)
	return &bulkLookupRequestError{
		statusCode: statusCode,
		errorCode:  errorCode,
		message:    message,
	}
}

func (t *IPUtilsStruct) resolveBulkLookupURL() string {
	if strings.TrimSpace(t.bulkBaseURL) != "" {
		return t.bulkBaseURL
	}
	return bulkLookupURL
}

func (t *IPUtilsStruct) resolveSingleLookupURL() string {
	if strings.TrimSpace(t.singleBaseURL) != "" {
		return t.singleBaseURL
	}
	return singleLookupURL
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

func buildOrderedAddresses(uniqueIPs []string, resultsByIP map[string]model.IPGeoAddress) ([]model.IPGeoAddress, error) {
	results := make([]model.IPGeoAddress, 0, len(uniqueIPs))
	for _, ip := range uniqueIPs {
		addr, ok := resultsByIP[ip]
		if !ok {
			return nil, fmt.Errorf("lookup returned no result for IP %s", ip)
		}
		results = append(results, addr)
	}
	return results, nil
}

func mapLookupResult(ip string, item bulkLookupResult) model.IPGeoAddress {
	return model.IPGeoAddress{
		IP:          ip,
		CountryName: item.CountryCode,
		CountryCode: item.CountryCode,
		//CountryName: item.CountryName,
		RegionName: item.RegionName,
		CityName:   item.CityName,
		Latitude:   item.Latitude,
		Longitude:  item.Longitude,
		ZipCode:    item.ZipCode,
		TimeZone:   item.TimeZone,
		ASName:     item.ASName,
		ASNumber:   item.ASNumber,
	}
}

func parseSingleLookupError(statusCode int, payload []byte) error {
	errorCode, message := parseLookupErrorPayload(statusCode, payload)
	if errorCode > 0 {
		return fmt.Errorf("single lookup failed with status %d (code %d): %s", statusCode, errorCode, message)
	}
	return fmt.Errorf("single lookup failed with status %d: %s", statusCode, message)
}

func parseLookupErrorPayload(statusCode int, payload []byte) (int, string) {
	var apiErr bulkLookupErrorResponse
	if err := json.Unmarshal(payload, &apiErr); err == nil && apiErr.Error.ErrorMessage != "" {
		return apiErr.Error.ErrorCode, apiErr.Error.ErrorMessage
	}

	message := strings.TrimSpace(string(payload))
	if message == "" {
		message = http.StatusText(statusCode)
	}
	return 0, message
}

func shouldFallbackToSingleLookup(err error) bool {
	var bulkErr *bulkLookupRequestError
	return errors.As(err, &bulkErr) && bulkErr.statusCode == http.StatusUnauthorized
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
