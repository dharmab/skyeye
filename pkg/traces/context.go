package traces

import (
	"context"
	"time"

	"github.com/lithammer/shortuuid/v3"
)

func NewRequestContext() context.Context {
	return WithTraceID(context.Background(), shortuuid.New())
}

type contextKey int

const (
	traceIDKey contextKey = iota
	errorKey
	callsignKey
	radioFrequencyKey
	clientNameKey
	requestKey
	requestTextKey
	callKey
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

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func GetTraceID(ctx context.Context) string {
	return getValue[string](ctx, traceIDKey)
}

func WithRequestError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, errorKey, err)
}

func GetRequestError(ctx context.Context) error {
	return getValue[error](ctx, errorKey)
}

func WithCallsign(ctx context.Context, callsign string) context.Context {
	return context.WithValue(ctx, callsignKey, callsign)
}

func GetCallsign(ctx context.Context) string {
	return getValue[string](ctx, callsignKey)
}

func WithClientName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, clientNameKey, name)
}

func GetClientName(ctx context.Context) string {
	return getValue[string](ctx, clientNameKey)
}

func WithRequest(ctx context.Context, request any) context.Context {
	return context.WithValue(ctx, requestKey, request)
}

func GetRequest(ctx context.Context) any {
	return ctx.Value(requestKey)
}

func WithRequestText(ctx context.Context, text string) context.Context {
	return context.WithValue(ctx, requestTextKey, text)
}

func GetRequestText(ctx context.Context) string {
	return getValue[string](ctx, requestTextKey)
}

func WithoutRequestText(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestTextKey, nil)
}

func WithCallText(ctx context.Context, text string) context.Context {
	return context.WithValue(ctx, callTextKey, text)
}

func GetCallText(ctx context.Context) string {
	return getValue[string](ctx, callTextKey)
}

func WithReceivedAt(ctx context.Context, receivedAt time.Time) context.Context {
	return context.WithValue(ctx, receivedAtKey, receivedAt)
}

func GetReceivedAt(ctx context.Context) time.Time {
	return getValue[time.Time](ctx, receivedAtKey)
}

func WithRecognizedAt(ctx context.Context, recognizedAt time.Time) context.Context {
	return context.WithValue(ctx, recognizedAtKey, recognizedAt)
}

func GetRecognizedAt(ctx context.Context) time.Time {
	return getValue[time.Time](ctx, recognizedAtKey)
}

func WithParsedAt(ctx context.Context, parsedAt time.Time) context.Context {
	return context.WithValue(ctx, parsedAt, parsedAt)
}

func GetParsedAt(ctx context.Context) time.Time {
	return getValue[time.Time](ctx, parsedAtKey)
}

func WithHandledAt(ctx context.Context, handledAt time.Time) context.Context {
	return context.WithValue(ctx, handledAtKey, handledAt)
}

func GetHandledAt(ctx context.Context) time.Time {
	return getValue[time.Time](ctx, handledAtKey)
}

func WithComposedAt(ctx context.Context, composedAt time.Time) context.Context {
	return context.WithValue(ctx, composedAtKey, composedAt)
}

func GetComposedAt(ctx context.Context) time.Time {
	return getValue[time.Time](ctx, composedAtKey)
}

func WithSynthesizedAt(ctx context.Context, synthesizedAt time.Time) context.Context {
	return context.WithValue(ctx, synthesizedAtKey, synthesizedAt)
}

func GetSynthesizedAt(ctx context.Context) time.Time {
	return getValue[time.Time](ctx, synthesizedAtKey)
}

func WithSubmittedAt(ctx context.Context, submittedAt time.Time) context.Context {
	return context.WithValue(ctx, submittedAtKey, submittedAt)
}

func GetSubmittedAt(ctx context.Context) time.Time {
	return getValue[time.Time](ctx, submittedAtKey)
}
