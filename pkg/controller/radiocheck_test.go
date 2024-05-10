package controller

import (
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/dcs"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/require"
)

// TestHandleRadioCheckFailure tests the controller's handling of a RADIO CHECK request for a callsign not found on the scope.
func (suite *ControllerTestSuite) TestHandleRadioCheckFailure() {
	callsign := "hornet 11"

	suite.radar.RunOnce()
	go suite.ctrl.HandleRadioCheck(&brevity.RadioCheckRequest{Callsign: callsign})
	response := <-suite.outCh

	require.IsType(suite.T(), brevity.RadioCheckResponse{}, response)
	require.Equal(suite.T(), callsign, response.(brevity.RadioCheckResponse).Callsign)
	require.False(suite.T(), response.(brevity.RadioCheckResponse).Status)
}

// TestHandleRadioCheckSuccess tests the controller's handling of a RADIO CHECK request for a callsign found on the scope.
func (suite *ControllerTestSuite) TestHandleRadioCheckSuccess() {
	callsign := "hornet 1 1"

	requestor := blueHornet(1, "hornet 11 | Requestor")
	suite.updates <- dcs.Updated{
		Aircraft: requestor,
		Frame: trackfile.Frame{
			Timestamp: time.Now(),
			Point:     orb.Point{3, 3},
			Altitude:  20000 * unit.Foot,
			Heading:   0 * unit.Degree,
			Speed:     300 * unit.Knot,
		},
	}
	suite.radar.RunOnce()

	go suite.ctrl.HandleRadioCheck(&brevity.RadioCheckRequest{Callsign: callsign})
	response := <-suite.outCh

	require.IsType(suite.T(), brevity.RadioCheckResponse{}, response)
	require.Equal(suite.T(), callsign, response.(brevity.RadioCheckResponse).Callsign)
	require.True(suite.T(), response.(brevity.RadioCheckResponse).Status)
}
