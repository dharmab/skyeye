package simpleradio

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/rs/zerolog/log"
)

// pingInterval is how often SRS pings should be sent.
const pingInterval = 15 * time.Second

// sendPings pings the SRS server at regular intervals.
func (c *client) sendPings(ctx context.Context) {
	log.Info().Stringer("interval", pingInterval).Msg("starting pings")
	c.SendPing()
	ticker := time.NewTicker(pingInterval)
	for {
		select {
		case <-ticker.C:
			c.SendPing()
		case <-ctx.Done():
			log.Info().Msg("stopping SRS pings due to context cancelation")
			return
		}
	}
}

// SendPing sends a single ping to the SRS server over both TCP and UDP.
func (c *client) SendPing() {
	guid := c.clientInfo.GUID
	logger := log.With().Str("GUID", string(guid)).Logger()

	if err := c.Send(c.newMessageWithClient(types.MessagePing)); err != nil {
		logger.Error().Err(err).Msg("error sending TCP ping")
	}

	_, err := c.udpConnection.Write([]byte(guid))
	if errors.Is(err, net.ErrClosed) {
		logger.Warn().Msg("ping skipped due to closed connection")
	} else if err != nil {
		logger.Error().Err(err).Msg("error sending UDP ping")
	}
}

// receivePings listens for incoming UDP ping packets and logs them at DEBUG level.
func (c *client) receivePings(ctx context.Context, in <-chan []byte) {
	for {
		select {
		case b := <-in:
			n := len(b)
			if n < types.GUIDLength {
				log.Debug().Int("bytes", n).Msg("received UDP ping smaller than expected")
			} else if n > types.GUIDLength {
				log.Debug().Int("bytes", n).Msg("received UDP ping larger than expected")
			} else {
				log.Trace().Str("GUID", string(b[0:types.GUIDLength])).Msg("received UDP ping")
				t := time.Now()
				func() {
					c.lastPingLock.Lock()
					defer c.lastPingLock.Unlock()
					c.lastPing = t
				}()
			}
		case <-ctx.Done():
			log.Info().Msg("stopping SRS ping receiver due to context cancellation")
			return
		}
	}
}
