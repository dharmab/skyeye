package voices

type Voice int

const (
	// FeminineVoice is the "Jenny" en-GB voice.
	// Origin: https://github.com/dioco-group/jenny-tts-dataset
	FeminineVoice Voice = iota
	// MasculineVoice is the "Alan" en-GB voice.
	// Origin: https://popey.me
	MasculineVoice
)
