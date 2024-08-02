package properties

const (
	// DataSource is the source simulator, control station or file format.
	DataSource = "DataSource"
	// DataRecorder is the oftware or hardware used to record the data.
	DataRecorder = "DataRecorder"
	// ReferenceTime is the UTC base time for the current mission.
	// This time is combined with each frame offset (in seconds) to get the final absolute UTC time for each data sample.
	ReferenceTime = "ReferenceTime"
	// RecordingTime is the file creation UTC time.
	RecordingTime = "RecordingTime"
	// Author is the creator of this recording.
	Author = "Author"
	// Title of the mission.
	Title = "Title"
	// Category of the mission.
	Category = "Category"
	// Mission briefing text.
	Briefing = "Briefing"
	// Mission debriefing text.
	Debriefing = "Debriefing"
	// Free comments about the flight. May contain escaped commas or EOL characters.
	Comments = "Comments"
	// ReferenceLongitude is the longitude of a median point in degrees.
	// Add this to each object's longtitude to get the actual longitude.
	ReferenceLongitude = "ReferenceLongitude"
	// ReferenceLatitude is the latitude of a median point in degrees.
	// Add this to each object's latitude to get the actual latitude.
	ReferenceLatitude = "ReferenceLatitude"
)
