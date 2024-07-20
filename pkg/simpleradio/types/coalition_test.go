package types

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/stretchr/testify/require"
)

// TestIsSpectator tests the IsSpectator function with valid and nonsense Coalition IDs.
func TestIsSpectator(t *testing.T) {
	require.True(t, IsSpectator(coalitions.Neutrals))
	require.False(t, IsSpectator(coalitions.Red))
	require.False(t, IsSpectator(coalitions.Blue))
}
