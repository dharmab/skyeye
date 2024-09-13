// package simpleradio contains a bespoke SimpleRadio-Standalone client.
package simpleradio

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	// IsOnFrequency checks if the named unit is on any of the client's frequencies.
	IsOnFrequency(string) bool
}

// client implements the SRS Client.
type client struct {
	// externalAWACSModePassword is the password for authenticating as an external AWACS in the SRS server.
	externalAWACSModePassword string

	// tcpConnection is the TCP connection to the SRS server used for messages.
	tcpConnection *net.TCPConn
	// udpConnection is the UDP connection to the SRS server used for audio and pings.
	udpConnection *net.UDPConn

	// clientInfo is the client information for this client. It is what players will see in the SRS client list, and the in-game overlay when this client transmits.
	clientInfo types.ClientInfo
	// clients is a map of GUIDs to client info, which the bot will use to filter out other clients that are not in the same coalition and frequency.
	clients map[types.GUID]types.ClientInfo
	// clientsLock controls access to the clients map.
	clientsLock sync.RWMutex

	// rxChan is a channel where received audio is published. A read-only version is available publicly.
	rxchan chan Audio
	// txChan is a channel where audio to be transmitted is buffered.
	txChan chan Audio
	// receivers tracks the state of each radio we are listening to.
	receivers map[types.Radio]*receiver
	// packetNumber is incremented for each voice packet transmitted.
	packetNumber uint64
	// busy indicates if there is a transmission in progress.
	busy sync.Mutex
	// mute suppresses audio transmission.
	mute bool

	// lastPing tracks the last time a ping was received so we can tell when the server is (probably) restarted or offline.
	lastPing time.Time
	// lastDataReceivedAt is the most recent time data was received. If this exceeds a data timeout, we have likely been disconnected from the server.
	lastDataReceivedAt time.Time
}

func NewClient(config types.ClientConfiguration) (Client, error) {
	guid := types.NewGUID()

	log.Info().Str("address", config.Address).Msg("connecting to SRS server")
	tcpConnection, err := connectTCP(config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SRS server: %w", err)
	}
	udpConnection, err := connectUDP(config.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SRS server: %w", err)
	}

	receivers := make(map[types.Radio]*receiver, len(config.Radios))
	for _, radio := range config.Radios {
		receivers[radio] = &receiver{}
	}

	client := &client{
		tcpConnection: tcpConnection,
		udpConnection: udpConnection,
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
		rxchan:       make(chan Audio),
		receivers:    receivers,
		packetNumber: 1,
		mute:         config.Mute,
		lastPing:     time.Now(),
	}

	return client, nil
}

func connectTCP(address string) (*net.TCPConn, error) {
	log.Info().Str("address", address).Msg("connecting to SRS server TCP socket")
	tcpAddress, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve SRS server address %v: %w", address, err)
	}
	connection, err := net.DialTCP("tcp", nil, tcpAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to data socket: %w", err)
	}
	return connection, nil
}

func connectUDP(address string) (*net.UDPConn, error) {
	log.Info().Str("address", address).Msg("connecting to SRS server UDP socket")
	udpAddress, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve SRS server address %v: %w", address, err)
	}
	connection, err := net.DialUDP("udp", nil, udpAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to UDP socket: %w", err)
	}
	return connection, nil
}

// Run implements [Client.Run].
func (c *client) Run(ctx context.Context, wg *sync.WaitGroup) error {
	log.Info().Msg("SRS client starting")

	// Ensure connections are closed when the context is canceled.
	defer func() {
		if err := c.close(); err != nil {
			log.Error().Err(err).Msg("error closing SRS client")
		}
	}()

	messageChan := make(chan types.Message)
	errorChan := make(chan error)

	wg.Add(1)
	go func() {
		defer wg.Done()
		reader := bufio.NewReader(c.tcpConnection)
		for {
			if ctx.Err() != nil {
				log.Info().Msg("stopping SRS client due to context cancellation")
				return
			}
			line, err := reader.ReadBytes(byte('\n'))
			if errors.Is(err, net.ErrClosed) {
				log.Error().Err(err).Msg("TCP connection closed")
				return
			}
			switch err {
			case nil:
				var message types.Message
				jsonErr := json.Unmarshal(line, &message)
				if jsonErr != nil {
					log.Warn().Str("text", string(line)).Err(jsonErr).Msg("failed to unmarshal message")
				} else {
					messageChan <- message
				}
			case io.EOF:
				log.Trace().Msg("EOF received from SRS server")
			default:
				log.Error().Err(err).Msg("error reading from SRS server")
				errorChan <- err
				return
			}
		}
	}()

	log.Info().Msg("sending initial sync message")
	if err := c.sync(); err != nil {
		errorChan <- fmt.Errorf("initial sync failed: %w", err)
	}

	log.Info().Msg("connecting to external AWACS mode")
	if err := c.connectExternalAWACSMode(); err != nil {
		return fmt.Errorf("external AWACS mode failed: %w", err)
	}
	// We need to send pings to the server to keep our connection alive. The server won't send us any audio until it receives a ping from us.
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.sendPings(ctx, wg)
	}()

	// udpPingRxChan is a channel for received ping packets.
	udpPingRxChan := make(chan []byte, 0xF)

	// Handle incoming pings - mostly for debugging. We don't need to echo them back.
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.receivePings(ctx, udpPingRxChan)
	}()

	// receive voice packets and decode them. This is the logic for receiving audio from the SRS server.
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

	// transmit queued audio. This is the logic for sending audio to the SRS server.
	voicePacketsTxChan := make(chan []voice.VoicePacket, 3)
	wg.Add(2)
	go func() {
		defer wg.Done()
		c.encodeVoice(ctx, voicePacketsTxChan)
	}()
	go func() {
		defer wg.Done()
		c.transmit(ctx, voicePacketsTxChan)
	}()

	// Start listening for incoming UDP packets and routing them to receivePings and receiveVoice.
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.receiveUDP(ctx, udpPingRxChan, udpVoiceRxChan)
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("stopping SRS client due to context cancelation")
			return fmt.Errorf("stopping client due to context cancelation: %w", ctx.Err())
		case m := <-messageChan:
			c.lastDataReceivedAt = time.Now()
			c.handleMessage(m)
		case err := <-errorChan:
			return fmt.Errorf("client error: %w", err)
		case <-ticker.C:
			if time.Since(c.lastPing) > 1*time.Minute {
				log.Warn().Msg("stopped receiving pings from SRS server")
				return errors.New("stopped receiving pings from SRS server")
			}
		}
	}
}

func (c *client) close() error {
	var err error
	if tcpErr := c.tcpConnection.Close(); tcpErr != nil {
		err = errors.Join(err, fmt.Errorf("error closing TCP connection to SRS: %w", tcpErr))
	}
	if udpErr := c.udpConnection.Close(); udpErr != nil {
		err = errors.Join(err, fmt.Errorf("error closing UDP connection to SRS: %w", udpErr))
	}
	return err
}
