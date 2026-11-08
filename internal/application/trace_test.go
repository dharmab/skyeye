package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/composer"
	"github.com/dharmab/skyeye/pkg/traces"
)

type recordingTracer struct {
	contexts []context.Context
}

func (t *recordingTracer) Trace(ctx context.Context) {
	t.contexts = append(t.contexts, ctx)
}

func TestAutomaticBroadcastTracing(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		call    any
		request any
		picture bool
		faded   bool
		threat  bool
		want    bool
	}{
		{name: "automatic picture disabled", call: brevity.PictureResponse{}, faded: true, threat: true},
		{name: "automatic picture enabled", call: brevity.PictureResponse{}, picture: true, want: true},
		{name: "requested picture unaffected", call: brevity.PictureResponse{}, request: &brevity.PictureRequest{}, want: true},
		{name: "faded disabled", call: brevity.FadedCall{}, picture: true, threat: true},
		{name: "faded enabled", call: brevity.FadedCall{}, faded: true, want: true},
		{name: "threat disabled", call: brevity.ThreatCall{}, picture: true, faded: true},
		{name: "threat enabled", call: brevity.ThreatCall{}, threat: true, want: true},
		{name: "merged unaffected", call: brevity.MergedCall{}, want: true},
		{name: "sunrise unaffected", call: brevity.SunriseCall{}, want: true},
		{name: "other response unaffected", call: brevity.RadioCheckResponse{}, want: true},
		{name: "no call unaffected", want: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			first, second := &recordingTracer{}, &recordingTracer{}
			app := Application{
				tracers:                       []traces.Tracer{first, second},
				enableAutomaticPictureTracing: tt.picture,
				enableAutomaticFadedTracing:   tt.faded,
				enableAutomaticThreatTracing:  tt.threat,
			}
			ctx := traces.NewRequestContext()
			if tt.call != nil {
				ctx = context.WithValue(ctx, traceCallKey{}, tt.call)
			}
			if tt.request != nil {
				ctx = traces.WithRequest(ctx, tt.request)
			}
			ctx = traces.WithRequestText(ctx, "private request")
			ctx = traces.WithCallText(ctx, "broadcast")
			ctx = traces.WithSynthesizedAt(ctx, time.Now())
			for _, err := range []error{nil, errors.New("synthesis failed")} {
				first.contexts, second.contexts = nil, nil
				app.trace(traces.WithRequestError(ctx, err))
				wantCount := 0
				if tt.want {
					wantCount = 1
				}
				for _, tracer := range []*recordingTracer{first, second} {
					if len(tracer.contexts) != wantCount {
						t.Fatalf("published %d traces, want %d (error: %v)", len(tracer.contexts), wantCount, err)
					}
					if tt.want && traces.GetRequestText(tracer.contexts[0]) != "" {
						t.Error("transcription privacy was not preserved")
					}
				}
			}
		})
	}
}

func TestComposePicturePreservesBroadcastWhenTracingDisabled(t *testing.T) {
	t.Parallel()
	tracer := &recordingTracer{}
	app := Application{tracers: []traces.Tracer{tracer}}
	out := make(chan Message[composer.NaturalLanguageResponse], 1)
	app.composeCall(traces.NewRequestContext(), brevity.PictureResponse{}, out)
	message := <-out
	if message.Data.Speech == "" {
		t.Fatal("expected picture broadcast to still be composed")
	}
	app.trace(message.Context)
	if len(tracer.contexts) != 0 {
		t.Fatal("automatic picture trace was published despite being disabled")
	}
}
