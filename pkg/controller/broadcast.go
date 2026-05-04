package controller

import (
	"slices"

	"github.com/dharmab/skyeye/pkg/callsigns"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/rs/zerolog/log"
)

func (c *Controller) addFriendlyToBroadcast(friendlyCallsigns []string, friendly *trackfiles.Trackfile) []string {
	logger := log.With().Str("callsign", friendly.Contact.Name).Logger()
	isOnFrequency := c.srsClient.IsOnFrequency(friendly.Contact.Name)
	if isOnFrequency {
		logger.Debug().Bool("isOnFrequency", isOnFrequency).Msg("friendly contact is on frequency")
	}

	shouldBroadcast := !c.threatMonitoringRequiresSRS || isOnFrequency
	if !shouldBroadcast {
		return friendlyCallsigns
	}
	if callsign, ok := callsigns.ParsePilotCallsign(friendly.Contact.Name); ok {
		if !slices.Contains(friendlyCallsigns, callsign) {
			friendlyCallsigns = append(friendlyCallsigns, callsign)
		}
	} else {
		logger.Debug().Msg("could not parse callsign")
	}
	return friendlyCallsigns
}
