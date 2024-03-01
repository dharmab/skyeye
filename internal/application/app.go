// package application is the main package for the SkyEye application.
package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/dharmab/skyeye/pkg/dcs"
	"github.com/dharmab/skyeye/pkg/recognizer"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/synthesizer"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

// Configuration for the SkyEye application.
type Configuration struct {
	// DCSGRPCAddress is the network address of the DCS-gRPC server (including port)
	DCSGRPCAddress string
	// GRPCConnectionTimeout is the connection timeout for connecting to DCS-gRPC
	GRPCConnectionTimeout time.Duration
	// SRSAddress is the network address of the SimpleRadio Standalone server (including port)
	SRSAddress string
	// SRSConnectionTimeout is the connection timeout for connecting to the SimpleRadio Standalone server
	SRSConnectionTimeout time.Duration
	// SRSClientName is the name of the bot that will appear in the client list and in in-game transmissions
	SRSClientName string
	// SRSExternalAWACSModePassword is the password for connecting to the SimpleRadio Standalone server using External AWACS Mode
	SRSExternalAWACSModePassword string
	// SRSFrequency is the radio frequency the bot will listen to and talk on in Hz
	SRSFrequency float64
	// SRSCoalition is the coalition that the bot will act on
	SRSCoalition srs.Coalition
	// WhisperModel is a whisper.cpp model used for Speech To Text
	WhisperModel whisper.Model
}

// Application is the interface for running the SkyEye application.
type Application interface {
	// Run runs the SkyEye application. It should be called exactly once.
	Run(context.Context) error
}

// app implements the Application.
type app struct {
	// dcsClient is a DCS-gRPC client
	dcsClient dcs.DCSClient
	// srsClient is a SimpleRadio Standalone client
	srsClient simpleradio.Client
	// recognizer provides speech-to-text recognition
	recognizer recognizer.Recognizer
	// synthesizer provides text-to-speech synthesis
	synthesizer synthesizer.Sythesizer
}

// NewApplication constructs a new Application.
func NewApplication(ctx context.Context, config Configuration) (Application, error) {
	slog.Info("constructing DCS client")
	dcsClient, err := dcs.NewDCSClient(
		ctx,
		dcs.ClientConfiguration{
			Address:           config.DCSGRPCAddress,
			ConnectionTimeout: config.GRPCConnectionTimeout,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	slog.Info("constructing SRS client")
	srsClient, err := simpleradio.NewClient(
		srs.ClientConfiguration{
			Address:                   config.SRSAddress,
			ConnectionTimeout:         config.SRSConnectionTimeout,
			ClientName:                config.SRSClientName,
			ExternalAWACSModePassword: config.SRSExternalAWACSModePassword,
			Coalition:                 config.SRSCoalition,
			Radio: srs.Radio{
				Frequency:        config.SRSFrequency,
				Modulation:       srs.ModulationAM,
				ShouldRetransmit: true,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	slog.Info("constructing speech-to-text recognizer")
	recognizer := recognizer.NewWhisperRecognizer(config.WhisperModel)

	slog.Info("constructing text-to-speech syhthesizer")
	synthesizer, err := synthesizer.NewPiperSpeaker(synthesizer.FeminineVoice)
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	app := &app{
		dcsClient:   dcsClient,
		srsClient:   srsClient,
		recognizer:  recognizer,
		synthesizer: synthesizer,
	}
	return app, nil
}

// Run implements Application.Run.
func (a *app) Run(ctx context.Context) error {
	defer func() {
		slog.Info("closing connection to DCS-gRPC server")
		err := a.dcsClient.Close()
		if err != nil {
			slog.Error("failed to close connection to DCS-gRPC server", "error", err)
		}
	}()

	go func() {
		slog.Info("running SRS client")
		if err := a.srsClient.Run(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				slog.Error("error running SRS client", "error", err)
			}
		}
	}()

	time.Sleep(2 * time.Second)
	slog.Info("generating sunrise message")
	sunriseSample, err := a.synthesizer.Say(fmt.Sprintf(
		"All players, GCI %s sunrise on %s",
		a.srsClient.Name(),
		synthesizer.PronounceDecimal(a.srsClient.FrequencyMHz(), 3, ""),
	))
	if err != nil {
		return fmt.Errorf("failed to generate sunrise message: %w", err)
	}
	//debug.PlayAudio(&a.otoCtx, audiotools.F32toS16LEBytes(sunriseSample))

	slog.Info("transmitting sunrise message")
	a.srsClient.Transmit(sunriseSample)

	time.Sleep(6 * time.Second)

	threatSample, err := a.synthesizer.Say("Raven fife one, threat, bullseye two fife niner, 59, 22000, track southwest, hostile, foxbat, two ship")
	if err != nil {
		return fmt.Errorf("failed to generate sunrise message: %w", err)
	}
	a.srsClient.Transmit(threatSample)

	for {
		select {
		case <-ctx.Done():
			return nil
		case sample := <-a.srsClient.Receive():
			slog.Info("recognizing audio sample received from SRS client")
			text, err := a.recognizer.Recognize(sample)
			if err != nil {
				slog.Error("error recongizing audio sample", "error", err)
			} else {
				slog.Info("recognized audio", "text", text)
			}
		}
	}
}
