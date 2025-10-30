package pcm

import "math"

// DecibelsToGain converts a decibel value to a linear gain value.
func DecibelsToGain(db float64) float64 {
	return math.Pow(10, db/20)
}

// GainToDecibels converts a linear gain value to a decibel value.
func GainToDecibels(gain float64) float64 {
	return 20 * math.Log10(gain)
}

// F32LEGain applies a linear gain to an F32LE PCM sample.
func F32LEGain(in []float32, gain float64) []float32 {
	if gain == 1.0 {
		return in
	}
	out := make([]float32, len(in))
	for i, s := range in {
		out[i] = float32(float64(s) * gain)
	}
	return out
}

// S16LEGain applies a linear gain to an S16LE PCM sample.
func S16LEGain(in []int16, gain float64) []int16 {
	out := make([]int16, len(in))
	for i, s := range in {
		v := float64(s) * gain
		if v > math.MaxInt16 {
			v = math.MaxInt16
		} else if v < math.MinInt16 {
			v = math.MinInt16
		}
		out[i] = int16(v)
	}
	return out
}
