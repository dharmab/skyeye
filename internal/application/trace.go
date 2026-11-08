package application

import (
	"context"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/traces"
)

// traceCallKey is a context value key used to store the brevity call or
// response for filtering purposes.
type traceCallKey struct{}

func (a *Application) shouldTrace(ctx context.Context) bool {
	switch ctx.Value(traceCallKey{}).(type) {
	case brevity.PictureResponse:
		// A player-requested PICTURE retains its request in the context;
		// only unsolicited broadcasts are subject to the automatic tracing toggle.
		isPlayerRequestResponse := traces.GetRequest(ctx) != nil
		return isPlayerRequestResponse || a.enableAutomaticPictureTracing
	case brevity.FadedCall:
		return a.enableAutomaticFadedTracing
	case brevity.ThreatCall:
		return a.enableAutomaticThreatTracing
	default:
		return true
	}
}
