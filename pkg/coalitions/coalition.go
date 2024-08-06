// package coalitions defines the coalitions in DCS World.
package coalitions

// Coalition is the ID of a coalition in DCS World.
type Coalition int

const (
	// Red is the ID of the red coalition.
	Red = 1
	// Blue is the ID of the blue coalition.
	Blue = 2
	// Neutrals is the ID of the neutral coalition.
	Neutrals = 3
)

// All returns all coalitions.
func All() []Coalition {
	return []Coalition{Red, Blue, Neutrals}
}
