package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewGUID tests that the NewGUID function returns a GUID of the correct length.
// Note: This test is not strictly deterministic.
func TestNewGUID(t *testing.T) {
	t.Parallel()
	for range 9999 {
		g := NewGUID()
		assert.Len(t, []byte(g), GUIDLength)
	}
}
