package controller

import (
	"context"

	"github.com/dharmab/skyeye/pkg/brevity"
)

type Controller interface {
	// Run starts the controller's control loops. It should be called exactly once. It blocks until the context is canceled.
	Run(context.Context)
	// HandleAlphaCheck handles an ALPHA CHECK by reporting the position of the requesting aircraft.
	HandleAlphaCheck(brevity.AlphaCheckRequest)
	// HandleBogeyDope handles a BOGEY DOPE by reporting the closest enemy group to the requesting aircraft.
	HandleBogeyDope(brevity.BogeyDopeRequest)
	// HandleDeclare handles a DECLARE by reporting information about the target group.
	HandleDeclare(brevity.DeclareRequest)
	// HandlePicture handles a PICTURE by reporting a tactical air picture.
	HandlePicture(brevity.PictureRequest)
	// HandleRadioCheck handles a RADIO CHECK by responding to the requesting aircraft.
	HandleRadioCheck(brevity.RadioCheckRequest)
	// HandleSnaplock handles a SNAPLOCK by reporting information about the target group.
	HandleSnaplock(brevity.SnaplockRequest)
	// HandleSpiked handles a SPIKED by reporting any enemy groups in the direction of the radar spike.
	HandleSpiked(brevity.SpikedRequest)
	// ResponseChannel returns a channel that the controller queues calls and responses on. The types received on this channel are Call and Response types defined in the brevity package.
	ResponseChannel() <-chan any
}

type controller struct {
	responseChan chan any
}

func New() Controller {
	return &controller{
		responseChan: make(chan any),
	}
}

// Run implements Controller.Run.
func (c *controller) Run(context.Context) {}

// ResponseChannel implements Controller.ResponseChannel.
func (c *controller) ResponseChannel() <-chan any {
	return c.responseChan
}
