package brevity

import "fmt"

// FadedCall reports a previously tracked group has not been updated by on or off-board sensors for 30 seconds.
// Reference: ATP 3-52.4 Chapter V section 19 subsection a.
type FadedCall struct {
	// Group which has faded.
	Group Group
}

func (c FadedCall) String() string {
	return fmt.Sprintf("FADED: group %s", c.Group)
}
