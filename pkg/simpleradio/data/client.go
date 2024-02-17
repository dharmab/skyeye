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

	srs "github.com/dharmab/skyeye/pkg/simpleradio/types"
)

type DataClient interface {
	Run(context.Context) error
	Send(srs.Message) error
}

type dataClient struct {
	connection                *net.TCPConn
	clientInfo                srs.ClientInfo
	externalAWACSModePassword string
	// otherClients is a map of GUIDs to client names
	// filtered to our coalition and frequency
	otherClients map[string]string
	// lastReceived is the most recent time data was received. If this exceeds a data timeout, we have likely been disconnected from the server.
	lastReceived time.Time
}

func NewClient(config srs.ClientConfiguration) (DataClient, error) {
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
		clientInfo: srs.ClientInfo{
			Name:           config.ClientName,
			GUID:           config.GUID,
			Seat:           0,
			Coalition:      config.Coalition,
			AllowRecording: false,
			RadioInfo: srs.RadioInfo{
				UnitID: 0,
				Unit:   "External AWACS",
				Radios: []srs.Radio{
					{
						Frequency:        config.Frequency.Frequency,
						Modulation:       config.Frequency.Modulation,
						IsEncrypted:      false,
						EncryptionKey:    0,
						GuardFrequency:   243.0,
						ShouldRetransmit: false,
					},
				},
				IFF:     srs.NewIFF(),
				Ambient: srs.NewAmbient(),
			},
			Position: &srs.Position{},
		},
		externalAWACSModePassword: config.ExternalAWACSModePassword,
		otherClients:              map[string]string{},
	}
	return client, nil
}

func (c *dataClient) Run(ctx context.Context) error {
	defer func() {
		if err := c.close(); err != nil {
			slog.Error("error closing data client", "error", err)
		}
	}()

	messageChan := make(chan srs.Message)
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
				var message srs.Message
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

	slog.Info("sending initial sync message")
	if err := c.sync(); err != nil {
		return fmt.Errorf("initial sync failed: %w", err)
	}
	slog.Info("sending initial radio update message")
	if err := c.updateRadios(); err != nil {
		return fmt.Errorf("initial radio update failed: %w", err)
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

func (c *dataClient) handleMessage(message srs.Message) {
	slog.Debug("handling message", "message", message)
	switch message.Type {
	case srs.MessagePing:
		logMessageAndIgnore(message)
	case srs.MessageServerSettings:
		logMessageAndIgnore(message)
	case srs.MessageVersionMismatch:
		logMessageAndIgnore(message)
	case srs.MessageExternalAWACSModeDisconnect:
		logMessageAndIgnore(message)
	case srs.MessageSync:
		c.syncClients(message.Clients)
	case srs.MessageUpdate:
		c.syncClient(message.Client)
	case srs.MessageRadioUpdate:
		c.syncClient(message.Client)
	case srs.MessageClientDisconnect:
		c.syncClient(message.Client)
	case srs.MessageExternalAWACSModePassword:
		// WTF is this???
		c.Send(srs.Message{
			Type:    srs.MessageRadioUpdate,
			Clients: []srs.ClientInfo{c.clientInfo},
		})
	default:
		slog.Warn("received unrecognized message", "payload", message)
	}
}

func logMessageAndIgnore(message srs.Message) {
	slog.Debug("received message", "payload", message)
}

func (c *dataClient) syncClients(others []srs.ClientInfo) {
	slog.Info("syncronizing clients", "count", len(others))
	for _, info := range others {
		c.syncClient(&info)
	}
}

func (c *dataClient) syncClient(other *srs.ClientInfo) {
	if other == nil {
		slog.Warn("syncClient called using nil client. ignoring...")
		return
	}
	clientLogger := slog.With("guid", other.GUID, "name", other.Name, "coalition", other.Coalition, "radios", other.RadioInfo)

	clientLogger.Debug("syncronizing client")

	if other.GUID == c.clientInfo.GUID {
		// why, of course I know him. he's me!
		clientLogger.Debug("skipped client due to same GUID")
		return
	}

	var isSameFrequency bool
	for _, otherRadio := range other.RadioInfo.Radios {
		for _, thisRadio := range c.clientInfo.RadioInfo.Radios {
			radioLogger := slog.With(
				"guid", other.GUID,
				"name", other.Name,
				"frequency", otherRadio.Frequency,
				"modulation", otherRadio.Modulation,
				"encryption", otherRadio.IsEncrypted,
			)

			doesFrequencyMatch := float64(thisRadio.Frequency) == float64(otherRadio.Frequency/1000000)
			doesModulationMatch := thisRadio.Modulation == otherRadio.Modulation
			doesEncryptionMatch := (!thisRadio.IsEncrypted && !otherRadio.IsEncrypted) || (thisRadio.IsEncrypted && otherRadio.IsEncrypted && thisRadio.EncryptionKey == otherRadio.EncryptionKey)
			radioLogger.Debug("checking client radio", "frequencyMatches", doesFrequencyMatch, "modulationMatches", doesModulationMatch, "encryptionMatches", doesEncryptionMatch)
			if doesFrequencyMatch && doesModulationMatch && doesEncryptionMatch {
				isSameFrequency = true
			}
		}
	}

	isSameCoalition := (c.clientInfo.Coalition == other.Coalition) || srs.IsSpectator(other.Coalition)
	clientLogger.Debug("checking client", "coalitionMatches", isSameCoalition, "frequencyMatches", isSameFrequency)
	if isSameCoalition && isSameFrequency {
		clientLogger.Debug("storing client with matching radio")
		c.otherClients[other.GUID] = other.Name
	} else {
		_, ok := c.otherClients[other.GUID]
		if ok {
			clientLogger.Debug("deleting client without matching radio")
			delete(c.otherClients, other.GUID)
		} else {
			clientLogger.Debug("skipped client without matching radio")
		}
	}
}

func (c *dataClient) Send(message srs.Message) error {
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

func (c *dataClient) newMessage(t srs.MessageType) srs.Message {
	return srs.Message{
		Version: "2.1.0.1", // stubbing fake SRS version
		Type:    t,
	}
}

func (c *dataClient) sync() error {
	message := c.newMessage(srs.MessageSync)
	message.Client = &c.clientInfo
	if err := c.Send(message); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}
	return nil
}

func (c *dataClient) update() error {
	message := c.newMessage(srs.MessageUpdate)
	message.Client = &c.clientInfo
	if err := c.Send(message); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}
	return nil
}

func (c *dataClient) updateRadios() error {
	message := c.newMessage(srs.MessageRadioUpdate)
	message.Client = &c.clientInfo
	if err := c.Send(message); err != nil {
		return fmt.Errorf("radio update failed: %w", err)
	}
	return nil
}

func (c *dataClient) connectExternalAWACSMode() error {
	message := c.newMessage(srs.MessageExternalAWACSModePassword)
	message.Client = &srs.ClientInfo{
		GUID:           c.clientInfo.GUID,
		Name:           c.clientInfo.Name,
		Coalition:      c.clientInfo.Coalition,
		AllowRecording: c.clientInfo.AllowRecording,
		Position:       c.clientInfo.Position,
	}
	message.ExternalAWACSModePassword = c.externalAWACSModePassword
	if err := c.Send(message); err != nil {
		return fmt.Errorf("failed to authenticate with EAM password: %w", err)
	}
	return nil
}

func (c *dataClient) close() error {
	if err := c.connection.Close(); err != nil {
		return fmt.Errorf("error closing TCP connection to SRS: %w", err)
	}
	return nil
}
