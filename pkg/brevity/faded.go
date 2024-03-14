package brevity

// FadedCall reports a previously tracked group has not been updated by on or off-board sensors for 30 seconds.
// Reference: ATP 3-52.4 Chapter V section 19 subsection a
type FadedCall interface {
	Faded()
	// Group which has faded.
	Group() Group
}
