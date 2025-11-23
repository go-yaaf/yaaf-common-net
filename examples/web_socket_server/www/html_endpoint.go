package www

import (
	"fmt"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"

	. "github.com/go-yaaf/yaaf-common-net/web"
)

// region Endpoint structure and factory method ------------------------------------------------------------------------

// HtmlEndPoint is a list of HTML generated pages from templates
type HtmlEndPoint struct {
	BaseEndPoint
}

// NewHtmlEndPoint factory method
func NewHtmlEndPoint() RestEndpoint {
	return &HtmlEndPoint{}
}

func (h *HtmlEndPoint) Path() string {
	return "/ui"
}

// RestEntries provide HTML methods configuration
func (h *HtmlEndPoint) RestEntries() (restEntries []RestEntry) {
	restEntries = []RestEntry{
		{Method: http.MethodGet, Handler: h.deviceMap, Path: "/index.html", Skip: TOKEN},
	}

	// Sort entries for best match
	sort.Slice(restEntries, func(i, j int) bool {
		return restEntries[i].Path > restEntries[j].Path
	})
	return
}

func (h *HtmlEndPoint) deviceMap(c *gin.Context) {
	// Generate HTML page from template: "index.html" with data
	fmt.Println(c.FullPath())
	c.HTML(http.StatusOK, "index.html", gin.H{"title": "Index", "version": "1.0.0"})
}

// endregion
