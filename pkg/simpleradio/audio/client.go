// package audio implements the SRS audio client. It is based on the OverlordBot audio client, but with some redesign.
// See also: https://gitlab.com/overlordbot/srs-bot/-/blob/master/OverlordBot.SimpleRadio/Network/AudioClient.cs
package audio

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
)

// Audio is a type alias for F32LE PCM data
// TODO turn this into a struct with trace id
type Audio []float32

// AudioClient is an SRS audio client configured to receive and transmit on a specific SRS frequency.
type AudioClient interface {
	// Run executes the control loops of the SRS audio client. It should be called exactly once. When the context is canceled or if the client encounters a non-recoverable error, the client will close its resources.
	Run(context.Context) error
	// Transmit plays the given audio on the audio client's SRS frequency.
	Transmit(Audio) error
	// Receive returns a channel which receives audio from the audio client's SRS frequency.
	Receive() <-chan Audio
}

// audioClient implements [AudioClient]
type audioClient struct {
	// guid is used to identify this client to the SRS server.
	guid types.GUID
	// radio is the SRS radio this client will receive and transmit on.
	radio types.Radio
	// connection is the UDP connection to the SRS server.
	connection *net.UDPConn // todo move connection mgmt into Run()
	// rxChan is a channel where received audio is published. A read-only version is available publicly.
	rxchan chan Audio
	// txChan is a channel where audio to be transmitted is bufffered.
	txChan chan Audio

	// lastRx is used to track the last received audio packet so we can tell when a transmission has (probably) ended.
	lastRx rxState
}

type rxState struct {
	origin       types.GUID
	deadline     time.Time
	packetNumber uint64
}

func NewClient(guid types.GUID, config types.ClientConfiguration) (AudioClient, error) {
	slog.Info("connecting to SRS server", "protocol", "udp", "address", config.Address)
	address, err := net.ResolveUDPAddr("udp", config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve SRS server address %v: %w", config.Address, err)
	}
	connection, err := net.DialUDP("udp", nil, address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SRS server %v over UDP: %w", config.Address, err)
	}
	return &audioClient{
		guid:       guid,
		radio:      config.Radio,
		connection: connection,
		txChan:     make(chan Audio),
		rxchan:     make(chan Audio),
		lastRx:     rxState{},
	}, nil
}

// Run implements AudioClient.Run
func (c *audioClient) Run(ctx context.Context) error {
	defer func() {
		if err := c.close(); err != nil {
			slog.Error("error closing SRS client", "error", err)
		}
	}()

	// We need to send pings to the server to keep our connection alive. The server won't send us any audio until it receives a ping from us.
	go c.sendPings(ctx)

	// udpPingRxChan is a channel for received ping packets.
	udpPingRxChan := make(chan []byte, 0xF)

	// Handle incoming pings - mostly for debugging. We don't need to echo them back.
	go c.receivePings(ctx, udpPingRxChan)

	// udpVoiceRxChan is a channel for received voice packets.
	udpVoiceRxChan := make(chan []byte, 64*0xFFFFF) // TODO configurable packet buffer size
	// voiceBytesChan is a channel for VoicePacket structs decoded from UDP voice packets.
	voiceBytesChan := make(chan []voice.VoicePacket, 0xFFFFF) // TODO configurable tranmission buffer size

	// receive voice packets and decode them. This is the logic for receiving audio from the SRS server.
	go c.receiveVoice(ctx, udpVoiceRxChan, voiceBytesChan)
	go c.decodeVoice(ctx, voiceBytesChan)

	// Start listening for incoming UDP packets and routing them to receivePings and receiveVoice.
	go c.receiveUDP(ctx, udpPingRxChan, udpVoiceRxChan)

	// Sit and wait, until the context is canceled.
	<-ctx.Done()
	c.close()
	return nil
}

// Receive implements AudioClient.Receive
func (c *audioClient) Receive() <-chan Audio {
	return c.rxchan
}

// Transmit implements AudioClient.Transmit
func (c *audioClient) Transmit(sample Audio) error {
	return nil
}

// close closes the UDP connection to the SRS server. This might be nonsensical because UDP is connectionless. \_(ツ)_/¯
func (c *audioClient) close() error {
	if err := c.connection.Close(); err != nil {
		return fmt.Errorf("error closing UDP connection to SRS: %w", err)
	}
	return nil
}
