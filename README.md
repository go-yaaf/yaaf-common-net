# yaaf-common-net

[![Build](https://github.com/go-yaaf/yaaf-common-net/actions/workflows/build.yml/badge.svg)](https://github.com/go-yaaf/yaaf-common-net/actions/workflows/build.yml)

`yaaf-common-net` is a Go library that provides a collection of networking utilities and web server implementations. It simplifies the creation of RESTful APIs, WebSocket services, and static file servers.

This library is built on top of `gin-gonic/gin` for the web framework and depends on the `go-yaaf/yaaf-common` library for common interfaces and utilities like logging.

## Features

-   **REST Server**: Easily create and manage RESTful APIs with a flexible endpoint and routing system.
-   **WebSocket Server**: Built-in support for WebSocket communication, including client management and message handling.
-   **Static File Server**: Serve static files and directories.
-   **Middleware Support**: Leverages Gin's middleware for request processing, authentication, and more.
-   **Token-Based Authentication**: Utilities for handling JWT and other token-based authentication schemes.

## Installation

To add `yaaf-common-net` to your project, use `go get`:

```bash
go get -u github.com/go-yaaf/yaaf-common-net
```

## Usage

### Creating a REST Server

Here is a simple example of how to set up a REST server:

```go
package main

import (
	"github.com/go-yaaf/yaaf-common-net/web"
	"github.com/go-yaaf/yaaf-common/logger"
)

func main() {
	// Initialize logger
	logger.Init()

	// Create a new web server instance
	restServer := web.NewWebServer()

	// Configure the server
	restServer.WithAPIVersion("1.0.0")
	restServer.WithAppName("my-rest-app")

	// Add REST endpoints (see next section)
	// restServer.AddRESTEndpoints(...)

	// Add a static file server
	restServer.AddStaticEndpoint("/static", "./public")

	// Start the server
	port := 8080
	logger.Info("Starting REST server on port %d", port)
	if err := restServer.Start(port); err != nil {
		logger.Fatal("Failed to start server: %v", err)
	}
}
```

### Defining a REST Endpoint

Endpoints are defined by creating a struct that embeds `web.BaseEndPoint` and implements the `web.RestEndpoint` interface.

```go
package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/go-yaaf/yaaf-common-net/web"
)

// MyEndpoint defines our sample endpoint
type MyEndpoint struct {
	web.BaseEndPoint
}

// NewMyEndpoint creates a new instance of the endpoint
func NewMyEndpoint() web.RestEndpoint {
	return &MyEndpoint{}
}

// Path returns the base path for this endpoint
func (e *MyEndpoint) Path() string {
	return "/api/v1/hello"
}

// RestEntries defines the routes for this endpoint
func (e *MyEndpoint) RestEntries() []web.RestEntry {
	return []web.RestEntry{
		{
			Method:  http.MethodGet,
			Path:    "",
			Handler: e.sayHello,
		},
	}
}

// sayHello is the handler function for our GET route
func (e *MyEndpoint) sayHello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Hello, World!"})
}

// In your main function, add the endpoint to the server:
// restServer.AddRESTEndpoints(NewMyEndpoint())
```

### Creating a WebSocket Server

Setting up a WebSocket server is similar to a REST server. You can also handle broadcasting messages to connected clients.

```go
package main

import (
	"github.com/go-yaaf/yaaf-common-net/web"
	"github.com/go-yaaf/yaaf-common/logger"
	"time"
)

func main() {
	// Create a new web server instance
	webServer := web.NewWebServer()

	// Add WebSocket endpoints
	// webServer.AddWebSocketEndpoints(...)

	// Start the server
	go webServer.Start(8080)

	// Example of broadcasting a message every 5 seconds
	registry := webServer.WebSocketRegistry("my_ws_group")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		logger.Info("Broadcasting message to clients")
		registry.Broadcast([]byte("This is a message from the server"))
	}
}
```

### Defining a WebSocket Endpoint

A WebSocket endpoint handles incoming messages based on their operation code (`OpCode`).

```go
package main

import (
	"github.com/go-yaaf/yaaf-common-net/web"
	"github.com/go-yaaf/yaaf-common/logger"
)

// MyWSEndpoint defines our WebSocket endpoint
type MyWSEndpoint struct{}

func NewMyWSEndpoint() web.IWSEndpointConfig {
	return &MyWSEndpoint{}
}

// Group returns the WebSocket group name
func (ws *MyWSEndpoint) Group() string {
	return "my_ws_group"
}

// Path returns the endpoint URL path
func (ws *MyWSEndpoint) Path() string {
	return "/ws/v1/updates"
}

// WSEntries defines handlers for incoming messages
func (ws *MyWSEndpoint) WSEntries() []web.WSEntry {
	return []web.WSEntry{
		{
			OpCode:  1, // Custom operation code for a specific message type
			Handler: ws.handleCustomMessage,
			Message: func() web.IWSMessage { return &web.WSMessage{} },
		},
	}
}

// handleCustomMessage processes incoming messages
func (ws *MyWSEndpoint) handleCustomMessage(m web.IWSMessage, c web.IWSClient) error {
	logger.Info("Received message from client %s: %v", c.ID(), m.Payload())
	// Echo the message back to the client
	return c.Push(m)
}

// In your main function, add the endpoint to the server:
// webServer.AddWebSocketEndpoints(NewMyWSEndpoint())
```

## Examples

For more detailed examples, please refer to the `examples` directory in this repository:
-   [REST Server Example](./examples/rest_server)
-   [WebSocket Server Example](./examples/web_socket_server)


## License

This project is licensed under the [MIT License](./LICENSE).
