// package composer converts brevity responses from structured forms into natural language.
package composer

import (
	"github.com/dharmab/skyeye/pkg/brevity"
)

// Composer converts brevity responses from structured forms into natural language.
// It is nondeterministic; the same input may randomly produce different output, to add variety and personality to the bot's respones.
type Composer interface {
	// ComposeAlphaCheckResponse constructs natural language brevity for responding to an ALPHA CHECK.
	ComposeAlphaCheckResponse(brevity.AlphaCheckResponse) NaturalLanguageResponse
	// ComposeBogeyDopeResponse constructs natural language brevity for responding to a BOGEY DOPE call.
	ComposeBogeyDopeResponse(brevity.BogeyDopeResponse) NaturalLanguageResponse
	// ComposeCheckInResponse constructs natural language brevity for responding to an ambiguous CHECK IN call.
	ComposeCheckInResponse(brevity.CheckInResponse) NaturalLanguageResponse
	// ComposeDeclareResponse constructs natural language brevity for responding to a DECLARE call.
	ComposeDeclareResponse(brevity.DeclareResponse) NaturalLanguageResponse
	// ComposeFadedCall constructs natural language brevity for announcing a contact has faded.
	ComposeFadedCall(brevity.FadedCall) NaturalLanguageResponse
	// ComposeNegativeRadarContactResponse constructs natural language brevity for saying the controller cannot find a contact on the radar.
	ComposeNegativeRadarContactResponse(brevity.NegativeRadarContactResponse) NaturalLanguageResponse
	// ComposePictureResponse constructs natural language brevity for responding to a PICTURE call.
	ComposePictureResponse(brevity.PictureResponse) NaturalLanguageResponse
	// ComposeRaygunResponse constructs natural language brevity for responding to a RADIO CHECK.
	ComposeRadioCheckResponse(brevity.RadioCheckResponse) NaturalLanguageResponse
	// ComposeShoppingResponse constructs natural language brevity for educating a caller about SHOPPING brevity.
	ComposeShoppingResponse(brevity.ShoppingResponse) NaturalLanguageResponse
	// ComposeSnaplockResponse constructs natural language brevity for responding to a SNAPLOCK call.
	ComposeSnaplockResponse(brevity.SnaplockResponse) NaturalLanguageResponse
	// ComposeSpikedResponse constructs natural language brevity for responding to a SPIKED call.
	ComposeSpikedResponse(brevity.SpikedResponse) NaturalLanguageResponse
	// ComposeSunriseCall constructs natural language brevity for announcing GCI services are online.
	ComposeSunriseCall(brevity.SunriseCall) NaturalLanguageResponse
	// ComposeThreatCall constructs natural language brevity for announcing a threat.
	ComposeThreatCall(brevity.ThreatCall) NaturalLanguageResponse
	// ComposeMergedCall constructs natural language brevity for announcing a merge.
	ComposeMergedCall(brevity.MergedCall) NaturalLanguageResponse
	// ComposeSayAgainResponse constructs natural language brevity for asking a caller to repeat their last transmission.
	ComposeSayAgainResponse(brevity.SayAgainResponse) NaturalLanguageResponse
	// ComposeTripwireResponse constructs natural language brevity for educating a caller about threat monitoring.
	ComposeTripwireResponse(brevity.TripwireResponse) NaturalLanguageResponse
}

// NaturalLanguageResponse contains the composer's responses in text form.
type NaturalLanguageResponse struct {
	// Subtitle is how the response will be displayed as in-game text.
	Subtitle string
	// Speech is the input to the TTS provider.
	Speech string
}

func (r *NaturalLanguageResponse) Write(speech, subtitle string) {
	r.Speech += speech
	r.Subtitle += subtitle
}

func (r *NaturalLanguageResponse) WriteBoth(s string) {
	r.Write(s, s)
}

func (r *NaturalLanguageResponse) WriteResponse(response NaturalLanguageResponse) {
	r.Write(response.Speech, response.Subtitle)
}

type composer struct {
	// callsign of the GCI controller
	callsign string
}

func New(callsign string) Composer {
	return &composer{callsign: callsign}
}

func applyToFirstCharacter(s string, f func(string) string) string {
	if len(s) == 0 {
		return s
	}
	return f(s[:1]) + s[1:]
}
