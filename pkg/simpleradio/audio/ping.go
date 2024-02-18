package audio

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
)

// TODO make PingInterval configurable
const PingInterval = 15 * time.Second

// sendPings is a loop which sends the client GUID to the server every 15 seconds to keep our connection alive.
func (c *audioClient) sendPings(ctx context.Context) {
	slog.Info("starting pings", "interval", PingInterval.String())
	go func() {
		time.Sleep(1 * time.Second)
		c.SendPing()
	}()

	ticker := time.NewTicker(PingInterval)
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
