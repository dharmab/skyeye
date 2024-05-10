package webeditor

const (
	BlueCoalitionName    = "blue"
	RedCoalitionName     = "red"
	NeutralCoalitionName = "neutrals"
)

type Mission struct {
	Coalition CoalitionMap `json:"coalition"`
	Theatre   string       `json:"theatre"`
}

type CoalitionMap struct {
	Blue     Coalition `json:"blue"`
	Neutrals Coalition `json:"neutrals"`
	Red      Coalition `json:"red"`
}

type Coalition struct {
	Name     string    `json:"name"`
	Bullseye Bullseye  `json:"bullseye"`
	Country  []Country `json:"country"`
}

type Bullseye struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Country struct {
	Plane Plane `json:"plane"`
}

type Plane struct {
	Group []PlaneGroup `json:"group"`
}

type PlaneGroup struct {
	Name  string      `json:"name"`
	Units []PlaneUnit `json:"units"`
}

type PlaneUnit struct {
	Name         string  `json:"name"`
	UnitID       uint32  `json:"unitId"`
	EditorType   string  `json:"type"`
	Altitude     float64 `json:"alt"`
	AltitudeType string  `json:"alt_type"`
	Heading      float64 `json:"heading"`
	Speed        float64 `json:"speed"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
}
