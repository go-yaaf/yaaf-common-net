package hero

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	. "github.com/go-yaaf/yaaf-common-net/web"

	. "github.com/go-yaaf/yaaf-common-net/examples/rest_server/model"
)

// region Endpoint structure and factory method ------------------------------------------------------------------------

// HeroesEndPoint is a list of deployment related actions for system support only
// @Path: /v1/hero/heroes
// @Context: heroes
// @RequestHeader: X-API-KEY     |  The key to identify the application (console)
// @RequestHeader: Authorization | The bearer token to identify the logged-in user
// @ResourceGroup: System Administrator
type HeroesEndPoint struct {
	BaseEndPoint
}

// NewHeroesEndPoint factory method
func NewHeroesEndPoint() RestEndpoint {
	return &HeroesEndPoint{}
}

func (h *HeroesEndPoint) Path() string {
	return heroApiVersion + "/heroes"
}

/*
RestEntries provide REST methods configuration
Each entry specifies the following attributes:
  - Method: The HTTP method type (VERB): GET | PUT | POST | PATCH | DELETE							[Mandatory]
  - Handler: Handler function pointer of type func (c *gin.Context)									[Mandatory]
  - Path: The method relative path (from the Endpoint base path)										[Mandatory]
  - Skip: Flag to indicate if the API KEY or TOKEN validations should be skipped						[Optional]
  - Role: Integer value (usually enum) to enable this entry only for token with the specified role	[Optional]
*/
func (h *HeroesEndPoint) RestEntries() (restEntries []RestEntry) {
	restEntries = []RestEntry{
		{Method: http.MethodPut, Handler: h.handler, Path: "", Role: Roles.SUPPORT + Roles.SALES},
		{Method: http.MethodPut, Handler: h.handler, Path: "/", Role: Roles.SUPPORT + Roles.SALES},

		{Method: http.MethodPatch, Handler: h.handler, Path: "", Role: Roles.SUPPORT + Roles.SALES},
		{Method: http.MethodPatch, Handler: h.handler, Path: "/", Role: Roles.SUPPORT + Roles.SALES},

		{Method: http.MethodDelete, Handler: h.handler, Path: "/:id", Role: Roles.MANAGEMENT},
		{Method: http.MethodGet, Handler: h.handler, Path: "/:id", Role: Roles.FINANCE},

		{Method: http.MethodGet, Handler: h.handler, Path: ""},
		{Method: http.MethodGet, Handler: h.handler, Path: "/"},

		{Method: http.MethodPatch, Handler: h.handler, Path: "/:id", Role: Roles.FINANCE},
		{Method: http.MethodPatch, Handler: h.handler, Path: "/logo"},
		{Method: http.MethodGet, Handler: h.handler, Path: "/logo/:id", Skip: APIKEY},

		{Method: http.MethodGet, Handler: h.handler, Path: "/export/:format", Skip: TOKEN},
	}

	// Sort entries for best match
	sort.Slice(restEntries, func(i, j int) bool {
		return restEntries[i].Path > restEntries[j].Path
	})
	return
}

func (h *HeroesEndPoint) handler(c *gin.Context) {
	// Get token data
	td := h.GetTokenData(c)
	if td == nil {
		return
	}

	path := c.FullPath()
	c.JSON(http.StatusOK, NewActionResponse(path, td.SubjectId))
}

// endregion
