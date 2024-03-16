package controller

import (
	"github.com/dharmab/skyeye/pkg/brevity"
)

// HandleAlphaCheck implements Controller.HandleAlphaCheck.
func (c *controller) HandleAlphaCheck(r *brevity.AlphaCheckRequest) {
	tf := c.scope.FindCallsign(r.Callsign)
	if tf == nil {
		c.out <- brevity.AlphaCheckResponse{
			Callsign: r.Callsign,
			Status:   false,
		}
	}
	location := tf.Bullseye(*c.scope.GetBullseye())
	c.out <- brevity.AlphaCheckResponse{
		Callsign: r.Callsign,
		Status:   true,
		Location: location,
	}
}
