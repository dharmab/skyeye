package controller

import (
	"slices"

	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/rs/zerolog/log"
)

func (c *controller) addFriendlyToBroadcast(callsigns []string, friendly *trackfiles.Trackfile) []string {
	logger := log.With().Str("callsign", friendly.Contact.Name).Logger()
	isOnFrequency := c.srsClient.IsOnFrequency(friendly.Contact.Name)
	if isOnFrequency {
		logger.Debug().Bool("isOnFrequency", isOnFrequency).Msg("friendly contact is on frequency")
	}

	shouldBroadcast := !c.threatMonitoringRequiresSRS || isOnFrequency
	if !shouldBroadcast {
		return callsigns
	}
	if callsign, ok := parser.ParsePilotCallsign(friendly.Contact.Name); ok {
		if !slices.Contains(callsigns, callsign) {
			callsigns = append(callsigns, callsign)
		}
	} else {
		logger.Debug().Msg("could not parse callsign")
	}
	return callsigns
}
