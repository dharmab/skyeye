package traces

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogTracer publishes traces as structured logs.
type LogTracer struct{}

var _ Tracer = (*LogTracer)(nil)

// Trace logs any trace information in the context.
func (*LogTracer) Trace(ctx context.Context) {
	loggerCtx := log.WithLevel(zerolog.TraceLevel)
	if traceID := GetTraceID(ctx); traceID != "" {
		loggerCtx = loggerCtx.Str("traceID", traceID)
	}
	if clientName := GetClientName(ctx); clientName != "" {
		loggerCtx = loggerCtx.Str("clientName", clientName)
	}
	if text := GetRequestText(ctx); text != "" {
		loggerCtx = loggerCtx.Str("requestText", text)
	}
	if request := GetRequest(ctx); request != nil {
		loggerCtx = loggerCtx.Type("requestType", request).Any("request", request)
	}
	if text := GetCallText(ctx); text != "" {
		loggerCtx = loggerCtx.Str("callText", text)
	}
	loggerCtx.Msg("workflow trace")
}
