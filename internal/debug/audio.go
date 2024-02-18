// package debug contains some temporary code to help debug the application. This code will be removed in the future.
package debug

import (
	"bytes"
	"encoding/binary"
	"log/slog"
	"math"
	"time"

	"github.com/ebitengine/oto/v3"
)

// NewOtoContext creates a new oto context, which is used to play audio.
func NewOtoContext() (*oto.Context, error) {
	ctx, readyChan, err := oto.NewContext(&oto.NewContextOptions{
		// hardcoded to prevent import cycle
		SampleRate:   16000,
		ChannelCount: 1,
		Format:       oto.FormatFloat32LE,
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

	slog.Info("playing sample")
	player.Play()
	for player.IsPlaying() {
		time.Sleep(1000 * time.Millisecond)
	}
	slog.Info("done playing sample")
}

// F32toBytes converts a slice of float32 to a slice of bytes. This is useful for converting from F32LE to S16Le.
func F32toBytes(in []float32) []byte {
	out := make([]byte, 0)
	for _, f := range in {
		out = binary.LittleEndian.AppendUint32(out, math.Float32bits(f))
	}
	return out
}
