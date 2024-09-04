package properties

const (
	// Name is the common notation for each object. It is strongly recommended to use ICAO or NATO names like: C172 or F/A-18C.
	Name = "Name"
	// Type is the object's type tags.
	Type = "Type"
	// Transform contains the object's coordinates, relative to the global reference point.
	Transform = "T"

	AdditionalType                    = "AdditionalType"
	Parent                            = "Parent"
	Next                              = "Next"
	ShortName                         = "ShortName"
	LongName                          = "LongName"
	FullName                          = "FullName"
	CallSign                          = "CallSign"
	Registration                      = "Registration"
	Squawk                            = "Squawk"
	ICAO24                            = "ICAO24"
	Pilot                             = "Pilot"
	Group                             = "Group"
	Country                           = "Country"
	Coalition                         = "Coalition"
	Color                             = "Color"
	Shape                             = "Shape"
	Debug                             = "Debug"
	Label                             = "Label"
	FocusedTarget                     = "FocusedTarget"
	LockedTarget                      = "LockedTarget"
	LockedTarget2                     = "LockedTarget2"
	LockedTarget3                     = "LockedTarget3"
	LockedTarget4                     = "LockedTarget4"
	LockedTarget5                     = "LockedTarget5"
	LockedTarget6                     = "LockedTarget6"
	LockedTarget7                     = "LockedTarget7"
	LockedTarget8                     = "LockedTarget8"
	LockedTarget0                     = "LockedTarget0"
	Importance                        = "Importance"
	Slot                              = "Slot"
	Disabled                          = "Disabled"
	Visible                           = "Visible"
	Health                            = "Health"
	Length                            = "Length"
	Width                             = "Width"
	Height                            = "Height"
	Radius                            = "Radius"
	IAS                               = "IAS"
	CAS                               = "CAS"
	TAS                               = "TAS"
	Mach                              = "Mach"
	AOA                               = "AOA"
	AOS                               = "AOS"
	AGL                               = "AGL"
	HDG                               = "HDG"
	HDM                               = "HDM"
	Throttle                          = "Throttle"
	Throttle2                         = "Throttle2"
	EngineRPM                         = "EngineRPM"
	EngineRPM2                        = "EngineRPM2"
	NR                                = "NR"
	NR2                               = "NR2"
	RotorRPM                          = "RotorRPM"
	RotorRPM2                         = "RotorRPM2"
	Afterburner                       = "Afterburner"
	AirBrakes                         = "AirBrakes"
	Flaps                             = "Flaps"
	LandingGear                       = "LandingGear"
	LandingGearHandle                 = "LandingGearHandle"
	Tailhook                          = "Tailhook"
	Parachute                         = "Parachute"
	DragChute                         = "DragChute"
	FuelWeight                        = "FuelWeight"
	FuelWeight2                       = "FuelWeight2"
	FuelWeight3                       = "FuelWeight3"
	FuelWeight4                       = "FuelWeight4"
	FuelWeight5                       = "FuelWeight5"
	FuelWeight6                       = "FuelWeight6"
	FuelWeight7                       = "FuelWeight7"
	FuelWeight8                       = "FuelWeight8"
	FuelWeight9                       = "FuelWeight9"
	FuelVolume                        = "FuelVolume"
	FuelFlowWeight                    = "FuelFlowWeight"
	FuelFlowWeight2                   = "FuelFlowWeight2"
	FuelFlowWeight3                   = "FuelFlowWeight3"
	FuelFlowWeight4                   = "FuelFlowWeight4"
	FuelFlowWeight5                   = "FuelFlowWeight5"
	FuelFlowWeight6                   = "FuelFlowWeight6"
	FuelFlowWeight7                   = "FuelFlowWeight7"
	FuelFlowWeight8                   = "FuelFlowWeight8"
	FuelFlowWeight9                   = "FuelFlowWeight9"
	FuelFlowVolume                    = "FuelFlowVolume"
	FuelFlowVolume2                   = "FuelFlowVolume2"
	FuelFlowVolume3                   = "FuelFlowVolume3"
	FuelFlowVolume4                   = "FuelFlowVolume4"
	FuelFlowVolume5                   = "FuelFlowVolume5"
	FuelFlowVolume6                   = "FuelFlowVolume6"
	FuelFlowVolume7                   = "FuelFlowVolume7"
	FuelFlowVolume8                   = "FuelFlowVolume8"
	FuelFlowVolume9                   = "FuelFlowVolume9"
	RadarMode                         = "RadarMode"
	RadarAzimuth                      = "RadarAzimuth"
	RadarElevation                    = "RadarElevation"
	RadarRoll                         = "RadarRoll"
	RadarRange                        = "RadarRange"
	RadarHorizontalBeamwidth          = "RadarHorizontalBeamwidth"
	RadarVerticalBeamwidth            = "RadarVerticalBeamwidth"
	RadarRangeGateAzimuth             = "RadarRangeGateAzimuth"
	RadarRangeGateElevation           = "RadarRangeGateElevation"
	RadarRangeGateRoll                = "RadarRangeGateRoll"
	RadarRangeGateMin                 = "RadarRangeGateMin"
	RadarRangeGateMax                 = "RadarRangeGateMax"
	RadarRangeGateHorizontalBeamwidth = "RadarRangeGateHorizontalBeamwidth"
	RadarRangeGateVerticalBeamwidth   = "RadarRangeGateVerticalBeamwidth"
	LockedTargetMode                  = "LockedTargetMode"
	LockedTargetAzimuth               = "LockedTargetAzimuth"
	LockedTargetElevation             = "LockedTargetElevation"
	LockedTargetRange                 = "LockedTargetRange"
	EngagementMode                    = "EngagementMode"
	EngagementMode2                   = "EngagementMode2"
	EngagementRange                   = "EngagementRange"
	EngagementRange2                  = "EngagementRange2"
	VerticalEngagementRange           = "VerticalEngagementRange"
	VerticalEngagementRange2          = "VerticalEngagementRange2"
	RollControlInput                  = "RollControlInput"
	PitchControlInput                 = "PitchControlInput"
	YawControlInput                   = "YawControlInput"
	RollControlPosition               = "RollControlPosition"
	PitchControlPosition              = "PitchControlPosition"
	YawControlPosition                = "YawControlPosition"
	RollTrimTab                       = "RollTrimTab"
	PitchTrimTab                      = "PitchTrimTab"
	YawTrimTab                        = "YawTrimTab"
	AileronLeft                       = "AileronLeft"
	AileronRight                      = "AileronRight"
	Elevator                          = "Elevator"
	Rudder                            = "Rudder"
	LocalizerLateralDeviation         = "LocalizerLateralDeviation"
	GlideslopeVerticalDeviation       = "GlideslopeVerticalDeviation"
	LocalizerAngularDeviation         = "LocalizerAngularDeviation"
	GlideslopeAngularDeviation        = "GlideslopeAngularDeviation"
	PilotHeadRoll                     = "PilotHeadRoll"
	PilotHeadPitch                    = "PilotHeadPitch"
	PilotHeadYaw                      = "PilotHeadYaw"
	PilotEyeGazePitch                 = "PilotEyeGazePitch"
	PilotEyeGazeYaw                   = "PilotEyeGazeYaw"
	VerticalGForce                    = "VerticalGForce"
	LongitudinalGForce                = "LongitudinalGForce"
	LateralGForce                     = "LateralGForce"
	TriggerPressed                    = "TriggerPressed"
	HeartRate                         = "HeartRate"
	SpO2                              = "SpO2"
)
