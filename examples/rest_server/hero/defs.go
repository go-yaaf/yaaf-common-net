package hero

import (
	"github.com/go-yaaf/yaaf-common-net/web"
)

const (
	heroApiVersion = "/v1/hero"
)

// NewListOfHeroEndPoints is a factory method for support endpoints list
func NewListOfHeroEndPoints() []web.RestEndpoint {
	list := make([]web.RestEndpoint, 0)
	list = append(list, NewHeroesEndPoint())
	list = append(list, NewVillainsEndPoint())

	return list
}
