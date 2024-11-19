package composer

import (
	"slices"
	"strings"
)

func (c *composer) ComposeCallsigns(callsigns ...string) string {
	for i, callsign := range callsigns {
		callsigns[i] = strings.ToUpper(callsign)
	}
	if len(callsigns) == 1 {
		return callsigns[0]
	}
	slices.Sort(callsigns)
	return strings.Join(callsigns, ", ")
}
