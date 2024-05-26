package controller

import (
	"context"
	"testing"

	"github.com/DCS-gRPC/go-bindings/dcs/v0/common"
	"github.com/dharmab/skyeye/pkg/dcs"
	"github.com/dharmab/skyeye/pkg/encyclopedia"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/trackfile"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	suite.Suite
	// Controller under test
	ctrl *controller
	// cancelFunc cancels the context, stopping the controller.
	cancelFunc context.CancelFunc
	// ctx is the context in which the controller runs.
	ctx context.Context
	// radar is the radar scope.
	radar radar.Radar
	// bullseyes is used to set the bullseye for the radar scope.
	bullseyes chan orb.Point
	// updates is used to send radar contacts to the radar scope.
	updates chan dcs.Updated
	// fades is used to fade radar contacts from the radar scope.
	fades chan dcs.Faded
	// outCh is the channel to which the controller publishes responses.
	outCh chan any
}

const (
	hornetPlatformDesignation  = "F/A-18"
	flankerPlatformDesignation = "Su-27"
)

var (
	hornetEditorType     = encyclopedia.New().AircraftByPlatformDesignation(hornetPlatformDesignation)[0].EditorType
	flankerEditorType    = encyclopedia.New().AircraftByPlatformDesignation(flankerPlatformDesignation)[0].EditorType
	flankerReportingName = encyclopedia.New().AircraftByPlatformDesignation(flankerPlatformDesignation)[0].NATOReportingName
)

func (suite *ControllerTestSuite) SetupSuite() {
	suite.updates = make(chan dcs.Updated, 16)
	suite.fades = make(chan dcs.Faded, 16)
	suite.bullseyes = make(chan orb.Point, 4)

	suite.radar = radar.New(suite.bullseyes, suite.updates, suite.fades)
	suite.ctrl = New(suite.radar, common.Coalition_COALITION_BLUE).(*controller)
	suite.outCh = make(chan any)
	ctx, cancel := context.WithCancel(context.Background())
	suite.ctx = ctx
	suite.cancelFunc = cancel

	go suite.ctrl.Run(suite.ctx, suite.outCh)
}

func (suite *ControllerTestSuite) TearDownSuite() {
	suite.cancelFunc()
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func blueHornet(unitID uint32, name string) trackfile.Aircraft {
	return trackfile.Aircraft{
		UnitID:     unitID,
		Name:       name,
		Coalition:  types.CoalitionBlue,
		EditorType: hornetEditorType,
	}
}

func redFlanker(unitID uint32, name string) trackfile.Aircraft {
	return trackfile.Aircraft{
		UnitID:     unitID,
		Name:       name,
		Coalition:  types.CoalitionRed,
		EditorType: flankerEditorType,
	}
}
