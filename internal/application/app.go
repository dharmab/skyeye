// package application is the main package for the SkyEye application.
package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	grpccoalition "github.com/DCS-gRPC/go-bindings/dcs/v0/coalition"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/mission"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/net"
	"github.com/dharmab/skyeye/internal/conf"
	secoalition "github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/commands"
	"github.com/dharmab/skyeye/pkg/composer"
	"github.com/dharmab/skyeye/pkg/controller"
	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/dharmab/skyeye/pkg/recognizer"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/synthesizer/speakers"
	"github.com/dharmab/skyeye/pkg/telemetry"
	"github.com/dharmab/skyeye/pkg/traces"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Application implements the SkyEye application.
type Application struct {
	// callsign of the GCI controller
	callsign string
	// srsClient is a SimpleRadio Standalone client
	srsClient *simpleradio.Client
	// telemetryClient streams ACMI data
	telemetryClient telemetry.Client
	// recognizer provides speech-to-text recognition
	recognizer recognizer.Recognizer
	// chatListener listens for chat messages
	chatListener *commands.ChatListener
	// parser converts English brevity text to internal representations
	parser *parser.Parser
	// radar tracks contacts and provides geometric computations
	radar *radar.Radar
	// controller publishes responses and calls
	controller *controller.Controller
	// composer converts responses and calls from internal representations to English brevity text
	composer composer.Composer
	// speaker provides text-to-speech synthesis
	speaker speakers.Speaker
	// enableTranscriptionLogging controls whether transcriptions are included in logs
	enableTranscriptionLogging bool
	// tracers are destinations where traces are sent when tracing is enabled
	tracers []traces.Tracer

	starts  chan sim.Started
	updates chan sim.Updated
	fades   chan sim.Faded

	// exitAfter is the duration after which the application should exit
	exitAfter time.Duration
}

// NewApplication constructs a new Application.
func NewApplication(config conf.Configuration) (*Application, error) {
	starts := make(chan sim.Started)
	updates := make(chan sim.Updated)
	fades := make(chan sim.Faded)

	var chatListener *commands.ChatListener
	if config.EnableGRPC {
		log.Info().Str("address", config.GRPCAddress).Msg("constructing gRPC clients")
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		if config.GRPCAPIKey != "" {
			log.Info().Msg("configuring gRPC client connection with provided API key")
			opts = append(opts, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				m := metadata.Pairs("X-API-Key", config.GRPCAPIKey)
				ctx = metadata.NewOutgoingContext(ctx, m)
				return invoker(ctx, method, req, reply, cc, opts...)
			}))
		}
		grpcClient, err := grpc.NewClient(config.GRPCAddress, opts...)
		if err != nil {
			return nil, err
		}
		missionClient := mission.NewMissionServiceClient(grpcClient)
		coalitionClient := grpccoalition.NewCoalitionServiceClient(grpcClient)
		netClient := net.NewNetServiceClient(grpcClient)

		log.Info().Msg("constructing chat listener")
		chatListener = commands.NewChatListener(
			config.Coalition,
			config.Callsign,
			missionClient,
			coalitionClient,
			netClient,
		)
	}

	radios := make([]srs.Radio, 0, len(config.SRSFrequencies))
	for _, radioFrequency := range config.SRSFrequencies {
		radios = append(radios, srs.Radio{
			Frequency:        radioFrequency.Frequency.Hertz(),
			Modulation:       radioFrequency.Modulation,
			ShouldRetransmit: true,
		})
	}

	log.Info().
		Str("address", config.SRSAddress).
		Stringer("timeout", config.SRSConnectionTimeout).
		Str("clientName", config.SRSClientName).
		Int("coalitionID", int(config.Coalition)).
		Int("modulationID", int(srs.ModulationAM)).
		Msg("constructing SRS client")
	srsClient, err := simpleradio.NewClient(srs.ClientConfiguration{
		Address:                   config.SRSAddress,
		ConnectionTimeout:         config.SRSConnectionTimeout,
		ClientName:                config.SRSClientName,
		ExternalAWACSModePassword: config.SRSExternalAWACSModePassword,
		Coalition:                 config.Coalition,
		Radios:                    radios,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	var telemetryClient telemetry.Client
	if config.ACMIFile != "" {
		log.Info().Str("file", config.ACMIFile).Msg("constructing ACMI file reader")
		telemetryClient = telemetry.NewFileClient(config.ACMIFile, config.RadarSweepInterval)
	} else {
		log.Info().Str("address", config.TelemetryAddress).Msg("constructing telemetry client")
		telemetryClient = telemetry.NewRealTimeClient(
			config.TelemetryAddress,
			config.Callsign,
			config.TelemetryPassword,
			config.TelemetryConnectionTimeout,
			config.RadarSweepInterval,
		)
	}

	log.Info().Msg("constructing speech-to-text recognizer")
	speechRecognizer := recognizer.NewWhisperRecognizer(config.WhisperModel, config.Callsign)

	log.Info().Msg("constructing request parser")
	requestParser := parser.New(config.Callsign, config.EnableTranscriptionLogging)

	log.Info().Msg("constructing radar scope")

	rdr := radar.New(starts, updates, fades, config.MandatoryThreatRadius)
	log.Info().Msg("constructing GCI controller")
	gciController := controller.New(
		rdr,
		srsClient,
		config.Coalition,
		config.EnableAutomaticPicture,
		config.PictureBroadcastInterval,
		config.EnableThreatMonitoring,
		config.ThreatMonitoringInterval,
		config.ThreatMonitoringRequiresSRS,
	)

	log.Info().Msg("constructing response composer")
	responseComposer := composer.Composer{Callsign: config.Callsign}

	log.Info().Msg("constructing text-to-speech synthesizer")
	synthesizer, err := speakers.NewPiperSpeaker(config.Voice, config.VoiceSpeed, config.VoicePauseLength)
	if err != nil {
		return nil, fmt.Errorf("failed to construct application: %w", err)
	}

	tracers := make([]traces.Tracer, 0)
	if config.EnableTracing {
		log.Info().Msg("constructing tracers")
		logTracer := traces.LogTracer{}
		tracers = append(tracers, &logTracer)
		if config.DiscordWebhookID != "" && config.DiscorbWebhookToken != "" {
			discordWebhook, err := traces.NewDiscordWebhook(config.DiscordWebhookID, config.DiscorbWebhookToken)
			if err != nil {
				return nil, fmt.Errorf("failed to construct application: %w", err)
			}
			tracers = append(tracers, discordWebhook)
		}
	}

	log.Info().Msg("constructing application")
	app := &Application{
		callsign:                   config.Callsign,
		enableTranscriptionLogging: config.EnableTranscriptionLogging,
		chatListener:               chatListener,
		srsClient:                  srsClient,
		telemetryClient:            telemetryClient,
		recognizer:                 speechRecognizer,
		parser:                     requestParser,
		radar:                      rdr,
		controller:                 gciController,
		composer:                   responseComposer,
		speaker:                    synthesizer,
		tracers:                    tracers,
		starts:                     starts,
		updates:                    updates,
		fades:                      fades,
		exitAfter:                  config.ExitAfter,
	}
	return app, nil
}

// Run implements Application.Run.
func (a *Application) Run(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) error {
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Msg("running telemetry client")
		if err := a.telemetryClient.Run(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Error().Err(err).Msg("error running telemetry client")
				cancel()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Msg("streaming telemetry data to radar")
		a.telemetryClient.Stream(ctx, wg, a.starts, a.updates, a.fades)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Msg("updating mission time and bullseye")
		ticker := time.NewTicker(2*time.Second + 100*time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("stopping mission time and bullseye updates due to context cancellation")
				return
			case <-ticker.C:
				a.updateMissionTime()
				a.updateBullseyes()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Msg("running SRS client")
		if err := a.srsClient.Run(ctx, wg); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Error().Err(err).Msg("error running SRS client")
				cancel()
			}
		}
	}()

	rxTextChan := make(chan Message[string])
	requestChan := make(chan Message[any])
	callChan := make(chan controller.Call)
	txTextChan := make(chan Message[composer.NaturalLanguageResponse])
	txAudioChan := make(chan Message[simpleradio.Audio])

	log.Info().Msg("starting subroutines")
	log.Info().Msg("starting speech recognition routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.recognize(ctx, rxTextChan)
	}()

	if a.chatListener != nil {
		requestChan := make(chan commands.Request)
		log.Info().Msg("starting chat listener routines")
		wg.Add(1)
		go func() {
			defer wg.Done()
			a.chatListener.Run(ctx, requestChan)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case request := <-requestChan:
					rCtx := traces.NewRequestContext()
					rCtx = traces.WithTraceID(rCtx, request.TraceID)
					rCtx = traces.WithPlayerName(rCtx, request.PlayerName)
					rCtx = traces.WithRequestText(rCtx, request.Text)
					rxTextChan <- Message[string]{Context: rCtx, Data: request.Text}
				}
			}
		}()
	}

	log.Info().Msg("starting request parsing routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.parse(ctx, rxTextChan, requestChan)
	}()
	log.Info().Msg("starting radar scope routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.radar.Run(ctx, wg)
	}()
	log.Info().Msg("starting GCI controller routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.control(ctx, wg, requestChan, callChan)
	}()
	log.Info().Msg("starting response composer routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.compose(ctx, callChan, txTextChan)
	}()
	log.Info().Msg("starting speech synthesis routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.synthesize(ctx, txTextChan, txAudioChan)
	}()
	log.Info().Msg("starting radio transmission routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.transmit(ctx, txAudioChan)
	}()

	log.Info().Dur("duration", a.exitAfter).Msg("starting exit timer routine")
	wg.Add(1)
	go func() {
		defer wg.Done()
		timer := time.NewTimer(a.exitAfter)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if a.srsClient.HumansOnFrequency() == 0 {
						log.Info().Msg("reached exit time and no clients are connected - exiting")
						cancel()
					} else {
						log.Warn().Msg("reached exit time but clients are still connected")
					}
				}
			}
		}
	}()

	return nil
}

// updateMissionTime updates the mission time on the radar.
func (a *Application) updateMissionTime() {
	missionTime := a.telemetryClient.Time()
	a.radar.SetMissionTime(missionTime)
}

// updateBullseyes updates the positions of the bullseyes on the radar.
func (a *Application) updateBullseyes() {
	for _, coalition := range []secoalition.Coalition{secoalition.Red, secoalition.Blue} {
		bullseye, err := a.telemetryClient.Bullseye(coalition)
		if err != nil {
			log.Warn().Err(err).Msg("error reading bullseye")
		} else {
			a.radar.SetBullseye(bullseye, coalition)
		}
	}
}

// trace the given request context using all configured tracers.
func (a *Application) trace(ctx context.Context) {
	if !a.enableTranscriptionLogging {
		ctx = traces.WithoutRequestText(ctx)
	}
	for _, tracer := range a.tracers {
		tracer.Trace(ctx)
	}
}
