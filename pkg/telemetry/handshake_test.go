package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordHash(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		password  string
		algorithm HashAlgorithm
		expected  string
	}{
		{
			password:  "",
			algorithm: CRC64WE,
			expected:  "0",
		},
		{
			password:  "",
			algorithm: CRC32ISOHDLC,
			expected:  "0",
		},
		{
			password:  "local",
			algorithm: CRC32ISOHDLC,
			expected:  "e528f7a6",
		},
		{
			password:  "local",
			algorithm: CRC64WE,
			expected:  "16d6497244e5930b",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.password, func(t *testing.T) {
			t.Parallel()
			actual := hashPassword(tc.password, tc.algorithm)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
