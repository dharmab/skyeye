package types

import (
	"slices"

	"github.com/dharmab/skyeye/pkg/coalitions"
)

// ClientInfo is information about the client included in messages.
type ClientInfo struct {
	// GUID a unique client ID.
	GUID GUID `json:"ClientGuid"`
	// Name is the name that will appear in the client list and in in-game transmissions
	Name string `json:"Name"`
	// Seat is the seat number for multicrew aircraft. For bots, set this to 0.
	Seat int `json:"Seat"`
	// Coalition is the side that the client will act on
	Coalition coalitions.Coalition `json:"Coalition"`
	// AllowRecording indicates consent to record audio server-side. For bots, this should usually be set to True.
	AllowRecording bool `json:"AllowRecord"`
	// RadioInfo contains the client's unit, radios, transponder and ambient audio settings.
	RadioInfo RadioInfo `json:"RadioInfo"`
	// Position is the unit's in-game location. This is omitted for external clients not bound to a unit.
	Position *Position `json:"LatLngPosition,omitempty"`
}

// RadioInfo is information about a client's radios.
type RadioInfo struct {
	// Radios is the inventory of radios operated by the client
	Radios []Radio `json:"radios,omitempty"`
	// Unit is the name of the unit that the client is bound to.
	Unit string `json:"unit"`
	// UnitID is the in-game ID of the unit that the client is bound to.
	UnitID uint64 `json:"unitId"`
	// IFF contains the client's transponder settings
	IFF Transponder `json:"iff"`
	// Ambient contains the client's ambient audio settings
	Ambient Ambient `json:"ambient"`
}

// IsOnFrequency is true if the other client has a radio with the same frequency, modulation, and encryption settings as this client.
func (i *RadioInfo) IsOnFrequency(other RadioInfo) bool {
	for _, thisRadio := range i.Radios {
		if slices.ContainsFunc(other.Radios, thisRadio.IsSameFrequency) {
			return true
		}
	}
	return false
}
