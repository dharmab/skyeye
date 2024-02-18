package types

// ClientInfo is information about the client included in the message
type ClientInfo struct {
	// GUID is some kind of unique ID for the client (???)
	GUID GUID `json:"ClientGuid"`
	// Name is the name that will appear in the client list and in in-game transmissions
	Name string `json:"Name"`
	// Seat is the seat number for multicrew aircraft. For bots, set this to 0.
	Seat int `json:"Seat"`
	// Coalition is the side that the client will act on
	Coalition Coalition `json:"Coalition"`
	// AllowRecording indicates consent to record audio server-side. For bots, this should usually be set to True.
	AllowRecording bool      `json:"AllowRecord"`
	RadioInfo      RadioInfo `json:"RadioInfo"`
	Position       *Position `json:"LatLngPosition,omitempty"`
}

type RadioInfo struct {
	Radios  []Radio `json:"radios,omitempty"`
	Unit    string  `json:"unit"`
	UnitID  uint64  `json:"unitId"`
	IFF     IFF     `json:"iff"`
	Ambient Ambient `json:"ambient"`
}
