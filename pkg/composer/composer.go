// Package composer converts brevity responses from structured forms into natural language.
package composer

import (
	"fmt"
	"strings"
	"unicode"
)

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
	r.Speech = join(r.Speech, speech)
	r.Subtitle = join(r.Subtitle, subtitle)
}

// WriteBoth appends the same text to the subtitle and speech fields.
func (r *NaturalLanguageResponse) WriteBoth(s string) {
	r.Write(s, s)
}

// WriteBothf appends the formatted string to the subtitle and speech fields.
func (r *NaturalLanguageResponse) WriteBothf(format string, a ...any) {
	r.WriteBoth(fmt.Sprintf(format, a...))
}

// WriteResponse appends the given response's subtitle and speech to this response.
func (r *NaturalLanguageResponse) WriteResponse(response NaturalLanguageResponse) {
	r.Write(response.Speech, response.Subtitle)
}

// join concatenates two strings, adding a space between them if not already present.
func join(a, b string) string {
	if len(a) == 0 {
		return b
	}
	preceding := rune(a[len(a)-1])
	if !unicode.IsSpace(preceding) {
		return a + addSpacing(b)
	}
	return a + b
}

// addSpacing prepends a space to the string if it starts with a letter or number.
func addSpacing(s string) string {
	if len(s) == 0 {
		return s
	}
	first := rune(s[0])
	if unicode.IsLetter(first) || unicode.IsNumber(first) {
		return " " + s
	}
	return s
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
