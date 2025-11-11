package types

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/stretchr/testify/assert"
)

func TestIsSpectator(t *testing.T) {
	t.Parallel()
	assert.True(t, IsSpectator(coalitions.Neutrals))
	assert.False(t, IsSpectator(coalitions.Red))
	assert.False(t, IsSpectator(coalitions.Blue))
}
