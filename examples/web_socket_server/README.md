# WEB Socket Server Example

This example demonstrate how to configure and run Web server serving web sockets using the **yaaf-common-net** package.
In this example, we will configure Web server in the ```main.go``` file using the following steps:

	// Create the new instance of the REST server
	restServer := web.NewRESTServer()

	// Set API Version ton inject in the X-API-VERSION header (usually it will be provided by the configuration)
	restServer.WithAPIVersion("1.0.1")

	// Set encryption secrets
	restServer.WithSecrets(secret, signing)

	// Enforce checking the name from the API Key
	restServer.WithAppName("rest-server-example")

	// Add REST endpoints for hero group
	restServer.AddEndpoints(hero.NewListOfHeroEndPoints()...)

	// Add REST endpoints for sample group
	restServer.AddEndpoints(sample.NewListOfSampleEndPoints()...)

	// Add static documentation endpoint
	restServer.AddStaticEndpoint("/doc", "./doc")

To configure REST endpoint, look at the file: **hero/heroes_endpoint.go**:

The function ```Path()``` returns the heroes endpoints group path.
The function ```RestEntries()``` returns array of REST entries (methods) configuration.
Each entry specifies the following attributes:
- Method: The HTTP method verb: GET | PUT | POST | PATCH | DELETE							        [Mandatory]
- Handler: Handler function pointer of type func (c *gin.Context)									[Mandatory]
- Path: The method relative path (from the Endpoint base path)										[Mandatory]
- Skip: Flag to indicate if the API KEY or TOKEN validations should be skipped						[Optional]
- Role: Integer value (usually enum) to enable this entry only for token with the specified role	[Optional]