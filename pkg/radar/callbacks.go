package radar

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/paulmach/orb"
)

type StartedCallback func()

func (s *scope) SetStartedCallback(callback StartedCallback) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.startedCallback = callback
}

// FadedCallback is a callback function that is called when a group has not been updated by sensors for a timeout period.
// The group and its coalition are provided.
type FadedCallback func(location orb.Point, group brevity.Group, coalition coalitions.Coalition)

func (s *scope) SetFadedCallback(callback FadedCallback) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.fadedCallback = callback
}

// RemovedCallback is a callback function that is called when a trackfile is aged out and removed.
// A copy of the trackfile is provided.
type RemovedCallback func(trackfile *trackfiles.Trackfile)

func (s *scope) SetRemovedCallback(callback RemovedCallback) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.removalCallback = callback
}
