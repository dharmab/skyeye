// Package voices defines the available voices for the synthesizer package.
// This package is split from speakers to avoid pulling C dependencies into
// half of SkyEye's unit tests.
package voices

// Voice for text-to-speech synthesis.
type Voice int

const (
	// FeminineVoice is the "Jenny" en-GB voice.
	// Origin: https://github.com/dioco-group/jenny-tts-dataset
	FeminineVoice Voice = iota
	// MasculineVoice is the "Alan" en-GB voice.
	// Origin: https://popey.me
	MasculineVoice
)
