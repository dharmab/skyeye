package commands

import (
	"context"
	"fmt"
	"time"

	secoalition "github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/lithammer/shortuuid/v3"

	grpccoalition "github.com/DCS-gRPC/go-bindings/dcs/v0/coalition"
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
	coalitionClient grpccoalition.CoalitionServiceClient
}

func NewChatListener(
	coalition secoalition.Coalition,
	callsign string,
	missionClient mission.MissionServiceClient,
	coalitionClient grpccoalition.CoalitionServiceClient,
	netClient net.NetServiceClient,
) *ChatListener {
	manager := &ChatListener{
		callsign:        callsign,
		missionClient:   missionClient,
		coalitionClient: coalitionClient,
		netClient:       netClient,
	}
	if coalition == secoalition.Red {
		manager.coalition = common.Coalition_COALITION_RED
	} else {
		manager.coalition = common.Coalition_COALITION_BLUE
	}
	return manager
}

func (l *ChatListener) Run(ctx context.Context, messages chan<- Request) {
	nextAttempt := time.Now()
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping chat listener due to context cancellation")
			return
		default:
			time.Sleep(time.Until(nextAttempt))
			nextAttempt = time.Now().Add(5 * time.Second)
			streamer, err := l.missionClient.StreamEvents(ctx, &mission.StreamEventsRequest{})
			if err != nil {
				log.Error().Err(err).Msg("error streaming chat events from DCS-gRPC")
				continue
			}
			if err = l.receive(ctx, streamer, messages); err != nil {
				continue
			}
			log.Error().Err(err).Msg("error streaming chat messages")
		}
	}
}

func (l *ChatListener) receive(ctx context.Context, client mission.MissionService_StreamEventsClient, messages chan<- Request) error {
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping chat listener due to context cancellation")
			return nil
		default:
			response, err := client.Recv()
			if err != nil {
				return fmt.Errorf("event stream error: %w", err)
			}
			chatEvent := response.GetPlayerSendChat()
			if chatEvent == nil {
				continue
			}
			if chatEvent.GetMessage() == "" {
				continue
			}
			logger := log.With().Uint32("unitID", chatEvent.GetPlayerId()).Str("text", chatEvent.GetMessage()).Logger()
			logger.Debug().Msg("received chat message")

			players, err := l.netClient.GetPlayers(ctx, &net.GetPlayersRequest{})
			if err != nil {
				return fmt.Errorf("error getting players: %w", err)
			}

			var playerInfo *net.GetPlayersResponse_GetPlayerInfo
			for _, player := range players.GetPlayers() {
				if player.GetId() == chatEvent.GetPlayerId() {
					playerInfo = player
				}
			}
			if playerInfo == nil {
				logger.Debug().Msg("player not found")
				continue
			}

			logger = logger.With().Str("name", playerInfo.GetName()).Stringer("coaltion", playerInfo.GetCoalition()).Logger()

			if playerInfo.GetCoalition() != l.coalition {
				logger.Debug().Msg("player not on coalition")
				continue
			}

			logger.Info().Msg("received chat message")

			messages <- Request{
				TraceID:    shortuuid.New(),
				PlayerName: playerInfo.GetName(),
				Text:       chatEvent.GetMessage(),
			}
		}
	}
}
