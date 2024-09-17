package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dharmab/skyeye/internal/cli"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	srstypes "github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/lithammer/shortuuid/v3"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	webhookURL                   string
	webhookTimeout               time.Duration
	logLevel                     string
	logFormat                    string
	srsAddress                   string
	srsConnectionTimeout         time.Duration
	srsExternalAWACSModePassword string
	srsFrequencies               []string
	scaleInterval                time.Duration
	stopDelay                    time.Duration
)

var scaler = &cobra.Command{
	Use:   "skyeye-scaler",
	Short: "SkyEye Autoscaler",
	Long:  "skyeye-scaler calls a webhook which can be used to create or destroy a SkyEye instance on demand.",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		v := viper.New()
		v.SetEnvPrefix("SKYEYE_SCALER")
		v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		v.AutomaticEnv()
		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if !f.Changed && v.IsSet(f.Name) {
				val := v.Get(f.Name)
				if err := cmd.Flags().Set(f.Name, fmt.Sprint(val)); err != nil {
					log.Warn().Str("flag", f.Name).Msg("Failed to set flag")
				}
			}
		})
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return run()
	},
}

var playersSeenAt time.Time

func init() {
	logLevelFlag := cli.NewEnum(&logLevel, "Level", "info", "error", "warn", "info", "debug", "trace")
	scaler.Flags().Var(logLevelFlag, "log-level", "Log level (error, warn, info, debug, trace)")
	logFormats := cli.NewEnum(&logFormat, "Format", "pretty", "json")
	scaler.Flags().Var(logFormats, "log-format", "Log format (pretty, json)")

	scaler.Flags().StringVar(&webhookURL, "webhook-url", "", "URL to call")
	scaler.MarkFlagRequired("webhook-url")
	scaler.Flags().DurationVar(&webhookTimeout, "webhook-timeout", 30*time.Second, "Webhook request timeout")
	scaler.Flags().DurationVar(&scaleInterval, "scale-interval", 1*time.Minute, "Interval at which to check SRS player count")
	scaler.Flags().DurationVar(&stopDelay, "stop-delay", 10*time.Minute, "Delay before sending stop requests after the SRS player count drops to 0")

	scaler.Flags().StringVar(&srsAddress, "srs-server-address", "localhost:5002", "Address of the SRS server")
	scaler.Flags().DurationVar(&srsConnectionTimeout, "srs-connection-timeout", 10*time.Second, "Connection timeout for SRS client")
	scaler.Flags().StringVar(&srsExternalAWACSModePassword, "srs-eam-password", "", "SRS external AWACS mode password")
	scaler.Flags().StringSliceVar(&srsFrequencies, "srs-frequencies", []string{"251.0AM", "133.0AM", "30.0FM"}, "List of SRS frequencies to use")
}

func main() {
	cobra.MousetrapDisplayDuration = 0
	if err := scaler.Execute(); err != nil {
		log.Fatal().Err(err).Msg("scaler exited eith error")
	}
}

func run() error {
	cli.SetupZerolog(logLevel, logFormat)
	log.Info().Msg("Starting SkyEye Autoscaler")

	if stopDelay > scaleInterval {
		log.Fatal().Msg("stop-delay must be shorter than scale-interval")
	}

	parsedFrequencies := cli.LoadFrequencies(srsFrequencies)
	radios := make([]srstypes.Radio, 0, len(parsedFrequencies))
	for _, radioFrequency := range parsedFrequencies {
		radios = append(radios, srstypes.Radio{
			Frequency:        radioFrequency.Frequency.Hertz(),
			Modulation:       radioFrequency.Modulation,
			ShouldRetransmit: true,
		})
	}

	clientName := fmt.Sprintf("SkyEye Scaler %s [BOT]", shortuuid.New())
	srsConfig := srstypes.ClientConfiguration{
		Address:                   srsAddress,
		ConnectionTimeout:         srsConnectionTimeout,
		ClientName:                clientName,
		ExternalAWACSModePassword: srsExternalAWACSModePassword,
		Coalition:                 coalitions.Blue,
		Radios:                    radios,
	}
	srsClient, err := simpleradio.NewClient(srsConfig)
	if err != nil {
		return fmt.Errorf("failed to create SRS client: %w", err)
	}

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	interuptChan := make(chan os.Signal, 1)
	signal.Notify(interuptChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := <-interuptChan
		log.Info().Any("signal", s).Msg("received shutdown signal")
		cancel()
		wg.Wait()
		os.Exit(0)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := srsClient.Run(ctx, &wg)
		if err != nil {
			log.Error().Err(err).Msg("failed to run SRS client")
			cancel()
		}
	}()

	client := &http.Client{Timeout: webhookTimeout}

	ticker := time.NewTicker(scaleInterval)
	defer ticker.Stop()
	callWebhook(client, srsClient)
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping application due to context cancelation")
			wg.Wait()
			return fmt.Errorf("stopping application due to context cancelation: %w", ctx.Err())
		case <-ticker.C:
			callWebhook(client, srsClient)
		}
	}
}

type Payload struct {
	Action      string   `json:"action"`
	Players     int      `json:"players"`
	Address     string   `json:"address"`
	Frequencies []string `json:"frequencies"`
}

func callWebhook(httpClient *http.Client, srsClient simpleradio.Client) {
	playerCount := srsClient.HumansOnFrequency()
	logger := log.With().Int("players", playerCount).Logger()
	action := "run"
	timeSincePlayersSeen := time.Since(playersSeenAt)
	if playerCount > 0 {
		playersSeenAt = time.Now()
	} else if timeSincePlayersSeen > stopDelay {
		action = "stop"
	}
	logger = logger.With().Str("action", action).Stringer("timeSincePlayers", timeSincePlayersSeen).Logger()

	frequencies := make([]float64, 0, len(srsClient.Frequencies()))
	for _, freq := range srsClient.Frequencies() {
		frequencies = append(frequencies, freq.Frequency.Megahertz())
	}

	payload := Payload{
		Action:      action,
		Players:     playerCount,
		Address:     srsAddress,
		Frequencies: srsFrequencies,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to marshal payload")
		return
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewBuffer(body))
	if err != nil {
		logger.Error().Err(err).Msg("failed to create request")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	logger.Info().Msg("calling webhook")
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to call webhook")
		return
	}
	defer resp.Body.Close()

	logger = logger.With().Int("status", resp.StatusCode).Logger()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Warn().Msg("received non-2XX response status from webhook")
	} else {
		logger.Info().Msg("webhook called successfully")
	}
}
