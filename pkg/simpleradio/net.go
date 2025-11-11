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

	skynet "github.com/dharmab/skyeye/pkg/net"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/rs/zerolog/log"
)

// connectTCP connects to the SRS server over TCP.
func (c *Client) connectTCP() error {
	log.Info().
		Str("address", c.address).
		Stringer("timeout", c.connectionTimeout).
		Msg("connecting to SRS server TCP socket")

	tcpAddress, err := net.ResolveTCPAddr("tcp", c.address)
	if err != nil {
		return fmt.Errorf("failed to resolve SRS server address %v: %w", c.address, err)
	}

	dialer := &net.Dialer{
		Timeout: c.connectionTimeout,
	}
	connection, err := dialer.Dial("tcp", tcpAddress.String())
	if err != nil {
		return fmt.Errorf("failed to connect to data socket: %w", err)
	}

	c.tcpConnection = connection.(*net.TCPConn)
	return nil
}

// connectUDP connects to the SRS server over UDP.
func (c *Client) connectUDP() error {
	log.Info().
		Str("address", c.address).
		Stringer("timeout", c.connectionTimeout).
		Msg("connecting to SRS server UDP socket")

	// Note: UDP is connectionless, so there's no actual "connection" to timeout.
	// The timeout here applies only to DNS resolution and local socket setup,
	// not to data transmission (which is handled by read deadlines).
	dialer := &net.Dialer{
		Timeout: c.connectionTimeout,
	}

	udpAddress, err := net.ResolveUDPAddr("udp", c.address)
	if err != nil {
		return fmt.Errorf("failed to resolve SRS server address %v: %w", c.address, err)
	}

	connection, err := dialer.Dial("udp", udpAddress.String())
	if err != nil {
		return fmt.Errorf("failed to connect to UDP socket: %w", err)
	}

	c.udpConnection = connection.(*net.UDPConn)
	return nil
}

// reconnect closes the existing connections and attempts to reconnect to the
// SRS server. It will retry until successful or the context is canceled.
func (c *Client) reconnect(ctx context.Context) error {
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
			backoff = min(time.Duration(float64(backoff)*math.Sqrt2), time.Minute)

			log.Error().Err(err).Stringer("retryIn", backoff).Msg("failed to reconnect to SRS server, retrying")
		}
	}
}

// receiveUDP listens for incoming UDP packets and routes them to the appropriate channel.
func (c *Client) receiveUDP(ctx context.Context, pingChan chan<- []byte, voiceChan chan<- []byte) {
	readTimeout := skynet.CalculateReadTimeout(c.connectionTimeout)

	// Set initial read deadline
	if err := c.udpConnection.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		log.Warn().
			Err(err).
			Stringer("readTimeout", readTimeout).
			Msg("failed to set initial UDP read deadline")
	}

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping SRS packet receiver due to context cancellation")
			return
		default:
			// Set read deadline before each read
			if err := c.udpConnection.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
				log.Warn().
					Err(err).
					Stringer("readTimeout", readTimeout).
					Msg("failed to set UDP read deadline")
			}

			buf := make([]byte, 1500)
			n, err := c.udpConnection.Read(buf)
			switch {
			case errors.Is(err, net.ErrClosed):
				if ctx.Err() == nil {
					log.Error().Err(err).Msg("UDP connection closed")
					time.Sleep(5 * time.Millisecond)
				}
			case errors.Is(err, io.EOF):
				log.Error().Err(err).Msg("UDP connection returned EOF")
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
					pingChan <- packet
				case n > types.GUIDLength:
					// Voice packet
					voiceChan <- packet
				}
			}
		}
	}
}

// receiveTCP listens for incoming TCP messages and routes them to the appropriate handler.
func (c *Client) receiveTCP(ctx context.Context) {
	reader := bufio.NewReader(c.tcpConnection)
	readTimeout := skynet.CalculateReadTimeout(c.connectionTimeout)

	// Set initial read deadline
	if err := c.tcpConnection.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
		log.Warn().
			Err(err).
			Stringer("readTimeout", readTimeout).
			Msg("failed to set initial TCP read deadline")
	}

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping SRS client due to context cancellation")
			return
		default:
			// Set read deadline before each read
			if err := c.tcpConnection.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
				log.Warn().
					Err(err).
					Stringer("readTimeout", readTimeout).
					Msg("failed to set TCP read deadline")
			}

			line, err := reader.ReadBytes(byte('\n'))
			if err != nil {
				if errors.Is(err, net.ErrClosed) && ctx.Err() != nil {
					continue
				}
				log.Error().Err(err).Msg("error reading from SRS server TCP socket")
				// Wait and try again in case it recovers by reconnecting
				time.Sleep(pingInterval)
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
