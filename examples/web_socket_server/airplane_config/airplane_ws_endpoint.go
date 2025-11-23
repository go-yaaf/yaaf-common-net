package airplane_config

import (
	. "github.com/go-yaaf/yaaf-common-net/web"
	"github.com/go-yaaf/yaaf-common/logger"
)

// region Web socket messages ------------------------------------------------------------------------------------------

// GetPlaneInfoMessage is control message to get specific airplane info
type GetPlaneInfoMessage struct {
	WSMessageHeader
	planeId string
}

func NewGetPlaneInfoMessage() IWSMessage {
	return &GetPlaneInfoMessage{WSMessageHeader: WSMessageHeader{OpCode: 1}}
}

func (m *GetPlaneInfoMessage) Payload() any {
	return m.planeId
}

// endregion

// region Endpoint structure and factory method ------------------------------------------------------------------------

// AirplanesSocketEndPoint is a device web-socket endpoint
// @Path: /v1/ws/airplanes
// @RequestHeader: X-API-KEY     |  The key to identify the application (console)
// @RequestHeader: Authorization | The bearer token to identify the logged-in user
// @ResourceGroup: Web Sockets
type AirplanesSocketEndPoint struct {
	handlers map[int]WSMessageHandler
}

func NewAirplanesSocketEndPoint() IWSEndpointConfig {
	return &AirplanesSocketEndPoint{}
}

// Group returns the web socket group
func (h *AirplanesSocketEndPoint) Group() string {
	return "airplanes"
}

// Path returns the endpoint path
func (h *AirplanesSocketEndPoint) Path() string {
	return "/v1/ws/airplanes"
}

// WSEntries returns the list of handlers for the endpoint
func (h *AirplanesSocketEndPoint) WSEntries() []WSEntry {
	return []WSEntry{
		{OpCode: NewGetPlaneInfoMessage().MessageCode(), Handler: h.handleGetPlaneInfo, Message: NewGetPlaneInfoMessage},
	}
}

func (h *AirplanesSocketEndPoint) handleGetPlaneInfo(m IWSMessage, rw IWSClient) error {
	logger.Info("Get airplane info for client: %s for plane :%v", rw.ID(), m.Payload())
	return nil
}

// endregion
