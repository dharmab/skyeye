package dcs

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/DCS-gRPC/go-bindings/dcs/v0/coalition"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/common"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/mission"
	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/trackfile"
	measure "github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
)

type Updated struct {
	Aircraft trackfile.Aircraft
	Frame    trackfile.Frame
}

type Faded struct {
	Timestamp time.Time
	UnitID    uint32
}

type Bullseye struct {
	Coalition types.Coalition
	Point     orb.Point
}

type Sim interface {
	// Stream aircraft updates from the sim to the provided channels.
	// The first channel receives updates for active aircraft.
	// The second channel receives messages when an aircraft disappears.
	// This function blocks until the context is cancelled.
	Stream(context.Context, chan<- Updated, chan<- Faded) error
	// Bullseye returns the coalition's bullseye center.
	Bullseye(context.Context) (*Bullseye, error)
}

type sim struct {
	missionClient   mission.MissionServiceClient
	coalitionClient coalition.CoalitionServiceClient
	coalition       common.Coalition
}

var _ Sim = &sim{}

func NewSim(missionClient mission.MissionServiceClient, coalitionClient coalition.CoalitionServiceClient, coalition common.Coalition) Sim {
	return &sim{
		missionClient:   missionClient,
		coalitionClient: coalitionClient,
		coalition:       coalition,
	}
}

const pollingInterval = 5 * time.Second

func (s *sim) Stream(ctx context.Context, aircraftChan chan<- Updated, fadedChan chan<- Faded) error {
	var wg sync.WaitGroup
	for _, category := range []common.GroupCategory{common.GroupCategory_GROUP_CATEGORY_AIRPLANE, common.GroupCategory_GROUP_CATEGORY_HELICOPTER} {
		wg.Add(1)
		go func() {
			err := s.StreamUnitCategory(ctx, category, aircraftChan, fadedChan)
			if err != nil {
				slog.Error("error streaming units", "error", err)
			}
			wg.Done()
		}()
	}

	<-ctx.Done()
	return nil
}

func (s *sim) StreamUnitCategory(ctx context.Context, category common.GroupCategory, aircraftChan chan<- Updated, fadedChan chan<- Faded) error {
	pollRate := uint32(pollingInterval.Seconds())
	streamRequest := mission.StreamUnitsRequest{
		Category:   category,
		PollRate:   &pollRate,
		MaxBackoff: &pollRate,
	}
	streamClient, err := s.missionClient.StreamUnits(ctx, &streamRequest)
	if err != nil {
		return fmt.Errorf("failed to stream units: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			response, err := streamClient.Recv()
			if err != nil {
				slog.Error("error receiving from stream", "error", err)
			} else {
				s.handleStreamUnitResponse(response, aircraftChan, fadedChan)
			}
		}
	}
}

func (s *sim) handleStreamUnitResponse(r *mission.StreamUnitsResponse, aircraftChan chan<- Updated, fadedChan chan<- Faded) {
	// Handle updates for faded aircraft
	goneResp := r.GetGone()
	if goneResp != nil {
		s.handleFaded(goneResp, fadedChan)
	} else if r.GetUnit() != nil {
		s.handleUpdate(r.GetUnit(), aircraftChan)
	} else {
		slog.Warn("unable to handle response", "response", r)
	}
}

func (s *sim) handleFaded(r *mission.StreamUnitsResponse_UnitGone, out chan<- Faded) {
	f := Faded{
		Timestamp: time.Now(),
		UnitID:    r.Id,
	}
	slog.Info("received faded aircraft update", "update", f)
	out <- f
}

func (s *sim) handleUpdate(u *common.Unit, out chan<- Updated) {
	var name string
	if u.PlayerName != nil && *u.PlayerName != "" {
		name = *u.PlayerName
	} else {
		name = fmt.Sprintf("%s ID%d", u.Name, u.Id)
	}

	position := u.Position
	point := orb.Point{position.Lon, position.Lat}
	alt := measure.Length(position.Alt) * measure.Meter
	hdg := measure.Angle(u.Velocity.Heading) * measure.Degree
	gs := measure.Speed(u.Velocity.Speed) * measure.MetersPerSecond

	var coalition types.Coalition
	if u.Coalition == common.Coalition_COALITION_BLUE {
		coalition = types.CoalitionBlue
	} else {
		coalition = types.CoalitionRed
	}

	update := Updated{
		Aircraft: trackfile.Aircraft{
			UnitID:     u.Id,
			Name:       name,
			Coalition:  coalition,
			EditorType: u.Type,
		},
		Frame: trackfile.Frame{
			Timestamp: time.Now(),
			Point:     point,
			Altitude:  alt,
			Heading:   hdg,
			Speed:     gs,
		},
	}
	slog.Info("received aircraft update", "update", update)
	out <- update
}

func (s *sim) Bullseye(ctx context.Context) (*Bullseye, error) {
	resp, err := s.coalitionClient.GetBullseye(ctx, &coalition.GetBullseyeRequest{
		Coalition: s.coalition,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get bullseye: %w", err)
	}
	position := resp.Position

	var coalition types.Coalition
	if s.coalition == common.Coalition_COALITION_BLUE {
		coalition = types.CoalitionBlue
	} else {
		coalition = types.CoalitionRed
	}

	return &Bullseye{
		Coalition: coalition,
		Point:     orb.Point{position.Lon, position.Lat},
	}, nil
}
