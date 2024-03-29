package audio

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
)

// pingInterval determines how often we should ping the SRS server over UDP.
const pingInterval = 15 * time.Second

// sendPings is a loop which sends the client GUID to the server at regular intervals to keep our connection alive.
func (c *audioClient) sendPings(ctx context.Context) {
	slog.Info("starting pings", "interval", pingInterval.String())
	go func() {
		time.Sleep(1 * time.Second)
		c.SendPing()
	}()

	ticker := time.NewTicker(pingInterval)
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

// SendPing sends a single ping to the SRS server. "One ping only, Vasily."
// The SRS server won't send us any audio until it receives a ping from us, so this is useful to initialize VoIP.
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
