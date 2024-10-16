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

func sanitize(text string) string {
	text = strings.ReplaceAll(text, "`", "\\`")
	return fmt.Sprintf("`%s`", text)
}

func createReport(ctx context.Context) (string, []*discord.MessageEmbedField) {
	header := "GCI Workflow Report"
	fields := make([]*discord.MessageEmbedField, 0)
	if clientName := GetClientName(ctx); clientName != "" {
		clientName = strings.ReplaceAll(clientName, "`", "\\`")
		field := &discord.MessageEmbedField{
			Name:  "SRS Client Name",
			Value: sanitize(clientName),
		}
		fields = append(fields, field)
	}
	if text := GetRequestText(ctx); text != "" {
		text = strings.ReplaceAll(text, "`", "\\`")
		field := &discord.MessageEmbedField{
			Name:  "Request",
			Value: fmt.Sprintf("%q", text),
		}
		fields = append(fields, field)
	}
	request := GetRequest(ctx)
	if request != nil {
		field := &discord.MessageEmbedField{
			Name:  "Parsed",
			Value: sanitize(fmt.Sprint(request)),
		}
		fields = append(fields, field)
	}
	if err := GetRequestError(ctx); err != nil {
		field := &discord.MessageEmbedField{
			Name:  "Error",
			Value: fmt.Sprintf("`%v`", err),
		}
		fields = append(fields, field)
	}
	if text := GetCallText(ctx); text != "" {
		fieldName := "Broadcast"
		if request != nil {
			fieldName = "Response"
		}
		field := &discord.MessageEmbedField{
			Name:  fieldName,
			Value: fmt.Sprintf("%q", text),
		}
		fields = append(fields, field)
	}
	if traceID := GetTraceID(ctx); traceID != "" {
		header += " " + sanitize(traceID)
	}
	if timings := formatTimings(ctx); len(timings) > 0 {
		fields = append(fields, timings...)
	}
	return header, fields
}

type timing struct {
	stage    string
	duration time.Duration
}

func formatTimings(ctx context.Context) []*discord.MessageEmbedField {
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
	fields := make([]*discord.MessageEmbedField, 0)
	if len(timings) > 0 {
		var totalDuration time.Duration
		for _, timing := range timings {
			field := &discord.MessageEmbedField{
				Name:   timing.stage,
				Value:  timing.duration.String(),
				Inline: true,
			}
			fields = append(fields, field)
			totalDuration += timing.duration
		}
		fields = append(fields, &discord.MessageEmbedField{
			Name:   "Total",
			Value:  totalDuration.String(),
			Inline: true,
		})
	}
	return fields
}

func (w *DiscordWebhook) Trace(ctx context.Context) {
	title, fields := createReport(ctx)
	params := &discord.WebhookParams{
		Embeds: []*discord.MessageEmbed{
			{
				Title:  title,
				Fields: fields,
			},
		},
		AllowedMentions: &discord.MessageAllowedMentions{},
	}
	_, err := w.session.WebhookExecute(w.id, w.token, false, params)
	if err != nil {
		log.Error().Err(err).Msg("failed to send Discord webhook")
		return
	}
}
