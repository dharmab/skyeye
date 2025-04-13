// Package controller implements high-level logic for Ground-Controlled Interception (GCI).
package controller

import (
	"context"
	"sync"
	"time"

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

// Call is an envelope for a GCI call.
type Call struct {
	Context context.Context
	Call    any
}

// NewCall creates a new Call message.
func NewCall(ctx context.Context, call any) Call {
	return Call{
		Context: ctx,
		Call:    call,
	}
}

// Controller handles requests for GCI service.
type Controller struct {
	// coalition this controller serves.
	coalition coalitions.Coalition

	// scope provides information about the airspace.
	scope *radar.Radar

	// srsClient is used to check if relevant friendly aircraft are on frequency before broadcasting calls.
	srsClient *simpleradio.Client

	// enableAutomaticPicture enables automatic picture broadcasts.
	enableAutomaticPicture bool
	// pictureBroadcastInterval is the interval at which the controller broadcasts a tactical air picture.
	pictureBroadcastInterval time.Duration
	// pictureBroadcastDeadline is the time at which the controller will broadcast the next tactical air picture.
	pictureBroadcastDeadline time.Time
	// pictureBroadcastDeadlineLock protects pictureBroadcastDeadline.
	pictureBroadcastDeadlineLock sync.RWMutex
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
	// mergeCooldowns tracks the next time a merge call may be published for each friendly.
	mergeCooldowns *cooldownTracker

	// calls is the channel to publish responses and calls to.
	calls chan<- Call
}

// New creates a new GCI controller.
func New(
	rdr *radar.Radar,
	srsClient *simpleradio.Client,
	coalition coalitions.Coalition,
	enableAutomaticPicture bool,
	pictureBroadcastInterval time.Duration,
	enableThreatMonitoring bool,
	threatMonitoringCooldown time.Duration,
	threatMonitoringRequiresSRS bool,
) *Controller {
	return &Controller{
		coalition:                   coalition,
		scope:                       rdr,
		srsClient:                   srsClient,
		enableAutomaticPicture:      enableAutomaticPicture,
		pictureBroadcastInterval:    pictureBroadcastInterval,
		pictureBroadcastDeadline:    time.Now().Add(pictureBroadcastInterval),
		enableThreatMonitoring:      enableThreatMonitoring,
		threatMonitoringCooldown:    threatMonitoringCooldown,
		threatCooldowns:             newCooldownTracker(threatMonitoringCooldown),
		threatMonitoringRequiresSRS: threatMonitoringRequiresSRS,
		merges:                      newMergeTracker(),
		mergeCooldowns:              newCooldownTracker(30 * time.Second),
	}
}

// Run starts the controller's control loops. It should be called exactly once. It blocks until the context is canceled.
// The controller publishes responses to the given channel.
func (c *Controller) Run(ctx context.Context, calls chan<- Call) {
	c.calls = calls

	log.Info().Msg("attaching callbacks")
	c.scope.SetFadedCallback(c.handleFaded)
	c.scope.SetRemovedCallback(c.handleRemoved)
	c.scope.SetStartedCallback(c.handleStarted)

	c.broadcastSunrise(ctx)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("detaching callbacks")
			c.scope.SetFadedCallback(nil)
			c.scope.SetRemovedCallback(nil)
			c.scope.SetStartedCallback(nil)
			return
		case <-ticker.C:
			c.broadcastMerges(traces.WithTraceID(ctx, shortuuid.New()))
			c.broadcastThreats(traces.WithTraceID(ctx, shortuuid.New()))
			c.broadcastAutomaticPicture(traces.WithTraceID(ctx, shortuuid.New()))
		}
	}
}
