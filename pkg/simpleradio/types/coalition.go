package types

// https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Common/Network/SRClient.cs
type Coalition int

const (
	CoalitionRed  = 1
	CoalitionBlue = 2
)

func IsSpectator(c Coalition) bool {
	return (c != CoalitionRed) && (c != CoalitionBlue)
}
