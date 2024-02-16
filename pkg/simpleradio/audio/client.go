// package audio implements the SRS audio client. It is based on the OverlordBot audio client, but with some redesign.
// See also: https://gitlab.com/overlordbot/srs-bot/-/blob/master/OverlordBot.SimpleRadio/Network/AudioClient.cs
package audio

// https://gitlab.com/overlordbot/srs-bot/-/blob/master/OverlordBot.SimpleRadio/Network/AudioClient.cs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
)

// Audio is a type alias for F32LE PCM data
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

// audioClient implements w
type audioClient struct {
	guid   string
	radios srs.ClientRadios

	connection *net.UDPConn
	// lastReceived is the most recent time audio was received. This is used to guess when a transmission is complete.
	lastReceived time.Time
	// TODO channel for received audio
	audioRxQueue chan Audio
	audioTxQueue chan Audio

	rxOverDuration time.Duration
}

func NewClient(config srs.ClientConfiguration, radios srs.ClientRadios) (AudioClient, error) {
	address, err := net.ResolveUDPAddr("udp", config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve SRS server address %v: %w", config.Address, err)
	}
	connection, err := net.DialUDP("udp", nil, address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SRS server %v over UDP: %w", config.Address, err)
	}
	return &audioClient{
		guid:         config.GUID,
		connection:   connection,
		radios:       radios,
		lastReceived: time.Now(),
		audioTxQueue: make(chan Audio),
		audioRxQueue: make(chan Audio),
	}, nil
}

const PingInterval = 15 * time.Second

// ping is a loop which sends the client GUID to the server every 15 seconds to keep our connection alive
//
// https://gitlab.com/overlordbot/srs-bot/-/blob/master/OverlordBot.SimpleRadio/Network/AudioClient.cs
func (c *audioClient) ping(ctx context.Context) {
	ticker := time.NewTicker(PingInterval)
	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping pings due to context cancelation")
			return
		case <-ticker.C:
			slog.Debug("sending UDP ping")
			n, err := c.connection.Write([]byte(c.guid))
			if err != nil {
				slog.Error("error writing ping", "error", err)
			}
			slog.Debug("sent UDP ping", "bytes", n)
		}
	}
}

// receive is a loop which attempts to receive audio packets from the SRS server. It buffers the data.
//
// https://gitlab.com/overlordbot/srs-bot/-/blob/master/OverlordBot.SimpleRadio/Network/AudioClient.cs
func (c *audioClient) receivePackets(ctx context.Context, ch chan<- VoicePacket) {
	for {
		if ctx.Err() != nil {
			slog.Error("stopping packet receiver due to context error", "error", ctx.Err())
			return
		}

		buf := make([]byte, 1500)
		n, err := c.connection.Read(buf)
		switch {
		case n == 0:
			slog.Debug("0 bytes read from UDP connection", "error", err)
		case err == io.EOF:
			// no op?
		case err != nil:
			slog.Warn("error reading from UDP connection", "error", err)
		case n < 22:
			slog.Debug("UDP packet smaller than expected", "bytes", n)
		case n == 22:
			slog.Debug("received UDP ping")
		case n > 22:
			ch <- newVoicePacketFrom(buf)
		}
	}
}

func (c *audioClient) decodePackets(ctx context.Context, ch <-chan VoicePacket) {
	decodeTicker := time.NewTicker(17 * time.Millisecond)
	defer decodeTicker.Stop()

	buf := new(bytes.Buffer)
	deadline := time.Now()

	for {
		select {
		case <-ctx.Done():
			slog.Info("stopping audio client due to context cancelation", "error", ctx.Err())
			return
		case p := <-ch:
			slog.Debug(
				"received packet",
				"length", p.PacketLength,
				"audio_segment_length", p.AudioSegmentLength,
				"audio_length", p.AudioLength,
				"unit_id", p.UnitID,
				"packet_id", p.PacketID,
				"retransmission_count", p.RetransmissionCount,
				"original_guid", string(p.OriginalGUID),
				"guid", string(p.GUID),
			)
			deadline = time.Now().Add(c.rxOverDuration)
			_, err := buf.Write(p.AudioBytes)
			if err != nil {
				// HOW DID YOU GET HERE
				slog.Error("failed to buffer received packet", "error", err)
			}
			// https://github.com/ciribob/DCS-SimpleRadioStandalone/blob/master/DCS-SR-Client/Network/UDPVoiceHandler.cs
		case <-decodeTicker.C:
			if buf.Len() > 0 && time.Now().After(deadline) {
				slog.Debug("decoding buffered audio", "buffer_length", buf.Len())
				b := make([]byte, buf.Len())
				n, err := buf.Read(b)
				if err != nil {
					slog.Error("error reading from rx buffer", "error", err, "len", buf.Len())
				} else {
					audio, err := decode(b)
					if err != nil {
						slog.Error("error decoding buffered audio", "error", err, "len", n)
					}
					c.audioRxQueue <- audio
				}
				buf.Reset()
			}
		}
	}
}

func (c *audioClient) Run(ctx context.Context) error {
	defer func() {
		if err := c.close(); err != nil {
			slog.Error("error closing SRS client", "error", err)
		}
	}()

	go c.ping(ctx)
	packetChan := make(chan VoicePacket)
	go c.receivePackets(ctx, packetChan)
	go c.decodePackets(ctx, packetChan)

	<-ctx.Done()
	return nil
	// https://gitlab.com/overlordbot/srs-bot/-/blame/master/OverlordBot.SimpleRadio/Network/AudioClient.cs?ref_type=heads#L88
}

func (c *audioClient) Receive() <-chan Audio {
	return c.audioRxQueue
}

func (c *audioClient) Transmit(sample Audio) error {
	return nil
}

func (c *audioClient) close() error {
	if err := c.connection.Close(); err != nil {
		return fmt.Errorf("error closing UDP connection to SRS: %w", err)
	}
	return nil
}
