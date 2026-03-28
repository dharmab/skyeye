// Package telemetry contains clients for reading ACMI data from files or
// real-time telemetry.
package telemetry

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	skynet "github.com/dharmab/skyeye/pkg/net"
	"github.com/rs/zerolog/log"
)

// RealTimeClient reads real-time telemetry data from a remote telemetry service.
type RealTimeClient struct {
	streamingClient
	// address of the telemetry service, including port
	address string
	// hostname of the client to use during handshake
	hostname string
	// password to use during handshake
	password string
	// connectionTimeout is the maximum time to wait for a connection to be established.
	connectionTimeout time.Duration
	hashAlgorithm     HashAlgorithm
}

// NewRealTimeClient creates a new telemetry client for reading real-time
// telemetry data.
func NewRealTimeClient(
	address,
	clientHostname,
	password string,
	connectionTimeout time.Duration,
	updateInterval time.Duration,
) *RealTimeClient {
	return &RealTimeClient{
		streamingClient:   *newStreamingClient(updateInterval),
		address:           address,
		hostname:          clientHostname,
		password:          password,
		connectionTimeout: connectionTimeout,
		hashAlgorithm:     CRC64WE,
	}
}

// Run reads telemetry data until the context is canceled, automatically
// reconnecting if the connection is lost, a read timeout occurs, or an EOF is
// received.
func (c *RealTimeClient) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			nextAttempt := time.Now().Add(skynet.ReconnectDelay)
			if err := c.read(ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				log.Error().Err(err).Msg("error reading telemetry, retrying")
				time.Sleep(time.Until(nextAttempt))
			}
		}
	}
}

func (c *RealTimeClient) read(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	connection, err := c.connect()
	if err != nil {
		return fmt.Errorf("error connecting to telemetry service: %w", err)
	}
	defer connection.Close()

	reader := bufio.NewReader(connection)

	if err := c.handshake(reader, connection); err != nil {
		return fmt.Errorf("error during client handhake: %w", err)
	}

	readTimeout := skynet.CalculateReadTimeout(c.connectionTimeout)

	// Set up periodic deadline refresh to keep connection alive during long reads
	deadlineCtx, deadlineCancel := context.WithCancel(ctx)
	defer deadlineCancel()

	go func() {
		ticker := time.NewTicker(skynet.CalculateDeadlineRefreshInterval(readTimeout))
		defer ticker.Stop()

		for {
			select {
			case <-deadlineCtx.Done():
				return
			case <-ticker.C:
				if err := connection.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
					log.Warn().
						Err(err).
						Stringer("readTimeout", readTimeout).
						Time("attemptedDeadline", time.Now().Add(readTimeout)).
						Msg("failed to refresh telemetry read deadline")
					// Connection may be in bad state, but read will fail
					// and trigger reconnect naturally
				}
			}
		}
	}()

	// Set initial deadline
	if err := connection.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		log.Warn().
			Err(err).
			Stringer("readTimeout", readTimeout).
			Msg("failed to set initial telemetry read deadline")
	}

	if err := c.handleLines(ctx, reader); err != nil {
		return fmt.Errorf("error reading updates: %w", err)
	}
	return nil
}

func (c *RealTimeClient) connect() (net.Conn, error) {
	dialer := &net.Dialer{Timeout: c.connectionTimeout}
	connection, err := dialer.Dial("tcp", c.address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %v: %w", c.address, err)
	}
	return connection, nil
}

func (c *RealTimeClient) handshake(reader *bufio.Reader, connection net.Conn) error {
	hostHandshakePacket, err := reader.ReadString('\x00')
	if err != nil {
		return fmt.Errorf("error reading host handshake: %w", err)
	}

	hostHandshake, err := DecodeHostHandshake(hostHandshakePacket)
	if err != nil {
		log.Debug().Str("packet", hostHandshakePacket).Msg("error decoding host handshake")
		return fmt.Errorf("error decoding host handshake: %w", err)
	}
	log.Info().
		Str("hostname", hostHandshake.Hostname).
		Str("lowLevelProtocolVersion", hostHandshake.LowLevelProtocolVersion).
		Str("highLevelProtocolVersion", hostHandshake.HighLevelProtocolVersion).
		Msg("received host handshake")

	if c.hashAlgorithm == CRC64WE {
		c.hashAlgorithm = CRC32ISOHDLC
	} else {
		c.hashAlgorithm = CRC64WE
	}

	clientHandshake := NewClientHandshake(c.hostname, c.password)
	log.Info().
		Str("hostname", clientHandshake.Hostname).
		Str("lowLevelProtocolVersion", clientHandshake.LowLevelProtocolVersion).
		Str("highLevelProtocolVersion", clientHandshake.HighLevelProtocolVersion).
		Stringer("algorithm", c.hashAlgorithm).
		Msg("sending client handshake")
	_, err = connection.Write([]byte(clientHandshake.Encode(c.hashAlgorithm)))
	if err != nil {
		return fmt.Errorf("error sending client handshake: %w", err)
	}

	return nil
}
