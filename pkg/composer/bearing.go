package composer

import "github.com/dharmab/skyeye/pkg/bearings"

func ComposeBearing(bearing bearings.Bearing) NaturalLanguageResponse {
	return NaturalLanguageResponse{
		Subtitle: bearing.String(),
		Speech:   bearing.String(),
	}
}
