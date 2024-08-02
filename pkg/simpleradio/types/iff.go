package types

// This file contains types related to the SRS Transponder: https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Common/DCSState/Transponder.cs
// These types are not used by Skyeye, but are included for completeness.

// IFFControlMode is used by the SRS client as part of the configuration for how the player sets Transponder codes.
type IFFControlMode int

const (
	// IFFControlModeCockpit corresponds to IFFControlMode.COCKPIT in SRS
	IFFControlModeCockpit = 0
	// IFFControlModeOverlay corresponds to IFFControlMode.OVERLAY in SRS
	IFFControlModeOverlay = 1
	// IFFControlModeDisabled corresponds to IFFControlMode.DISABLED in SRS
	IFFControlModeDisabled = 2
)

// IFFStatus is used by the SRS client to indicate the output of the IFF system.
type IFFStatus int

const (
	// IFFStatusOff corresponds to IFFStatus.OFF in SRS
	IFFStatusOff = 0
	// IFFStatusNormal corresponds to IFFStatus.NORMAL in SRS
	IFFStatusNormal = 1
	// IFFStatusIdent corresponds to IFFStatus.IDENT in SRS
	IFFStatusIdent = 2
)

// IFFMode is used by the SRS client to indicate the mode of the IFF system.
type IFFMode int

// IFFModeDisabled is a special value used by the SRS client to indicate that a given transponder mode is disabled.
const IFFModeDisabled = -1

// IFFMicDisabled is a special value used by the SRS client to indicate that the mic-triggered ident mode is disabled.
const IFFMicDisabled = -1

// Transponder represents an aircraft's transponder.
type Transponder struct {
	// ControlMode is the mode in which the player sets Transponder codes.
	ControlMode IFFControlMode `json:"control"`
	// Status is the Transponder output state.
	Status IFFStatus `json:"status"`
	// Mode1 is a two digit military IFF code.
	Mode1 IFFMode `json:"mode1"`
	// Mode 2 is a four digit military IFF code.
	Mode2 IFFMode `json:"mode2"`
	// Mode 3 is a four digit military/civilian transponder code, also known as Mode A or Mode 3/A. This is the code that ATC uses to identify aircraft on radar.
	Mode3 IFFMode `json:"mode3"`
	// Mode 4 is an encrypted military IFF code. In SRS, it's a simple on/off state.
	Mode4 bool `json:"mode4"`
	// Mic is used by some aircraft that can auto-ident while the Mic switch is pressed.
	Mic int `json:"mic"`
}

// NewIFF returns a new Transponder with all fields set to reasonable defaults.
func NewIFF() Transponder {
	return Transponder{
		ControlMode: IFFControlModeDisabled,
		Status:      IFFStatusOff,
		Mode1:       IFFModeDisabled,
		Mode2:       IFFModeDisabled,
		Mode3:       IFFModeDisabled,
		Mode4:       false,
		Mic:         IFFMicDisabled,
	}
}
