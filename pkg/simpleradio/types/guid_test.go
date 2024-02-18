package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
