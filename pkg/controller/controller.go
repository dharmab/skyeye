// package controller implements high-level logic for Ground-Controlled Interception (GCI)
package controller

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

var (
	lowestAltitude  = unit.Length(0)
	highestAltitude = unit.Length(100000) * unit.Foot
)

// Controller handles requests for GCI service.
type Controller interface {
	// Run starts the controller's control loops. It should be called exactly once. It blocks until the context is canceled.
	// The controller publishes responses to the given channel.
	Run(ctx context.Context, out chan<- any)
	// HandleAlphaCheck handles an ALPHA CHECK by reporting the position of the requesting aircraft.
	HandleAlphaCheck(*brevity.AlphaCheckRequest)
	// HandleBogeyDope handles a BOGEY DOPE by reporting the closest enemy group to the requesting aircraft.
	HandleBogeyDope(*brevity.BogeyDopeRequest)
	// HandleDeclare handles a DECLARE by reporting information about the target group.
	HandleDeclare(*brevity.DeclareRequest)
	// HandlePicture handles a PICTURE by reporting a tactical air picture.
	HandlePicture(*brevity.PictureRequest)
	// HandleRadioCheck handles a RADIO CHECK by responding to the requesting aircraft.
	HandleRadioCheck(*brevity.RadioCheckRequest)
	// HandleSnaplock handles a SNAPLOCK by reporting information about the target group.
	HandleSnaplock(*brevity.SnaplockRequest)
	// HandleSpiked handles a SPIKED by reporting any enemy groups in the direction of the radar spike.
	HandleSpiked(*brevity.SpikedRequest)
	// HandleTripwire handles a TRIPWIRE... by not implementing it LOL
	HandleTripwire(*brevity.TripwireRequest)
	// HandleUnableToUnderstand handles requests where the wake word was recognized but the request could not be understood, by asking players on the channel to repeat their message.
	HandleUnableToUnderstand(*brevity.UnableToUnderstandRequest)
}

type controller struct {
	out                         chan<- any
	scope                       radar.Radar
	coalition                   coalitions.Coalition
	frequencies                 []simpleradio.RadioFrequency
	pictureBroadcastInterval    time.Duration
	pictureBroadcastDeadline    time.Time
	threatCooldowns             *cooldownTracker
	warmupTime                  time.Time
	srsClient                   simpleradio.Client
	enableThreatMonitoring      bool
	threatMonitoringRequiresSRS bool
	threatMonitoringCooldown    time.Duration
	wasLastPictureClean         bool
}

func New(
	rdr radar.Radar,
	srsClient simpleradio.Client,
	coalition coalitions.Coalition,
	frequencies []simpleradio.RadioFrequency,
	pictureBroadcastInterval time.Duration,
	enableThreatMonitoring bool,
	threatMonitoringCooldown time.Duration,
	threatMonitoringRequiresSRS bool,
) Controller {
	return &controller{
		scope:                       rdr,
		coalition:                   coalition,
		frequencies:                 frequencies,
		pictureBroadcastInterval:    pictureBroadcastInterval,
		pictureBroadcastDeadline:    time.Now().Add(pictureBroadcastInterval),
		threatCooldowns:             newCooldownTracker(threatMonitoringCooldown),
		warmupTime:                  time.Now().Add(15 * time.Second),
		srsClient:                   srsClient,
		enableThreatMonitoring:      enableThreatMonitoring,
		threatMonitoringCooldown:    threatMonitoringCooldown,
		threatMonitoringRequiresSRS: threatMonitoringRequiresSRS,
	}
}

// Run implements [Controller.Run].
func (c *controller) Run(ctx context.Context, out chan<- any) {
	c.out = out

	log.Info().Msg("attaching FADED callback")
	c.scope.SetFadedCallback(func(group brevity.Group, coalition coalitions.Coalition) {
		for _, id := range group.ObjectIDs() {
			c.threatCooldowns.remove(id)
		}
		if coalition == c.coalition.Opposite() {
			group.SetDeclaration(brevity.Hostile)
			if c.srsClient.ClientsOnFrequency() > 0 {
				log.Info().Stringer("group", group).Msg("broadcasting FADED call")
				c.out <- brevity.FadedCall{Group: group}
			} else {
				log.Debug().Msg("skipping FADED call because no clients are on frequency")
			}
		}
	})

	log.Info().Msg("broadcasting SUNRISE call")
	frequencies := make([]unit.Frequency, 0, len(c.frequencies))
	for _, f := range c.frequencies {
		frequencies = append(frequencies, unit.Frequency(f.Frequency)*unit.Hertz)
	}
	log.Info().Any("cf", c.frequencies).Any("frequencies", frequencies).Msg("WTF")
	c.out <- brevity.SunriseCall{Frequencies: frequencies}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("detaching FADED callback")
			c.scope.SetFadedCallback(nil)
			return
		case <-ticker.C:
			if time.Now().After(c.pictureBroadcastDeadline) {
				logger := log.With().Logger()
				c.broadcastPicture(&logger, false)
			}
			c.broadcastThreats()
		}
	}
}
