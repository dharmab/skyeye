package simpleradio

import (
	"context"
	"errors"
	"math/rand/v2"
	"net"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
	"github.com/rs/zerolog/log"
)

// Transmit implements [Client.Transmit].
func (c *client) Transmit(sample Audio) {
	c.txChan <- sample
}

// transmit voice packets from queued transmissions to the SRS server.
func (c *client) transmit(ctx context.Context, packetChan <-chan []voice.VoicePacket) {
	for {
		select {
		case packets := <-packetChan:
			func() {
				c.txLock.Lock()
				defer c.txLock.Unlock()
				c.waitForClearChannel()
				if !c.mute {
					c.writePackets(packets)
				}
			}()
			// Pause between transmissions to sound more natural.
			pause := time.Duration(500+rand.IntN(500)) * time.Millisecond
			time.Sleep(pause)
		case <-ctx.Done():
			log.Info().Msg("stopping SRS audio transmitter due to context cancellation")
			return
		}
	}
}

// waitForClearChannel waits for incoming transmissions to finish.
func (c *client) waitForClearChannel() {
	for {
		isReceiving := false
		deadline := time.Now()
		for _, receiver := range c.receivers {
			if receiver.isReceivingTransmission() {
				isReceiving = true
				if receiver.deadline.After(deadline) {
					deadline = receiver.deadline
				}
			}
		}
		if isReceiving {
			delay := time.Until(deadline) + 250*time.Millisecond
			log.Info().Stringer("delay", delay).Msg("delaying outgoing transmission to avoid interrupting incoming transmission")
			time.Sleep(delay)
		} else {
			return
		}
	}
}

// writePackets writes voice packets to the UDP connection.
func (c *client) writePackets(packets []voice.VoicePacket) {
	startTime := time.Now()
	for i, packet := range packets {
		b := packet.Encode()
		// Tight timing is important here - don't write the next packet until halfway through the previous packet's frame.
		// Write too quickly, and the server will skip audio to play the latest packet.
		// Write too slowly, and the transmission will stutter.
		delay := time.Until(
			startTime.
				Add(time.Duration(i) * frameLength).
				Add(-frameLength / 2),
		)
		time.Sleep(delay)
		_, err := c.udpConnection.Write(b)
		if errors.Is(err, net.ErrClosed) {
			log.Error().Err(err).Msg("UDP connection closed during transmission")
			return
		}
		if err != nil {
			log.Error().Err(err).Msg("failed to transmit voice packet")
		}
	}
}
