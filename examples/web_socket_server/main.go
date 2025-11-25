package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-yaaf/yaaf-common-net/web"
	"github.com/go-yaaf/yaaf-common/logger"

	"github.com/go-yaaf/yaaf-common-net/examples/web_socket_server/airplane_config"
	. "github.com/go-yaaf/yaaf-common-net/examples/web_socket_server/model"
	"github.com/go-yaaf/yaaf-common-net/examples/web_socket_server/www"
)

var secret = "put your secret string. It must be at least 32 characters long"
var signing = "put your signing key string. It must be at least 32 characters long"

func init() {
	logLevel := "INFO"
	logger.SetLevel(logLevel)
	logger.EnableJsonFormat(true)
	logger.Init()
}

// Main entry point
func main() {

	// Create the new instance of the WEB server
	webServer := web.NewWebServer()

	// Set API Version (usually it will be provided by the configuration)
	webServer.WithAPIVersion("1.0.1")

	// Set encryption secrets
	webServer.WithSecrets(secret, signing)

	// Enforce checking the name from the API Key
	webServer.WithAppName("websocket-server-example")

	// Add REST endpoints
	webServer.AddWebSocketEndpoints(airplane_config.NewAirplanesSocketEndPoint())

	// Add HTML Templates endpoint
	webServer.WithHtmlTemplates("examples/web_socket_server/templates/**")
	webServer.AddRESTEndpoints(www.NewListOfHtmlEndPoints()...)

	port := 8080
	logger.Info("Starting REST server, listening on port: %d", port)

	// Start Web server
	go func() {
		if err := webServer.Start(port); err != nil {
			logger.Warn("error starting REST server: %s", err.Error())
		} else {
			logger.Info("Closing the REST server...")
		}
	}()

	registry := webServer.WebSocketRegistry("airplanes")
	go startPoller(registry)

	<-make(chan struct{})
}

// start polling open-sky status every 5 seconds
func startPoller(reg web.IWSClientRegistry) {
	ticker := time.NewTicker(10 * time.Second) // every 5s for demo
	defer ticker.Stop()

	for {
		planes, err := fetchOpenSkyStates()
		if err != nil {
			log.Println("fetch error:", err)
			<-ticker.C
			continue
		}

		payload := struct {
			Type   string       `json:"type"`
			Time   time.Time    `json:"time"`
			Planes []PlaneState `json:"planes"`
		}{
			Type:   "planes",
			Time:   time.Now().UTC(),
			Planes: planes,
		}

		if data, er := json.Marshal(payload); er != nil {
			logger.Warn("marshal error: %v", er)
			<-ticker.C
			continue
		} else {
			reg.Broadcast(data)
			<-ticker.C
		}
	}
}

func fetchOpenSkyStates() ([]PlaneState, error) {
	// For anonymous access, bounding box helps reduce data size.
	// Here: around Israel as example; change for your region or remove bbox.
	url := "https://opensky-network.org/api/states/all?lamin=29&lomin=33&lamax=34.5&lomax=36"

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var osr OpenSkyResponse
	if err = json.NewDecoder(resp.Body).Decode(&osr); err != nil {
		return nil, err
	}

	planes := make([]PlaneState, 0, len(osr.States))
	for _, s := range osr.States {
		// Docs: https://opensky-network.org/apidoc/rest.html#response
		// [0] icao24, [1] callsign, [2] originCountry,
		// [5] longitude, [6] latitude, [7] baroAltitude,
		// [10] velocity, [10] heading is actually [10]? in docs it's [10] trueTrack, [9] geoAltitude etc.
		// To keep it safe we cast carefully.

		// need at least lat/lon
		if len(s) < 11 {
			continue
		}

		getString := func(idx int) string {
			if idx >= len(s) || s[idx] == nil {
				return ""
			}
			if v, ok := s[idx].(string); ok {
				return v
			}
			return ""
		}

		getFloat := func(idx int) float64 {
			if idx >= len(s) || s[idx] == nil {
				return 0
			}
			switch v := s[idx].(type) {
			case float64:
				return v
			case float32:
				return float64(v)
			default:
				return 0
			}
		}

		getBool := func(idx int) bool {
			if idx >= len(s) || s[idx] == nil {
				return false
			}
			if v, ok := s[idx].(bool); ok {
				return v
			}
			return false
		}

		lat := getFloat(6)
		lon := getFloat(5)
		if lat == 0 && lon == 0 {
			continue
		}

		p := PlaneState{
			Icao24:    getString(0),
			Callsign:  getString(1),
			Country:   getString(2),
			Longitude: lon,
			Latitude:  lat,
			Altitude:  getFloat(7),
			Heading:   getFloat(10),
			Velocity:  getFloat(9),
			OnGround:  getBool(8),
			// Raw:       s,
		}
		planes = append(planes, p)
	}

	return planes, nil
}
