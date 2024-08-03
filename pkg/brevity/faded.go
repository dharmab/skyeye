package brevity

// FadedCall reports a previously tracked group has not been updated by on or off-board sensors for 30 seconds.
// Reference: ATP 3-52.4 Chapter V section 19 subsection a
type FadedCall struct {
	// Group which has faded.
	Group Group
}

// FadedCall reports a previously tracked group has not been updated by on or off-board sensors for 30 seconds,
// and disappeared within weapons range of a friendly aircraft.
// Reference: ATP 3-52.4 Chapter V section 19 subsection a
type VanishedCall struct {
	// Group which has vanished.
	Group Group
}
