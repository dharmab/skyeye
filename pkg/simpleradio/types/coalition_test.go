package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIsSpectator tests the IsSpectator function with valid and nonsense Coalition IDs.
func TestIsSpectator(t *testing.T) {
	require.True(t, IsSpectator(0))
	require.False(t, IsSpectator(CoalitionBlue))
	require.False(t, IsSpectator(CoalitionRed))
	for i := 3; i < 1024; i++ {
		require.True(t, IsSpectator(Coalition(i)))
	}
}
