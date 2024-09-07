// package simpleradio contains a bespoke SimpleRadio-Standalone client.
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

type Audio []float32

// Client is a SimpleRadio-Standalone client.
type Client interface {
	// Run starts the SimpleRadio-Standalone client. It should be called exactly once.
	Run(context.Context, *sync.WaitGroup) error
	// Send sends a message to the SRS server.
	Send(types.Message) error
	// Receive returns a channel that receives transmissions over the radio. Each transmission is F32LE PCM audio data.
	Receive() <-chan Audio
	// Transmit queues a transmission to send over the radio. The audio data should be in F32LE PCM format.
	Transmit(Audio)
	// Frequencies returns the frequencies the client is listening on.
	Frequencies() []RadioFrequency
	// ClientsOnFrequency returns the number of peers on the client's frequencies.
	ClientsOnFrequency() int
	// HumansOnFrequency returns the number of human peers on the client's frequencies.
	// A human peer is any client whose name does not end with "[BOT]".
	HumansOnFrequency() int
	// BotsOnFrequency returns the number of bot peers on the client's frequencies.
	// A bot peer is any client whose name ends with "[BOT]".
	BotsOnFrequency() int
	// IsOnFrequency checks if the named unit is on any of the client's frequencies.
	IsOnFrequency(string) bool
}

// client implements the SRS Client.
type client struct {
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

	// rxChan is a channel where received audio is published. A read-only version is available publicly.
	rxChan chan Audio
	// txChan is a channel where audio to be transmitted is buffered.
	txChan chan Audio
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
	lastPing time.Time
}

func NewClient(config types.ClientConfiguration) (Client, error) {
	guid := types.NewGUID()

	receivers := make(map[types.Radio]*receiver, len(config.Radios))
	for _, radio := range config.Radios {
		receivers[radio] = &receiver{}
	}

	client := &client{
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
			Position: &types.Position{},
		},
		externalAWACSModePassword: config.ExternalAWACSModePassword,
		clients:                   make(map[types.GUID]types.ClientInfo),

		txChan:       make(chan Audio),
		rxChan:       make(chan Audio),
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
func (c *client) initialize() error {
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
func (c *client) autoheal(ctx context.Context) {
	ticker := time.NewTicker(pingInterval / 3)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if time.Since(c.lastPing) > pingInterval*3 {
				log.Warn().Msg("stopped receiving traffic from SRS server")

				log.Warn().Msg("attempting to reconnect to SRS server")
				if reconnectErr := c.reconnect(ctx); reconnectErr != nil {
					log.Err(reconnectErr).Msg("failed to reconnect to SRS server")
					continue
				}
				if initErr := c.initialize(); initErr != nil {
					log.Err(initErr).Msg("failed to reinitialize SRS client")
					continue
				}
				c.lastPing = time.Now()
			}
		}
	}
}

// Run implements [Client.Run].
func (c *client) Run(ctx context.Context, wg *sync.WaitGroup) error {
	log.Info().Msg("SRS client starting")

	defer c.close()

	wg.Add(1)
	go func() {
		defer wg.Done()
		c.receiveTCP(ctx)
	}()

	if initErr := c.initialize(); initErr != nil {
		return initErr
	}

	// We need to send pings to the server to keep our connection alive.
	// The server won't send us any audio until it receives a ping from us.
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.sendPings(ctx)
	}()

	udpPingRxChan := make(chan []byte, 0xF)
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.receivePings(ctx, udpPingRxChan)
	}()

	udpVoiceRxChan := make(chan []byte, 64*0xFFFFF)
	voiceBytesRxChan := make(chan []voice.VoicePacket, 0xFFFFF)
	wg.Add(2)
	go func() {
		defer wg.Done()
		c.receiveVoice(ctx, udpVoiceRxChan, voiceBytesRxChan)
	}()
	go func() {
		defer wg.Done()
		c.decodeVoice(ctx, voiceBytesRxChan)
	}()

	voicePacketsTxChan := make(chan []voice.VoicePacket, 3)
	wg.Add(4)
	go func() {
		defer wg.Done()
		c.encodeVoice(ctx, voicePacketsTxChan)
	}()
	go func() {
		defer wg.Done()
		c.transmit(ctx, voicePacketsTxChan)
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

// close the client's connections. Should be called after the autoheal goroutine has completed.
func (c *client) close() {
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
