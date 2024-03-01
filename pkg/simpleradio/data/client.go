package data

// https://gitlab.com/overlordbot/srs-bot/-/blob/master/OverlordBot.SimpleRadio/Network/DataClient.cs

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
)

// DataClient is a client for the SRS data protocol.
type DataClient interface {
	// Name returns the name of the client as it appears in the SRS client list and in in-game transmissions.
	Name() string
	// Run starts the SRS data client. It should be called exactly once. The given channel will be closed when the client is ready.
	Run(context.Context, chan<- any) error
	// Send sends a message to the SRS server.
	Send(types.Message) error
}

type dataClient struct {
	// connection is the TCP connection to the SRS server.
	connection *net.TCPConn
	// clientInfo is the client information for this client. It is what players will see in the SRS client list, and the in-game overlay when this client transmits.
	clientInfo types.ClientInfo
	// externalAWACSModePassword is the password for authenticating as an external AWACS in the SRS server.
	externalAWACSModePassword string
	// otherClients is a map of GUIDs to client names, which the bot will use to filter out other clients that are not in the same coalition and frequency.
	otherClients map[types.GUID]string
	// lastReceived is the most recent time data was received. If this exceeds a data timeout, we have likely been disconnected from the server.
	lastReceived time.Time
}

func NewClient(guid types.GUID, config types.ClientConfiguration) (DataClient, error) {
	slog.Info("connecting to SRS server", "protocol", "tcp", "address", config.Address)
	address, err := net.ResolveTCPAddr("tcp", config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve SRS server address %v: %w", config.Address, err)
	}
	connection, err := net.DialTCP("tcp", nil, address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SRS server %v over TCP: %w", config.Address, err)
	}

	client := &dataClient{
		connection: connection,
		clientInfo: types.ClientInfo{
			Name:           config.ClientName,
			GUID:           guid,
			Seat:           0,
			Coalition:      config.Coalition,
			AllowRecording: false,
			RadioInfo: types.RadioInfo{
				UnitID:  100000002,
				Unit:    "External AWACS",
				Radios:  []types.Radio{config.Radio},
				IFF:     types.NewIFF(),
				Ambient: types.NewAmbient(),
			},
			Position: &types.Position{},
		},
		externalAWACSModePassword: config.ExternalAWACSModePassword,
		otherClients:              map[types.GUID]string{},
	}
	return client, nil
}

// Name implements DataClient.Name.
func (c *dataClient) Name() string {
	return c.clientInfo.Name
}

// Run implements DataClient.Run.
func (c *dataClient) Run(ctx context.Context, readyCh chan<- any) error {
	defer func() {
		if err := c.close(); err != nil {
			slog.Error("error closing data client", "error", err)
		}
	}()

	messageChan := make(chan types.Message)
	errorChan := make(chan error)

	go func() {
		reader := bufio.NewReader(c.connection)
		for {
			if ctx.Err() != nil {
				return
			}
			line, err := reader.ReadBytes(byte('\n'))
			switch err {
			case nil:
				var message types.Message
				jsonErr := json.Unmarshal(line, &message)
				if jsonErr != nil {
					slog.Warn("failed to unmarshal message", "text", line, "error", jsonErr)
				} else {
					messageChan <- message
				}
			case io.EOF:
				// no-op
			default:
				slog.Error("receive error", "error", err)
				errorChan <- err
				return
			}
		}
	}()

	close(readyCh)

	slog.Info("sending initial sync message")
	if err := c.sync(); err != nil {
		return fmt.Errorf("initial sync failed: %w", err)
	}

	slog.Info("connecting to external AWACS mode")
	if err := c.connectExternalAWACSMode(); err != nil {
		return fmt.Errorf("external AWACS mode failed: %w", err)
	}

	for {
		select {
		case m := <-messageChan:
			c.lastReceived = time.Now()
			c.handleMessage(m)
		case <-ctx.Done():
			slog.Info("stopping data client due to context cancellation", "error", ctx.Err())
			select {
			case <-messageChan:
			case <-errorChan:
			}
			return nil
		case err := <-errorChan:
			return fmt.Errorf("data client error: %w", err)
		}
	}
}

// handleMessage routes a given message to the appropriate handler.
func (c *dataClient) handleMessage(message types.Message) {
	slog.Debug("handling message", "message", message)
	switch message.Type {
	case types.MessagePing:
		logMessageAndIgnore(message)
	case types.MessageServerSettings:
		logMessageAndIgnore(message)
	case types.MessageVersionMismatch:
		logMessageAndIgnore(message)
	case types.MessageExternalAWACSModeDisconnect:
		logMessageAndIgnore(message)
	case types.MessageSync:
		c.syncClients(message.Clients)
	case types.MessageUpdate:
		c.syncClient(message.Client)
	case types.MessageRadioUpdate:
		c.syncClient(message.Client)
	case types.MessageClientDisconnect:
		c.syncClient(message.Client)
	case types.MessageExternalAWACSModePassword:
		// WTF is this???
		if message.Client.Coalition == c.clientInfo.Coalition {
			c.updateRadios()
		}
	default:
		slog.Warn("received unrecognized message", "payload", message)
	}
}

// logMessageAndIgnore logs a message at DEBUG level.
func logMessageAndIgnore(message types.Message) {
	slog.Debug("received message", "payload", message)
}

// syncClients calls syncClient for each client in the given slice.
func (c *dataClient) syncClients(others []types.ClientInfo) {
	slog.Info("syncronizing clients", "count", len(others))
	for _, info := range others {
		c.syncClient(info)
	}
}

// syncClient checks if the given client matches this client's coalition and radios, and if so, stores it in the otherClients map. Non-matching clients are removed from the map if previously stored.
func (c *dataClient) syncClient(other types.ClientInfo) {
	clientLogger := slog.With("guid", other.GUID, "name", other.Name, "coalition", other.Coalition, "radios", other.RadioInfo)

	clientLogger.Debug("syncronizing client")

	if other.GUID == c.clientInfo.GUID {
		// why, of course I know him. he's me!
		clientLogger.Debug("skipped client due to same GUID")
		return
	}

	isSameFrequency := c.clientInfo.RadioInfo.IsOnFrequency(other.RadioInfo)
	// isSameCoalition is true if the other client is in the same coalition as this client, or if the other client is a spectator.
	isSameCoalition := (c.clientInfo.Coalition == other.Coalition) || types.IsSpectator(other.Coalition)
	clientLogger.Debug("checking client", "coalitionMatches", isSameCoalition, "frequencyMatches", isSameFrequency)

	// if the other client has a matching radio and is not in an opposing coalition, store it in otherClients. Otherwise, banish it to the shadow realm.
	if isSameCoalition && isSameFrequency {
		clientLogger.Debug("storing client with matching radio")
		c.otherClients[other.GUID] = other.Name
	} else {
		_, ok := c.otherClients[other.GUID]
		if ok {
			clientLogger.Debug("deleting client without matching radio")
			delete(c.otherClients, other.GUID)
			// TODO memory leak here due to continually adding and removing clients from the map. https://100go.co/28-maps-memory-leaks/
		} else {
			clientLogger.Debug("skipped client without matching radio")
		}
	}
}

// Send implements DataClient.Send.
func (c *dataClient) Send(message types.Message) error {
	// Sending a message means writing a JSON-serialized message to the TCP connection, followed by a newline.
	if message.Version == "" {
		return fmt.Errorf("message Version is required")
	}
	b, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message to JSON: %w", err)
	}
	b = append(b, byte('\n'))
	slog.Debug("sending message", "message", message)
	_, err = c.connection.Write(b)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
}

// newMessage is a helper that initializes a new message with the client's version and the given message type.
func (c *dataClient) newMessage(t types.MessageType) types.Message {
	return types.Message{
		Version: "2.1.0.2", // stubbing fake SRS version, TODO add flag
		Type:    t,
	}
}

// sync sends a sync message to the SRS server containing this client's information.
func (c *dataClient) sync() error {
	message := c.newMessageWithClient(types.MessageSync)
	if err := c.Send(message); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}
	return nil
}

func (c *dataClient) newMessageWithClient(t types.MessageType) types.Message {
	message := c.newMessage(t)
	message.Client = c.clientInfo
	return message
}

func (c *dataClient) update() error {
	message := c.newMessageWithClient(types.MessageUpdate)
	if err := c.Send(message); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	return nil
}

// updateRadios sends a radio update message to the SRS server containing this client's information.
func (c *dataClient) updateRadios() error {
	message := c.newMessageWithClient(types.MessageRadioUpdate)
	if err := c.Send(message); err != nil {
		return fmt.Errorf("radio update failed: %w", err)
	}
	return nil
}

// connectExternalAWACSMode sends an external AWACS mode password message to the SRS server to authenticate as an external AWACS.
func (c *dataClient) connectExternalAWACSMode() error {
	message := c.newMessageWithClient(types.MessageExternalAWACSModePassword)
	message.ExternalAWACSModePassword = c.externalAWACSModePassword
	if err := c.Send(message); err != nil {
		return fmt.Errorf("failed to authenticate with EAM password: %w", err)
	}
	return nil
}

// close closes the TCP connection to the SRS server. This is anti-idomatic Go and should be refactored.
func (c *dataClient) close() error {
	if err := c.connection.Close(); err != nil {
		return fmt.Errorf("error closing TCP connection to SRS: %w", err)
	}
	return nil
}
