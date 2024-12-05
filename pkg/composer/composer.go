// Package composer converts brevity responses from structured forms into natural language.
package composer

// Composer converts brevity responses from structured forms into natural language.
// It is nondeterministic; the same input may randomly produce different output, to add variety and personality to the bot's responses.
type Composer struct {
	// Callsign of the GCI controller
	Callsign string
}

// NaturalLanguageResponse contains the composer's responses in text form.
type NaturalLanguageResponse struct {
	// Subtitle is how the response will be displayed as in-game text.
	Subtitle string
	// Speech is the input to the TTS provider.
	Speech string
}

// Write appends text to the subtitle and speech fields.
func (r *NaturalLanguageResponse) Write(speech, subtitle string) {
	r.Speech += speech
	r.Subtitle += subtitle
}

// WriteBoth appends the same text to the subtitle and speech fields.
func (r *NaturalLanguageResponse) WriteBoth(s string) {
	r.Write(s, s)
}

// WriteResponse appends the given response's subtitle and speech to this response.
func (r *NaturalLanguageResponse) WriteResponse(response NaturalLanguageResponse) {
	r.Write(response.Speech, response.Subtitle)
}

func applyToFirstCharacter(s string, f func(string) string) string {
	if len(s) == 0 {
		return s
	}
	return f(s[:1]) + s[1:]
}
