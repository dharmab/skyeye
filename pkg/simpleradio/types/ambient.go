package types

type Ambient struct {
	Volume float64 `json:"vol"`
	Type   string  `json:"abType"`
}

func NewAmbient() Ambient {
	return Ambient{
		Volume: 1.0,
	}
}
