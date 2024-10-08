package radar

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/trackfiles"
)

// FadedCallback is a callback function that is called when a group has not been updated by sensors for a timeout period.
// The group and its coalition are provided.
type FadedCallback func(group brevity.Group, coalition coalitions.Coalition)

func (s *scope) SetFadedCallback(callback FadedCallback) {
	s.fadedCallback = callback
}

// RemovedCallback is a callback function that is called when a trackfile is aged out and removed.
// A copy of the trackfile is provided.
type RemovedCallback func(trackfile trackfiles.Trackfile)

func (s *scope) SetRemovedCallback(callback RemovedCallback) {
	s.removalCallback = callback
}
