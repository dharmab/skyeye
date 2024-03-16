package dcs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/DCS-gRPC/go-bindings/dcs/v0/common"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/mission"
	measure "github.com/martinlindhe/unit"
)

type Sim interface {
}

type sim struct {
	missionClient mission.MissionServiceClient
}

var _ Sim = &sim{}

func NewSim(missionClient mission.MissionServiceClient) Sim {
	return &sim{
		missionClient: missionClient,
	}
}

const pollingInterval = 5 * time.Second

func (s *sim) StreamAircraft(ctx context.Context, aircraftChan chan<- Aircraft, fadedChan chan<- Faded) error {
	for _, category := range []common.GroupCategory{common.GroupCategory_GROUP_CATEGORY_AIRPLANE, common.GroupCategory_GROUP_CATEGORY_HELICOPTER} {
		go func() {
			err := s.StreamUnitCategory(ctx, category, aircraftChan, fadedChan)
			if err != nil {
				// TODO surface error
				slog.Error("error streaming units", "error", err)
			}
		}()
	}

	return nil
}

func (s *sim) StreamUnitCategory(ctx context.Context, category common.GroupCategory, aircraftChan chan<- Aircraft, fadedChan chan<- Faded) error {
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

func (s *sim) handleStreamUnitResponse(r *mission.StreamUnitsResponse, aircraftChan chan<- Aircraft, fadedChan chan<- Faded) {
	goneResp := r.GetGone()
	if goneResp != nil {
		fadedChan <- &faded{
			timestamp: time.Now(),
			unitID:    goneResp.GetId(),
		}
		return
	}
	unitResp := r.GetUnit()
	if unitResp == nil {
		slog.Warn("unable to handle nil unit response")
		return
	}

	position := unitResp.GetPosition()

	aircraftChan <- &aircraft{
		timestamp: time.Now(),
		unitID:    unitResp.GetId(),
		name:      unitResp.GetName(),
		coalition: unitResp.GetCoalition(),
		platform:  unitResp.GetType(),
		latitude:  measure.Angle(position.Lat) * measure.Degree,
		longitude: measure.Angle(position.Lon) * measure.Degree,
		altitude:  measure.Length(position.Alt) * measure.Meter,
		heading:   measure.Angle(unitResp.Velocity.Heading) * measure.Degree,
		speed:     measure.Speed(unitResp.Velocity.Speed) * measure.MetersPerSecond,
	}
}
