package types

// Ambient is related to the ambient audio feature introduced in SRS 2.0.9.1. It is included for completeness but is otherwise unused in SkyEye.
type Ambient struct {
	Volume float64 `json:"vol"`
	Type   string  `json:"abType"`
}

// NewAmbient returns a new Ambient with all fields set to reasonable defaults.
func NewAmbient() Ambient {
	return Ambient{
		Volume: 1.0,
	}
}
