package types

// https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Common/DCSState/Transponder.cs
type IFFControlMode int

const (
	IFFControlModeCockpit  = 0
	IFFControlModeOverlay  = 1
	IFFControlModeDisabled = 2
)

// https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Common/DCSState/Transponder.cs
type IFFStatus int

const (
	IFFStatusOff    = 0
	IFFStatusNormal = 1
	IFFStatusIdent  = 2
)

type IFFMode int

const IFFModeDisabled = -1
const IFFMicDisabled = -1

// https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Common/DCSState/IFF.cs
type IFF struct {
	ControlMode IFFControlMode `json:"control"`
	Status      IFFStatus      `json:"status"`
	Mode1       IFFMode        `json:"mode1"`
	Mode2       IFFMode        `json:"mode2"`
	Mode3       IFFMode        `json:"mode3"`
	Mode4       bool           `json:"mode4"`
	Mic         int            `json:"mic"`
}

func NewIFF() IFF {
	return IFF{
		ControlMode: IFFControlModeDisabled,
		Status:      IFFStatusOff,
		Mode1:       IFFModeDisabled,
		Mode2:       IFFModeDisabled,
		Mode3:       IFFModeDisabled,
		Mode4:       false,
		Mic:         IFFMicDisabled,
	}
}
