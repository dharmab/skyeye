package audio

import (
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/dharmab/skyeye/pkg/simpleradio/types"
	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
)

// maxRxGap is a duration after which the receiver will assume the end of a transmission if no packets are received.
// TODO make this configurable.
const maxRxGap = 300 * time.Millisecond

// receiveUDP listens for incoming UDP packets and routes them to the appropriate channel.
func (c *audioClient) receiveUDP(ctx context.Context, pingCh chan<- []byte, voiceCh chan<- []byte) {
	for {
		if ctx.Err() != nil {
			slog.Error("stopping packet receiver due to context error", "error", ctx.Err())
			return
		}

		udpPacketBuf := make([]byte, 1500)
		n, err := c.connection.Read(udpPacketBuf)
		udpPacket := make([]byte, n)
		copy(udpPacket, udpPacketBuf[0:n])

		switch {
		case err == io.EOF:
			slog.Error("UDP connection closed?", "error", err)
		case err != nil:
			slog.Error("UDP connection read error", "error", err)
		case n == 0:
			slog.Warn("0 bytes read from UDP connection", "error", err)
		case n < types.GUIDLength:
			slog.Debug("UDP packet smaller than expected", "bytes", n)
		case n == types.GUIDLength:
			slog.Debug("routing UDP ping packet", "bytes", n)
			pingCh <- udpPacket
		case n > types.GUIDLength:
			deadline := time.Now().Add(maxRxGap)
			slog.Debug("extending transmission receive deadline", "deadline", deadline)
			c.lastRx.deadline = deadline
			slog.Debug("routing UDP voice packet", "bytes", n)
			voiceCh <- udpPacket
		}
	}
}

// receivePings listens for incoming UDP ping packets and logs them at DEBUG level.
func (c *audioClient) receivePings(ctx context.Context, in <-chan []byte) {
	for {
		select {
		case b := <-in:
			n := len(b)
			if n < types.GUIDLength {
				slog.Debug("ping packet smaller than expected", "bytes", n)
			} else if n > types.GUIDLength {
				slog.Debug("ping packet larger than expected", "bytes", n)
			} else {
				slog.Debug("received UDP ping", "guid", b[0:types.GUIDLength])
			}
		case <-ctx.Done():
			slog.Info("stopping ping receiver due to context cancellation")
			return
		}
	}
}

// receiveVoice listens for incoming UDP voice packets, decodes them into VoicePacket structs, and routes them to the out channel for audio decoding.
func (c *audioClient) receiveVoice(ctx context.Context, in <-chan []byte, out chan<- []voice.VoicePacket) {
	// buf is a buffer of voice packets which are collected until the end of a transmission is detected.
	buf := make([]voice.VoicePacket, 0)
	// t is a ticker which triggers the check for the end of a transmission.
	// TODO make this duration configurable.
	t := time.NewTicker(50 * time.Millisecond)
	for {
		select {
		case b := <-in:
			slog.Debug("decoding voice packet")
			vp, err := decodeVoicePacket(b)
			if err != nil {
				slog.Debug("error while decoding voice packet", "error", err)
				continue
			}
			if vp == nil {
				slog.Warn("nil pointer returned from decodeVoicePacket")
				continue
			}

			slog.Debug(
				"checking voice packet",
				"originGUID", vp.OriginGUID,
				"packetID", vp.PacketID,
				"frequencies", vp.Frequencies,
			)
			// isNewPacket is true if the packet is the first packet of a new transmission. This is the case if c.lastRx's fields are zero values.
			isNewPacket := c.lastRx.origin == "" && c.lastRx.packetNumber == 0
			// isSameOrigin is true if the packet's origin GUID matches the last received packet's origin GUID.
			isSameOrigin := c.lastRx.origin == types.GUID(vp.OriginGUID)
			// isNewerPacket is true if the packet's packet number is greater than the last received packet's packet number.
			isNewerPacket := vp.PacketID > uint64(c.lastRx.packetNumber)

			// hasMatchingRadio is true if the packet's frequencies contain a frequency which matches the client's radio's frequency, modulation, and encryption settings.
			hasMatchingRadio := false
			for _, f := range vp.Frequencies {
				doesFrequencyMatch := f.Frequency == c.radio.Frequency
				doesModulationMatch := types.Modulation(f.Modulation) == c.radio.Modulation
				doesEncryptionMatch := f.Encryption == c.radio.EncryptionKey
				if doesFrequencyMatch && doesModulationMatch && doesEncryptionMatch {
					hasMatchingRadio = true
				}
			}

			// isMatchingPacket is true if the packet is either:
			// - the first packet of a new transmission
			// - a newer packet from the same origin and with matching radio frequencies as the last received packet
			isMatchingPacket := hasMatchingRadio && (isNewPacket || (isNewerPacket && isSameOrigin))
			slog.Debug("checked packet", "isMatchingPacket", isMatchingPacket, "isNewPacket", isNewPacket, "isNewerPacket", isNewerPacket, "isSameOrigin", isSameOrigin, "hasMatchingRadio", hasMatchingRadio)

			// If the packet fits, buffer it and update the lastRx state.
			if isMatchingPacket {
				slog.Debug("appending packet to voice buffer", "originGUID", vp.OriginGUID, "packetID", vp.PacketID)
				buf = append(buf, *vp)
				c.updateLastRX(vp)
			}
		case <-t.C:
			// Check if there is anything in the buffer and that we've consumed all queued packets. Then check if we've passed the receive deadline.
			// If so, we have a tranmission ready to publish for audio decoding.
			if len(buf) > 0 && len(in) == 0 && time.Now().After(c.lastRx.deadline) {
				slog.Debug("passed receive deadline with packets in buffer", "bufferLength", len(buf), "lastPacketID", c.lastRx.packetNumber, "lastOrigin", c.lastRx.origin)
				audio := make([]voice.VoicePacket, len(buf))
				copy(audio, buf)
				slog.Debug("publishing audio bytes to audio channel")
				out <- audio
				slog.Debug("reseting receiver state")
				buf = make([]voice.VoicePacket, 0)
				c.resetLastRx()
			}
		case <-ctx.Done():
			slog.Info("stopping voice receiver due to context cancellation")
			return
		}
	}
}

// updateLastRX updates the lastRx state with the origin and packet number of the given voice packet.
func (c *audioClient) updateLastRX(vp *voice.VoicePacket) {
	c.lastRx.origin = types.GUID(vp.OriginGUID)
	c.lastRx.packetNumber = vp.PacketID
}

// resetLastRx resets the lastRx state to zero values.
func (c *audioClient) resetLastRx() {
	c.lastRx.packetNumber = 0
	c.lastRx.origin = ""
}
