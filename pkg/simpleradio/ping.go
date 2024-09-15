package simpleradio

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/rs/zerolog/log"
)

// pingInterval determines how often we should ping the SRS server over UDP.
const pingInterval = 15 * time.Second

// sendPings is a loop which sends the client GUID to the server at regular intervals to keep our connection alive.
func (c *client) sendPings(ctx context.Context, wg *sync.WaitGroup) {
	log.Info().Stringer("interval", pingInterval).Msg("starting pings")
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second)
		c.SendPing()
	}()

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

// SendPing sends a single ping to the SRS server. "One ping only, Vasily."
// The SRS server won't send us any audio until it receives a ping from us, so this is useful to initialize VoIP.
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
