package properties

import "github.com/dharmab/skyeye/pkg/coalitions"

// RED IS ALLIES - DCS descends from the Flanker games where the player was red by default

const AlliesCoalition = "Allies"   // RED
const EnemiesCoalition = "Enemies" // BLUE

func CoaliationToProperty(coalition coalitions.Coalition) string {
	switch coalition {
	case coalitions.Red:
		return AlliesCoalition
	case coalitions.Blue:
		return EnemiesCoalition
	default:
		return "Neutrals"
	}
}

func PropertyToCoalition(property string) coalitions.Coalition {
	switch property {
	case AlliesCoalition:
		return coalitions.Red
	case EnemiesCoalition:
		return coalitions.Blue
	default:
		return coalitions.Neutrals
	}
}
