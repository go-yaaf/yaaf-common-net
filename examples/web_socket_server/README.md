# WEB Socket Server Example

This example demonstrate how to configure and run Web server serving web sockets using the **yaaf-common-net** package.

## Use Case
This example demonstrates a web server static HTML content for the frontend and a web-socket endpoint to accept and push messages
from the server to the frontend.
In this example, the server is periodically querying a public service provided by OpenSky network to get realtime location and info about all
the airplanes in a specific region.
This information is pushed to the client via the websocket to continuously update the airplanes location and data on the map.
* The frontend is based on a simple Bootstrap 5 framework with Leaflet map component and is using html templates to render the static content:
    * ```/index.html``` The static HTML with all the styling and script including the web-socket client
* The server also exposes the web-socket endpoint for the client to connect and get frequent airplanes updates
    * ```/v1/ws/airplanes``` The web-socket endpoint

## REST Server configuration
In this example, we will configure Web server in the ```main.go``` file using the following steps:

	// Create the new instance of the Web server
	webServer := web.NewWebServer()

	// Set API Version ton inject in the X-API-VERSION header (usually it will be provided by the configuration)
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

To configure Web Socket endpoint, look at the file: **airplane_config/airplane_ws_endpoint.go**:

The function ```Group()``` returns the web-socket group name (the web server supports multiple groups of web-sockets).
The function ```Path()``` returns the web-socket endpoint path.

The function ```WSEntries()``` returns array of web-socket entries (message-handlers) configuration.
Each entry specifies the following attributes:
- OpCode: The message op-code (each op-code is handled by a different handler)		                [Mandatory]
- Handler: Handler function pointer of type: ```func (m IWSMessage, rw IWSClient) error``` 			[Mandatory]
- Message: The message factory method                        										[Mandatory]
