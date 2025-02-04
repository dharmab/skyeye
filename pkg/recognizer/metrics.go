package recognizer

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const statusAttributeKey = "status"

var (
	recognizerCounter           metric.Int64Counter
	recognizerErrorCounter      metric.Int64Counter
	recognizerDurationCounter   metric.Float64Counter
	recognizerDurationHistogram metric.Float64Histogram
)

func init() {
	meter := otel.Meter("")
	var err error
	recognizerCounter, err = meter.Int64Counter(
		"recognizer.counter",
		metric.WithDescription("Number of Recognize function calls"),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		panic(err)
	}
	recognizerErrorCounter, err = meter.Int64Counter(
		"recognizer.error.counter",
		metric.WithDescription("Number of Recognize function errors"),
		metric.WithUnit("{error}"),
	)
	recognizerDurationCounter, err = meter.Float64Counter(
		"recognizer.duration.counter",
		metric.WithDescription("Time spent recognizing speech"),
		metric.WithUnit("s"),
	)
	if err != nil {
		panic(err)
	}
	recognizerDurationHistogram, err = meter.Float64Histogram(
		"recognizer.duration.histogram",
		metric.WithDescription("Aggregated time spent recognizing speech"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0, 0.25, 0.5, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 30, 60),
	)
	if err != nil {
		panic(err)
	}
}

func record(ctx context.Context, ok bool, startedAt time.Time) {
	recognizerCounter.Add(ctx, 1)
	if !ok {
		recognizerErrorCounter.Add(ctx, 1)
	}

	sec := time.Since(startedAt).Seconds()
	recognizerDurationCounter.Add(ctx, sec)
	recognizerDurationHistogram.Record(ctx, sec)
}
