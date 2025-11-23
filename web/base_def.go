package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

// RestEntry represent a single HTTP REST call
type RestEntry struct {
	Path    string          // Rest method path
	Method  string          // HTTP method verb
	Handler gin.HandlerFunc // Handler function
	Skip    int             // Skip validation
	Role    int             // Role flags
}

func (b *RestEntry) ID(base string) string {
	return fmt.Sprintf("%s %s%s", b.Method, base, b.Path)
}

// RestEndpoint is a group of RestEntry
type RestEndpoint interface {
	Path() string             // Rest method path
	RestEntries() []RestEntry // List of REST entries
}
