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

// TestHandleBogeyDopeNoContact tests the controller's handling of a BOGEY DOPE request when the requestor callsign is not found on the scope.
func (suite *ControllerTestSuite) TestHandleBogeyDopeNoContact() {
	callsign := "hornet 11"

	suite.radar.EXPECT().
		FindCallsign(callsign).
		Return(nil)

	go suite.ctrl.HandleBogeyDope(&brevity.BogeyDopeRequest{Callsign: callsign})
	response := <-suite.outCh

	require.IsType(suite.T(), brevity.BogeyDopeResponse{}, response)
	require.Equal(suite.T(), callsign, response.(brevity.BogeyDopeResponse).Callsign)
	require.Nil(suite.T(), response.(brevity.BogeyDopeResponse).Group)
}

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

// TestHandleBogeyDopeSingleContact tests the case:
// - Blue requestor is at origin
// - A single red contact is at 0, 40
func (suite *ControllerTestSuite) TestHandleBogeyDopeSingleContact() {
	callsign := "hornet 11"

	blueTrackfile := quickTrackFile(1, "hornet 11 | Requestor", common.Coalition_COALITION_BLUE)
	blueTrackfile.Update(trackfile.Frame{
		Timestamp: time.Now(),
		Point:     orb.Point{0, 0},
		Altitude:  20000 * unit.Foot,
		Heading:   0 * unit.Degree,
		Speed:     300 * unit.Knot,
	})

	redTrackfile := quickTrackFile(2, "bug 11 | Enemy 1", common.Coalition_COALITION_RED)
	redTrackfile.Contact.EditorType = flankerEditorType
	redTrackfile.Update(trackfile.Frame{
		Timestamp: time.Now(),
		Point:     orb.Point{0, 40},
		Altitude:  18211 * unit.Foot,
		Heading:   180 * unit.Degree,
		Speed:     350 * unit.Knot,
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
		Return(redTrackfile)

	go suite.ctrl.HandleBogeyDope(&brevity.BogeyDopeRequest{Callsign: callsign})
	response := <-suite.outCh

	require.IsType(suite.T(), brevity.BogeyDopeResponse{}, response)
	require.Equal(suite.T(), callsign, response.(brevity.BogeyDopeResponse).Callsign)
	group := response.(brevity.BogeyDopeResponse).Group
	require.NotNil(suite.T(), group)
	require.False(suite.T(), group.Threat())
	require.Equal(suite.T(), 1, group.Contacts())
	require.Nil(suite.T(), group.Bullseye())
	require.Equal(suite.T(), 18000*unit.Foot, group.Altitude())
	require.Equal(suite.T(), brevity.South, group.Track())
	require.Equal(suite.T(), brevity.Hot, group.Aspect())
	braa := group.BRAA()
	require.NotNil(suite.T(), braa)
	require.Equal(suite.T(), 0, braa.Bearing)
	require.Equal(suite.T(), 40, braa.Range)
	require.Equal(suite.T(), brevity.Hostile, group.Declaration())
	require.False(suite.T(), group.Heavy())
	require.Equal(suite.T(), flankerReportingName, group.Platform())
	require.False(suite.T(), group.High())
	require.False(suite.T(), group.Fast())
	require.False(suite.T(), group.VeryFast())
}
