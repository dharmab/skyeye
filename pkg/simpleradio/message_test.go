package simpleradio

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SRS servers are inconsistent about the case of boolean settings - a live server sends both
// "true" and "True" in the same payload.
func TestUpdateServerSettings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		settings map[string]string
		initial  bool
		expected bool
	}{
		{
			name:     "enabled lowercase",
			settings: map[string]string{string(types.CoalitionAudioSecurity): "true"},
			expected: true,
		},
		{
			name:     "enabled titlecase",
			settings: map[string]string{string(types.CoalitionAudioSecurity): "True"},
			expected: true,
		},
		{
			name:     "disabled",
			settings: map[string]string{string(types.CoalitionAudioSecurity): "false"},
			initial:  true,
			expected: false,
		},
		{
			name:     "disabled titlecase",
			settings: map[string]string{string(types.CoalitionAudioSecurity): "False"},
			initial:  true,
			expected: false,
		},
		{
			name:     "absent leaves the setting alone",
			settings: map[string]string{string(types.ExternalAWACSMode): "True"},
			initial:  true,
			expected: true,
		},
		{
			name:     "no settings at all",
			settings: nil,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			c := &Client{}
			c.secureCoalitionRadios.Store(test.initial)

			c.updateServerSettings(types.Message{ServerSettings: test.settings})

			assert.Equal(t, test.expected, c.secureCoalitionRadios.Load())
		})
	}
}

func TestHandleMessage(t *testing.T) {
	t.Parallel()
	peer := testPeer("peer000000000000000001", "Eagle 1", coalitions.Blue, testRadio)

	t.Run("sync stores matching clients", func(t *testing.T) {
		t.Parallel()
		c := newSyncTestClient()
		c.handleMessage(types.Message{Type: types.MessageSync, Clients: []types.ClientInfo{peer}})
		assert.Contains(t, c.clients, peer.GUID)
	})

	t.Run("update stores a client", func(t *testing.T) {
		t.Parallel()
		c := newSyncTestClient()
		c.handleMessage(types.Message{Type: types.MessageUpdate, Client: &peer})
		assert.Contains(t, c.clients, peer.GUID)
	})

	t.Run("radio update stores a client", func(t *testing.T) {
		t.Parallel()
		c := newSyncTestClient()
		c.handleMessage(types.Message{Type: types.MessageRadioUpdate, Client: &peer})
		assert.Contains(t, c.clients, peer.GUID)
	})

	t.Run("disconnect removes a client", func(t *testing.T) {
		t.Parallel()
		c := newSyncTestClient()
		c.syncClient(peer)
		require.Contains(t, c.clients, peer.GUID)

		c.handleMessage(types.Message{Type: types.MessageClientDisconnect, Client: &peer})
		assert.NotContains(t, c.clients, peer.GUID)
	})

	t.Run("server settings are applied", func(t *testing.T) {
		t.Parallel()
		c := newSyncTestClient()
		c.handleMessage(types.Message{
			Type:           types.MessageServerSettings,
			ServerSettings: map[string]string{string(types.CoalitionAudioSecurity): "true"},
		})
		assert.True(t, c.secureCoalitionRadios.Load())
	})

	// Messages that carry no client must not panic. These arrive in practice - a ping message has
	// no Client field at all.
	t.Run("messages without a client are ignored", func(t *testing.T) {
		t.Parallel()
		c := newSyncTestClient()
		for _, messageType := range []types.MessageType{
			types.MessagePing,
			types.MessageUpdate,
			types.MessageRadioUpdate,
			types.MessageClientDisconnect,
			types.MessageVersionMismatch,
			types.MessageExternalAWACSModeDisconnect,
			types.MessageExternalAWACSModePassword,
			types.MessageType(999),
		} {
			c.handleMessage(types.Message{Type: messageType})
		}
		assert.Empty(t, c.clients)
	})
}
