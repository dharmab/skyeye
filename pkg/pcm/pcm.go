// package pcm converts between different representations of PCM audio data.
// Ideally the only representations we would need would be []float32 for F32LE and []int16 for S16LE.
// Sadly, many modules require us to provide raw byte arrays, so we also need conversion functions for []byte.
package pcm

import (
	"encoding/binary"
	"math"
)

// F32ToS16 converts a float32 in range -1, 1 to an int16 in range -32768, 32767.
func F32ToS16(f float32) int16 {
	return int16(f * math.MaxInt16)
}

// S16ToF32 converts an int16 in range -32768, 32767 to a float32 in range -1, 1.
func S16ToF32(s int16) float32 {
	return float32(s) / math.MaxInt16
}

// F32toS16LE converts a slice of float32 to a slice of int16. This is useful for converting from F32LE to S16LE.
func F32toS16LE(in []float32) []int16 {
	out := make([]int16, 0)
	for _, f := range in {
		s := F32ToS16(f)
		out = append(out, s)
	}
	return out
}

// F32toS16LEBytes converts a slice of float32 to a slice of bytes. This is useful for converting from F32LE to S16LE.
func F32toS16LEBytes(in []float32) []byte {
	out := make([]byte, 0)
	for _, f := range in {
		s := F32ToS16(f)
		out = binary.LittleEndian.AppendUint16(out, uint16(s))
	}
	return out
}

// S16LEtoF32 converts a slice of int16 bytes to a slice of float32 bytes. This is useful for converting from S16LE to F32LE.
func F32LEBytesToS16LEBytes(in []byte) []byte {
	out := make([]byte, 0)
	for i := 0; i < len(in); i += 4 {
		f := math.Float32frombits(binary.LittleEndian.Uint32(in[i : i+4]))
		s := F32ToS16(f)
		out = binary.LittleEndian.AppendUint16(out, uint16(s))
	}
	return out
}

// S16LEToF32LE converts a slice of int16 to a slice of float32. This is useful for converting from S16LE to F32LE.
func S16LEToF32LE(in []int16) []float32 {
	out := make([]float32, 0)
	for _, s := range in {
		f := S16ToF32(s)
		out = append(out, f)
	}
	return out
}

// S16LEBytesToF32 converts a slice of bytes to a slice of float32. This is useful for converting from S16LE to F32LE.
func S16LEBytesToF32LE(in []byte) []float32 {
	out := make([]float32, 0)
	for i := 0; i < len(in); i += 2 {
		u := binary.LittleEndian.Uint16(in[i : i+2])
		s := int16(u)
		f := S16ToF32(s)
		out = append(out, f)
	}
	return out
}
