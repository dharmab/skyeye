// package controller implements high-level logic for Ground-Controlled Interception (GCI)
package controller

import (
	"context"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/rs/zerolog/log"
)

var lowestAltitude = unit.Length(0)
var highestAltitude = unit.Length(100000) * unit.Foot

// Controller handles requests for GCI service.
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

	log.Info().Msg("attaching FADED callback")
	c.scope.SetFadedCallback(func(group brevity.Group, coalition coalitions.Coalition) {
		if coalition == c.hostileCoalition() {
			group.SetDeclaration(brevity.Hostile)
			log.Info().Str("group", group.String()).Msg("broadcasting FADED call")
			c.out <- brevity.FadedCall{Group: group}
		}
	})

	log.Info().Int("frequency", int(c.frequency.Megahertz())).Msg("broadcasting SUNRISE call")
	c.out <- brevity.SunriseCall{Frequency: c.frequency}

	// TODO control loop for THREAT
	<-ctx.Done()
}

// SetTime implements [Controller.SetTime].
func (c *controller) SetTime(t time.Time) {
	c.scope.SetMissionTime(t)
}

// SetBullseye implements [Controller.SetBullseye].
func (c *controller) SetBullseye(bullseye orb.Point) {
	c.scope.SetBullseye(bullseye, c.coalition)
}

// hostileCoalition returns the coalition that is hostile to the controller's coalition.
func (c *controller) hostileCoalition() coalitions.Coalition {
	if c.coalition == coalitions.Blue {
		return coalitions.Red
	}
	return coalitions.Red
}
