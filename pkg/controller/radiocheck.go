package controller

import "github.com/dharmab/skyeye/pkg/brevity"

// HandleRadioCheck implements Controller.HandleRadioCheck.
func (c *controller) HandleRadioCheck(r *brevity.RadioCheckRequest) {
	tf := c.scope.FindCallsign(r.Callsign)
	c.out <- brevity.RadioCheckResponse{Callsign: r.Callsign, Status: tf == nil}
}
