package types

import (
	"time"
)

// ClientConfiguration is configuration used to construct the audio and data clients
type ClientConfiguration struct {
	// GUID corresponds to [ClientInfo.GUID]
	GUID string
	// Address is the network address of the SRS server, including port
	Address string
	// ConnectionTimeout is the connection timeout for connecting to the SRS server
	ConnectionTimeout time.Duration
	// ClientName corresponds to [ClientInfo.Name]
	ClientName string
	// ExternalAWACSModePassword is the password for External AWACS Mode
	ExternalAWACSModePassword string
	// Coalition corresponds to [ClientInfo.Coalition]
	Coalition Coalition
	// Raduio is the [Radio] to listen and talk on
	Radio Radio
	// AllowRecording corresponds to [ClientInfo.AllowRecording]
	AllowRecording bool
}
