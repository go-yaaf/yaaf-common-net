package web

import (
	"sync"
)

// DefaultClientRegistry is the basic implementation of web socket client registry
// It is used by web socket server to manage and track all the connected web socket clients
type DefaultClientRegistry struct {
	sync.RWMutex
	Connections map[string]IWSClient
	register    chan IWSClient
	unregister  chan IWSClient
	broadcast   chan []byte
}

// NewClientRegistry factory method
func NewClientRegistry(group string) IWSClientRegistry {
	return &DefaultClientRegistry{
		Connections: make(map[string]IWSClient),
	}
}

// Start Initialize registry
func (r *DefaultClientRegistry) Start() {
	for {
		select {
		case c := <-r.register:
			r.Lock()
			r.Connections[c.ID()] = c
			r.Unlock()
		case c := <-r.unregister:
			r.Lock()
			r.removeClient(c.ID())
			r.Unlock()
		case msg := <-r.broadcast:
			r.Lock()
			for _, c := range r.Connections {
				_ = c.SendRaw(msg)
			}
			r.Unlock()
		}
	}
}

func (r *DefaultClientRegistry) removeClient(id string) {
	if conn, ok := r.Connections[id]; ok {
		delete(r.Connections, id)
		_ = conn.Close()
	}
}

// RegisterClient Register new connected client
func (r *DefaultClientRegistry) RegisterClient(wsc IWSClient) {

	if r.Connections == nil || wsc == nil {
		return
	}

	r.Lock()
	r.Connections[wsc.ID()] = wsc
	r.Unlock()
}

// UnregisterClient Unregister disconnected client
func (r *DefaultClientRegistry) UnregisterClient(wsc IWSClient) {
	r.Lock()
	if conn, ok := r.Connections[wsc.ID()]; ok {
		delete(r.Connections, wsc.ID())
		_ = conn.Close()
	}
	r.Unlock()
}

// ConnectedClients Get number of current connected clients
func (r *DefaultClientRegistry) ConnectedClients() int {
	r.Lock()
	defer r.Unlock()
	return len(r.Connections)
}

// Client returns Web-socket client by ID
func (r *DefaultClientRegistry) Client(id string) IWSClient {
	r.Lock()
	defer r.Unlock()
	if conn, ok := r.Connections[id]; ok {
		return conn
	} else {
		return nil
	}
}

// Broadcast send message to all clients
func (r *DefaultClientRegistry) Broadcast(msg []byte) {
	r.Lock()
	for _, c := range r.Connections {
		_ = c.SendRaw(msg)
	}
	r.Unlock()
}
