package traces

import (
	"context"
	"time"

	"github.com/lithammer/shortuuid/v3"
)

// NewRequestContext returns a new context with a new trace ID.
func NewRequestContext() context.Context {
	return WithTraceID(context.Background(), shortuuid.New())
}

type contextKey int

const (
	traceIDKey contextKey = iota
	errorKey
	radioFrequencyKey
	clientNameKey
	playerNameKey
	requestKey
	requestTextKey
	callTextKey
	receivedAtKey
	recognizedAtKey
	parsedAtKey
	handledAtKey
	composedAtKey
	synthesizedAtKey
	submittedAtKey
)

func getValue[T any](ctx context.Context, key contextKey) T {
	v := ctx.Value(key)
	if v == nil {
		var t T
		return t
	}
	return v.(T)
}

// WithTraceID returns a new context with the given trace ID.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID returns the trace ID from the context, or an empty string if no trace ID is set.
func GetTraceID(ctx context.Context) string {
	return getValue[string](ctx, traceIDKey)
}

// WithRequestError returns a new context with the given error.
func WithRequestError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, errorKey, err)
}

// GetRequestError returns the error from the context, or nil if no error is set.
func GetRequestError(ctx context.Context) error {
	return getValue[error](ctx, errorKey)
}

// WithClientName returns a new context with the given client name.
func WithClientName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, clientNameKey, name)
}

// GetClientName returns the client name from the context, or an empty string if no client name is set.
func GetClientName(ctx context.Context) string {
	return getValue[string](ctx, clientNameKey)
}

// WithPlayerName returns a new context with the given player name.
func WithPlayerName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, playerNameKey, name)
}

// GetPlayerName returns the player name from the context, or an empty string if no player name is set.
func GetPlayerName(ctx context.Context) string {
	return getValue[string](ctx, playerNameKey)
}

// WithRequest returns a new context with the given request.
func WithRequest(ctx context.Context, request any) context.Context {
	return context.WithValue(ctx, requestKey, request)
}

// GetRequest returns the request from the context, or nil if no request is set.
func GetRequest(ctx context.Context) any {
	return ctx.Value(requestKey)
}

// WithRequestText returns a new context with the given request text.
func WithRequestText(ctx context.Context, text string) context.Context {
	return context.WithValue(ctx, requestTextKey, text)
}

// GetRequestText returns the request text from the context, or an empty string if no request text is set.
func GetRequestText(ctx context.Context) string {
	return getValue[string](ctx, requestTextKey)
}

// WithoutRequestText returns a new context without any request text. This is useful for removing user input whe enhanced privacy is desired.
func WithoutRequestText(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestTextKey, nil)
}

// WithCallText returns a new context with the given call text.
func WithCallText(ctx context.Context, text string) context.Context {
	return context.WithValue(ctx, callTextKey, text)
}

// GetCallText returns the call text from the context, or an empty string if no call text is set.
func GetCallText(ctx context.Context) string {
	return getValue[string](ctx, callTextKey)
}

// WithReceivedAt returns a new context with the given time the request was received.
func WithReceivedAt(ctx context.Context, receivedAt time.Time) context.Context {
	return context.WithValue(ctx, receivedAtKey, receivedAt)
}

// GetReceivedAt returns the time the request was received, or the zero time if no received at time is set.
func GetReceivedAt(ctx context.Context) time.Time {
	return getValue[time.Time](ctx, receivedAtKey)
}

// WithRecognizedAt returns a new context with the given time the request was recognized.
func WithRecognizedAt(ctx context.Context, recognizedAt time.Time) context.Context {
	return context.WithValue(ctx, recognizedAtKey, recognizedAt)
}

// GetRecognizedAt returns the time the request was recognized, or the zero time if no recognized at time is set.
func GetRecognizedAt(ctx context.Context) time.Time {
	return getValue[time.Time](ctx, recognizedAtKey)
}

// WithParsedAt returns a new context with the given time the request was parsed.
func WithParsedAt(ctx context.Context, parsedAt time.Time) context.Context {
	return context.WithValue(ctx, parsedAt, parsedAt)
}

// GetParsedAt returns the time the request was parsed, or the zero time if no parsed at time is set.
func GetParsedAt(ctx context.Context) time.Time {
	return getValue[time.Time](ctx, parsedAtKey)
}

// WithHandledAt returns a new context with the given time the request was handled by the controller.
func WithHandledAt(ctx context.Context, handledAt time.Time) context.Context {
	return context.WithValue(ctx, handledAtKey, handledAt)
}

// GetHandledAt returns the time the request was handled by the controller, or the zero time if no handled at time is set.
func GetHandledAt(ctx context.Context) time.Time {
	return getValue[time.Time](ctx, handledAtKey)
}

// WithComposedAt returns a new context with the given time the response was composed.
func WithComposedAt(ctx context.Context, composedAt time.Time) context.Context {
	return context.WithValue(ctx, composedAtKey, composedAt)
}

// GetComposedAt returns the time the response was composed, or the zero time if no composed at time is set.
func GetComposedAt(ctx context.Context) time.Time {
	return getValue[time.Time](ctx, composedAtKey)
}

// WithSynthesizedAt returns a new context with the given time the response audio was synthesized.
func WithSynthesizedAt(ctx context.Context, synthesizedAt time.Time) context.Context {
	return context.WithValue(ctx, synthesizedAtKey, synthesizedAt)
}

// GetSynthesizedAt returns the time the response audio was synthesized, or the zero time if no synthesized at time is set.
func GetSynthesizedAt(ctx context.Context) time.Time {
	return getValue[time.Time](ctx, synthesizedAtKey)
}

// WithSubmittedAt returns a new context with the given time the response was submitted to the SRS client.
func WithSubmittedAt(ctx context.Context, submittedAt time.Time) context.Context {
	return context.WithValue(ctx, submittedAtKey, submittedAt)
}

// GetSubmittedAt returns the time the response was submitted to the SRS client, or the zero time if no submitted at time is set.
func GetSubmittedAt(ctx context.Context) time.Time {
	return getValue[time.Time](ctx, submittedAtKey)
}
