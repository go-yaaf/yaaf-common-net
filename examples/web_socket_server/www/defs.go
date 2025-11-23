package www

import (
	"github.com/go-yaaf/yaaf-common-net/web"
)

// NewListOfHtmlEndPoints is a factory method for support endpoints list
func NewListOfHtmlEndPoints() []web.RestEndpoint {
	list := make([]web.RestEndpoint, 0)
	list = append(list, NewHtmlEndPoint())

	return list
}
