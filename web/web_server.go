// Copyright 2022. Motty Cohen
//
// REST server with Web Socket endpoint and Static files endpoints support
//
package web

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/cors"

	"github.com/go-yaaf/yaaf-common-net/rest"
	"github.com/go-yaaf/yaaf-common-net/socket"
)

type WebServer struct {
	port                string                            // Port number to listen to
	router              *mux.Router                       // The path router
	readTimeout         time.Duration                     // Socket read time-out in milliseconds
	writeTimeout        time.Duration                     // Socket write time-out in milliseconds
	enableMetrics       bool                              // Enables metrics endpoint
	restEndPoints       []rest.IRestEndpoint              // List of REST endpoints
	webSocketEndPoints  []socket.IWebSocketEndpoint       // List of Web Socket endpoints
	staticEndPoints     []rest.IStaticEndpoint            // List of static files endpoints
	restAdapterFunc     rest.RestHandlerAdaptorFunc       // External adaptor function
	middlewareFunctions []func(http.Handler) http.Handler // List of middleware functions
}

// NewWebServer is a factory method to create a new instance of Rest Server
func NewWebServer(port string) *WebServer {
	srv := &WebServer{
		port:                port,
		readTimeout:         3000,
		writeTimeout:        3000,
		enableMetrics:       false,
		router:              mux.NewRouter(),
		restEndPoints:       make([]rest.IRestEndpoint, 0),
		webSocketEndPoints:  make([]socket.IWebSocketEndpoint, 0),
		staticEndPoints:     make([]rest.IStaticEndpoint, 0),
		restAdapterFunc:     rest.DefaultRestHandlerWrapperFunc,
		middlewareFunctions: make([]func(http.Handler) http.Handler, 0),
	}
	return srv
}

// RestHandlerAdaptor injects external adaptor function to perform request pre-processing
func (s *WebServer) RestHandlerAdaptor(adaptorFunc rest.RestHandlerAdaptorFunc) *WebServer {
	if adaptorFunc != nil {
		s.restAdapterFunc = adaptorFunc
	}
	return s
}

// AddHandler adds path handler to the server
func (s *WebServer) AddHandler(path string, handler http.HandlerFunc) *WebServer {
	s.router.HandleFunc(path, handler)
	return s
}

// AddRestEntries adds list of REST entries to the server
func (s *WebServer) AddRestEntries(list ...rest.RestEntry) *WebServer {
	endPoint := rest.NewRestEndpoint(list)
	s.restEndPoints = append(s.restEndPoints, endPoint)
	return s
}

// AddWebSocketEntries adds list of Web Socket entries to the server
func (s *WebServer) AddWebSocketEntries(list ...socket.WSEndpointConfig) *WebServer {
	endPoint := socket.NewWebSocketEndpoint(list)
	s.webSocketEndPoints = append(s.webSocketEndPoints, endPoint)
	return s
}

// AddStaticEntries adds list of static file entries to the server
func (s *WebServer) AddStaticEntries(list ...rest.StaticFilesEntry) *WebServer {
	endPoint := rest.NewStaticEndpoint(list)
	s.staticEndPoints = append(s.staticEndPoints, endPoint)
	return s
}

// EnableMetrics enables/disables the /metrics endpoint
func (s *WebServer) EnableMetrics(enable bool) *WebServer {
	s.enableMetrics = enable
	return s
}

// Start configure and start the web server
func (s *WebServer) Start() (err error) {

	// Redirect trailing slash
	s.router.StrictSlash(true)

	// Configure REST endpoints
	s.configRestEndpoints()

	// Configure Web Socket endpoints
	s.configWsEndpoints()

	// Configure static files endpoints
	s.configStaticFilesHandler()

	// Configure middleware functions
	s.configMiddlewareFunctions()

	// If enabled, configure the metrics endpoint
	if s.enableMetrics {
		s.router.Handle("/metrics", promhttp.Handler())
	}

	// Add configured router to http
	http.Handle("/", s.router)

	// test if port only specified, with preceding ":"
	addr, err := s.getAddress()
	if err != nil {
		return err
	}

	// Build CORS configuration - CHECK IF IT CAN BE DONE USING MIDDLEWARE
	corsConfig := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"X-ACCESS-TOKEN", "X-TIMEZONE", "X-API-KEY, Content-Disposition", "Content-Filename"},
		MaxAge:           1209600,
	})

	// Start listen and serve requests
	return http.ListenAndServe(addr, corsConfig.Handler(s.router))
}

// Get HTTP listener address from port
func (s *WebServer) getAddress() (string, error) {

	// test if port only specified, with preceding ":"
	addr := ""
	if s.port[0] == ':' {
		addr = s.port
	} else {
		// test if port only without preceding ":"
		if _, err := strconv.Atoi(s.port); err == nil {
			addr = fmt.Sprintf(":%s", s.port)
		} else {
			//assume that SERVER_PORT specified as "host_name_or_ip:port"
			if !strings.Contains(s.port, ":") {
				return "", errors.New("invalid SERVER_PORT parameter specified: " + s.port)
			} else {
				addr = s.port
			}
		}
	}
	return addr, nil
}

// configure server with all the REST endpoints
func (s *WebServer) configRestEndpoints() {

	// Collect all rest entries from all the endpoints
	entries := rest.RestEntries{}
	for _, ep := range s.restEndPoints {
		for _, restEntry := range ep.Entries() {
			entry := restEntry
			entries = append(entries, &entry)
		}
	}

	// Sort entries to ensure proper URL pattern matching (static routes before dynamic routes)
	sort.Sort(entries)

	// Add handlers to mux router (in order)
	for _, e := range entries {
		s.router.HandleFunc(e.Path, s.restAdapterFunc(*e)).Methods(e.Method)
	}
	return
}

// configure server with all Web Socket endpoints
func (s *WebServer) configWsEndpoints() {
	//for _, ep := range s.webSocketEndPoints {
	//	for _, entry := range ep.Entries() {
	//		listener := socket.NewListener(entry)
	//		s.router.HandleFunc(entry.Path, listener.ListenForWSConnections)
	//	}
	//}
}

// configure server with all static file endpoints
func (s *WebServer) configStaticFilesHandler() {
	for _, sep := range s.staticEndPoints {
		for _, entry := range sep.Entries() {
			if len(entry.Folder) > 0 {
				subDirectory := fmt.Sprintf("./%s/", entry.Folder)
				staticHandler := http.StripPrefix(fmt.Sprintf("%s/", entry.Path), http.FileServer(http.Dir(subDirectory)))
				s.router.PathPrefix(fmt.Sprintf("%s", entry.Path)).Handler(staticHandler)
			}
		}

	}
}

// configure middleware functions
func (s *WebServer) configMiddlewareFunctions() {
	for _, mf := range s.middlewareFunctions {
		if mf != nil {
			s.router.Use(mf)
		}
	}
}
