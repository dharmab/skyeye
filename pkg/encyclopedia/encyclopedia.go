package encyclopedia

type Encyclopedia interface {
	Aircraft() map[string]Aircraft
	AircraftByPlatformDesignation(string) []Aircraft
}

type encyclopedia struct {
}

var _ Encyclopedia = &encyclopedia{}

func New() Encyclopedia {
	return &encyclopedia{}
}

func (e *encyclopedia) Aircraft() map[string]Aircraft {
	var out = make(map[string]Aircraft)
	for _, a := range aircraftData {
		out[a.EditorType] = a
	}
	return out
}

func (e *encyclopedia) AircraftByPlatformDesignation(platform string) (out []Aircraft) {
	for _, a := range aircraftData {
		if a.PlatformDesignation == platform {
			out = append(out, a)
		}
	}
	return
}
