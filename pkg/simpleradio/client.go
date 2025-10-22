// Package simpleradio contains a SimpleRadio-Standalone client.
package simpleradio

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
	"github.com/rs/zerolog/log"
)

// Audio is a sample of audio data in F32LE PCM format.
type Audio []float32

// Transmission is an envelope containing a trace ID, SRS client name, and audio sample for a transmission.
type Transmission struct {
	// TraceID of the transmission.
	TraceID string
	// ClientName is the name of the SRS client that transmitted the audio.
	ClientName string
	// Audio sample for the transmission.
	Audio Audio
}

// Client is a SimpleRadio-Standalone Client.
type Client struct {
	// externalAWACSModePassword is the password for authenticating as an external AWACS in the SRS server.
	externalAWACSModePassword string

	// address is the address of the SRS server, including the port.
	address string
	// tcpConnection is the TCP connection to the SRS server used for messages.
	tcpConnection *net.TCPConn
	// udpConnection is the UDP connection to the SRS server used for audio and pings.
	udpConnection *net.UDPConn

	// clientInfo is the client information for this client. It is what players will see in the SRS client list, and in
	/// the in-game overlay when this client transmits.
	clientInfo types.ClientInfo
	// clients is a map of GUIDs to client info, which the bot will use to filter out other clients that are not in the
	// same coalition and frequency.
	clients map[types.GUID]types.ClientInfo
	// clientsLock controls access to the clients map.
	clientsLock sync.RWMutex

	// secureCoalitionRadios indicates if the client should only receive transmissions from the same coalition.
	secureCoalitionRadios bool

	// rxChan is a channel where received transmission are published. A read-only version is available publicly.
	rxChan chan Transmission
	// txChan is a channel where outgoing transmissions are buffered.
	txChan chan Transmission
	// receivers tracks the state of each radio we are listening to.
	receivers map[types.Radio]*receiver
	// packetNumber is incremented for each voice packet transmitted.
	packetNumber uint64
	// txLock prevents multiple outgoing transmissions from occurring simultaneously. It must be acquired before writing
	// voice packets to the UDP connection.
	txLock sync.Mutex
	// mute suppresses audio transmission.
	mute bool

	// lastPing tracks the last time a ping was received. If no pings are received for a period of time, the client will
	// attempt to reconnect.
	lastPing     time.Time
	lastPingLock sync.RWMutex
}

// NewClient creates a new SimpleRadio-Standalone client.
func NewClient(config types.ClientConfiguration) (*Client, error) {
	guid := types.NewGUID()

	receivers := make(map[types.Radio]*receiver, len(config.Radios))
	for _, radio := range config.Radios {
		receivers[radio] = &receiver{}
	}

	client := &Client{
		address: config.Address,
		clientInfo: types.ClientInfo{
			Name:      config.ClientName,
			GUID:      guid,
			Coalition: config.Coalition,
			RadioInfo: types.RadioInfo{
				UnitID:  100000002,
				Unit:    "External AWACS",
				Radios:  config.Radios,
				IFF:     types.NewIFF(),
				Ambient: types.NewAmbient(),
			},
			AllowRecording: true,
			Position:       &types.Position{},
		},
		externalAWACSModePassword: config.ExternalAWACSModePassword,
		clients:                   make(map[types.GUID]types.ClientInfo),

		txChan:       make(chan Transmission),
		rxChan:       make(chan Transmission),
		receivers:    receivers,
		packetNumber: 1,
		mute:         config.Mute,
		lastPing:     time.Now(),
	}

	err := client.connectTCP()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SRS server: %w", err)
	}
	err = client.connectUDP()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SRS server: %w", err)
	}

	return client, nil
}

// initialize must be called after (re)connecting to the SRS server to synchronize the client and server state.
func (c *Client) initialize() error {
	log.Info().Msg("syncing with SRS server")
	if err := c.sync(); err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	log.Info().Msg("reconnecting to external AWACS mode")
	if err := c.connectExternalAWACSMode(); err != nil {
		return fmt.Errorf("connecting external AWACS mode failed: %w", err)
	}

	for _, receiver := range c.receivers {
		receiver.reset()
	}

	c.SendPing()

	return nil
}

// autoheal attempts to reconnect and reinitialize the SRS client if it stops receiving traffic from the SRS server.
func (c *Client) autoheal(ctx context.Context) {
	ticker := time.NewTicker(pingInterval / 3)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			func() {
				c.lastPingLock.Lock()
				defer c.lastPingLock.Unlock()
				if time.Since(c.lastPing) > pingInterval*3 {
					log.Warn().Msg("stopped receiving traffic from SRS server")

					log.Warn().Msg("attempting to reconnect to SRS server")
					if reconnectErr := c.reconnect(ctx); reconnectErr != nil {
						log.Err(reconnectErr).Msg("failed to reconnect to SRS server")
						return
					}
					if initErr := c.initialize(); initErr != nil {
						log.Err(initErr).Msg("failed to reinitialize SRS client")
						return
					}
					c.lastPing = time.Now()
				}
			}()
		}
	}
}

// Run starts the SimpleRadio-Standalone client. It should be called exactly once.
func (c *Client) Run(ctx context.Context, wg *sync.WaitGroup) error {
	log.Info().Msg("SRS client starting")

	defer c.close()

	wg.Go(func() {
		c.receiveTCP(ctx)
	})

	if initErr := c.initialize(); initErr != nil {
		return initErr
	}

	// We need to send pings to the server to keep our connection alive.
	// The server won't send us any audio until it receives a ping from us.
	wg.Go(func() {
		c.sendPings(ctx)
	})

	udpPingRxChan := make(chan []byte, 0xF)
	wg.Go(func() {
		c.receivePings(ctx, udpPingRxChan)
	})

	udpVoiceRxChan := make(chan []byte, 64*0xFFFFF)
	voiceBytesRxChan := make(chan []voice.Packet, 0xFFFFF)
	wg.Add(2)
	go func() {
		defer wg.Done()
		c.receiveVoice(ctx, udpVoiceRxChan, voiceBytesRxChan)
	}()
	go func() {
		defer wg.Done()
		c.decodeVoice(ctx, voiceBytesRxChan)
	}()

	voicePacketsTxChan := make(chan []voice.Packet, 3)
	wg.Add(4)
	go func() {
		defer wg.Done()
		c.encodeVoice(ctx, voicePacketsTxChan)
	}()
	go func() {
		defer wg.Done()
		c.transmitPackets(ctx, voicePacketsTxChan)
	}()
	go func() {
		defer wg.Done()
		c.receiveUDP(ctx, udpPingRxChan, udpVoiceRxChan)
	}()
	go func() {
		defer wg.Done()
		c.autoheal(ctx)
	}()

	<-ctx.Done()
	return nil
}

func (c *Client) getPeerName(guid types.GUID) (string, bool) {
	c.clientsLock.RLock()
	defer c.clientsLock.RUnlock()
	info, ok := c.clients[guid]
	if ok {
		return info.Name, true
	}
	return "", false
}

// close the client's connections. Should be called after the autoheal goroutine has completed.
func (c *Client) close() {
	var err error
	if tcpErr := c.tcpConnection.Close(); tcpErr != nil {
		err = errors.Join(err, fmt.Errorf("error closing TCP connection to SRS: %w", tcpErr))
	}
	if udpErr := c.udpConnection.Close(); udpErr != nil {
		err = errors.Join(err, fmt.Errorf("error closing UDP connection to SRS: %w", udpErr))
	}
	if err != nil {
		log.Error().Err(err).Msg("error closing SRS client connections")
	}
}
