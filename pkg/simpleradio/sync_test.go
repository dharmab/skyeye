package simpleradio

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSelfGUID = types.GUID("SelfGUID00000000000000")

var (
	testRadio      = types.Radio{Frequency: 251_000_000, Modulation: types.ModulationAM}
	testOtherRadio = types.Radio{Frequency: 133_000_000, Modulation: types.ModulationAM}
)

// newSyncTestClient builds a blue client listening on 251.000AM, without touching the network.
func newSyncTestClient() *Client {
	return &Client{
		clientInfo: types.ClientInfo{
			GUID:      testSelfGUID,
			Name:      "GCI Test [BOT]",
			Coalition: coalitions.Blue,
			RadioInfo: types.RadioInfo{Radios: []types.Radio{testRadio}},
		},
		clients: make(map[types.GUID]types.ClientInfo),
	}
}

func testPeer(guid types.GUID, name string, coalition coalitions.Coalition, radios ...types.Radio) types.ClientInfo {
	return types.ClientInfo{
		GUID:      guid,
		Name:      name,
		Coalition: coalition,
		RadioInfo: types.RadioInfo{Radios: radios},
	}
}

func TestSyncClient(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		peer     types.ClientInfo
		expected bool
	}{
		{
			name:     "same coalition on frequency",
			peer:     testPeer("peer00000000000000000a", "Eagle 1", coalitions.Blue, testRadio),
			expected: true,
		},
		{
			name:     "opposing coalition",
			peer:     testPeer("peer00000000000000000b", "Bandit 1", coalitions.Red, testRadio),
			expected: false,
		},
		{
			// Spectators are admitted regardless of coalition so that they can talk to the GCI.
			name:     "spectator",
			peer:     testPeer("peer00000000000000000c", "Spectator", coalitions.Neutrals, testRadio),
			expected: true,
		},
		{
			name:     "same coalition off frequency",
			peer:     testPeer("peer00000000000000000d", "Eagle 2", coalitions.Blue, testOtherRadio),
			expected: false,
		},
		{
			// Observed on a live server: ATIS clients appear with no radios at all.
			name:     "no radios",
			peer:     testPeer("peer00000000000000000e", "ATIS Incirlik", coalitions.Blue),
			expected: false,
		},
		{
			name:     "self",
			peer:     testPeer(testSelfGUID, "GCI Test [BOT]", coalitions.Blue, testRadio),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			c := newSyncTestClient()
			c.syncClient(test.peer)

			_, ok := c.clients[test.peer.GUID]
			assert.Equal(t, test.expected, ok)
		})
	}
}

// A peer that retunes away from our frequency must stop being tracked, otherwise it would keep
// counting towards the on-frequency totals that gate broadcasts.
func TestSyncClientForgetsPeerThatRetunes(t *testing.T) {
	t.Parallel()
	c := newSyncTestClient()
	guid := types.GUID("peer000000000000000001")

	c.syncClient(testPeer(guid, "Eagle 1", coalitions.Blue, testRadio))
	require.Contains(t, c.clients, guid)

	c.syncClient(testPeer(guid, "Eagle 1", coalitions.Blue, testOtherRadio))
	assert.NotContains(t, c.clients, guid)
}

func TestSyncClients(t *testing.T) {
	t.Parallel()
	c := newSyncTestClient()
	c.syncClients([]types.ClientInfo{
		testPeer("peer000000000000000001", "Eagle 1", coalitions.Blue, testRadio),
		testPeer("peer000000000000000002", "Bandit 1", coalitions.Red, testRadio),
		testPeer("peer000000000000000003", "Eagle 2", coalitions.Blue, testRadio),
	})

	assert.Len(t, c.clients, 2)
}

func TestRemoveClient(t *testing.T) {
	t.Parallel()
	c := newSyncTestClient()
	peer := testPeer("peer000000000000000001", "Eagle 1", coalitions.Blue, testRadio)

	c.syncClient(peer)
	require.Contains(t, c.clients, peer.GUID)

	c.removeClient(peer)
	assert.NotContains(t, c.clients, peer.GUID)

	// Removing a peer that was never tracked is harmless.
	c.removeClient(testPeer("peer000000000000000009", "Ghost", coalitions.Blue, testRadio))
}
