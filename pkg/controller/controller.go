// package controller implements high-level logic for Ground-Controlled Interception (GCI)
package controller

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	"github.com/dharmab/skyeye/pkg/trackfiles"
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
	Run(context.Context, chan<- any)
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
	// coalition this controller serves.
	coalition coalitions.Coalition

	// scope provides information about the airspace.
	scope radar.Radar

	// srsClient is used to to check if relevant friendly aircraft are on frequency before broadcasting calls.
	srsClient simpleradio.Client

	// warmupTime is when the controller is ready to broadcast tactical information. This provides time
	// for the radar scope to populate with data.
	warmupTime time.Time

	// pictureBroadcastInterval is the interval at which the controller broadcasts a tactical air picture.
	pictureBroadcastInterval time.Duration
	// pictureBroadcastDeadline is the time at which the controller will broadcast the next tactical air picture.
	pictureBroadcastDeadline time.Time
	// wasLastPictureClean tracks if the most recently broadcast picture was clean, so that the controller can avoid
	// repeatedly broadcasting clean pictures.
	wasLastPictureClean bool

	// enableThreatMonitoring enables automatic threat calls.
	enableThreatMonitoring bool
	// threatCooldowns tracks the next time a threat call should be published for each threat.
	threatCooldowns *cooldownTracker
	// threatMonitoringCooldown is the interval between threat calls for the same threat.
	threatMonitoringCooldown time.Duration
	// threatMonitoringRequiresSRS enforces that threat calls are only broadcast when the relevant friendly aircraft are on frequency.
	threatMonitoringRequiresSRS bool

	// merges tracks which contacts are in the merge.
	merges *mergeTracker

	// out is the channel to publish responses and calls to.
	out chan<- any
}

func New(
	rdr radar.Radar,
	srsClient simpleradio.Client,
	coalition coalitions.Coalition,
	pictureBroadcastInterval time.Duration,
	enableThreatMonitoring bool,
	threatMonitoringCooldown time.Duration,
	threatMonitoringRequiresSRS bool,
) Controller {
	return &controller{
		coalition:                   coalition,
		scope:                       rdr,
		srsClient:                   srsClient,
		warmupTime:                  time.Now().Add(15 * time.Second),
		pictureBroadcastInterval:    pictureBroadcastInterval,
		pictureBroadcastDeadline:    time.Now().Add(pictureBroadcastInterval),
		enableThreatMonitoring:      enableThreatMonitoring,
		threatMonitoringCooldown:    threatMonitoringCooldown,
		threatCooldowns:             newCooldownTracker(threatMonitoringCooldown),
		threatMonitoringRequiresSRS: threatMonitoringRequiresSRS,
		merges:                      newMergeTracker(),
	}
}

// Run implements [Controller.Run].
func (c *controller) Run(ctx context.Context, out chan<- any) {
	c.out = out

	log.Info().Msg("attaching callbacks")
	c.scope.SetFadedCallback(func(group brevity.Group, coalition coalitions.Coalition) {
		for _, id := range group.ObjectIDs() {
			c.remove(id)
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
	c.scope.SetRemovedCallback(func(trackfile trackfiles.Trackfile) {
		c.remove(trackfile.Contact.ID)
	})

	frequencies := make([]unit.Frequency, 0)
	for _, rf := range c.srsClient.Frequencies() {
		frequencies = append(frequencies, rf.Frequency)
	}
	c.out <- brevity.SunriseCall{Frequencies: frequencies}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("detaching callbacks")
			c.scope.SetFadedCallback(nil)
			c.scope.SetRemovedCallback(nil)
			return
		case <-ticker.C:
			c.broadcastMerges()
			c.broadcastThreats()
			if time.Now().After(c.pictureBroadcastDeadline) {
				logger := log.With().Logger()
				c.broadcastPicture(&logger, false)
			}
		}
	}
}

func (c *controller) remove(id uint64) {
	log.Debug().Uint64("id", id).Msg("removing ID from controller state tracking")
	c.threatCooldowns.remove(id)
	c.merges.remove(id)
}
