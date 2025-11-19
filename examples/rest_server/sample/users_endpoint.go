package sample

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	. "github.com/go-yaaf/yaaf-common-net/web"
)

// region Endpoint structure and factory method ------------------------------------------------------------------------

// UsersEndPoint is a list of deployment related actions for system support only
// @Path: /v1/sample/users
// @Context: users
// @RequestHeader: X-API-KEY     |  The key to identify the application (console)
// @RequestHeader: Authorization | The bearer token to identify the logged-in user
// @ResourceGroup: System Administrator
type UsersEndPoint struct {
	BaseEndPoint
}

// NewUsersEndPoint factory method
func NewUsersEndPoint() RestEndpoint {
	return &UsersEndPoint{}
}

func (h *UsersEndPoint) Path() string {
	return sampleApiVersion + "/users"
}

func (h *UsersEndPoint) RestEntries() (restEntries []RestEntry) {
	restEntries = []RestEntry{
		{Method: http.MethodPut, Handler: h.handler, Path: ""},
		{Method: http.MethodPut, Handler: h.handler, Path: "/"},

		{Method: http.MethodPatch, Handler: h.handler, Path: ""},
		{Method: http.MethodPatch, Handler: h.handler, Path: "/"},

		{Method: http.MethodDelete, Handler: h.handler, Path: "/:id"},
		{Method: http.MethodGet, Handler: h.handler, Path: "/:id"},

		{Method: http.MethodGet, Handler: h.handler, Path: ""},
		{Method: http.MethodGet, Handler: h.handler, Path: "/"},

		{Method: http.MethodPatch, Handler: h.handler, Path: "/:id"},
		{Method: http.MethodPatch, Handler: h.handler, Path: "/logo"},
		{Method: http.MethodGet, Handler: h.handler, Path: "/logo/:id"},

		{Method: http.MethodGet, Handler: h.handler, Path: "/export/:format"},
	}

	// Sort entries for best match
	sort.Slice(restEntries, func(i, j int) bool {
		return restEntries[i].Path > restEntries[j].Path
	})
	return
}

func (h *UsersEndPoint) handler(c *gin.Context) {
	// Get token data
	td := h.GetTokenData(c)
	if td == nil {
		return
	}

	path := c.FullPath()
	c.JSON(http.StatusOK, NewActionResponse(path, td.SubjectId))
}

// endregion
