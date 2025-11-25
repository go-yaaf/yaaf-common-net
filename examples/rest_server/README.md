# REST Server Example

This example demonstrate how to configure and run REST server using the **yaaf-common-net** package.

## Use Case
This example demonstrates a web server exposing 2 groups of REST endpoints:
* Group 1: hero - exposing the following endpoints:
  * ```/v1/hero//heroes``` endpoint with several REST methods (GET | POST | PUT | PATCH | DELETE) to manage Heroes
  * ```/v1/hero/villains``` endpoint with several REST methods (GET | POST | PUT | PATCH | DELETE) to manage Villains
* Group 2: sample - exposing the following endpoints:
    * ```/v1/sample/accounts``` endpoint with several REST methods (GET | POST | PUT | PATCH | DELETE) to manage accounts
    * ```/v1/sample/users``` endpoint with several REST methods (GET | POST | PUT | PATCH | DELETE) to manage users
  
## REST Server configuration
In this example, we will configure REST server in the ```main.go``` file using the following steps:

	// Create the new instance of the Web server
	webServer := web.NewWebServer()

	// Set API Version ton inject in the X-API-VERSION header (usually it will be provided by the configuration)
	webServer.WithAPIVersion("1.0.1")

	// Set encryption secrets
	webServer.WithSecrets(secret, signing)

	// Enforce checking the name from the API Key
	webServer.WithAppName("rest-server-example")

	// Add REST endpoints for hero group
	webServer.AddRESTEndpoints(hero.NewListOfHeroEndPoints()...)

	// Add REST endpoints for sample group
	webServer.AddRESTEndpoints(sample.NewListOfSampleEndPoints()...)

	// Add static documentation endpoint
	webServer.AddStaticEndpoint("/doc", "./doc")

To configure REST endpoint, look at the file: **hero/heroes_endpoint.go**:

The function ```Path()``` returns the heroes endpoints group path.
The function ```RestEntries()``` returns array of REST entries (methods) configuration.
Each entry specifies the following attributes:
- Method: The HTTP method verb: GET | PUT | POST | PATCH | DELETE							        [Mandatory]
- Handler: Handler function pointer of type func (c *gin.Context)									[Mandatory]
- Path: The method relative path (from the Endpoint base path)										[Mandatory]
- Skip: Flag to indicate if the API KEY or TOKEN validations should be skipped						[Optional]
- Role: Integer value (usually enum) to enable this entry only for token with the specified role	[Optional]