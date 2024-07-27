package encyclopedia

type Encyclopedia interface {
	Aircraft() map[string]Aircraft
}

type encyclopedia struct {
}

var _ Encyclopedia = &encyclopedia{}

func New() Encyclopedia {
	return &encyclopedia{}
}

// Aircraft returns a map of aircraft data keyed by the name they appear by in ACMI telemetry.
func (e *encyclopedia) Aircraft() map[string]Aircraft {
	var out = make(map[string]Aircraft)
	for _, a := range aircraftData {
		out[a.ACMIShortName] = a
	}
	return out
}
