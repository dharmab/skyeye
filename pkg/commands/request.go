package commands

// Request is an envelope containing a trace ID, player name and text message.
type Request struct {
	// TraceID of the request.
	TraceID string
	// PlayerName is the name of the player that sent the message.
	PlayerName string
	// Text message.
	Text string
}
