package composer

import (
	"slices"
	"strings"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (*Composer) composeCallsigns(callsigns ...string) string {
	for i, callsign := range callsigns {
		if callsign != brevity.LastCaller {
			callsigns[i] = strings.ToUpper(callsign)
		}
	}
	if len(callsigns) == 1 {
		return callsigns[0]
	}
	slices.Sort(callsigns)
	return strings.Join(callsigns, ", ")
}
