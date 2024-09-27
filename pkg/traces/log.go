package traces

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LogTracer struct {
	Level zerolog.Level
}

var _ Tracer = (*LogTracer)(nil)

func NewLogTracer() *LogTracer {
	return &LogTracer{Level: zerolog.TraceLevel}
}

func (t *LogTracer) Trace(ctx context.Context) {
	loggerCtx := log.WithLevel(t.Level)
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
	if call := GetCall(ctx); call != nil {
		loggerCtx = loggerCtx.Type("callType", call).Any("call", call)
	}
	loggerCtx.Msg("workflow trace")
}
