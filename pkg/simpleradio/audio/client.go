// package audio implements the SRS audio client. It is based on the OverlordBot audio client, but with some redesign.
// See also: https://gitlab.com/overlordbot/srs-bot/-/blob/master/OverlordBot.SimpleRadio/Network/AudioClient.cs
package audio

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
	"github.com/rs/zerolog/log"
)

// Audio is a type alias for F32LE PCM data
// TODO turn this into a struct with trace id
type Audio []float32

// AudioClient is an SRS audio client configured to receive and transmit on a specific SRS frequency.
type AudioClient interface {
	// Frequency returns the SRS frequency this client is configured to receive and transmit on in Hz.
	Frequency() float64
	// Run executes the control loops of the SRS audio client. It should be called exactly once. When the context is canceled or if the client encounters a non-recoverable error, the client will close its resources.
	Run(context.Context, *sync.WaitGroup) error
	// Transmit queues the given audio to play on the audio client's SRS frequency.
	Transmit(Audio)
	// Receive returns a channel which receives audio from the audio client's SRS frequency.
	Receive() <-chan Audio
}

// audioClient implements AudioClient.
type audioClient struct {
	// guid is used to identify this client to the SRS server.
	guid types.GUID
	// radio is the SRS radio this client will receive and transmit on.
	radio types.Radio
	// connection is the UDP connection to the SRS server.
	connection *net.UDPConn // todo move connection mgmt into Run()
	// rxChan is a channel where received audio is published. A read-only version is available publicly.
	rxchan chan Audio
	// txChan is a channel where audio to be transmitted is buffered.
	txChan chan Audio

	// lastRx tracks the last received audio packet so we can tell when a transmission has (probably) ended.
	lastRx rxState
	// packetNumber is incremented for each voice packet transmitted.
	packetNumber uint64

	// busy indicates if there is a transmission in progress.
	busy sync.Mutex
}

// rxState contains the state of the current received transmission.
type rxState struct {
	// origin is the GUID of a client we are currently listening to. We can only listen to one client at a time, and whoever started broadcasting first wins.
	origin types.GUID
	// deadline is extended every time another voice packet is received. When we pass the deadline, the transmission is considered over.
	deadline time.Time
	// packetNumber is the number of the last received voice packet. We only record a packet if its packet number is larger than the last received packet's, and skip any that were dropped or delivered out of order.
	// If we were more ambitious we would reassemble the packets and use Opus's forward error correction to recover from lost packets... too bad!
	packetNumber uint64
}

func NewClient(guid types.GUID, config types.ClientConfiguration) (AudioClient, error) {
	log.Info().Str("protocol", "udp").Str("address", config.Address).Msg("connecting to SRS server")
	address, err := net.ResolveUDPAddr("udp", config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve SRS server address %v: %w", config.Address, err)
	}
	connection, err := net.DialUDP("udp", nil, address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SRS server %v over UDP: %w", config.Address, err)
	}
	return &audioClient{
		guid:         guid,
		radio:        config.Radio,
		connection:   connection,
		txChan:       make(chan Audio),
		rxchan:       make(chan Audio),
		lastRx:       rxState{},
		packetNumber: 1,
		busy:         sync.Mutex{},
	}, nil
}

// Frequency implements AudioClient.Frequency
func (c *audioClient) Frequency() float64 {
	return c.radio.Frequency
}

// Run implements AudioClient.Run
func (c *audioClient) Run(ctx context.Context, wg *sync.WaitGroup) error {
	defer func() {
		if err := c.close(); err != nil {
			log.Error().Err(err).Msg("error closing SRS client")
		}
	}()
	defer c.close()

	// We need to send pings to the server to keep our connection alive. The server won't send us any audio until it receives a ping from us.
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.sendPings(ctx, wg)
	}()

	// udpPingRxChan is a channel for received ping packets.
	udpPingRxChan := make(chan []byte, 0xF)

	// Handle incoming pings - mostly for debugging. We don't need to echo them back.
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.receivePings(ctx, udpPingRxChan)
	}()

	// udpVoiceRxChan is a channel for received voice packets.
	udpVoiceRxChan := make(chan []byte, 64*0xFFFFF) // TODO configurable packet buffer size
	// voiceBytesRxChan is a channel for VoicePackets deserialized from UDP voice packets.
	voiceBytesRxChan := make(chan []voice.VoicePacket, 0xFFFFF) // TODO configurable tranmission buffer size

	// receive voice packets and decode them. This is the logic for receiving audio from the SRS server.
	wg.Add(2)
	go func() {
		defer wg.Done()
		c.receiveVoice(ctx, udpVoiceRxChan, voiceBytesRxChan)
	}()
	go func() {
		defer wg.Done()
		c.decodeVoice(ctx, voiceBytesRxChan)
	}()

	// voicePacketsTxChan is a channel for transmissions which are ready to send.
	voicePacketsTxChan := make(chan []voice.VoicePacket, 3)

	// transmit queued audio. This is the logic for sending audio to the SRS server.
	wg.Add(2)
	go func() {
		defer wg.Done()
		c.encodeVoice(ctx, voicePacketsTxChan)
	}()
	go func() {
		defer wg.Done()
		c.transmit(ctx, voicePacketsTxChan)
	}()

	// Start listening for incoming UDP packets and routing them to receivePings and receiveVoice.
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.receiveUDP(ctx, udpPingRxChan, udpVoiceRxChan)
	}()

	// Sit and wait, until the context is canceled.
	<-ctx.Done()
	return nil
}

// Receive implements AudioClient.Receive
func (c *audioClient) Receive() <-chan Audio {
	return c.rxchan
}

// Transmit implements AudioClient.Transmit
func (c *audioClient) Transmit(sample Audio) {
	c.txChan <- sample
}

// close closes the UDP connection to the SRS server. This might be nonsensical because UDP is connectionless. \_(ツ)_/¯
func (c *audioClient) close() error {
	if err := c.connection.Close(); err != nil {
		return fmt.Errorf("error closing UDP connection to SRS: %w", err)
	}
	return nil
}
