package recognizer

type recognizerOptions struct {
	locations []string
}

// Option configures a recognizer.
type Option func(*recognizerOptions)

// WithLocations adds location names to the recognizer's initial prompt,
// improving transcription accuracy for place names.
func WithLocations(locations []string) Option {
	return func(o *recognizerOptions) {
		o.locations = locations
	}
}
