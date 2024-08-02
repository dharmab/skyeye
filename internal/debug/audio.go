// package debug contains some temporary code to help debug the application. This code will be removed in the future.
package debug

import (
	"bytes"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/rs/zerolog/log"
)

// NewOtoContext creates a new oto context, which is used to play audio.
func NewOtoContext() (*oto.Context, error) {
	ctx, readyChan, err := oto.NewContext(&oto.NewContextOptions{
		// hardcoded to prevent import cycle
		SampleRate:   16000,
		ChannelCount: 1,
		Format:       oto.FormatSignedInt16LE,
	})
	<-readyChan
	return ctx, err
}

// MustNewOtoContext is like NewOtoContext, but panics if an error occurs.
func MustNewOtoContext() *oto.Context {
	ctx, err := NewOtoContext()
	if err != nil {
		panic(err)
	}
	return ctx
}

// PlayAudio plays the given PCM audio data using the given oto context. The PCM data should be in S16LE format.
func PlayAudio(ctx *oto.Context, pcm []byte) {
	player := ctx.NewPlayer(bytes.NewReader(pcm))
	defer player.Close()

	log.Debug().Int("length", len(pcm)).Msg("playing sample")
	player.Play()
	for player.IsPlaying() {
		time.Sleep(1000 * time.Millisecond)
	}
	log.Debug().Msg("done playing sample")
}
