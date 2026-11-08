//go:build darwin && cgo

package recognizer

/*
#cgo CFLAGS: -fblocks
#cgo LDFLAGS: -framework Foundation -framework Speech -framework AVFoundation
#include <stdlib.h>
#include "apple_native.h"
*/
import "C"

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"unsafe"

	"github.com/dharmab/skyeye/pkg/pcm/rate"
	"github.com/rs/zerolog/log"
)

type appleSpeechRecognizer struct{ hints string }

var _ Recognizer = (*appleSpeechRecognizer)(nil)

// NewAppleSpeechRecognizer creates a local-only en-US buffer recognizer.
// It checks configuration without requesting permission. The first valid Recognize
// call requests Speech authorization asynchronously. The executable must embed a
// nonempty NSSpeechRecognitionUsageDescription in its Info.plist.
func NewAppleSpeechRecognizer(callsign string, opts ...Option) (Recognizer, error) {
	if message := C.apple_preflight(); message != nil {
		defer C.free(unsafe.Pointer(message))
		return nil, errors.New(C.GoString(message))
	}
	hints, err := json.Marshal(appleContextualStrings(callsign, opts...))
	if err != nil {
		return nil, err
	}
	return &appleSpeechRecognizer{hints: string(hints)}, nil
}

func (r *appleSpeechRecognizer) Recognize(ctx context.Context, pcm []float32, enableTranscriptionLogging bool) (string, error) {
	if err := appleValidate(ctx, pcm); err != nil {
		return "", err
	}
	hints := C.CString(r.hints)
	defer C.free(unsafe.Pointer(hints))
	// Native code copies PCM and hints before returning; no Go pointers escape.
	handle := C.apple_begin((*C.float)(unsafe.Pointer(&pcm[0])), C.size_t(len(pcm)), C.double(rate.Wideband.Hertz()), hints)
	defer C.apple_release(handle)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		var text, message *C.char
		if C.apple_poll(handle, &text, &message) != 0 {
			if message != nil {
				err := errors.New(C.GoString(message))
				C.free(unsafe.Pointer(message))
				return "", err
			}
			result := C.GoString(text)
			C.free(unsafe.Pointer(text))
			if enableTranscriptionLogging {
				log.Debug().Str("text", result).Msg("Apple Speech transcription")
			}
			return result, nil
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
		}
	}
}
