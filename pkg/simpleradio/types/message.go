package types

// MessageType is an enum indicating the type of an SRS data protocol message.
type MessageType int

const (
	MessageUpdate MessageType = iota
	MessagePing
	MessageSync
	MessageRadioUpdate
	MessageServerSettings
	MessageClientDisconnect
	MessageVersionMismatch
	MessageExternalAWACSModePassword
	MessageExternalAWACSModeDisconnect
)

// Message is the JSON schema of SRS protocol messages. The SRS data protocol sends these messages, one per line, in JSON format over the TCP connection.
// The order of fields in this type matches the order of fields in the official SRS client, just in case a different order were to trigger some obscure bug.
type Message struct {
	// Version is the SRS client version.
	Version string `json:"Version"`
	// Client is used in messages that reference a single client.
	Client ClientInfo `json:"Client,omitempty"`
	// Clients is used in messages that reference multiple clients.
	Clients []ClientInfo `json:"Clients,omitempty"`
	// ServerSettings is a map of server settings and their values. It sometimes appears in Sync messages.
	ServerSettings map[string]string `json:"ServerSettings,omitempty"`
	// ExternalAWACSModePassword is the External AWACS Mode password, used in ExternalAWACSModePassword messages to authenticate a client as an AWACS.
	ExternalAWACSModePassword string `json:"ExternalAWACSModePassword,omitempty"`
	// Type is the type of the message.
	Type MessageType `json:"MsgType"`
}
