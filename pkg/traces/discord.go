package traces

import (
	"context"
	"fmt"
	"strings"
	"time"

	discord "github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
)

type DiscordWebhook struct {
	session *discord.Session
	id      string
	token   string
}

var _ Tracer = (*DiscordWebhook)(nil)

func NewDiscordWebhook(webhookID, token string) (*DiscordWebhook, error) {
	session, err := discord.New("")
	if err != nil {
		return nil, fmt.Errorf("failed to create Discord session: %w", err)
	}
	webhook := &DiscordWebhook{
		session: session,
		id:      webhookID,
		token:   token,
	}
	return webhook, nil
}

func createReport(ctx context.Context) string {
	content := "GCI Workflow Report"
	if traceID := GetTraceID(ctx); traceID != "" {
		content += fmt.Sprintf(" (Trace ID: `%s`)", traceID)
	}
	content += "\n"
	if callsign := GetCallsign(ctx); callsign != "" {
		content += fmt.Sprintf("GCI Callsign: %s\n", callsign)
	}
	if clientName := GetClientName(ctx); clientName != "" {
		clientName = strings.ReplaceAll(clientName, "`", "\\`")
		content += fmt.Sprintf("SRS Client Name: `%s`\n", clientName)
	}
	if text := GetRequestText(ctx); text != "" {
		content += fmt.Sprintf("Recognized: %q\n", text)
	}
	request := GetRequest(ctx)
	if request != nil {
		content += fmt.Sprintf("Parsed: `%s`\n", request)
	}
	if text := GetCallText(ctx); text != "" {
		if request != nil {
			content += "Responded: "
		} else {
			content += "Broadcast: "
		}
		content += fmt.Sprintf("%q\n", text)
	}
	if err := GetRequestError(ctx); err != nil {
		content += fmt.Sprintf("Error: `%v`\n", err)
	}

	if timings := formatTimings(ctx); timings != "" {
		content += fmt.Sprintf("Timings: %s\n", timings)
	}

	if call := GetCall(ctx); call != nil {
		content += fmt.Sprintf("Internal Response: `%s`\n", call)
	}
	return content
}

type timing struct {
	stage    string
	duration time.Duration
}

func formatTimings(ctx context.Context) string {
	timings := make([]timing, 0)

	addTiming := func(stage string, start time.Time, end time.Time) {
		if !start.IsZero() && !end.IsZero() {
			duration := end.Sub(start).Round(time.Millisecond)
			if duration > time.Millisecond {
				timings = append(timings, timing{stage, duration})
			}
		}
	}

	receivedAt := GetReceivedAt(ctx)
	recogizedAt := GetRecognizedAt(ctx)
	parsedAt := GetParsedAt(ctx)
	handledAt := GetHandledAt(ctx)
	composedAt := GetComposedAt(ctx)
	synthesizedAt := GetSynthesizedAt(ctx)
	submittedAt := GetSubmittedAt(ctx)

	addTiming("Recognition", receivedAt, recogizedAt)
	addTiming("Parsing", recogizedAt, parsedAt)
	addTiming("Handling", parsedAt, handledAt)
	addTiming("Composition", handledAt, composedAt)
	addTiming("Synthesis", composedAt, synthesizedAt)
	addTiming("Submission", synthesizedAt, submittedAt)
	var s string
	if len(timings) > 0 {
		var totalDuration time.Duration
		for _, timing := range timings {
			s += fmt.Sprintf("%s: %s, ", timing.stage, timing.duration)
			totalDuration += timing.duration
		}
		s += fmt.Sprintf("Total: %s", totalDuration)
	}
	return s
}

func (w *DiscordWebhook) Trace(ctx context.Context) {
	params := &discord.WebhookParams{
		Content:         createReport(ctx),
		AllowedMentions: &discord.MessageAllowedMentions{},
	}
	_, err := w.session.WebhookExecute(w.id, w.token, false, params)
	if err != nil {
		log.Error().Err(err).Msg("failed to send Discord webhook")
		return
	}
}
