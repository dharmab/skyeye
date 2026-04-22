package simpleradio

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/stretchr/testify/assert"
)

func TestUpdateServerSettingsSecureCoalitionRadios(t *testing.T) {
	t.Parallel()

	t.Run("enabled", func(t *testing.T) {
		t.Parallel()
		client := &Client{}

		client.updateServerSettings(types.Message{
			ServerSettings: map[string]string{
				string(types.CoalitionAudioSecurity): "TrUe",
			},
		})

		assert.True(t, client.secureCoalitionRadios.Load())
	})

	t.Run("disabled", func(t *testing.T) {
		t.Parallel()
		client := &Client{}
		client.secureCoalitionRadios.Store(true)

		client.updateServerSettings(types.Message{
			ServerSettings: map[string]string{
				string(types.CoalitionAudioSecurity): "false",
			},
		})

		assert.False(t, client.secureCoalitionRadios.Load())
	})
}
