package recognizer

import "fmt"

// prompt is a prompt for OpenAI's Whisper model. See https://platform.openai.com/docs/guides/speech-to-text#prompting
func prompt(callsign string) string {
	return fmt.Sprintf("Either ANYFACE or %s, PILOT CALLSIGN, DIGITS, one of 'RADIO' 'ALPHA' 'BOGEY' 'PICTURE' 'DECLARE' 'SNAPLOCK' 'SPIKED', ARGUMENTS such as BULLSEYE, BRAA, numbers or digits.", callsign)
}
