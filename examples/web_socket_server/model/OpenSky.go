package model

// This file contains the OpenSky sample data models: the response and the payload (PlaneState)

// PlaneState is a simplified representation of a single aircraft status
type PlaneState struct {
	Icao24    string  `json:"icao24"`
	Callsign  string  `json:"callsign"`
	Country   string  `json:"country"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	Altitude  float64 `json:"altitude"`
	Heading   float64 `json:"heading"`
	Velocity  float64 `json:"velocity"`
	OnGround  bool    `json:"on_ground"`
	Raw       []any   `json:"raw,omitempty"` // optional: full row if you want
}

// OpenSkyResponse matches the /states/all response
type OpenSkyResponse struct {
	Time   int64           `json:"time"`
	States [][]interface{} `json:"states"`
}
