package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type TelemetryClient struct {
	client
	// address of the telemetry service, including port
	address string
	// hostname of the client to use during handshake
	hostname string
	// password to use during handshake
	password string
	// connectionTimeout is the maximum time to wait for a connection to be established.
	connectionTimeout time.Duration
}

func NewTelemetryClient(
	address,
	clientHostname,
	password string,
	connectionTimeout time.Duration,
	updateInterval time.Duration,
) *TelemetryClient {
	return &TelemetryClient{
		client:            *NewClient(updateInterval),
		address:           address,
		hostname:          clientHostname,
		password:          password,
		connectionTimeout: connectionTimeout,
	}
}

func (c *TelemetryClient) Run(ctx context.Context, wg *sync.WaitGroup) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			nextAttempt := time.Now().Add(10 * time.Second)
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

func (c *TelemetryClient) read(ctx context.Context) error {
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

	if err := c.handleLines(ctx, reader); err != nil {
		return fmt.Errorf("error reading updates: %w", err)
	}
	return nil
}

func (c *TelemetryClient) connect() (net.Conn, error) {
	dialer := &net.Dialer{Timeout: c.connectionTimeout}
	connection, err := dialer.Dial("tcp", c.address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %v: %w", c.address, err)
	}
	return connection, nil
}

func (c *TelemetryClient) handshake(reader *bufio.Reader, connection net.Conn) error {
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

	clientHandshake := NewClientHandshake(c.hostname, c.password)
	_, err = connection.Write([]byte(clientHandshake.Encode()))
	if err != nil {
		return fmt.Errorf("error sending client handshake: %w", err)
	}
	log.Info().Str("hostname", c.hostname).Msg("sent client handshake")
	return nil
}
