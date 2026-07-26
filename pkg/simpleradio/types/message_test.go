package types

import (
	"encoding/json"
	"testing"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// syncFixture is a sync message captured from a live SRS server, with client names and GUIDs
// replaced. It carries fields SkyEye does not model (Model, Name, Gateway, DISEntityId) on purpose,
// so that the test fails if unknown fields ever stop being tolerated.
const syncFixture = `{
 "Version": "2.1.0.2",
 "Clients": [
  {
   "ClientGuid": "0000000000000000000000",
   "Name": "Example Client",
   "Coalition": 2,
   "AllowRecord": true,
   "Seat": 0,
   "RadioInfo": {
    "ambient": {"abType": "", "vol": 1},
    "iff": {"control": 2, "mic": -1, "mode1": -1, "mode2": -1, "mode3": -1, "mode4": false, "status": 0},
    "radios": [
     {"Model": "", "Name": "", "enc": false, "encKey": 0, "freq": 136000000, "modulation": 0, "retransmit": true, "secFreq": 0},
     {"Model": "", "Name": "", "enc": false, "encKey": 0, "freq": 255000000, "modulation": 0, "retransmit": true, "secFreq": 0}
    ],
    "unit": "External AWACS",
    "unitId": 100000002
   },
   "LatLngPosition": {"alt": 0, "lat": 0, "lng": 0},
   "Gateway": false,
   "DISEntityId": -1
  },
  {
   "ClientGuid": "1111111111111111111111",
   "Name": "ATIS Example",
   "Coalition": 2,
   "AllowRecord": true,
   "Seat": 0
  }
 ],
 "MsgType": 2
}`

func TestUnmarshalSyncMessage(t *testing.T) {
	t.Parallel()
	var message Message
	require.NoError(t, json.Unmarshal([]byte(syncFixture), &message))

	assert.Equal(t, "2.1.0.2", message.Version)
	assert.Equal(t, MessageSync, message.Type)
	require.Len(t, message.Clients, 2)

	client := message.Clients[0]
	assert.Equal(t, GUID("0000000000000000000000"), client.GUID)
	assert.Equal(t, "Example Client", client.Name)
	assert.Equal(t, coalitions.Coalition(coalitions.Blue), client.Coalition)
	assert.True(t, client.AllowRecording)
	assert.Equal(t, "External AWACS", client.RadioInfo.Unit)
	assert.Equal(t, uint64(100000002), client.RadioInfo.UnitID)
	require.Len(t, client.RadioInfo.Radios, 2)
	assert.InDelta(t, 136_000_000.0, client.RadioInfo.Radios[0].Frequency, 0.5)
	assert.True(t, client.RadioInfo.Radios[0].ShouldRetransmit)
	require.NotNil(t, client.Position)

	// A client with no RadioInfo at all is normal - ATIS clients appear this way.
	assert.Empty(t, message.Clients[1].RadioInfo.Radios)
}

func TestMessageRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		message Message
	}{
		{
			name:    "ping carries only a version and type",
			message: Message{Version: "2.1.0.2", Type: MessagePing},
		},
		{
			name: "update carries one client",
			message: Message{
				Version: "2.1.0.2",
				Type:    MessageUpdate,
				Client: &ClientInfo{
					GUID:      "0000000000000000000000",
					Name:      "Example Client",
					Coalition: coalitions.Blue,
					RadioInfo: RadioInfo{
						Radios:  []Radio{{Frequency: 251_000_000, Modulation: ModulationAM}},
						Unit:    "External AWACS",
						UnitID:  100000002,
						IFF:     NewIFF(),
						Ambient: NewAmbient(),
					},
					Position: &Position{},
				},
			},
		},
		{
			name: "external AWACS mode password",
			message: Message{
				Version:                   "2.1.0.2",
				Type:                      MessageExternalAWACSModePassword,
				ExternalAWACSModePassword: "hunter2",
			},
		},
		{
			name: "server settings",
			message: Message{
				Version:        "2.1.0.2",
				Type:           MessageServerSettings,
				ServerSettings: map[string]string{string(CoalitionAudioSecurity): "false"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			encoded, err := json.Marshal(test.message)
			require.NoError(t, err)

			var decoded Message
			require.NoError(t, json.Unmarshal(encoded, &decoded))
			assert.Equal(t, test.message, decoded)
		})
	}
}

// Optional fields must stay absent rather than appearing as nulls, because the SRS server is
// strict about the messages it accepts.
func TestMessageOmitsEmptyFields(t *testing.T) {
	t.Parallel()
	encoded, err := json.Marshal(Message{Version: "2.1.0.2", Type: MessagePing})
	require.NoError(t, err)

	assert.JSONEq(t, `{"Version":"2.1.0.2","MsgType":1}`, string(encoded))
}
