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

// TestHandleBogeyDopeNoContact tests the controller's handling of a BOGEY DOPE request when the requestor callsign is not found on the scope.
func (suite *ControllerTestSuite) TestHandleBogeyDopeNoContact() {
	callsign := "hornet 1 1"

	suite.radar.RunOnce()

	go suite.ctrl.HandleBogeyDope(&brevity.BogeyDopeRequest{Callsign: callsign, Filter: brevity.Everything})
	response := <-suite.outCh

	require.IsType(suite.T(), brevity.NegativeRadarContactResponse{}, response)
	require.Equal(suite.T(), callsign, response.(brevity.NegativeRadarContactResponse).Callsign)
}

// TestHandleBogeyDopeClean tests the controller's handling of a BOGEY DOPE request when there are no red contacts in the airspace.
func (suite *ControllerTestSuite) TestHandleBogeyDopeClean() {
	callsign := "hornet 1 1"
	requestor := blueHornet(1, "hornet 11 | Requestor")
	suite.updates <- dcs.Updated{
		Aircraft: requestor,
		Frame: trackfile.Frame{
			Timestamp: time.Now(),
			Point:     orb.Point{0, 0},
			Altitude:  20000 * unit.Foot,
			Heading:   0 * unit.Degree,
			Speed:     300 * unit.Knot,
		},
	}
	suite.radar.RunOnce()

	go suite.ctrl.HandleBogeyDope(&brevity.BogeyDopeRequest{Callsign: callsign})
	response := <-suite.outCh

	require.IsType(suite.T(), brevity.BogeyDopeResponse{}, response)
	require.Equal(suite.T(), callsign, response.(brevity.BogeyDopeResponse).Callsign)
	require.Nil(suite.T(), response.(brevity.BogeyDopeResponse).Group)
}

// TestHandleBogeyDopeSingleContact tests the case:
// - Blue requestor is near bulleye
// - A single red contact is 40 nmi north (outside threat range)
func (suite *ControllerTestSuite) TestHandleBogeyDopeSingleContact() {
	callsign := "hornet 1 1"
	requestor := blueHornet(1, "hornet 11 | Requestor")
	suite.updates <- dcs.Updated{
		Aircraft: requestor,
		Frame: trackfile.Frame{
			Timestamp: time.Now(),
			Point:     orb.Point{3, 0},
			Altitude:  20000 * unit.Foot,
			Heading:   0 * unit.Degree,
			Speed:     300 * unit.Knot,
		},
	}

	contact := redFlanker(2, "bug 11 | Enemy 1")
	suite.updates <- dcs.Updated{
		Aircraft: contact,
		Frame: trackfile.Frame{
			Timestamp: time.Now(),
			Point:     orb.Point{3, 40},
			Altitude:  18211 * unit.Foot,
			Heading:   180 * unit.Degree,
			Speed:     350 * unit.Knot,
		},
	}
	suite.radar.RunOnce()

	go suite.ctrl.HandleBogeyDope(&brevity.BogeyDopeRequest{Callsign: callsign})
	response := <-suite.outCh

	require.IsType(suite.T(), brevity.BogeyDopeResponse{}, response)
	require.Equal(suite.T(), callsign, response.(brevity.BogeyDopeResponse).Callsign)
	group := response.(brevity.BogeyDopeResponse).Group
	require.NotNil(suite.T(), group)
	require.False(suite.T(), group.Threat(), "group should not be classified as a threat")
	require.Equal(suite.T(), 1, group.Contacts())
	require.Equal(suite.T(), 18000*unit.Foot, group.Altitude()) // Rounds to nearest 1000 feet
	require.Equal(suite.T(), brevity.South, group.Track())
	require.Equal(suite.T(), brevity.Hot, group.Aspect())
	braa := group.BRAA()
	require.NotNil(suite.T(), braa)
	require.Equal(suite.T(), 0, braa.Bearing)
	require.Equal(suite.T(), 40, braa.Range)
	require.Equal(suite.T(), brevity.Hostile, group.Declaration())
	require.False(suite.T(), group.Heavy(), "group should not be classified as heavy")
	require.Equal(suite.T(), flankerReportingName, group.Platform())
	require.False(suite.T(), group.High(), "group should not be classified as high")
	require.False(suite.T(), group.Fast(), "group should not be classified as fast")
	require.False(suite.T(), group.VeryFast(), "group should not be classified as very fast")
}
