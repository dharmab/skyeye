// package controller implements high-level logic for Ground-Controlled Interception (GCI)
package controller

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/dharmab/skyeye/pkg/simpleradio"
	"github.com/dharmab/skyeye/pkg/traces"
	"github.com/lithammer/shortuuid/v3"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

var (
	lowestAltitude  = unit.Length(0)
	highestAltitude = unit.Length(100000) * unit.Foot
)

type Call struct {
	Context context.Context
	Call    any
}

func NewCall(ctx context.Context, call any) Call {
	return Call{
		Context: ctx,
		Call:    call,
	}
}

// Controller handles requests for GCI service.
type Controller interface {
	// Run starts the controller's control loops. It should be called exactly once. It blocks until the context is canceled.
	// The controller publishes responses to the given channel.
	Run(ctx context.Context, out chan<- Call)
	// HandleAlphaCheck handles an ALPHA CHECK by reporting the position of the requesting aircraft.
	HandleAlphaCheck(context.Context, *brevity.AlphaCheckRequest)
	// HandleBogeyDope handles a BOGEY DOPE by reporting the closest enemy group to the requesting aircraft.
	HandleBogeyDope(context.Context, *brevity.BogeyDopeRequest)
	// HandleDeclare handles a DECLARE by reporting information about the target group.
	HandleDeclare(context.Context, *brevity.DeclareRequest)
	// HandlePicture handles a PICTURE by reporting a tactical air picture.
	HandlePicture(context.Context, *brevity.PictureRequest)
	// HandleRadioCheck handles a RADIO CHECK by responding to the requesting aircraft.
	HandleRadioCheck(context.Context, *brevity.RadioCheckRequest)
	// HandleSnaplock handles a SNAPLOCK by reporting information about the target group.
	HandleSnaplock(context.Context, *brevity.SnaplockRequest)
	// HandleSpiked handles a SPIKED by reporting any enemy groups in the direction of the radar spike.
	HandleSpiked(context.Context, *brevity.SpikedRequest)
	// HandleTripwire handles a TRIPWIRE... by not implementing it LOL
	HandleTripwire(context.Context, *brevity.TripwireRequest)
	// HandleUnableToUnderstand handles requests where the wake word was recognized but the request could not be understood, by asking players on the channel to repeat their message.
	HandleUnableToUnderstand(context.Context, *brevity.UnableToUnderstandRequest)
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

	// enableAutomaticPicture enables automatic picture broadcasts.
	enableAutomaticPicture bool
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

	// calls is the channel to publish responses and calls to.
	calls chan<- Call
}

func New(
	rdr radar.Radar,
	srsClient simpleradio.Client,
	coalition coalitions.Coalition,
	enableAutomaticPicture bool,
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
		enableAutomaticPicture:      enableAutomaticPicture,
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
func (c *controller) Run(ctx context.Context, calls chan<- Call) {
	c.calls = calls

	log.Info().Msg("attaching callbacks")
	c.scope.SetFadedCallback(c.handleFaded)
	c.scope.SetRemovedCallback(c.handleRemoved)

	c.broadcastSunrise(ctx)

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
			c.broadcastMerges(traces.WithTraceID(ctx, shortuuid.New()))
			c.broadcastThreats(traces.WithTraceID(ctx, shortuuid.New()))
			if c.enableAutomaticPicture && time.Now().After(c.pictureBroadcastDeadline) {
				c.broadcastPicture(traces.WithTraceID(ctx, shortuuid.New()), &log.Logger, false)
			}
		}
	}
}

func (c *controller) broadcastSunrise(ctx context.Context) {
	frequencies := make([]unit.Frequency, 0)
	for _, rf := range c.srsClient.Frequencies() {
		frequencies = append(frequencies, rf.Frequency)
	}
	c.calls <- NewCall(traces.WithTraceID(ctx, shortuuid.New()), brevity.SunriseCall{Frequencies: frequencies})
}

func (c *controller) remove(id uint64) {
	log.Debug().Uint64("id", id).Msg("removing ID from controller state tracking")
	c.threatCooldowns.remove(id)
	c.merges.remove(id)
}
