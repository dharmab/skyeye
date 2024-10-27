package telemetry

import (
	acmi "github.com/dharmab/goacmi/properties/coalitions"
	skyeye "github.com/dharmab/skyeye/pkg/coalitions"
)

func propertyToCoalition(v string) skyeye.Coalition {
	switch v {
	case string(acmi.Allies):
		return skyeye.Red
	case string(acmi.Enemies):
		return skyeye.Blue
	default:
		return skyeye.Neutrals
	}
}
