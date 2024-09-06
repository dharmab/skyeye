package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/tacview/acmi"
	"github.com/dharmab/skyeye/pkg/tacview/types"
	"github.com/rs/zerolog/log"
)

type telemetryClient struct {
	address  string
	hostname string
	password string
	*tacviewClient
}

var _ Client = &telemetryClient{}

func NewTelemetryClient(
	address,
	clientHostname,
	password string,
	coalition coalitions.Coalition,
	starts chan<- sim.Started,
	updates chan<- sim.Updated,
	fades chan<- sim.Faded,
	updateInterval time.Duration,
) (Client, error) {
	log.Info().Str("protocol", "tcp").Str("address", address).Msg("connecting to telemetry service")

	tacviewClient := newTacviewClient(starts, updates, fades, updateInterval)
	return &telemetryClient{
		address:       address,
		hostname:      clientHostname,
		password:      password,
		tacviewClient: tacviewClient,
	}, nil
}

// Run implements [Client.Run].
func (c *telemetryClient) Run(ctx context.Context, wg *sync.WaitGroup) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := c.run(ctx, wg); err != nil {
				log.Error().Err(err).Msg("telemetry error, attempting to reconnect")
				time.Sleep(10 * time.Second)
			} else {
				return nil
			}
		}
	}
}

// Time implements [Client.Time].
func (c *telemetryClient) Time() time.Time {
	return c.missionTime
}

func (c *telemetryClient) run(ctx context.Context, wg *sync.WaitGroup) error {
	addr, err := net.ResolveTCPAddr("tcp", c.address)
	if err != nil {
		return fmt.Errorf("failed to resolve telemetry service address %v: %w", c.address, err)
	}
	connection, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to connect to telemetry service %v: %w", c.address, err)
	}
	defer connection.Close()

	reader := bufio.NewReader(connection)

	if err := c.handshake(reader, connection, c.hostname, c.password); err != nil {
		err := fmt.Errorf("handshake error: %w", err)
		log.Error().Err(err).Msg("error during handshake, attempting to reconnect")
	}

	source := acmi.New(reader, c.updateInterval)

	if err := c.stream(ctx, wg, source); err != nil {
		if errors.Is(err, io.EOF) {
			return err
		} else {
			return fmt.Errorf("error running telemetry: %w", err)
		}
	}
	return nil
}

func (c *telemetryClient) handshake(reader *bufio.Reader, connection *net.TCPConn, hostname, password string) error {
	hostHandshakePacket, err := reader.ReadString('\000')
	if err != nil {
		return fmt.Errorf("error reading handshake: %w", err)
	}

	hostHandshake, err := types.DecodeHostHandshake(hostHandshakePacket)
	if err != nil {
		log.Debug().Str("packet", hostHandshakePacket).Msg("error decoding host handshake")
		return fmt.Errorf("error decoding host handshake: %w", err)
	}
	log.Info().Str("hostname", hostHandshake.Hostname).Msg("received host handshake")

	clientHandshake := types.NewClientHandshake(hostname, password)
	_, err = connection.Write([]byte(clientHandshake.Encode()))
	if err != nil {
		return fmt.Errorf("error sending client handshake: %w", err)
	}
	return nil
}

// Close implements [Client.Close].
func (c *telemetryClient) Close() error {
	return nil
}
