package simpleradio

import (
	"fmt"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// syncClients calls syncClient for each client in the given slice.
func (c *client) syncClients(others []types.ClientInfo) {
	log.Info().Int("count", len(others)).Msg("syncronizing clients")
	for _, info := range others {
		c.syncClient(info)
	}
}

// syncClient checks if the given client matches this client's coalition and radios, and if so, stores it in the clients map. Non-matching clients are removed from the map if previously stored.
func (c *client) syncClient(other types.ClientInfo) {
	if other.GUID == c.clientInfo.GUID {
		// why, of course I know him. he's me!
		return
	}

	if len(other.RadioInfo.Radios) == 0 {
		return
	}

	frequencies := make([]string, 0)
	for _, radio := range other.RadioInfo.Radios {
		frequency := unit.Frequency(radio.Frequency) * unit.Hertz
		if frequency.Megahertz() > 8 {
			frequencies = append(frequencies, fmt.Sprint(frequency.Megahertz()))
		}
	}
	log.Debug().
		Str("name", other.Name).
		Uint64("unitID", other.RadioInfo.UnitID).
		Strs("frequencies", frequencies).
		Msgf("synced with SRS client %q", other.Name)

	isSameCoalition := c.clientInfo.Coalition == other.Coalition || types.IsSpectator(other.Coalition)
	isOnFrequency := c.clientInfo.RadioInfo.IsOnFrequency(other.RadioInfo)

	// if the other client has a matching radio and is not in an opposing coalition, store it in the clients map. Otherwise, banish it to the shadow realm.
	c.clientsLock.Lock()
	defer c.clientsLock.Unlock()
	if isSameCoalition && isOnFrequency {
		c.clients[other.GUID] = other
	} else {
		delete(c.clients, other.GUID)
	}
}

// removeClient removes the client with the given GUID from the clients map.
func (c *client) removeClient(info types.ClientInfo) {
	c.clientsLock.Lock()
	defer c.clientsLock.Unlock()
	delete(c.clients, info.GUID)
}

// sync sends a sync message to the SRS server containing this client's information.
func (c *client) sync() error {
	message := c.newMessageWithClient(types.MessageSync)
	if err := c.Send(message); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}
	return nil
}
