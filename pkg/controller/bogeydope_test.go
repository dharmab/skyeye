package controller

import (
	"time"

	"github.com/DCS-gRPC/go-bindings/dcs/v0/common"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/require"
)

// TestHandleBogeyDopeClean tests the controller's handling of a BOGEY DOPE request when there are no red contacts in the airspace.
func (suite *ControllerTestSuite) TestHandleBogeyDopeClean() {
	callsign := "hornet 11"

	blueTrackfile := quickTrackFile(1, "hornet 11 | Sample", common.Coalition_COALITION_BLUE)
	blueTrackfile.Update(trackfile.Frame{
		Timestamp: time.Now(),
		Point:     orb.Point{0, 0},
		Altitude:  20000 * unit.Foot,
		Heading:   0 * unit.Degree,
		Speed:     300 * unit.Knot,
	})

	suite.radar.EXPECT().
		FindCallsign(callsign).
		Return(blueTrackfile)

	suite.radar.EXPECT().
		FindNearestGroup(
			blueTrackfile.LastKnown().Point,
			common.Coalition_COALITION_RED,
			brevity.Everything,
		).
		Return(nil)

	go suite.ctrl.HandleBogeyDope(&brevity.BogeyDopeRequest{Callsign: callsign})
	response := <-suite.outCh

	require.IsType(suite.T(), brevity.BogeyDopeResponse{}, response)
	require.Equal(suite.T(), callsign, response.(brevity.BogeyDopeResponse).Callsign)
	require.Nil(suite.T(), response.(brevity.BogeyDopeResponse).Group)
}
