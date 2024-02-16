package types

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

type Message struct {
	Client  *ClientInfo  `json:"Client,omitempty"`
	Clients []ClientInfo `json:"Clients,omitempty"`
	// ServerSettings is a map of server settings and their values. It sometimes appears in Sync messages.
	ServerSettings            map[string]string `json:"ServerSettings,omitempty"`
	ExternalAWACSModePassword string            `json:"ExternalAWACSModePassword,omitempty"`
	Type                      MessageType       `json:"MsgType"`
	Version                   string            `json:"Version"`
}
