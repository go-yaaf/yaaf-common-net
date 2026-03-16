package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestLookupGeoAndAddressPreservesOrderAndRemovesDuplicates(t *testing.T) {
	var requests int32

	client := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&requests, 1)
			require.Equal(t, http.MethodPost, r.Method)
			require.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
			require.Equal(t, "json", r.URL.Query().Get("format"))
			require.Equal(t, bulkLookupFields, r.URL.Query().Get("fields"))

			var ips []string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&ips))
			require.ElementsMatch(t, []string{"1.1.1.1", "8.8.8.8"}, ips)

			payload := map[string]bulkLookupResult{
				"1.1.1.1": {
					CountryCode: "AU",
					CountryName: "Australia",
					RegionName:  "Queensland",
					CityName:    "South Brisbane",
					Latitude:    -27.4748,
					Longitude:   153.017,
					ZipCode:     "4101",
					TimeZone:    "Australia/Brisbane",
					ASNumber:    "13335",
					ASName:      "Cloudflare, Inc.",
				},
				"8.8.8.8": {
					CountryCode: "US",
					CountryName: "United States of America",
					RegionName:  "California",
					CityName:    "Mountain View",
					Latitude:    37.4056,
					Longitude:   -122.0775,
					ZipCode:     "94043",
					TimeZone:    "America/Los_Angeles",
					ASNumber:    "15169",
					ASName:      "Google LLC",
				},
			}

			buffer, err := json.Marshal(payload)
			require.NoError(t, err)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(buffer)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	ipUtils := &IPUtilsStruct{
		apiKey:      "test-key",
		httpClient:  client,
		bulkBaseURL: "https://bulk.ip2location.io/",
	}

	result, err := ipUtils.LookupGeoAndAddress([]string{"1.1.1.1", "8.8.8.8", "1.1.1.1"})
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, int32(1), atomic.LoadInt32(&requests))

	require.Equal(t, "AU", result[0].CountryName)
	require.Equal(t, "US", result[1].CountryName)
	require.Equal(t, "1.1.1.1", result[0].IP)
	require.Equal(t, "8.8.8.8", result[1].IP)
	require.Equal(t, -27.4748, result[0].Latitude)
	require.Equal(t, 153.017, result[0].Longitude)
}

func TestLookupGeoAndAddressSplitsLargeRequests(t *testing.T) {
	var requests int32

	client := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&requests, 1)

			var ips []string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&ips))
			require.LessOrEqual(t, len(ips), bulkLookupBatchSize)

			payload := make(map[string]bulkLookupResult, len(ips))
			for _, ip := range ips {
				payload[ip] = bulkLookupResult{
					CountryCode: "US",
					CountryName: "United States",
					Latitude:    1,
					Longitude:   2,
				}
			}

			buffer, err := json.Marshal(payload)
			require.NoError(t, err)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(buffer)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	ipUtils := &IPUtilsStruct{
		apiKey:      "test-key",
		httpClient:  client,
		bulkBaseURL: "https://bulk.ip2location.io/",
	}

	ips := make([]string, 0, bulkLookupBatchSize+1)
	for i := 0; i <= bulkLookupBatchSize; i++ {
		ips = append(ips, fmt.Sprintf("11.%d.%d.%d", (i/65536)%256, (i/256)%256, i%256))
	}
	ips = append(ips, "12.16.0.1", "13.16.0.2")

	result, err := ipUtils.LookupGeoAndAddress(ips)
	require.NoError(t, err)
	require.Len(t, result, len(ips))
	require.GreaterOrEqual(t, atomic.LoadInt32(&requests), int32(2))
}

func TestLookupGeoAndAddressSkipsPrivateIPs(t *testing.T) {
	var requests int32

	client := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&requests, 1)

			var ips []string
			require.NoError(t, json.NewDecoder(r.Body).Decode(&ips))
			require.Equal(t, []string{"1.1.1.1", "8.8.8.8"}, ips)

			payload := map[string]bulkLookupResult{
				"1.1.1.1": {
					CountryCode: "AU",
					CountryName: "Australia",
				},
				"8.8.8.8": {
					CountryCode: "US",
					CountryName: "United States of America",
				},
			}

			buffer, err := json.Marshal(payload)
			require.NoError(t, err)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(buffer)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	ipUtils := &IPUtilsStruct{
		apiKey:      "test-key",
		httpClient:  client,
		bulkBaseURL: "https://bulk.ip2location.io/",
	}

	result, err := ipUtils.LookupGeoAndAddress([]string{"1.1.1.1", "192.168.1.10", "8.8.8.8", "10.0.0.5"})
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, int32(1), atomic.LoadInt32(&requests))
	require.Equal(t, "1.1.1.1", result[0].IP)
	require.Equal(t, "8.8.8.8", result[1].IP)
	require.Equal(t, "AU", result[0].CountryName)
	require.Equal(t, "US", result[1].CountryName)
}

func TestLookupGeoAndAddressReturnsEmptyForOnlyPrivateIPs(t *testing.T) {
	var requests int32

	client := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&requests, 1)
			return nil, fmt.Errorf("request should not be executed")
		}),
	}

	ipUtils := &IPUtilsStruct{
		apiKey:      "test-key",
		httpClient:  client,
		bulkBaseURL: "https://bulk.ip2location.io/",
	}

	result, err := ipUtils.LookupGeoAndAddress([]string{"192.168.1.10", "10.0.0.5", "172.16.0.1"})
	require.NoError(t, err)
	require.Empty(t, result)
	require.Equal(t, int32(0), atomic.LoadInt32(&requests))
}

func TestLookupGeoAndAddressRejectsInvalidInput(t *testing.T) {
	ipUtils := &IPUtilsStruct{apiKey: "test-key"}

	result, err := ipUtils.LookupGeoAndAddress([]string{"1.1.1.1", "not-an-ip"})
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "invalid IP")
}

func TestLookupGeoAndAddressReturnsAPIError(t *testing.T) {
	client := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":{"error_code":401,"error_message":"Invalid API key."}}`))),
				Header:     make(http.Header),
			}, nil
		}),
	}

	ipUtils := &IPUtilsStruct{
		apiKey:      "test-key",
		httpClient:  client,
		bulkBaseURL: "https://bulk.ip2location.io/",
	}

	result, err := ipUtils.LookupGeoAndAddress([]string{"1.1.1.1"})
	require.Error(t, err)
	require.Nil(t, result)
	require.Contains(t, err.Error(), "Invalid API key.")
}

func TestLookupGeoAndAddressFallsBackToSingleLookupOnBulkUnauthorized(t *testing.T) {
	var bulkRequests int32
	var singleRequests int32

	client := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			switch r.Method {
			case http.MethodPost:
				atomic.AddInt32(&bulkRequests, 1)

				var ips []string
				require.NoError(t, json.NewDecoder(r.Body).Decode(&ips))
				require.Equal(t, []string{"1.1.1.1", "8.8.8.8"}, ips)

				return &http.Response{
					StatusCode: http.StatusUnauthorized,
					Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":{"error_code":401,"error_message":"Invalid API key or insufficient query."}}`))),
					Header:     make(http.Header),
				}, nil
			case http.MethodGet:
				atomic.AddInt32(&singleRequests, 1)

				ip := r.URL.Query().Get("ip")
				payload := bulkLookupResult{}
				switch ip {
				case "1.1.1.1":
					payload = bulkLookupResult{
						CountryCode: "AU",
						CountryName: "Australia",
					}
				case "8.8.8.8":
					payload = bulkLookupResult{
						CountryCode: "US",
						CountryName: "United States of America",
					}
				default:
					t.Fatalf("unexpected fallback lookup for IP %s", ip)
				}

				buffer, err := json.Marshal(payload)
				require.NoError(t, err)

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(buffer)),
					Header:     make(http.Header),
				}, nil
			default:
				t.Fatalf("unexpected method %s", r.Method)
				return nil, nil
			}
		}),
	}

	ipUtils := &IPUtilsStruct{
		apiKey:        "test-key",
		httpClient:    client,
		bulkBaseURL:   "https://bulk.ip2location.io/",
		singleBaseURL: "https://api.ip2location.io/",
	}

	result, err := ipUtils.LookupGeoAndAddress([]string{"1.1.1.1", "192.168.1.10", "8.8.8.8", "1.1.1.1"})
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, int32(1), atomic.LoadInt32(&bulkRequests))
	require.Equal(t, int32(2), atomic.LoadInt32(&singleRequests))
	require.Equal(t, "1.1.1.1", result[0].IP)
	require.Equal(t, "8.8.8.8", result[1].IP)
	require.Equal(t, "AU", result[0].CountryName)
	require.Equal(t, "US", result[1].CountryName)
}

func TestLookupGeoAndAddressDoesNotFallbackOnBulkServerError(t *testing.T) {
	var singleRequests int32

	client := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method == http.MethodGet {
				atomic.AddInt32(&singleRequests, 1)
			}

			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"error":{"error_code":500,"error_message":"Internal error"}}`))),
				Header:     make(http.Header),
			}, nil
		}),
	}

	ipUtils := &IPUtilsStruct{
		apiKey:        "test-key",
		httpClient:    client,
		bulkBaseURL:   "https://bulk.ip2location.io/",
		singleBaseURL: "https://api.ip2location.io/",
	}

	result, err := ipUtils.LookupGeoAndAddress([]string{"1.1.1.1"})
	require.Error(t, err)
	require.Nil(t, result)
	require.Equal(t, int32(0), atomic.LoadInt32(&singleRequests))
	require.Contains(t, err.Error(), "bulk lookup failed")
}
