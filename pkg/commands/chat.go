package commands

import (
	"context"

	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/lithammer/shortuuid/v3"

	"github.com/DCS-gRPC/go-bindings/dcs/v0/coalition"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/common"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/mission"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/net"
	"github.com/rs/zerolog/log"
)

type ChatListener struct {
	coalition       common.Coalition
	callsign        string
	netClient       net.NetServiceClient
	missionClient   mission.MissionServiceClient
	coalitionClient coalition.CoalitionServiceClient
}

func NewChatListener(
	coalition coalitions.Coalition,
	callsign string,
	missionClient mission.MissionServiceClient,
	coalitionClient coalition.CoalitionServiceClient,
) *ChatListener {
	manager := &ChatListener{
		callsign:        callsign,
		missionClient:   missionClient,
		coalitionClient: coalitionClient,
	}
	if coalition == coalitions.Red {
		manager.coalition = common.Coalition_COALITION_RED
	} else {
		manager.coalition = common.Coalition_COALITION_BLUE
	}
	return manager
}

func (m *ChatListener) isPlayerOnCoalition(ctx context.Context, id uint32) (bool, error) {
	request := &coalition.GetPlayerUnitsRequest{
		Coalition: m.coalition,
	}
	units, err := m.coalitionClient.GetPlayerUnits(ctx, request)
	if err != nil {
		return false, err
	}
	for _, u := range units.GetUnits() {
		if u.GetId() == id {
			return true, nil
		}
	}
	return false, nil
}

func (l *ChatListener) Run(ctx context.Context, messages chan<- Request) {
	streamer, err := l.missionClient.StreamEvents(ctx, &mission.StreamEventsRequest{})
	if err != nil {
		log.Error().Err(err).Msg("error creating event stream")
		return
	}
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping chat listener due to context cancellation")
			return
		default:
			response, err := streamer.Recv()
			if err != nil {
				log.Error().Err(err).Msg("event stream error")
				continue
			}
			event := response.GetPlayerSendChat()
			logger := log.With().Uint32("unitID", event.GetPlayerId()).Str("text", event.GetMessage()).Logger()
			isSameCoalition, err := l.isPlayerOnCoalition(ctx, event.GetPlayerId())
			if err != nil {
				logger.Error().Err(err).Msg("error checking player coalition")
				continue
			}
			if !isSameCoalition {
				logger.Debug().Msg("player is not on the same coalition")
				continue
			}

			players, err := l.netClient.GetPlayers(ctx, &net.GetPlayersRequest{})
			if err != nil {
				logger.Error().Err(err).Msg("error getting players")
				continue
			}

			var playerName string
			for _, player := range players.GetPlayers() {
				if player.GetId() == event.GetPlayerId() {
					playerName = player.GetName()
					logger = logger.With().Str("player", playerName).Logger()
				}
			}

			logger.Info().Msg("received chat message")

			messages <- Request{
				TraceID:    shortuuid.New(),
				PlayerName: playerName,
				Text:       event.GetMessage(),
			}
		}
	}
}
