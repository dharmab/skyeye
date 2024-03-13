package composer

import (
	"github.com/dharmab/skyeye/pkg/brevity"
)

// Composer converts brevity responses from structured forms into natural language.
// It is nondeterministic; the same input may randomly produce different output, to add variety and personality to the bot's respones.
type Composer interface {
	ComposeAlphaCheckResponse(brevity.AlphaCheckResponse) NaturalLanguageResponse
	ComposeBogeyDopeResponse(brevity.BogeyDopeResponse) NaturalLanguageResponse
	ComposeDeclareResponse(brevity.DeclareResponse) NaturalLanguageResponse
	ComposeFadedCall(brevity.FadedCall) NaturalLanguageResponse
	ComposePictureResponse(brevity.PictureResponse) NaturalLanguageResponse
	ComposeRadioCheckResponse(brevity.RadioCheckResponse) NaturalLanguageResponse
	ComposeSnaplockResponse(brevity.SnaplockResponse) NaturalLanguageResponse
	ComposeSpikedResponse(brevity.SpikedResponse) NaturalLanguageResponse
	ComposeSunriseCall(brevity.SunriseCall) NaturalLanguageResponse
	ComposeThreatCall(brevity.ThreatCall) NaturalLanguageResponse
}

// NaturalLanguageResponse contains the composer's responses in text form.
type NaturalLanguageResponse struct {
	// Subtitle is how the response will be displayed as in-game text.
	Subtitle string
	// Speech is the input to the TTS provider.
	Speech string
}

type composer struct {
	callsign string
}

func New() Composer {
	return &composer{}
}
