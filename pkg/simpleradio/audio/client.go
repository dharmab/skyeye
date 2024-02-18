// package audio implements the SRS audio client. It is based on the OverlordBot audio client, but with some redesign.
// See also: https://gitlab.com/overlordbot/srs-bot/-/blob/master/OverlordBot.SimpleRadio/Network/AudioClient.cs
package audio

// https://gitlab.com/overlordbot/srs-bot/-/blob/master/OverlordBot.SimpleRadio/Network/AudioClient.cs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/pion/opus"
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

// audioClient implements w
type audioClient struct {
	guid srs.GUID

	connection *net.UDPConn // todo move connection mgmt into Run()
	// lastRxTime is the most recent time audio was received. This is used to guess when a transmission is complete.
	lastRxTime time.Time
	// rxChan is a channel where received audio is published. A read-only version is available publicly.
	rxchan chan Audio
	// txChan is a channel where audio to be transmitted is bufffered.
	txChan chan Audio

	rxDeadline time.Time
}

func NewClient(guid srs.GUID, config srs.ClientConfiguration, radios srs.RadioInfo) (AudioClient, error) {
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
		connection: connection,
		lastRxTime: time.Now(),
		txChan:     make(chan Audio),
		rxchan:     make(chan Audio),
	}, nil
}

const PingInterval = 15 * time.Second

// sendPings is a loop which sends the client GUID to the server every 15 seconds to keep our connection alive
//
// https://gitlab.com/overlordbot/srs-bot/-/blob/master/OverlordBot.SimpleRadio/Network/AudioClient.cs
func (c *audioClient) sendPings(ctx context.Context) {
	slog.Info("starting pings", "interval", PingInterval.String())
	go func() {
		time.Sleep(1 * time.Second)
		c.SendPing()
	}()

	ticker := time.NewTicker(PingInterval)
	for {
		select {
		case <-ticker.C:
			c.SendPing()
		case <-ctx.Done():
			slog.Info("stopping pings due to context cancelation")
			return

		}
	}
}

func (c *audioClient) SendPing() {
	slog.Debug("sending UDP ping", "guid", c.guid)
	n, err := c.connection.Write([]byte(c.guid))
	if errors.Is(err, net.ErrClosed) {
		slog.Warn("ping skipped due to closed connection")
	} else if err != nil {
		slog.Error("error writing ping", "error", err)
	} else if n != srs.GUIDLength {
		slog.Warn("wrote unexpected number of bytes while sending UDP ping", "guid", c.guid, "bytes", n, "expectedBytes", srs.GUIDLength)
	} else {
		slog.Debug("sent UDP ping", "guid", c.guid)
	}
}

func (c *audioClient) Run(ctx context.Context) error {
	defer func() {
		if err := c.close(); err != nil {
			slog.Error("error closing SRS client", "error", err)
		}
	}()

	go c.sendPings(ctx)

	udpPingRxChan := make(chan []byte, 0xF)
	udpVoiceRxChan := make(chan []byte, 64*0xFFFFF) // TODO configurable audio buffer size
	voiceRxChan := make(chan []byte, 0xF)

	go c.receivePings(ctx, udpPingRxChan)
	go c.receiveVoice(ctx, udpVoiceRxChan, voiceRxChan)
	go c.decodeVoice(ctx, voiceRxChan)
	go c.receiveUDP(ctx, udpPingRxChan, udpVoiceRxChan)

	<-ctx.Done()
	return nil
	// https://gitlab.com/overlordbot/srs-bot/-/blame/master/OverlordBot.SimpleRadio/Network/AudioClient.cs?ref_type=heads#L88
}

func (c *audioClient) receiveUDP(ctx context.Context, pingCh chan<- []byte, voiceCh chan<- []byte) {
	for {
		if ctx.Err() != nil {
			slog.Error("stopping packet receiver due to context error", "error", ctx.Err())
			return
		}

		udpPacketBuf := make([]byte, 1500)
		n, err := c.connection.Read(udpPacketBuf)
		udpPacket := make([]byte, n)
		copy(udpPacket, udpPacketBuf[0:n])

		switch {
		case err == io.EOF:
			slog.Error("UDP connection closed?", "error", err)
		case err != nil:
			slog.Error("UDP connection read error", "error", err)
		case n == 0:
			slog.Warn("0 bytes read from UDP connection", "error", err)
		case n < srs.GUIDLength:
			slog.Debug("UDP packet smaller than expected", "bytes", n)
		case n == srs.GUIDLength:
			slog.Debug("routing UDP ping packet", "bytes", n)
			pingCh <- udpPacket
		case n > srs.GUIDLength:
			deadline := time.Now().Add(300 * time.Millisecond)
			slog.Debug("extending transmission receive deadline", "deadline", deadline)
			c.rxDeadline = deadline
			slog.Debug("routing UDP voice packet", "bytes", n)
			voiceCh <- udpPacket
		}
	}
}

func (c *audioClient) receivePings(ctx context.Context, ch <-chan []byte) {
	for {
		select {
		case b := <-ch:
			n := len(b)
			if n < srs.GUIDLength {
				slog.Debug("ping packet smaller than expected", "bytes", n)
			} else if n > srs.GUIDLength {
				slog.Debug("ping packet larger than expected", "bytes", n)
			} else {
				slog.Debug("received UDP ping", "guid", b[0:srs.GUIDLength])
			}
		case <-ctx.Done():
			slog.Info("stopping ping receiver due to context cancellation")
			return
		}
	}
}

func (c *audioClient) receiveVoice(ctx context.Context, packetChan <-chan []byte, audioChan chan<- []byte) {
	buf := make([]byte, 0)
	t := time.NewTicker(100 * time.Millisecond)
	for {
		select {
		case b := <-packetChan:
			voicePacket := newVoicePacketFrom(b)
			buf = append(buf, voicePacket.AudioBytes...)
		case <-t.C:
			slog.Debug("checking if we should send buffer to decoding...", "bufferLength", len(buf), "deadline", c.rxDeadline.String())
			if len(buf) > 0 && time.Now().After(c.rxDeadline) {
				slog.Debug("passed receive deadline with packets in buffer", "bufferLength", len(buf), "deadline", c.rxDeadline.String())
				audio := make([]byte, len(buf))
				copy(buf, audio)
				audioChan <- audio
				buf = make([]byte, 0)
			}
		case <-ctx.Done():
			slog.Info("stopping voice receiver due to context cancellation")
			return
		}
	}
}

func (c *audioClient) decodeVoice(ctx context.Context, opusChan <-chan []byte) {
	decoder := opus.NewDecoder()
	for {
		select {
		case b := <-opusChan:
			audio, err := decode(decoder, b)
			if err != nil {
				slog.Error("failed to decode audio", "error", err)
			} else {
				c.rxchan <- audio
			}
		case <-ctx.Done():
			slog.Info("stopping voice decoder due to context cancellation")
			return
		}
	}
}

func (c *audioClient) Receive() <-chan Audio {
	return c.rxchan
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
