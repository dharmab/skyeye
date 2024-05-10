package controller

import (
	"context"
	"log/slog"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/trackfile"
)

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
}

type controller struct {
	out        chan<- any
	scope      radar.Radar
	coalition  types.Coalition
	trackfiles map[string]*trackfile.Trackfile
}

func New(r radar.Radar, coalition types.Coalition) Controller {
	return &controller{
		scope:      r,
		coalition:  coalition,
		trackfiles: make(map[string]*trackfile.Trackfile),
	}
}

// Run implements Controller.Run.
func (c *controller) Run(ctx context.Context, out chan<- any) {
	c.out = out

	gcTicker := time.NewTicker(30 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return
		case <-gcTicker.C:
			c.expireTrackfiles()
		}
	}

	// TODO control loops for FADED and THREAT
}

func (c *controller) hostileCoalition() types.Coalition {
	if c.coalition == types.CoalitionBlue {
		return types.CoalitionRed
	}
	return types.CoalitionRed
}

func (c *controller) expireTrackfiles() {
	for name, tf := range c.trackfiles {
		if tf.LastKnown().Timestamp.Before(time.Now().Add(3 * time.Minute)) {
			slog.Info("removing aged out trackfile", "name", name, "last_updated", tf.LastKnown().Timestamp)
			delete(c.trackfiles, name)
		}
	}
}
