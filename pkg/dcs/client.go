// package dcs provides a client for the DCS-gRPC server. This client can be used to query and manipulate the sim.
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

// ClientConfiguration configures the DCS-gRPC client.
type ClientConfiguration struct {
	Address           string
	ConnectionTimeout time.Duration
}

// DCSClient wraps the DCS-gRPC client.
type DCSClient interface {
	// WorldClient returns a WorldServiceClient, which can be used to query the sim world state. This includes the theatre name, airbases, and map markers.
	WorldClient() world.WorldServiceClient
	// MissionClient returns a MissionServiceClient, which can be used to query and manipulate the mission scenario. This includes commands to manipulate the Communications menu.
	MissionClient() mission.MissionServiceClient
	// CoalitionClient returns a CoalitionServiceClient, which can be used to query coalition data. This includes the bullseye, groups and player units on the blue and red coalitions.
	CoalitionClient() coalition.CoalitionServiceClient
	// GroupClient returns a GroupServiceClient, which can be used to query the units within a group.
	GroupClient() group.GroupServiceClient
	// UnitClient returns a UnitServiceClient, which can be used to query a unit's name, description and position.
	UnitClient() unit.UnitServiceClient
	// Close closes the DCS-gRPC client connection. This is anti-idomatic Go and should be refactored...
	Close() error
}

// dcsClient implements DCSClient.
type dcsClient struct {
	// connection is the gRPC client connection to DCS-gRPC, and indirectly, the DCS World server.
	connection *grpc.ClientConn
	// worldClient a client for the sim world state. This includes the theatre name, airbases, and map markers.
	worldClient world.WorldServiceClient
	// missionClient is a client for mission scenario data. This includes commands to manipulate the Communications menu.
	missionClient mission.MissionServiceClient
	// coalitionClient is a client for coalition data. This includes the bullseye, groups and player units on the blue and red coalitions..
	coalitionClient coalition.CoalitionServiceClient
	// groupClient is a client for getting the units within a group.
	groupClient group.GroupServiceClient
	// unitClient is a client for querying a unit's name, description and position.
	unitClient unit.UnitServiceClient
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

// connectToDCSGRPC connects to the DCS-gRPC server at the given address. This is kind of anti-idomatic Go and should be refactored...
func connectToDCSGRPC(ctx context.Context, address string, timeout time.Duration) (*grpc.ClientConn, error) {
	connectionCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return grpc.DialContext(connectionCtx, address, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// WorldClient implements DCSClient.WorldClient
func (c *dcsClient) WorldClient() world.WorldServiceClient {
	return c.worldClient
}

// MissionClient implements DCSClient.MissionClient
func (c *dcsClient) MissionClient() mission.MissionServiceClient {
	return c.missionClient
}

// CoalitionClient implements DCSClient.CoalitionClient
func (c *dcsClient) CoalitionClient() coalition.CoalitionServiceClient {
	return c.coalitionClient
}

// GroupClient implements DCSClient.GroupClient
func (c *dcsClient) GroupClient() group.GroupServiceClient {
	return c.groupClient
}

// UnitClient implements DCSClient.UnitClient
func (c *dcsClient) UnitClient() unit.UnitServiceClient {
	return c.unitClient
}

// Close implements DCSClient.Close
func (c *dcsClient) Close() error {
	return c.connection.Close()
}
