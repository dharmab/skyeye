package controller

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

type Controller interface {
	// Run starts the controller's control loops. It should be called exactly once. It blocks until the context is canceled.
	// The controller publishes responses to the given channel.
	Run(context.Context, chan<- any)
	// SetTime updates the mission time used for computing magnetic declination.
	SetTime(time.Time)
	// SetBullseye updates the bullseye point.
	SetBullseye(orb.Point)
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
	// HandleUnableToUnderstand handles requests where the wake word was recognized but the request could not be understood, by asking players on the channel to repeat their message.
	HandleUnableToUnderstand(*brevity.UnableToUnderstandRequest)
}

type controller struct {
	out       chan<- any
	scope     radar.Radar
	coalition coalitions.Coalition
	frequency unit.Frequency
}

func New(rdr radar.Radar, coalition coalitions.Coalition, frequency unit.Frequency) Controller {
	return &controller{
		scope:     rdr,
		coalition: coalition,
		frequency: frequency,
	}
}

// Run implements [Controller.Run].
func (c *controller) Run(ctx context.Context, out chan<- any) {
	c.out = out

	log.Info().Int("frequency", int(c.frequency.Megahertz())).Msg("broadcasting SUNRISE call")
	c.out <- brevity.SunriseCall{Frequency: c.frequency}
	<-ctx.Done()

	// TODO control loops for FADED and THREAT
}

// SetTime implements [Controller.SetTime].
func (c *controller) SetTime(t time.Time) {
	c.scope.SetMissionTime(t)
}

// SetBullseye implements [Controller.SetBullseye].
func (c *controller) SetBullseye(bullseye orb.Point) {
	c.scope.SetBullseye(bullseye)
}

// hostileCoalition returns the coalition that is hostile to the controller's coalition.
func (c *controller) hostileCoalition() coalitions.Coalition {
	if c.coalition == coalitions.Blue {
		return coalitions.Red
	}
	return coalitions.Red
}

// findCallsign searches the controller's scope for a trackfile matching the given callsign.
func (c *controller) findCallsign(callsign string) *trackfiles.Trackfile {
	logger := log.With().Str("callsign", callsign).Logger()
	logger.Debug().Msg("searching scope for trackfile matching callsign")
	trackfile := c.scope.FindCallsign(callsign)
	if trackfile == nil {
		logger.Debug().Msg("no trackfile found for callsign")
	} else {
		logger.Debug().Str("name", trackfile.Contact.Name).Str("aircraft", trackfile.Contact.ACMIName).Int("unitID", int(trackfile.Contact.UnitID)).Msg("trackfile found for callsign")
	}
	return trackfile
}
