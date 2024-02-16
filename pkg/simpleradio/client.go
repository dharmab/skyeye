package simpleradio

// https://gitlab.com/overlordbot/srs-bot/-/blob/master/OverlordBot.SimpleRadio/Client.cs

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dharmab/skyeye/pkg/simpleradio/audio"
	"github.com/dharmab/skyeye/pkg/simpleradio/data"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
)

type Client interface {
	Run(context.Context) error
	Receive() <-chan audio.Audio
	Transmit(audio.Audio) error
}

type client struct {
	audioClient audio.AudioClient
	dataClient  data.DataClient
}

func NewClient(config types.ClientConfiguration, radios types.RadioInfo) (Client, error) {
	dataClient, err := data.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to construct SRS data client: %w", err)
	}

	audioClient, err := audio.NewClient(config, radios)
	if err != nil {
		return nil, fmt.Errorf("failed to construct SRS audio client: %w", err)
	}

	client := &client{
		dataClient:  dataClient,
		audioClient: audioClient,
	}

	return client, nil
}

func (c *client) Run(ctx context.Context) error {
	errorChan := make(chan error)

	go func() {
		slog.Info("running SRS data client")
		if err := c.dataClient.Run(ctx); err != nil {
			errorChan <- err
		}
	}()
	go func() {
		slog.Info("running SRS audio client")
		if err := c.audioClient.Run(ctx); err != nil {
			errorChan <- err
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("stopping client due to context cancelation: %w", ctx.Err())
		case err := <-errorChan:
			return fmt.Errorf("client error: %w", err)
		}

	}
}

func (c *client) Receive() <-chan audio.Audio {
	return c.audioClient.Receive()
}

func (c *client) Transmit(sample audio.Audio) error {
	return c.audioClient.Transmit(sample)
}
