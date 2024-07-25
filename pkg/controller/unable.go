package controller

import "github.com/dharmab/skyeye/pkg/brevity"

func (c *controller) HandleUnableToUnderstand(r *brevity.UnableToUnderstandRequest) {
	c.out <- brevity.SayAgainResponse{Callsign: r.Callsign}
}
