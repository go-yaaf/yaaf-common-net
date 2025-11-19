package main

import (
	"github.com/go-yaaf/yaaf-common-net/examples/rest_server/hero"
	"github.com/go-yaaf/yaaf-common-net/examples/rest_server/sample"
	"github.com/go-yaaf/yaaf-common-net/web"
	"github.com/go-yaaf/yaaf-common/logger"
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

	// Create the new instance of the REST server
	restServer := web.NewRESTServer()

	// Set API Version (usually it will be provided by the configuration)
	restServer.WithAPIVersion("1.0.1")

	// Set encryption secrets
	restServer.WithSecrets(secret, signing)

	// Enforce checking the name from the API Key
	restServer.WithAppName("rest-server-example")

	// Add REST endpoints
	restServer.AddEndpoints(hero.NewListOfHeroEndPoints()...)
	restServer.AddEndpoints(sample.NewListOfSampleEndPoints()...)

	// Add static documentation endpoint
	restServer.AddStaticEndpoint("/doc", "./doc")

	port := 8080
	logger.Info("Starting REST server, listening on port: %d", port)

	// Start REST server for prometheus metrics endpoint
	go func() {
		if err := restServer.Start(port); err != nil {
			logger.Warn("error starting REST server: %s", err.Error())
		} else {
			logger.Info("Closing the REST server...")
		}
	}()

	<-make(chan struct{})
}
