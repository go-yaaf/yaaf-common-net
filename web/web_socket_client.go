package web

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-yaaf/yaaf-common/logger"
	"github.com/go-yaaf/yaaf-common/utils"
	"github.com/gorilla/websocket"
)

// region Web Socket client structure and fluent API configuration -----------------------------------------------------

// WSClient represent single web socket client handler
type WSClient struct {
	id             string          // Web socket client unique ID
	conn           *websocket.Conn // Pointer to the underlying web socket connection
	decoder        IMessageDecoder // Message decoder (if empty use default JSON decoder)
	handlers       map[int]WSEntry // Map of Web Socket entries
	onDisconnected DisconnectedCb  // Client disconnect callback
	send           chan []byte
}

// WSClientConfig is the configuration for a web socket client
type WSClientConfig struct {
	Id           string
	WsConn       *websocket.Conn
	Handlers     map[int]WSEntry
	onDisconnect DisconnectedCb
}

// NewWsClient creates a new web socket client
func NewWsClient(clientId string, conn *websocket.Conn, onDisconnect DisconnectedCb) IWSClient {

	ws := &WSClient{
		id:             clientId,
		conn:           conn,
		onDisconnected: onDisconnect,
	}

	ws.decoder = NewJsonDecoder()

	if ws.conn != nil {
		go ws.run()
	}
	return ws
}

// ID returns the client ID
func (c *WSClient) ID() string {
	return c.id
}

// Send typed message
func (c *WSClient) Send(msg IWSMessage) error {
	if buffer, err := json.Marshal(msg); err == nil {
		return c.SendRaw(buffer)
	} else {
		return fmt.Errorf("websocket client [%s]: message marshal failed: %v", c.id, err)
	}
}

// SendRaw send raw message
func (c *WSClient) SendRaw(buffer []byte) error {
	// Set write deadline to 60 seconds
	deadLine := time.Now().Add(time.Second * time.Duration(60))
	_ = c.conn.SetWriteDeadline(deadLine)

	if err := c.conn.WriteMessage(websocket.TextMessage, buffer); err != nil {
		if c.onDisconnected != nil {
			c.onDisconnected(c)
		}
		return fmt.Errorf("websocket client [%s]: send failed: %v", c.id, err)
	} else {
		return nil
	}
}

// Close closes the connection
func (c *WSClient) Close() error {
	close(c.send)
	return c.conn.Close()
}

// RemoteAddress returns the remote address of the client
func (c *WSClient) RemoteAddress() (ra string) {
	if c.conn != nil {
		ra = c.conn.RemoteAddr().String()
	}
	return
}

func (c *WSClient) run() {

	defer utils.RecoverAll(func(err interface{}) {
		logger.Error("WSClient::run error: %s", err)
	})

	for {
		if _, rawMessage, err := c.conn.ReadMessage(); err != nil {
			continue
		} else {
			if msg, fe := c.decoder.Decode(rawMessage); fe != nil {
				logger.Error("error decoding received message from: [%s]: error: %s message dump: %s", c.id, fe.Error(), string(rawMessage))
			} else {
				if len(c.handlers) > 0 {
					if mh, ok := c.handlers[msg.MessageCode()]; ok {
						go func() { _ = mh.Handler(msg, c) }()
						return
					}
				}
			}
		}
	}
}
