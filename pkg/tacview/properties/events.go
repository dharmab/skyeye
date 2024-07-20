package properties

const (
	// Mesage is a generic event.
	MessageEvent = "Message"
	// Bookmark events are highlighted in the time line and in the event log.
	// They are handy to highlight parts of the flight, like a bombing run, or when the trainee was in her final approach for landing.
	BookmarkEvent = "Bookmark"
	// Debug events are highlighted and easy to spot in the timeline and event log.
	// Because they must be used for development purposes, they are displayed only when launching Tacview with the command line argument /Debug:on
	DebugEvent = "Debug"
	// LeftArea events specify when an aircraft (or any object) is cleanly removed from the battlefield (not destroyed).
	LeftAreaEvent = "LeftArea"
	// Destroyed events specify when an object has been officially destroyed.
	DestroyedEvent = "Destroyed"
	// TakenOff injects a take-off event into the flight recording.
	TakenOffEvent = "TakenOff"
	// Landed injects a landed event into the flight recording.
	LandedEvent = "Landed"
	// Timeout is mainly used for real-life training debriefing to specify when a weapon (typically a missile) reaches or misses its target.
	// Tacview will report in the shot log as well as in the 3D view the result of the shot.
	//
	// Most parameters are optional.
	// SourceId designates the object which has fired the weapon, while TargetId designates the target.
	// Even if the displayed result may be in nautical miles, bullseye coordinates must be specified in meters.
	// The target must be explicitly (manually) destroyed or disabled using the appropriate properties independently from this event.
	TimeoutEvent = "Timeout"
)
