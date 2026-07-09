package web

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-yaaf/yaaf-common/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// allowedWSOrigins is the explicit allowlist of browser Origins permitted to open
// a web-socket connection. When empty, only same-host and non-browser (no Origin
// header) connections are accepted. Use SetAllowedWSOrigins to configure it.
var allowedWSOrigins []string

// SetAllowedWSOrigins configures the web-socket Origin allowlist (exact matches,
// e.g. "https://app.example.com"). This prevents Cross-Site WebSocket Hijacking,
// which the previous "always allow" policy exposed.
func SetAllowedWSOrigins(origins ...string) {
	allowedWSOrigins = origins
}

// checkWSOrigin enforces the web-socket Origin policy.
func checkWSOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")

	// Non-browser clients (native apps, server-to-server) send no Origin header.
	if origin == "" {
		return true
	}

	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	// Same-host requests are always allowed.
	if strings.EqualFold(u.Host, r.Host) {
		return true
	}

	// Otherwise the Origin must be explicitly allowlisted.
	for _, allowed := range allowedWSOrigins {
		if strings.EqualFold(origin, allowed) {
			return true
		}
	}
	return false
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
	CheckOrigin:       checkWSOrigin,
}

// WSListener is a wrapper for web socket listener
type WSListener struct {
	registry IWSClientRegistry
	decoder  IMessageDecoder
	handlers map[int]WSEntry
}

// NewListener factory method
func NewListener(registry IWSClientRegistry, cfg IWSEndpointConfig) (wsh *WSListener) {
	wsh = &WSListener{
		registry: registry,
		handlers: make(map[int]WSEntry, len(cfg.WSEntries())),
	}

	// Set message decoder
	wsh.decoder = NewJsonDecoder()

	for k, handlerEntry := range cfg.WSEntries() {
		wsh.handlers[k] = handlerEntry
	}
	return
}

// ListenForWSConnections Listen for web socket connections
func (h *WSListener) ListenForWSConnections(w http.ResponseWriter, r *http.Request) {

	connectedClients := h.registry.ConnectedClients()
	maxConnectedClients := 10000

	if connectedClients == maxConnectedClients {
		http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("error upgrading connection from %s to Web Socket: %s", r.RemoteAddr, err.Error())
		return
	}

	// Fetch value of request's ctx for "client_id" key/value pair.
	// if found, will be used as client connection id. otherwise, generated id will be used.
	clientId := uuid.New().String()
	qParams := make(map[string]string)

	// Get client id from context
	if r.Context().Value("clientId") != nil {
		clientId = r.Context().Value("clientId").(string)
	}

	// Get extra query params from context
	if r.Context().Value("params") != nil {
		qParams = r.Context().Value("params").(map[string]string)
	}

	// Get query params
	for k, v := range r.URL.Query() {
		qParams[k] = v[0]
	}

	// Inject HTTP headers to the params
	for k, v := range r.Header {
		qParams[k] = fmt.Sprintf("%v", v)
	}

	//conn.EnableWriteCompression(true)
	conn.EnableWriteCompression(false)

	tcpConn := conn.NetConn().(*net.TCPConn)
	_ = tcpConn.SetLinger(0)
	_ = tcpConn.SetNoDelay(true)
	_ = tcpConn.SetWriteBuffer(1048576)
	_ = tcpConn.SetReadBuffer(1048576)

	wsClient := NewWsClient(clientId, conn, h.onDisconnected)
	h.registry.RegisterClient(wsClient)
	return
}

func (h *WSListener) onDisconnected(ws IWSClient) {
	h.registry.UnregisterClient(ws)
}
