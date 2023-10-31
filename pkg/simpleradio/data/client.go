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
			GUID:           config.GUID.String(),
			Seat:           0,
			Coalition:      int(config.Coalition),
			AllowRecording: true,
			Radios: srs.ClientRadios{
				UnitID: 0,
				Unit:   "",
				Radios: []srs.Radio{
					{
						Frequency:        config.Frequency.Frequency,
						Modulation:       config.Frequency.Modulation,
						IsEncrypted:      false,
						EncryptionKey:    0,
						GuardFrequency:   243.0,
						ShouldRetransmit: false,
						Volume:           1.0,
					},
				},
			},
		},
		externalAWACSModePassword: config.ExternalAWACSModePassword,
		otherClients:              map[string]string{},
	}

	if err := client.sync(); err != nil {
		defer client.close()
		return nil, err
	}
	if err := client.connectExternalAWACSMode(); err != nil {
		defer client.close()
		return nil, err
	}

	return client, nil
}

func (c *dataClient) Run(ctx context.Context) error {
	defer func() {
		if err := c.close(); err != nil {
			slog.Error("error closing data client", "error", err)
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
		slog.Info("syncronizing clients")
		for _, info := range message.Clients {
			c.syncClient(info)
		}
	case srs.MessageUpdate:
		c.syncClient(message.Client)
	case srs.MessageRadioUpdate:
		c.syncClient(message.Client)
	case srs.MessageClientDisconnect:
		c.syncClient(message.Client)
	case srs.MessageExternalAWACSModePassword:
		// WTF is this???
		c.Send(srs.Message{
			Type:   srs.MessageRadioUpdate,
			Client: c.clientInfo,
		})
	default:
		slog.Warn("received unrecognized message", "payload", message)
	}
}

func logMessageAndIgnore(message srs.Message) {
	slog.Debug("received message", "payload", message)
}

func (c *dataClient) syncClient(other srs.ClientInfo) {
	logger := slog.With("guid", other.GUID, "name", other.Name, "coalition", other.Coalition, "radios", other.Radios)

	logger.Debug("syncronizing client")

	if other.GUID == c.clientInfo.GUID {
		// why, of course I know him. he's me!
		logger.Debug("skipped client due to same GUID")
		return
	}

	isSameCoalition := other.Coalition == c.clientInfo.Coalition

	var isSameFrequency bool
	for _, otherRadio := range other.Radios.Radios {
		for _, thisRadio := range c.clientInfo.Radios.Radios {
			slog.Debug(
				"checking client radio",
				"guid", other.GUID,
				"name", other.Name,
				"frequency", otherRadio.Frequency,
				"modulation", otherRadio.Modulation,
				"encryption", otherRadio.IsEncrypted,
			)
			if thisRadio.Frequency == otherRadio.Frequency && thisRadio.Modulation == otherRadio.Modulation {
				// Frequency and modulation matches
				if !thisRadio.IsEncrypted && !otherRadio.IsEncrypted {
					// No encryption
					isSameFrequency = true
				} else if thisRadio.IsEncrypted && otherRadio.IsEncrypted && thisRadio.EncryptionKey == otherRadio.EncryptionKey {
					// Encryption enabled on both radios and key matches
					isSameFrequency = true
				}
			}
		}
	}

	if isSameCoalition && isSameFrequency {
		logger.Debug("storing client with matching radio")
		c.otherClients[other.GUID] = other.Name

	} else {
		_, ok := c.otherClients[other.GUID]
		if ok {
			logger.Debug("deleting client without matching radio")
			delete(c.otherClients, other.GUID)
		} else {
			logger.Debug("skipped client without matching radio")
		}
	}
}

func (c *dataClient) Send(message srs.Message) error {
	b, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message to JSON: %w", err)
	}
	slog.Debug("sending message", "message", message)
	_, err = c.connection.Write(b)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
}

func (c *dataClient) newMessage(t srs.MessageType) srs.Message {
	return srs.Message{
		Version: "v0.0.0-dev",
		Type:    t,
		Client:  c.clientInfo,
	}
}

func (c *dataClient) sync() error {
	message := c.newMessage(srs.MessageSync)
	message.Client = c.clientInfo
	if err := c.Send(message); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}
	return nil
}

func (c *dataClient) updateRadios() error {
	message := c.newMessage(srs.MessageRadioUpdate)
	message.Client = c.clientInfo
	if err := c.Send(message); err != nil {
		return fmt.Errorf("radio update failed: %w", err)
	}
	return nil
}

func (c *dataClient) connectExternalAWACSMode() error {
	message := c.newMessage(srs.MessageExternalAWACSModePassword)
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
