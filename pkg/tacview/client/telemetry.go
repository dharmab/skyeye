package client

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"time"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/tacview/acmi"
	"github.com/dharmab/skyeye/pkg/tacview/types"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

type telemetryClient struct {
	connection *net.TCPConn
	hostname   string
	password   string
	*tacviewClient
}

var _ Client = &telemetryClient{}

func NewTelemetryClient(
	address,
	clientHostname,
	password string,
	coalition coalitions.Coalition,
	updates chan<- sim.Updated,
	fades chan<- sim.Faded,
	bullseyes chan<- orb.Point,
	updateInterval time.Duration,
) (Client, error) {
	log.Info().Str("protocol", "tcp").Str("address", address).Msg("connecting to telemetry service")
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve telemetry service address %v: %w", address, err)
	}
	connection, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to telemetry service %v: %w", address, err)
	}
	return &telemetryClient{
		connection: connection,
		hostname:   clientHostname,
		password:   password,
		tacviewClient: &tacviewClient{
			coalition:      coalition,
			updates:        updates,
			fades:          fades,
			bullseyes:      bullseyes,
			updateInterval: updateInterval,
		},
	}, nil
}

func (c *telemetryClient) Run(ctx context.Context) error {
	reader := bufio.NewReader(c.connection)

	if err := c.handshake(reader, c.hostname, c.password); err != nil {
		return fmt.Errorf("handshake error: %w", err)
	}

	source := acmi.New(c.coalition, reader, c.updateInterval)
	return c.run(ctx, source)
}

func (c *telemetryClient) handshake(reader *bufio.Reader, hostname, password string) error {
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
	_, err = c.connection.Write([]byte(clientHandshake.Encode()))
	if err != nil {
		return fmt.Errorf("error sending client handshake: %w", err)
	}
	return nil
}

func (c *telemetryClient) Close() error {
	if err := c.connection.Close(); err != nil {
		return fmt.Errorf("error closing connection to telemetry service: %w", err)
	}
	return nil
}
