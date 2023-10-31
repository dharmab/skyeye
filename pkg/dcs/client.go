package dcs

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/DCS-gRPC/go-bindings/dcs/v0/coalition"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/group"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/mission"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/unit"
	"github.com/DCS-gRPC/go-bindings/dcs/v0/world"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConfiguration struct {
	Address           string
	ConnectionTimeout time.Duration
}

type DCSClient interface {
	WorldClient() world.WorldServiceClient
	MissionClient() mission.MissionServiceClient
	CoalitionClient() coalition.CoalitionServiceClient
	GroupClient() group.GroupServiceClient
	UnitClient() unit.UnitServiceClient
	Close() error
}

type dcsClient struct {
	connection      *grpc.ClientConn
	worldClient     world.WorldServiceClient
	missionClient   mission.MissionServiceClient
	coalitionClient coalition.CoalitionServiceClient
	groupClient     group.GroupServiceClient
	unitClient      unit.UnitServiceClient
}

func NewDCSClient(ctx context.Context, config ClientConfiguration) (DCSClient, error) {
	slog.Info("connecting to DCS-gRPC server", "address", config.Address)
	connection, err := connectToDCSGRPC(ctx, config.Address, config.ConnectionTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DCS-gRPC server %v: %w", config.Address, err)
	}

	return &dcsClient{
		connection:      connection,
		worldClient:     world.NewWorldServiceClient(connection),
		missionClient:   mission.NewMissionServiceClient(connection),
		coalitionClient: coalition.NewCoalitionServiceClient(connection),
		groupClient:     group.NewGroupServiceClient(connection),
		unitClient:      unit.NewUnitServiceClient(connection),
	}, nil
}

func connectToDCSGRPC(ctx context.Context, address string, timeout time.Duration) (*grpc.ClientConn, error) {
	connectionCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return grpc.DialContext(connectionCtx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func (c *dcsClient) WorldClient() world.WorldServiceClient {
	return c.worldClient
}

func (c *dcsClient) MissionClient() mission.MissionServiceClient {
	return c.missionClient
}

func (c *dcsClient) CoalitionClient() coalition.CoalitionServiceClient {
	return c.coalitionClient
}

func (c *dcsClient) GroupClient() group.GroupServiceClient {
	return c.groupClient
}

func (c *dcsClient) UnitClient() unit.UnitServiceClient {
	return c.unitClient
}

func (c *dcsClient) Close() error {
	return c.connection.Close()
}
