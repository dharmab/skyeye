package controller

import (
	"github.com/DCS-gRPC/go-bindings/dcs/v0/common"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/stretchr/testify/require"
)

// TestHandleRadioCheckFailure tests the controller's handling of a RADIO CHECK request for a callsign not found on the scope.
func (suite *ControllerTestSuite) TestHandleRadioCheckFailure() {
	callsign := "hornet 11"
	suite.radar.EXPECT().FindCallsign(callsign).Return(nil)

	go suite.ctrl.HandleRadioCheck(&brevity.RadioCheckRequest{Callsign: callsign})
	response := <-suite.outCh

	require.IsType(suite.T(), brevity.RadioCheckResponse{}, response)
	require.Equal(suite.T(), callsign, response.(brevity.RadioCheckResponse).Callsign)
	require.False(suite.T(), response.(brevity.RadioCheckResponse).Status)
}

// TestHandleRadioCheckSuccess tests the controller's handling of a RADIO CHECK request for a callsign found on the scope.
func (suite *ControllerTestSuite) TestHandleRadioCheckSuccess() {
	editorType := encyclopedia.New().AircraftByPlatformDesignation("F/A-18")[0].EditorType
	callsign := "hornet 11"

	suite.radar.EXPECT().FindCallsign(callsign).Return(&trackfile.Trackfile{
		Contact: trackfile.Aircraft{
			UnitID:     1,
			Name:       "hornet 11 | Sample",
			Coalition:  common.Coalition_COALITION_BLUE,
			EditorType: editorType,
		},
	})

	go suite.ctrl.HandleRadioCheck(&brevity.RadioCheckRequest{Callsign: callsign})
	response := <-suite.outCh

	require.IsType(suite.T(), brevity.RadioCheckResponse{}, response)
	require.Equal(suite.T(), callsign, response.(brevity.RadioCheckResponse).Callsign)
	require.True(suite.T(), response.(brevity.RadioCheckResponse).Status)
}
