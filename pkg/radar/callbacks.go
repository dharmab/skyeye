package radar

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	"github.com/paulmach/orb"
)

// StartedCallback is a callback function that is called when a new mission is started.
type StartedCallback func()

// SetStartedCallback sets the callback function that is called when a new mission is started.
func (r *Radar) SetStartedCallback(callback StartedCallback) {
	r.callbackLock.Lock()
	defer r.callbackLock.Unlock()
	r.startedCallback = callback
}

// FadedCallback is a callback function that is called when a group has not been updated by sensors for a timeout period.
// The group and its coalition are provided.
type FadedCallback func(location orb.Point, group brevity.Group, coalition coalitions.Coalition)

// SetFadedCallback sets the callback function to be called when a trackfile fades.
func (r *Radar) SetFadedCallback(callback FadedCallback) {
	r.callbackLock.Lock()
	defer r.callbackLock.Unlock()
	r.fadedCallback = callback
}

// RemovedCallback is a callback function that is called when a trackfile is aged out and removed.
// A copy of the trackfile is provided.
type RemovedCallback func(trackfile *trackfiles.Trackfile)

// SetRemovedCallback sets the callback function that is called when a trackfile is removed.
func (r *Radar) SetRemovedCallback(callback RemovedCallback) {
	r.callbackLock.Lock()
	defer r.callbackLock.Unlock()
	r.removalCallback = callback
}
