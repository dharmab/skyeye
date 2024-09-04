package types

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/stretchr/testify/require"
)

func TestIsSpectator(t *testing.T) {
	t.Parallel()
	require.True(t, IsSpectator(coalitions.Neutrals))
	require.False(t, IsSpectator(coalitions.Red))
	require.False(t, IsSpectator(coalitions.Blue))
}
