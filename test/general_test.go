package test

import (
	"fmt"
	"strings"
	"testing"
)

func TestRestPathExtraction(t *testing.T) {

	p := "/go/to/dashboard"

	entries := make(map[string]string)
	entries["GET /users/add"] = "/users/add"
	entries["POST  /users/add"] = "/users/add/new"

	for id := range entries {
		idx := strings.LastIndex(id, " ")
		restPath := id[idx+1:]

		if strings.HasPrefix(p, restPath) {
			fmt.Println(p)
			return
		}
	}
}
