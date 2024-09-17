package simpleradio

import (
	"context"

	"github.com/dharmab/skyeye/pkg/simpleradio/voice"
	"github.com/rs/zerolog/log"
	"gopkg.in/hraban/opus.v2"
)

// Mirror of OPUS_APPLICATION_VOIP from the Opus API.
const opusApplicationVoIP = 2048

// deocdeVoice decodes incoming voice packets from voicePacketsChan into F32LE PCM audio data published to the client's rxChan.
func (c *client) decodeVoice(ctx context.Context, voicePacketsChan <-chan []voice.VoicePacket) {
	for {
		select {
		case voicePackets := <-voicePacketsChan:
			decoder, err := opus.NewDecoder(int(sampleRate.Hertz()), channels)
			if err != nil {
				log.Error().Err(err).Msg("failed to create Opus decoder")
				continue
			}
			transmissionPCM := make([]float32, 0)
			for _, packet := range voicePackets {
				packetPCM, err := c.decodeFrame(decoder, packet.AudioBytes)
				if err != nil {
					log.Error().Err(err).Msg("failed to decode audio")
				} else {
					transmissionPCM = append(transmissionPCM, packetPCM...)
				}
			}

			if len(transmissionPCM) > 0 {
				log.Info().Int("len", len(transmissionPCM)).Msg("publishing received audio to receiving channel")
				c.rxChan <- transmissionPCM
			} else {
				log.Debug().Msg("decoded transmission PCM is empty")
			}
		case <-ctx.Done():
			log.Info().Msg("stopping voice decoder due to context cancellation")
			return
		}
	}
}

// encodeVoice encodes audio from the client's txChan and publishes an entire transmission's worth of voice packets to packetCh.
func (c *client) encodeVoice(ctx context.Context, packetChan chan<- []voice.VoicePacket) {
	frequencyList := make([]voice.Frequency, 0, len(c.clientInfo.RadioInfo.Radios))
	for _, radio := range c.clientInfo.RadioInfo.Radios {
		frequencyList = append(frequencyList, voice.Frequency{
			Frequency:  radio.Frequency,
			Modulation: byte(radio.Modulation),
			Encryption: 0,
		})
	}
	for {
		select {
		case audio := <-c.txChan:
			log.Trace().Msg("encoding transmission from PCM data")
			encoder, err := opus.NewEncoder(int(sampleRate.Hertz()), channels, opusApplicationVoIP)
			if err != nil {
				log.Error().Err(err).Msg("failed to create Opus encoder")
				continue
			}

			txPackets := make([]voice.VoicePacket, 0)
			for i := 0; i < len(audio); i += int(frameSize) {
				logger := log.With().Int("index", i).Logger()
				var frameAudio []float32
				// pad frame to frame size
				if i+int(frameSize) < len(audio) {
					frameAudio = audio[i : i+int(frameSize)]
				} else {
					frameAudio = audio[i:]
				}
				// Align audio to Opus frame size
				if len(frameAudio) < int(frameSize) {
					padding := make([]float32, int(frameSize)-len(frameAudio))
					frameAudio = append(frameAudio, padding...)
				}
				audioBytes, err := c.encodeFrame(encoder, frameAudio)
				if err != nil {
					logger.Error().Err(err).Msg("failed to encode audio")
					continue
				}

				guid := c.clientInfo.GUID
				voicePacket := voice.NewVoicePacket(
					audioBytes,
					frequencyList,
					100000002,
					c.packetNumber,
					0,
					[]byte(guid),
					[]byte(guid),
				)
				c.packetNumber++
				// TODO transmission struct with attached text and trace id
				txPackets = append(txPackets, voicePacket)
			}
			log.Trace().Int("count", len(txPackets)).Msg("encoded transmission packets")
			packetChan <- txPackets
		case <-ctx.Done():
			log.Info().Msg("stopping voice encoder due to context cancellation")
			return
		}
	}
}