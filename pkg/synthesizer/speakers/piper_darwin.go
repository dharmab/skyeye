//go:build darwin

package speakers

import (
	"time"

	"github.com/dharmab/skyeye/pkg/synthesizer/voices"
)

func NewPiperSpeaker(_ voices.Voice, _ float64, _ time.Duration) (Speaker, error) {
	panic("unreachable: piper is not used on macOS")
}
