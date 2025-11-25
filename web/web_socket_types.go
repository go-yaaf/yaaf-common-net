package web

import (
	"encoding/json"
	"net/http"
)

const (
	WsPingOpCode = 0
)

// IWSMessage is a Web socket message header interface:
type IWSMessage interface {
	MessageCode() int  // Get message op-code
	MessageID() uint64 // Get message unique ID
	SessionID() string // Get session ID
	Payload() any      // Get arbitrary message payload
}

// region Web Socket message header ------------------------------------------------------------------------------------

// WSMessageHeader is a common structure for all web socket messages
type WSMessageHeader struct {
	OpCode    int
	MessageId uint64
	SessionId string
}

// MessageCode get web-socket message op-code
func (mb WSMessageHeader) MessageCode() int { return mb.OpCode }

// MessageID get web-socket message ID
func (mb WSMessageHeader) MessageID() uint64 { return mb.MessageId }

// SessionID get web-socket session ID
func (mb WSMessageHeader) SessionID() string { return mb.SessionId }

// endregion

// region Web Socket Ping Pong messages --------------------------------------------------------------------------------

// WSPingMessage message sent from client (for keep-alive)
type WSPingMessage struct {
	WSMessageHeader
}

// Payload returns the message payload
func (mp WSPingMessage) Payload() interface{} { return nil }

// NewWsPingMessage creates a new ping message
func NewWsPingMessage() IWSMessage {
	return WSPingMessage{WSMessageHeader: WSMessageHeader{OpCode: WsPingOpCode}}
}

var pingMessage = NewWsPingMessage()

// endregion

// region Web Socket Raw message ---------------------------------------------------------------------------------------

// WSRawMessage is a raw message structure
type WSRawMessage struct {
	WSMessageHeader
	Body []byte
}

// Payload returns the message payload
func (m *WSRawMessage) Payload() any { return m.Body }

// endregion

// region Web Socket message decoder -----------------------------------------------------------------------------------

// IMessageDecoder is a message decoder interface
type IMessageDecoder interface {
	Encode(message IWSMessage) ([]byte, error)
	Decode(buffer []byte) (IWSMessage, error)
}

// endregion

// region Web Socket Client Callbacks ----------------------------------------------------------------------------------

// PongReceivedCb PONG message handler received from client
type PongReceivedCb func(sessionId, pongMessage string, latencyMs int64)

// MessageReceivedCb message received callback
type MessageReceivedCb func(IWSClient, IWSMessage, int)

// DisconnectedCb called when client disconnected
type DisconnectedCb func(IWSClient)

// endregion

// region Web Socket client --------------------------------------------------------------------------------------------

// WSConnectParams is the configuration for a web socket connection
type WSConnectParams struct {
	Url                string      // Full url (int is case path and host are ignored)
	Path               string      // URL path segment
	Host               string      // url host + port
	WriteBufferSize    int         // Write buffer size (if not provided use the default 8K buffer)
	ReadBufferSize     int         // Read buffer size (if not provided use the default 8K buffer)
	CompressionEnabled bool        // Try to enable compression
	Header             http.Header // List of HTTP headers
}

// IWSClient is a Web socket client interface
type IWSClient interface {
	ID() string              // Socket client unique ID
	Send(m IWSMessage) error // Send message through the socket
	SendRaw(m []byte) error  // Send arbitrary data through the socket
	Close() error            // Close connection
}

// endregion

// region Message factory and default message decoder (JSON) -----------------------------------------------------------

// MessageFactoryFunc is a function that creates a new message instance
type MessageFactoryFunc func() IWSMessage

var messageFactories = map[int]MessageFactoryFunc{}

// AddMessageFactory adds a message factory for a given opcode
func AddMessageFactory(opcode int, f MessageFactoryFunc) {
	messageFactories[opcode] = f
}

// GetMessageFactoryFunc returns the message factory for a given opcode
func GetMessageFactoryFunc(opcode int) MessageFactoryFunc {
	return messageFactories[opcode]
}

// JsonDecoder is a JSON message decoder
type JsonDecoder struct{}

// NewJsonDecoder creates a new JSON message decoder
func NewJsonDecoder() IMessageDecoder {
	return &JsonDecoder{}
}

// Encode encodes a message to JSON
func (_ JsonDecoder) Encode(m IWSMessage) (result []byte, err error) {
	return json.Marshal(m)
}

// Decode decodes a JSON message
func (_ JsonDecoder) Decode(buffer []byte) (msg IWSMessage, err error) {

	bm := &WSMessageHeader{}

	if err = json.Unmarshal(buffer, bm); err != nil {
		return nil, err
	}

	if mf, ok := messageFactories[bm.MessageCode()]; ok {
		msg = mf()
		if err = json.Unmarshal(buffer, msg); err != nil {
			return nil, err
		}
	} else {
		msg = &WSRawMessage{
			WSMessageHeader: WSMessageHeader{
				OpCode:    bm.MessageCode(),
				MessageId: bm.MessageID(),
				SessionId: bm.SessionID(),
			},
			Body: buffer,
		}
	}
	return
}

// endregion

// WSMessageHandler is a function that handles a web socket message
type WSMessageHandler func(m IWSMessage, rw IWSClient) error

// WSEntry is a web socket entry configuration
type WSEntry struct {
	OpCode  int                // Message op-code
	Message MessageFactoryFunc // Message factory function
	Handler WSMessageHandler   // Message handler function
}

// IWSEndpointConfig ia a Web socket endpoint configuration interface
type IWSEndpointConfig interface {
	Group() string        // Web socket registry group
	Path() string         // Web socket endpoint path
	WSEntries() []WSEntry // List of Web socket entries configuration
}

// IWSClientRegistry is aWeb socket client registry
type IWSClientRegistry interface {
	Start()
	RegisterClient(c IWSClient)
	UnregisterClient(c IWSClient)
	ConnectedClients() int
	Client(id string) IWSClient
	Broadcast(msg []byte)
}

// WSClientFactory is a function that creates a new web socket client
type WSClientFactory func() IWSClient
