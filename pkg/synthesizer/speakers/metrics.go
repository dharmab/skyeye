package speakers

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var (
	synthesizerCounter           metric.Int64Counter
	synthesizerErrorCounter      metric.Int64Counter
	synthesizerDurationCounter   metric.Float64Counter
	synthesizerDurationHistogram metric.Float64Histogram
)

func init() {
	meter := otel.Meter("")
	var err error
	synthesizerCounter, err = meter.Int64Counter(
		"synthesizer.counter",
		metric.WithDescription("Number of Synthesize function calls"),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		panic(err)
	}
	synthesizerErrorCounter, err = meter.Int64Counter(
		"synthesizer.error.counter",
		metric.WithDescription("Number of Synthesize function errors"),
		metric.WithUnit("{error}"),
	)
	if err != nil {
		panic(err)
	}
	synthesizerDurationCounter, err = meter.Float64Counter(
		"synthesizer.duration.counter",
		metric.WithDescription("Time spent synthesizing speech"),
		metric.WithUnit("s"),
	)
	if err != nil {
		panic(err)
	}
	synthesizerDurationHistogram, err = meter.Float64Histogram(
		"synthesizer.duration.histogram",
		metric.WithDescription("Aggregated time spent synthesizing speech"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0, 0.25, 0.5, 1, 2.0, 3, 4, 5, 6, 7, 8, 9, 10),
	)
	if err != nil {
		panic(err)
	}
}

func record(ctx context.Context, ok bool, startedAt time.Time) {
	synthesizerCounter.Add(ctx, 1)
	if !ok {
		synthesizerErrorCounter.Add(ctx, 1)
	}
	sec := time.Since(startedAt).Seconds()
	synthesizerDurationCounter.Add(ctx, sec)
	synthesizerDurationHistogram.Record(ctx, sec)
}
