package audio

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/rs/zerolog/log"
)

// pingInterval determines how often we should ping the SRS server over UDP.
const pingInterval = 15 * time.Second

// sendPings is a loop which sends the client GUID to the server at regular intervals to keep our connection alive.
func (c *audioClient) sendPings(ctx context.Context, wg *sync.WaitGroup) {
	log.Info().Dur("interval", pingInterval).Msg("starting pings")
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
func (c *audioClient) SendPing() {
	logger := log.With().Str("GUID", string(c.guid)).Logger()
	logger.Trace().Msg("sending UDP ping")
	n, err := c.connection.Write([]byte(c.guid))
	if errors.Is(err, net.ErrClosed) {
		logger.Warn().Msg("ping skipped due to closed connection")
	} else if err != nil {
		logger.Error().Err(err).Msg("error writing ping")
	} else if n != srs.GUIDLength {
		logger.Warn().Int("bytes", n).Int("expectedBytes", srs.GUIDLength).Str("comment", "HOW DID YOU GET HERE").Msg("wrote unexpected number of bytes while sending UDP ping")
	} else {
		logger.Trace().Msg("sent UDP ping")
	}
}
