package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-yaaf/yaaf-common/entity"

	. "github.com/go-yaaf/yaaf-common-net/model"
	"github.com/go-yaaf/yaaf-common-net/utils"
)

var whiteList []string = []string{
	"/com.chrome.devtools.json",
	"/favicon.ico",
}

// region REST server structure and factory method ---------------------------------------------------------------------
var serverInst *Server = nil

// Server is the main web server structure
type Server struct {
	engine        *gin.Engine
	version       string
	appName       string
	templatesPath string
	entries       map[string]RestEntry         // Map of REST path to REST Entries
	registries    map[string]IWSClientRegistry // Map of web-socket groups name to web-socket client registry
	skipList      map[string]int               // Skip validation (API KEY and TOKEN) list
	headers       map[string]string            // Custom headers
}

// NewWebServer Factory method
func NewWebServer() *Server {

	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	engine.Use(
		corsMiddleware(),
		disableCache(),
		gin.CustomRecovery(customRecovery),
		apiKeyValidator(),
		tokenValidator(),
		apiVersion(),
	)
	server := &Server{
		engine:     engine,
		version:    "1.0.0",
		entries:    make(map[string]RestEntry),
		registries: make(map[string]IWSClientRegistry),
		skipList:   make(map[string]int),
		headers:    make(map[string]string),
	}
	serverInst = server
	return serverInst
}

// WithAppName sets the application name to check after parsing the API Key
func (s *Server) WithAppName(appName string) *Server {
	s.appName = appName
	return s
}

// WithAPIVersion set API version to inject to the header: X-API-VERSION
func (s *Server) WithAPIVersion(version string) *Server {
	s.version = version
	return s
}

// WithSecrets set token encryption secrets
func (s *Server) WithSecrets(apiSecret, signingKey string) *Server {
	utils.TokenUtils().WithSecrets(apiSecret, signingKey)
	return s
}

// WithHeader override HTTP headers for CORS manipulation
func (s *Server) WithHeader(header, value string) *Server {
	s.headers[header] = value
	return s
}

// WithHtmlTemplates sets the global HTML templates path
func (s *Server) WithHtmlTemplates(path string) *Server {
	s.templatesPath = path
	return s
}

// Extract SKIP validations flag from the current entry
func (s *Server) getEntrySkipFlag(method, path string) int {

	// First path: try match static path segments first
	for k, v := range s.entries {
		if v.Method == method {
			parts := strings.Fields(k)
			pt := parts[1]
			if strings.Contains(pt, ":") {
				continue
			}
			if s.matchExactPath(pt, path) {
				return v.Skip
			}
		}
	}

	// Now, try match parametrized path
	for k, v := range s.entries {
		if v.Method == method {
			parts := strings.Fields(k)
			pt := parts[1]
			if strings.Contains(pt, ":") {
				if s.matchTemplatePath(pt, path) {
					return v.Skip
				}
			}
		}
	}

	// Now, try match skip list path
	if flag, ok := s.skipList[path]; ok {
		return flag
	}
	return 0
}

// Extract Role flags from the current entry
func (s *Server) getEntryRoleFlag(method, path string) int {

	// First path: try match static path segments first
	for k, v := range s.entries {
		if v.Method == method {
			parts := strings.Fields(k)
			pt := parts[1]
			if strings.Contains(pt, ":") {
				continue
			}
			if s.matchExactPath(pt, path) {
				return v.Role
			}
		}
	}

	// Now, try match parametrized path
	for k, v := range s.entries {
		if v.Method == method {
			parts := strings.Fields(k)
			pt := parts[1]
			if strings.Contains(pt, ":") {
				if s.matchTemplatePath(pt, path) {
					return v.Role
				}
			}
		}
	}
	return 0
}

// Check if path is matching the whitelist
func (s *Server) matchWhiteList(path string) bool {
	for _, pt := range whiteList {
		if strings.HasSuffix(path, pt) {
			return true
		}
	}
	return false
}

// match exact path
func (s *Server) matchExactPath(path, actual string) bool {
	return path == actual
}

// match template path
func (s *Server) matchTemplatePath(template, actual string) bool {

	templates := strings.Split(template, "/")
	segments := strings.Split(actual, "/")
	if len(templates) != len(segments) {
		return false
	}
	for i := 0; i < len(templates); i++ {
		if templates[i] == segments[i] {
			continue
		} else if strings.HasPrefix(templates[i], ":") {
			continue
		} else {
			return false
		}
	}

	// All parts feet
	return true
}

// Start web server
func (s *Server) Start(port int) error {

	_ = s.engine.SetTrustedProxies(nil)

	if len(s.templatesPath) > 0 {
		s.engine.LoadHTMLGlob(s.templatesPath)
	}

	if port == 0 {
		port = 8080
	}

	return s.engine.Run(fmt.Sprintf(":%d", port))
}

// WebSocketRegistry returns the provided group's client registry
func (s *Server) WebSocketRegistry(name string) IWSClientRegistry {
	return s.registries[name]
}

// endregion

// region REST server fluent API configuration -------------------------------------------------------------------------

// AddRESTEndpoints add REST endpoints
func (s *Server) AddRESTEndpoints(endpoints ...RestEndpoint) *Server {

	var group *gin.RouterGroup
	for _, ep := range endpoints {

		if len(ep.Path()) > 0 {
			group = s.engine.Group(ep.Path())
		} else {
			group = s.engine.Group("/")
		}

		for _, entry := range ep.RestEntries() {
			group.Handle(entry.Method, entry.Path, entry.Handler)
			s.entries[entry.ID(group.BasePath())] = entry
		}
		group.OPTIONS("/", CorsOptions)
	}
	return s
}

// CorsOptions handles CORS options
func CorsOptions(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, DELETE, POST, PUT, PATCH")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	c.Next()
}

// AddStaticEndpoint add static file endpoint (for documentation)
func (s *Server) AddStaticEndpoint(path, folder string) *Server {
	s.engine.Static(path, folder)
	return s
}

// AddStaticFile registers a single route in order to serve a single file of the local filesystem.
func (s *Server) AddStaticFile(path, relativePath string) *Server {
	s.engine.StaticFile(path, relativePath)
	return s
}

// AddWebSocketEndpoints registers a single route as a web-socket listener
func (s *Server) AddWebSocketEndpoints(endpoints ...IWSEndpointConfig) *Server {

	for _, ep := range endpoints {
		// Create web-socket registry for each endpoint

		var registry IWSClientRegistry = nil

		if reg, ok := s.registries[ep.Group()]; ok {
			registry = reg
		} else {
			registry = NewClientRegistry(ep.Group())
			s.registries[ep.Group()] = registry
			go registry.Start()
		}

		// Create listener for each endpoint
		listener := NewListener(registry, ep)

		// WS should not include API Key and Token
		s.skipList[ep.Path()] = TOKEN

		// Register path and start listener for each endpoint
		s.engine.GET(ep.Path(), func(c *gin.Context) {
			w, r := c.Writer, c.Request
			listener.ListenForWSConnections(w, r)
		})
	}

	return s
}

// endregion

// region Server Middlewares -------------------------------------------------------------------------------------------

// Fetch API key from the header and check it
func apiKeyValidator() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Skip OPTIONS
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Skip whitelist
		if serverInst.matchWhiteList(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get path and check if we need to skip API key
		restPath := strings.ToLower(c.Request.URL.Path)
		flag := serverInst.getEntrySkipFlag(c.Request.Method, restPath)
		if flag&APIKEY == APIKEY {
			c.Next()
			return
		}

		// Extract the API KEY from the header
		apiKey := c.GetHeader("X-API-KEY")

		// Parse API KEY and check if app name should be valid
		if appName, err := utils.TokenUtils().ParseApiKey(apiKey); err != nil {
			_ = c.AbortWithError(http.StatusForbidden, fmt.Errorf("invalid API key for path: %s", restPath))
		} else {
			if len(serverInst.appName) > 0 && serverInst.appName != appName {
				_ = c.AbortWithError(http.StatusForbidden, fmt.Errorf("invalid API key for path: %s", restPath))
			} else {
				c.Next()
			}
		}
	}
}

// Fetch and check token, after processing, renew token
func tokenValidator() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Skip OPTIONS
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Skip whitelist
		if serverInst.matchWhiteList(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get path and check if we need to skip Access Token
		restPath := strings.ToLower(c.Request.URL.Path)
		flag := serverInst.getEntrySkipFlag(c.Request.Method, restPath)
		if flag&TOKEN == TOKEN {
			c.Next()
			return
		}

		// Extract the user token from the header: X-ACCESS-TOKEN
		td := getTokenData(c)
		if td == nil {
			_ = c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid auth token for path: %s", restPath))
			return
		}

		// Get path and check if role guard exists
		roles := serverInst.getEntryRoleFlag(c.Request.Method, restPath)
		if roles > 0 {
			if roles&td.SubjectRole == 0 {
				_ = c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("user role not authorized for path: %s", restPath))
				return
			}
		}

		// Rewrite new token with new expiration time (30 minutes)
		if td.ExpiresIn > 0 {
			td.ExpiresIn = int64(entity.Now() + 1000*60*30)
		}

		if token, err := utils.TokenUtils().CreateToken(td); err != nil {
			return
		} else {
			c.Header("X-ACCESS-TOKEN", token)
		}
		c.Next()
	}
}

// GetTokenData extract security token data from Authorization header
func getTokenData(c *gin.Context) *TokenData {

	token := c.GetHeader("X-ACCESS-TOKEN")
	if len(token) == 0 {
		_ = c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid auth token"))
		return nil
	}
	if td, err := utils.TokenUtils().ParseToken(token); err != nil {
		_ = c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid auth token"))
		return nil
	} else {
		return td
	}
}

// Add response header to disable cache
func disableCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store")
	}
}

// Add custom recovery from any error
func customRecovery(c *gin.Context, recovered any) {
	if err, ok := recovered.(string); ok {
		c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
	}
	c.AbortWithStatus(http.StatusInternalServerError)
}

// Enable CORS
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token, X-API-KEY, X-ACCESS-TOKEN, X-TIMEZONE, accept, origin, Cache-Control, X-Requested-With, Content-Disposition, Content-Filename")
		c.Writer.Header().Set("Access-Control-Exposed-Headers", "X-API-KEY, X-ACCESS-TOKEN, X-TIMEZONE, Content-Disposition, Content-Filename")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// Add / Override custom headers
		for k, v := range serverInst.headers {
			c.Writer.Header().Set(k, v)
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Add response header with API version
func apiVersion() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-API-VERSION", "s.version")
	}
}

// endregion
