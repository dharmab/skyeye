package simpleradio

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/rs/zerolog/log"
)

func (c *client) connectTCP() error {
	log.Info().Str("address", c.address).Msg("connecting to SRS server TCP socket")
	tcpAddress, err := net.ResolveTCPAddr("tcp", c.address)
	if err != nil {
		return fmt.Errorf("failed to resolve SRS server address %v: %w", c.address, err)
	}
	connection, err := net.DialTCP("tcp", nil, tcpAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to data socket: %w", err)
	}
	c.tcpConnection = connection
	//c.tcpReader = bufio.NewReader(connection)
	return nil
}

func (c *client) connectUDP() error {
	log.Info().Str("address", c.address).Msg("connecting to SRS server UDP socket")
	udpAddress, err := net.ResolveUDPAddr("udp", c.address)
	if err != nil {
		return fmt.Errorf("failed to resolve SRS server address %v: %w", c.address, err)
	}
	connection, err := net.DialUDP("udp", nil, udpAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to UDP socket: %w", err)
	}
	c.udpConnection = connection
	return nil
}

// reconnect attempts to reconnect the TCP and UDP connections indefinitely.
func (c *client) reconnect(ctx context.Context) error {
	var err error
	backoff := frameLength

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			log.Info().Msg("attempting to reconnect to SRS server")
			_ = c.tcpConnection.Close()
			err = c.connectTCP()
			if err == nil {
				log.Info().Msg("successfully reconnected to SRS server over TCP")
				_ = c.udpConnection.Close()
				err = c.connectUDP()
				if err == nil {
					log.Info().Msg("successfully reconnected to SRS server over UDP")
					return nil
				}
			}
			time.Sleep(backoff)
			backoff = time.Duration(float64(backoff) * math.Sqrt2)
			if backoff > time.Minute {
				backoff = time.Minute
			}

			log.Error().Err(err).Stringer("retryIn", backoff).Msg("failed to reconnect to SRS server, retrying")
		}
	}
}

// receiveUDP listens for incoming UDP packets and routes them to the appropriate channel.
func (c *client) receiveUDP(ctx context.Context, pingCh chan<- []byte, voiceCh chan<- []byte) {
	for {
		if ctx.Err() != nil {
			if ctx.Err() == context.Canceled {
				log.Info().Msg("stopping SRS packet receiver due to context cancellation")
			} else {
				log.Error().Err(ctx.Err()).Msg("stopping packet receiver due to context error")
			}
			return
		}

		buf := make([]byte, 1500)
		n, err := c.udpConnection.Read(buf)
		switch {
		case errors.Is(err, net.ErrClosed):
			log.Error().Err(err).Msg("UDP connection closed")
		case errors.Is(err, io.EOF):
			log.Error().Err(err).Msg("UDP connection closed")
		case err != nil:
			log.Error().Err(err).Msg("UDP connection read error")
		case n == 0:
			log.Warn().Err(err).Msg("0 bytes read from UDP connection")
		default:
			packet := make([]byte, n)
			copy(packet, buf)
			switch {
			case n < types.GUIDLength:
				log.Debug().Int("bytes", n).Msg("UDP packet smaller than expected")
			case n == types.GUIDLength:
				// Ping packet
				pingCh <- packet
			case n > types.GUIDLength:
				// Voice packet
				voiceCh <- packet
			}
		}
	}
}

func (c *client) receiveTCP(ctx context.Context) {
	reader := bufio.NewReader(c.tcpConnection)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping SRS client due to context cancellation")
			return
		default:
			line, err := reader.ReadBytes(byte('\n'))
			if err != nil {
				log.Error().Err(err).Msg("error reading from SRS server TCP socket")
				// Wait and try again in case it recovers by reconnecting
				time.Sleep(5 * time.Second)
				reader = bufio.NewReader(c.tcpConnection)
				continue
			}

			var message types.Message
			jsonErr := json.Unmarshal(line, &message)
			if jsonErr != nil {
				log.Warn().Str("text", string(line)).Err(jsonErr).Msg("failed to unmarshal message")
			} else {
				c.handleMessage(message)
			}
		}
	}
}
