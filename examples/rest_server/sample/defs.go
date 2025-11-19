package sample

import (
	"github.com/go-yaaf/yaaf-common-net/web"
)

const (
	sampleApiVersion = "/v1/sample"
)

// NewListOfSampleEndPoints is a factory method for support endpoints list
func NewListOfSampleEndPoints() []web.RestEndpoint {
	list := make([]web.RestEndpoint, 0)
	list = append(list, NewAccountsEndPoint())
	list = append(list, NewUsersEndPoint())

	return list
}
