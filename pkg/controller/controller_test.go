package controller

import (
	"context"
	"testing"

	"github.com/DCS-gRPC/go-bindings/dcs/v0/common"
	"github.com/dharmab/skyeye/pkg/mocks"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type ControllerTestSuite struct {
	suite.Suite
	// Controller under test
	ctrl *controller
	// cancelFunc cancels the context, stopping the controller.
	cancelFunc context.CancelFunc
	// ctx is the context in which the controller runs.
	ctx context.Context
	// mctrl is the gomock controller.
	mctrl *gomock.Controller
	// radar is the mock radar scope.
	radar *mocks.MockRadar
	// outCh is the channel to which the controller publishes responses.
	outCh chan any
}

func (suite *ControllerTestSuite) SetupSuite() {
	suite.mctrl = gomock.NewController(suite.T())
	suite.radar = mocks.NewMockRadar(suite.mctrl)
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
