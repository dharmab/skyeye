package radar

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
)

// FadedCallback is a callback function that is called when a group has not been updated by sensors for a timeout period.
// The group and it's coalition are provided.
type FadedCallback func(group brevity.Group, coalition coalitions.Coalition)

func (s *scope) SetFadedCallback(callback FadedCallback) {
	s.fadedCallback = callback
}
