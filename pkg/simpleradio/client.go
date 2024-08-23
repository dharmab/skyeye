// package simpleradio contains a bespoke SimpleRadio-Standalone client.
package simpleradio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/audio"
	"github.com/dharmab/skyeye/pkg/simpleradio/data"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/rs/zerolog/log"
)

// Client is a SimpleRadio-Standalone client.
type Client interface {
	// Name returns the name of the client as it appears in the SRS client list and in in-game transmissions.
	Name() string
	// Frequency returns the radio frequency the client is configured to receive and transmit on in Hz.
	Frequency() float64
	// FrequencyMHz returns Client.Frequency in MHz.
	FrequencyMHz() float64
	// Run starts the SimpleRadio-Standalone client. It should be called exactly once.
	Run(context.Context, *sync.WaitGroup) error
	// Receive returns a channel that receives transmissions over the radio. Each transmission is F32LE PCM audio data.
	Receive() <-chan audio.Audio
	// Transmit queues a transmission to send over the radio. The audio data should be in F32LE PCM format.
	Transmit(audio.Audio)
	// IsOnFrequency checks if the named unit is on the client's frequency.
	IsOnFrequency(string) bool
}

// client implements the SRS Client.
type client struct {
	// dataClient is a client for the SRS data protocol.
	dataClient data.DataClient
	// audioClient is a client for the SRS audio protocol.
	audioClient audio.AudioClient
}

func NewClient(config types.ClientConfiguration) (Client, error) {
	guid := types.NewGUID()
	dataClient, err := data.NewClient(guid, config)
	if err != nil {
		return nil, fmt.Errorf("failed to construct SRS data client: %w", err)
	}

	audioClient, err := audio.NewClient(guid, config)
	if err != nil {
		return nil, fmt.Errorf("failed to construct SRS audio client: %w", err)
	}

	client := &client{
		dataClient:  dataClient,
		audioClient: audioClient,
	}

	return client, nil
}

// Name implements [Client.Name].
func (c *client) Name() string {
	return c.dataClient.Name()
}

// Frequency implements [Client.Frequency].
func (c *client) Frequency() float64 {
	return c.audioClient.Frequency()
}

// FrequencyMHz implements [Client.FrequencyMHz].
func (c *client) FrequencyMHz() float64 {
	return c.Frequency() / 1e6
}

// Run implements [Client.Run].
func (c *client) Run(ctx context.Context, wg *sync.WaitGroup) error {
	errorChan := make(chan error)

	// TODO return a ready channel and wait for each. This resolves a minor race condition on startup
	dataReadyCh := make(chan any)
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Msg("running SRS data client")
		if err := c.dataClient.Run(ctx, wg, dataReadyCh); err != nil {
			errorChan <- err
		}
	}()
	<-dataReadyCh

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Msg("running SRS audio client")
		if err := c.audioClient.Run(ctx, wg); err != nil {
			errorChan <- err
		}
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping SRS client due to context cancelation")
			return fmt.Errorf("stopping client due to context cancelation: %w", ctx.Err())
		case err := <-errorChan:
			return fmt.Errorf("client error: %w", err)
		case <-ticker.C:
			if time.Since(c.audioClient.LastPing()) > 1*time.Minute {
				log.Warn().Msg("stopped receiving pings from SRS data client")
				return fmt.Errorf("stopped receiving pings from SRS data client")
			}

		}
	}
}

// Receive implements [Client.Receive].
func (c *client) Receive() <-chan audio.Audio {
	return c.audioClient.Receive()
}

// Transmit implements [Client.Transmit].
func (c *client) Transmit(sample audio.Audio) {
	c.audioClient.Transmit(sample)
}

// IsOnFrequency implements [Client.IsOnFrequency].
func (c *client) IsOnFrequency(name string) bool {
	return c.dataClient.IsOnFrequency(name)
}
