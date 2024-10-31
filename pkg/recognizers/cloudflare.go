package recognizers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type cloudRecognizer struct {
	client    *http.Client
	accountID string
	token     string
	callsign  string
}

var _ Recognizer = &cloudRecognizer{}

// NewCloudflareRecognizer creates a new recognizer using the Cloudflare Workers AI API.
func NewCloudflareRecognizer(accountID, token, callsign string) (Recognizer, error) {
	return &cloudRecognizer{
		client:    &http.Client{},
		accountID: accountID,
		token:     token,
		callsign:  callsign,
	}, nil
}

type cloudflareRequest struct {
	Audio          []uint `json:"audio"`
	SourceLanguage string `json:"source_lang"`
	TargetLanguage string `json:"target_lang"`
}

type cloudflareResponse struct {
	Text string `json:"text"`
}

// Recognize implements [Recognizer.Recognize] using the Cloudflare Workers AI API.
func (r *cloudRecognizer) Recognize(ctx context.Context, sample []float32, enableTranscriptionLogging bool) (string, error) {
	rCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/@cf/openai/whisper", r.accountID)
	audio := make([]uint, len(sample))
	for i, f := range sample {
		audio[i] = uint(uint8(127 + f*128))
	}
	input := cloudflareRequest{
		Audio:          audio,
		SourceLanguage: "en",
		TargetLanguage: "en",
	}
	body, err := json.Marshal(input)
	log.Debug().Str("body", string(body)).Msg("request body")
	if err != nil {
		return "", fmt.Errorf("error marshaling request body: %w", err)
	}
	request, err := http.NewRequestWithContext(rCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("error creating POST request: %w", err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.token))
	request.Header.Set("Content-Type", "application/json")

	logger := log.With().Str("accountID", r.accountID).Str("url", url).Logger()
	logger.Info().Msg("attempting to POST audio data to Cloudflare Workers AI")
	response, err := r.client.Do(request)
	if err != nil {
		logger.Error().Err(err).Msg("failure to POST audio data to Cloudflare Workers AI")
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		event := log.Error().Int("status", response.StatusCode)
		if body, err := io.ReadAll(response.Body); err == nil {
			event = event.Str("body", string(body))
		}
		event.Msg("unexpected status code returned from Cloudflare Workers AI")
		return "", fmt.Errorf("unexpected status code returned from Cloudflare Workers AI: %d", response.StatusCode)
	}
	var o cloudflareResponse
	if err := json.NewDecoder(response.Body).Decode(&o); err != nil {
		return "", fmt.Errorf("error decoding response from Cloudflare Workers AI: %w", err)
	}

	if enableTranscriptionLogging {
		logger.Info().Str("text", o.Text).Msg("transcription received from Cloudflare Workers AI")
	}

	return o.Text, nil
}
