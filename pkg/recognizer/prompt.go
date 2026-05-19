package recognizer

import (
	"fmt"
	"strings"
)

// prompt constructs a prompt for OpenAI's audio transcription models. See https://platform.openai.com/docs/guides/speech-to-text#prompting
func prompt(callsign string, locations []string) string {
	s := fmt.Sprintf("Either ANYFACE or %s, PILOT CALLSIGN, DIGITS, one of 'RADIO' 'ALPHA' 'BOGEY' 'PICTURE' 'DECLARE' 'SNAPLOCK' 'SPIKED', ARGUMENTS such as BULLSEYE, BRAA, numbers or digits.", callsign)
	if len(locations) > 0 {
		s += " Locations: " + strings.Join(locations, ", ") + "."
	}
	return s
}
