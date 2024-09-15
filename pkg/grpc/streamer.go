package grpc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/DCS-gRPC/go-bindings/dcs/v0/coalition"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/common"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/mission"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/unit"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/sim"
	"github.com/dharmab/skyeye/pkg/trackfiles"
	measures "github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/relvacode/iso8601"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type Streamer struct {
	missionService   mission.MissionServiceClient
	coalitionService coalition.CoalitionServiceClient
	unitService      unit.UnitServiceClient
	interval         time.Duration
	missonStartTime  time.Time
}

var _ sim.Sim = &Streamer{}

func NewStreamer(address string, interval time.Duration) (*Streamer, error) {
	grpcClient, err := grpc.NewClient(address)
	if err != nil {
		return nil, err
	}
	return &Streamer{
		missionService:   mission.NewMissionServiceClient(grpcClient),
		coalitionService: coalition.NewCoalitionServiceClient(grpcClient),
		unitService:      unit.NewUnitServiceClient(grpcClient),
		interval:         interval,
	}, nil
}

func (s *Streamer) Stream(ctx context.Context, starts chan<- sim.Started, updates chan<- sim.Updated, fades chan<- sim.Faded) {
	for {
		err := s.setStartTime(ctx)
		if err != nil {
			log.Error().Err(err).Msg("failed to set mission start time")
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	var wg sync.WaitGroup
	streamCtx, cancel := context.WithCancel(ctx)
	wg.Add(3)
	go func() {
		defer wg.Done()
		defer cancel()
		s.streamMissionStartEvents(streamCtx, starts)
	}()
	go func() {
		defer wg.Done()
		defer cancel()
		s.streamCategory(streamCtx, common.GroupCategory_GROUP_CATEGORY_AIRPLANE, updates, fades)
	}()
	go func() {
		defer wg.Done()
		defer cancel()
		s.streamCategory(streamCtx, common.GroupCategory_GROUP_CATEGORY_HELICOPTER, updates, fades)
	}()
	wg.Wait()
}

func (s *Streamer) setStartTime(ctx context.Context) error {
	request := &mission.GetScenarioCurrentTimeRequest{}
	response, err := s.missionService.GetScenarioCurrentTime(ctx, request)
	if err != nil {
		return err
	}
	s.missonStartTime, err = iso8601.ParseString(response.GetDatetime())
	if err != nil {
		return err
	}
	return nil
}

func (s *Streamer) streamMissionStartEvents(ctx context.Context, starts chan<- sim.Started) {
	request := &mission.StreamEventsRequest{}
	stream, err := s.missionService.StreamEvents(ctx, request)
	if err != nil {
		log.Error().Err(err).Msg("failed to stream events")
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			response, err := stream.Recv()
			if err != nil {
				log.Error().Err(err).Msg("received error from events stream")
				return
			}
			if response.GetMissionStart() != nil {
				err := s.setStartTime(ctx)
				if err != nil {
					log.Error().Err(err).Msg("failed to update mission start time")
					continue
				}
				starts <- sim.Started{
					Timestamp:        time.Now(),
					MissionTimestamp: s.missonStartTime.Add(time.Duration(response.GetTime()) * time.Second),
				}
			}
		}
	}
}

func (s *Streamer) streamCategory(ctx context.Context, category common.GroupCategory, updates chan<- sim.Updated, fades chan<- sim.Faded) {
	rate := uint32(s.interval.Seconds())
	request := &mission.StreamUnitsRequest{
		PollRate:   &rate,
		MaxBackoff: &rate,
		Category:   category,
	}
	stream, err := s.missionService.StreamUnits(ctx, request)
	if err != nil {
		log.Error().Err(err).Msg("failed to stream units")
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			response, err := stream.Recv()
			if err != nil {
				log.Error().Err(err).Msg("received error from units stream")
				return
			}
			missionTimestamp := s.missonStartTime.Add(time.Duration(response.GetTime()) * time.Second)
			logger := log.With().Time("missionTimestamp", missionTimestamp).Logger()
			gone := response.GetGone()
			if gone != nil {
				fades <- sim.Faded{
					Timestamp:        time.Now(),
					MissionTimestamp: missionTimestamp,
					ID:               uint64(gone.GetId()),
				}
			} else {
				aircraft := response.GetUnit()
				logger = logger.With().Uint32("id", aircraft.GetId()).Str("name", aircraft.GetName()).Logger()
				labels := trackfiles.Labels{
					ID:       uint64(aircraft.GetId()),
					Name:     aircraft.GetName(),
					ACMIName: aircraft.GetType(),
				}

				if aircraft.GetCoalition() == common.Coalition_COALITION_BLUE {
					labels.Coalition = coalitions.Blue
				} else if aircraft.GetCoalition() == common.Coalition_COALITION_RED {
					labels.Coalition = coalitions.Red
				} else {
					logger.Warn().Msgf("unit coalition not recognized: %v", aircraft.GetCoalition())
					continue
				}

				position := aircraft.GetPosition()
				if position == nil {
					logger.Warn().Msg("unit has no position")
					continue
				}
				frame := trackfiles.Frame{
					Time:  missionTimestamp,
					Point: orb.Point{position.GetLon(), position.GetLat()},
				}
				orientation := aircraft.GetOrientation()
				if orientation == nil {
					logger.Warn().Msg("unit has no orientation")
					continue
				}
				frame.Heading = measures.Angle(orientation.GetHeading()) * measures.Degree
				updates <- sim.Updated{Labels: labels, Frame: frame}
			}
		}
	}
}

func (s *Streamer) Bullseye(side coalitions.Coalition) (orb.Point, error) {
	request := &coalition.GetBullseyeRequest{}
	switch side {
	case coalitions.Blue:
		request.Coalition = common.Coalition_COALITION_BLUE
	case coalitions.Red:
		request.Coalition = common.Coalition_COALITION_RED
	default:
		return orb.Point{}, fmt.Errorf("unknown coalition: %v", side)
	}
	response, err := s.coalitionService.GetBullseye(context.Background(), request)
	if err != nil {
		return orb.Point{}, err
	}
	position := response.GetPosition()
	if position == nil {
		return orb.Point{}, errors.New("bullseye has no position")
	}
	bullseye := orb.Point{position.GetLon(), position.GetLat()}
	return bullseye, nil
}

func (s *Streamer) Time() (time.Time, error) {
	request := &mission.GetScenarioCurrentTimeRequest{}
	response, err := s.missionService.GetScenarioCurrentTime(context.Background(), request)
	if err != nil {
		return time.Time{}, err
	}
	t, err := iso8601.ParseString(response.GetDatetime())
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}
