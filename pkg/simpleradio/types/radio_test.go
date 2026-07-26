package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRadioIsSameFrequency(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		this     Radio
		other    Radio
		expected bool
	}{
		{
			name:     "identical",
			this:     Radio{Frequency: 251_000_000, Modulation: ModulationAM},
			other:    Radio{Frequency: 251_000_000, Modulation: ModulationAM},
			expected: true,
		},
		{
			// SRS clients report frequencies with a little slop, so a tolerance is intended.
			name:     "within tolerance",
			this:     Radio{Frequency: 251_000_000, Modulation: ModulationAM},
			other:    Radio{Frequency: 251_000_499, Modulation: ModulationAM},
			expected: true,
		},
		{
			name:     "exactly at tolerance",
			this:     Radio{Frequency: 251_000_000, Modulation: ModulationAM},
			other:    Radio{Frequency: 251_000_500, Modulation: ModulationAM},
			expected: true,
		},
		{
			name:     "just outside tolerance",
			this:     Radio{Frequency: 251_000_000, Modulation: ModulationAM},
			other:    Radio{Frequency: 251_000_501, Modulation: ModulationAM},
			expected: false,
		},
		{
			name:     "tolerance applies below too",
			this:     Radio{Frequency: 251_000_000, Modulation: ModulationAM},
			other:    Radio{Frequency: 250_999_500, Modulation: ModulationAM},
			expected: true,
		},
		{
			name:     "different frequency",
			this:     Radio{Frequency: 251_000_000, Modulation: ModulationAM},
			other:    Radio{Frequency: 133_000_000, Modulation: ModulationAM},
			expected: false,
		},
		{
			name:     "same frequency different modulation",
			this:     Radio{Frequency: 30_000_000, Modulation: ModulationAM},
			other:    Radio{Frequency: 30_000_000, Modulation: ModulationFM},
			expected: false,
		},
		{
			name:     "both encrypted with the same key",
			this:     Radio{Frequency: 251_000_000, Modulation: ModulationAM, IsEncrypted: true, EncryptionKey: 3},
			other:    Radio{Frequency: 251_000_000, Modulation: ModulationAM, IsEncrypted: true, EncryptionKey: 3},
			expected: true,
		},
		{
			name:     "both encrypted with different keys",
			this:     Radio{Frequency: 251_000_000, Modulation: ModulationAM, IsEncrypted: true, EncryptionKey: 3},
			other:    Radio{Frequency: 251_000_000, Modulation: ModulationAM, IsEncrypted: true, EncryptionKey: 4},
			expected: false,
		},
		{
			name:     "only one encrypted",
			this:     Radio{Frequency: 251_000_000, Modulation: ModulationAM, IsEncrypted: true, EncryptionKey: 3},
			other:    Radio{Frequency: 251_000_000, Modulation: ModulationAM},
			expected: false,
		},
		{
			// The key is only meaningful when encryption is on.
			name:     "neither encrypted but keys differ",
			this:     Radio{Frequency: 251_000_000, Modulation: ModulationAM, EncryptionKey: 3},
			other:    Radio{Frequency: 251_000_000, Modulation: ModulationAM, EncryptionKey: 4},
			expected: true,
		},
		{
			// Only the primary frequency is matched. A guard frequency is not received.
			name:     "guard frequency is not matched",
			this:     Radio{Frequency: 251_000_000, Modulation: ModulationAM},
			other:    Radio{Frequency: 133_000_000, Modulation: ModulationAM, GuardFrequency: 251_000_000},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, test.expected, test.this.IsSameFrequency(test.other))
			assert.Equal(t, test.expected, test.other.IsSameFrequency(test.this), "must be symmetric")
		})
	}
}
