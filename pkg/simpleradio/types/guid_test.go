package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNewGUID tests that the NewGUID function returns a GUID of the correct length.
// Note: This test is not strictly deterministic.
func TestNewGUID(t *testing.T) {
	for range 9999 {
		g := NewGUID()
		require.Equal(
			t,
			GUIDLength, len([]byte(g)),
			"GUID %v is not %v bytes in length", g, GUIDLength,
		)
	}
}
