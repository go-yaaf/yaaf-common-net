package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/go-yaaf/yaaf-common/entity"

	. "github.com/go-yaaf/yaaf-common-net/model"
	"github.com/go-yaaf/yaaf-common-net/utils"
)

var whiteList map[string]int

const (
	Unknown  int = 0
	NoToken      = 1
	NoApiKey     = 2
)

func init() {
	whiteList = make(map[string]int)

	// The following methods don't require API Key or Token validations
	//whiteList["/health"] = NoApiKey + NoToken
	//whiteList["/health/"] = NoApiKey + NoToken
	//whiteList["/doc"] = NoApiKey + NoToken
	//whiteList["/doc/"] = NoApiKey + NoToken
	//whiteList["/my/auth"] = NoApiKey + NoToken

	// The following methods require API Key but not Token validations
	//whiteList["/user/authorize"] = NoToken

	// The following methods require API Key but not Token validations
}

// region REST server structure and factory method ---------------------------------------------------------------------

type Server struct {
	engine *gin.Engine
}

// NewRESTServer Factory method
func NewRESTServer() *Server {

	gin.SetMode(gin.ReleaseMode)
	engine := gin.Default()

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "X-API-KEY", "X-ACCESS-TOKEN", "X-TIMEZONE-OFFSET"},
		ExposeHeaders:    []string{"Content-Length", "X-API-KEY", "X-ACCESS-TOKEN", "X-TIMEZONE-OFFSET"},
		AllowCredentials: true,
		AllowWebSockets:  true,
		AllowWildcard:    true,
		MaxAge:           12 * time.Hour,
	}))

	engine.Use(
		corsMiddleware(),
		disableCache(),
		gin.CustomRecovery(customRecovery),
		apiKeyValidator(),
		tokenValidator(),
		apiVersion(),
	)

	return &Server{
		engine: engine,
	}
}

// endregion

// region REST server fluent API configuration -------------------------------------------------------------------------

// AddEndpoints add REST endpoints
func (s *Server) AddEndpoints(endpoints ...RestEndpoint) *Server {

	var group *gin.RouterGroup
	for _, ep := range endpoints {

		if len(ep.Path()) > 0 {
			group = s.engine.Group(ep.Path())
		} else {
			group = s.engine.Group("/")
		}

		for _, entry := range ep.RestEntries() {
			group.Handle(entry.Method, entry.Path, entry.Handler)
		}
		//group.OPTIONS("/", CorsOptions)
	}
	return s
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

// endregion

// region REST server builder and starter ------------------------------------------------------------------------------

// Start web server
func (s *Server) Start(port int) error {

	_ = s.engine.SetTrustedProxies(nil)

	if port == 0 {
		port = 8080
	}

	return s.engine.Run(fmt.Sprintf(":%d", port))
}

// endregion

// region REST server Middlewares --------------------------------------------------------------------------------------

// Fetch API key from the header and check it
func apiKeyValidator() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Skip OPTIONS
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get path and strip version
		restPath := strings.ToLower(c.Request.URL.Path)
		if strings.HasPrefix(restPath, "/v1/") {
			restPath = strings.Replace(restPath, "/v1/", "/", 1)
		}

		// Handle root
		if restPath == "/" || len(restPath) == 0 {
			c.Next()
			return
		}

		// Handle white-listed prefix (no need for token or API key)
		for path, value := range whiteList {
			if strings.HasPrefix(restPath, path) {
				if value > 1 {
					c.Next()
					return
				}
			}
		}

		apiKey := c.GetHeader("X-API-KEY")
		if _, err := utils.TokenUtils().ParseApiKey(apiKey); err != nil {
			_ = c.AbortWithError(http.StatusForbidden, fmt.Errorf("invalid API key for path: %s", restPath))
		} else {
			c.Next()
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

		// Get path and strip version
		restPath := strings.ToLower(c.Request.URL.Path)
		if strings.HasPrefix(restPath, "/v1/") {
			restPath = strings.Replace(restPath, "/v1/", "/", 1)
		}

		// Handle root
		if restPath == "/" || len(restPath) == 0 {
			c.Next()
			return
		}

		// Handle white-listed prefix (no need for token or API key)
		for path, value := range whiteList {
			if strings.HasPrefix(restPath, path) {
				if value > 0 {
					c.Next()
					return
				}
			}
		}

		td := getTokenData(c)
		if td == nil {
			_ = c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid auth token for path: %s", restPath))
			return
		}

		// Set new token
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
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, X-API-KEY, X-ACCESS-TOKEN, X-TIMEZONE, accept, origin, Cache-Control, X-Requested-With, Content-Disposition, Content-Filename")
		c.Writer.Header().Set("Access-Control-Exposed-Headers", "X-API-KEY, X-ACCESS-TOKEN, X-TIMEZONE, Content-Disposition, Content-Filename")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, HEAD")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

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
		c.Header("X-API-VERSION", "1.x")
	}
}

// endregion
