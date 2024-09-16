package simpleradio

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/rs/zerolog/log"
)

// Send implements [Client.Send].
func (c *client) Send(message types.Message) error {
	// Sending a message means writing a JSON-serialized message to the TCP connection, followed by a newline.
	if message.Version == "" {
		return errors.New("message Version is required")
	}
	b, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message to JSON: %w", err)
	}
	b = append(b, byte('\n'))
	_, err = c.tcpConnection.Write(b)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	return nil
}

// newMessage creates a new message with the client's version and the given message type.
func (c *client) newMessage(t types.MessageType) types.Message {
	return types.Message{
		Version: "2.1.0.2", // stubbing fake SRS version, TODO add flag
		Type:    t,
	}
}

// newMessageWithClient creates a new message with the client's version, the given message type, and the client's info.
func (c *client) newMessageWithClient(t types.MessageType) types.Message {
	message := c.newMessage(t)
	message.Client = c.clientInfo
	return message
}

// handleMessage routes a given message to the appropriate handler.
func (c *client) handleMessage(message types.Message) {
	switch message.Type {
	case types.MessagePing:
		logMessageAndIgnore(message)
	case types.MessageServerSettings:
		c.updateServerSettings(message)
	case types.MessageVersionMismatch:
		log.Warn().Any("message", message).Msg("received version mismatch message from SRS server")
	case types.MessageExternalAWACSModeDisconnect:
		logMessageAndIgnore(message)
	case types.MessageSync:
		c.syncClients(message.Clients)
	case types.MessageUpdate:
		c.syncClient(message.Client)
	case types.MessageRadioUpdate:
		c.syncClient(message.Client)
	case types.MessageClientDisconnect:
		c.removeClient(message.Client)
	case types.MessageExternalAWACSModePassword:
		if message.Client.Coalition == c.clientInfo.Coalition {
			log.Debug().Any("remoteClient", message.Client).Msg("received external AWACS mode password message")
			// TODO is the update necessary?
			if err := c.updateRadios(); err != nil {
				log.Error().Err(err).Msg("failed to update radios")
			}
		}
	default:
		log.Warn().Any("message", message).Msg("received unrecognized message")
	}
}

// updateServerSettings updates the client's settings to match the server's settings.
func (c *client) updateServerSettings(message types.Message) {
	log.Debug().Any("serverSettings", message.ServerSettings).Msg("received server settings")
	if enabled, ok := message.ServerSettings[string(types.CoalitionAudioSecurity)]; ok {
		if strings.ToLower(enabled) == "true" {
			log.Info().Msg("enabling secure coalition radios")
			c.secureCoalitionRadios = true
		} else {
			log.Info().Msg("disabling secure coalition radios")
			c.secureCoalitionRadios = false
		}
	}
}

// updateRadios sends a radio update message to the SRS server containing this client's information.
func (c *client) updateRadios() error {
	message := c.newMessageWithClient(types.MessageRadioUpdate)
	if err := c.Send(message); err != nil {
		return fmt.Errorf("radio update failed: %w", err)
	}
	return nil
}

// connectExternalAWACSMode sends an external AWACS mode password message to the SRS server to authenticate as an external AWACS.
func (c *client) connectExternalAWACSMode() error {
	message := c.newMessageWithClient(types.MessageExternalAWACSModePassword)
	message.ExternalAWACSModePassword = c.externalAWACSModePassword
	if err := c.Send(message); err != nil {
		return fmt.Errorf("failed to authenticate with EAM password: %w", err)
	}
	return nil
}

// logMessageAndIgnore logs a message at DEBUG level.
func logMessageAndIgnore(message types.Message) {
	log.Debug().Any("message", message).Msg("received message")
}
