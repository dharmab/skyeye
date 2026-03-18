//go:build integration

package pocket_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/radar"
	"github.com/dharmab/skyeye/pkg/recognizer/parakeet"
	parakeetmodel "github.com/dharmab/skyeye/pkg/recognizer/parakeet/model"
	"github.com/dharmab/skyeye/pkg/synthesizer/pocket"
	pocketmodel "github.com/dharmab/skyeye/pkg/synthesizer/pocket/model"
	fuzz "github.com/hbollon/go-edlib"
	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const gciCallsign = "Magic"

// callsignCandidates returns all "word X Y" callsigns from 1 1 to 9 9
// for the given callsign word, simulating what would exist in the radar database.
func callsignCandidates(word string) []string {
	candidates := make([]string, 0, 81)
	for f := 1; f <= 9; f++ {
		for s := 1; s <= 9; s++ {
			candidates = append(candidates, fmt.Sprintf("%s %d %d", word, f, s))
		}
	}
	return candidates
}

// newPipeline sets up the TTS speaker, STT recognizer, and parser for integration tests.
// It skips the test if model files are not available.
func newPipeline(t *testing.T) (*pocket.Speaker, *parser.Parser, func(string) string) {
	t.Helper()

	modelsPath := os.Getenv("SKYEYE_MODELS_PATH")
	if modelsPath == "" {
		modelsPath = "models"
	}

	pocketDir := filepath.Join(modelsPath, pocketmodel.DirName)
	require.NoError(t, pocketmodel.Verify(pocketDir), "Pocket TTS model files must be present")

	parakeetDir := filepath.Join(modelsPath, parakeetmodel.DirName)
	require.NoError(t, parakeetmodel.Verify(parakeetDir), "Parakeet model files must be present")

	speaker, err := pocket.New(pocketDir)
	require.NoError(t, err)

	rec, err := parakeet.NewRecognizer(parakeetDir)
	require.NoError(t, err)

	p := parser.New(gciCallsign, true)

	// synthesizeAndRecognize runs TTS→STT and returns the recognized text.
	synthesizeAndRecognize := func(text string) string {
		t.Helper()
		audio, err := speaker.Say(context.Background(), text)
		require.NoError(t, err)
		require.NotEmpty(t, audio)

		recognized, err := rec.Recognize(context.Background(), audio, false)
		require.NoError(t, err)
		t.Logf("Input:      %q", text)
		t.Logf("Recognized: %q", recognized)
		return recognized
	}

	return speaker, p, synthesizeAndRecognize
}

func TestRoundTripRadioCheck(t *testing.T) {
	t.Parallel()
	speaker, p, recognize := newPipeline(t)
	defer speaker.Close()

	recognized := recognize("Magic, Falcon 2 1, radio check")
	request := p.Parse(recognized)
	require.IsType(t, &brevity.RadioCheckRequest{}, request)
	actual := request.(*brevity.RadioCheckRequest)
	assert.Equal(t, "falcon 2 1", actual.Callsign)
}

func TestRoundTripAlphaCheck(t *testing.T) {
	t.Parallel()
	speaker, p, recognize := newPipeline(t)
	defer speaker.Close()

	recognized := recognize("Magic, Viper 3 1, alpha check")
	request := p.Parse(recognized)
	require.IsType(t, &brevity.AlphaCheckRequest{}, request)
	actual := request.(*brevity.AlphaCheckRequest)
	assert.Equal(t, "viper 3 1", actual.Callsign)
}

func TestRoundTripBogeyDope(t *testing.T) {
	t.Parallel()
	speaker, p, recognize := newPipeline(t)
	defer speaker.Close()

	recognized := recognize("Magic, Hornet 4 1, bogey dope")
	request := p.Parse(recognized)
	require.IsType(t, &brevity.BogeyDopeRequest{}, request)
	actual := request.(*brevity.BogeyDopeRequest)
	assert.Equal(t, "hornet 4 1", actual.Callsign)
	assert.Equal(t, brevity.Aircraft, actual.Filter)
}

func TestRoundTripPicture(t *testing.T) {
	t.Parallel()
	speaker, p, recognize := newPipeline(t)
	defer speaker.Close()

	recognized := recognize("Magic, Eagle 2 1, picture")
	request := p.Parse(recognized)
	require.IsType(t, &brevity.PictureRequest{}, request)
	actual := request.(*brevity.PictureRequest)
	// Snap callsign using edit distance against a multi-flight candidate list,
	// mirroring the real radar database.
	var candidates []string
	for _, w := range []string{"eagle", "mobius", "wardog"} {
		candidates = append(candidates, callsignCandidates(w)...)
	}
	snapped, err := fuzz.FuzzySearchThreshold(actual.Callsign, candidates, radar.CallsignSimilarityThreshold, fuzz.Levenshtein)
	require.NoError(t, err)
	assert.Equal(t, "eagle 2 1", snapped, "callsign=%q did not snap to eagle 2 1", actual.Callsign)
}

func TestRoundTripSpiked(t *testing.T) {
	t.Parallel()
	speaker, p, recognize := newPipeline(t)
	defer speaker.Close()

	recognized := recognize("Magic, Cobra 3 1, spiked, one eight zero")
	request := p.Parse(recognized)
	require.IsType(t, &brevity.SpikedRequest{}, request)
	actual := request.(*brevity.SpikedRequest)
	assert.Equal(t, "cobra 3 1", actual.Callsign)
	assert.Equal(t, bearings.NewMagneticBearing(180*unit.Degree), actual.Bearing)
}

// TestRoundTripCallsignNumbers tests TTS→STT→parser round trips across many
// callsign words, number combinations, and request phrasings. Since TTS→STT
// is inherently lossy, individual permutations may fail — the test uses a
// probabilistic approach and requires an overall success rate above 99%.
func TestRoundTripCallsignNumbers(t *testing.T) {
	t.Parallel()

	callsignWords := []string{"Eagle", "Mobius", "Wardog"}
	requestPhrases := []string{"bogey dope", "request bogey dope"}

	// Build a combined candidate list with all callsign words, simulating
	// a mission with multiple flights in the radar database.
	var allCandidates []string
	for _, word := range callsignWords {
		allCandidates = append(allCandidates, callsignCandidates(strings.ToLower(word))...)
	}

	type testInput struct {
		input            string
		expectedCallsign string
	}

	// Build the full list of inputs.
	var inputs []testInput
	for _, word := range callsignWords {
		wordLower := strings.ToLower(word)
		for first := 1; first <= 9; first++ {
			for second := 1; second <= 9; second++ {
				expectedCallsign := fmt.Sprintf("%s %d %d", wordLower, first, second)
				for _, phrase := range requestPhrases {
					inputs = append(inputs, testInput{
						input:            fmt.Sprintf("Magic, %s %d %d, %s", word, first, second, phrase),
						expectedCallsign: expectedCallsign,
					})
				}
			}
		}
	}

	type result struct {
		input      string
		recognized string
		success    bool
		detail     string
	}

	// Create a worker pool of TTS→STT pipelines.
	numWorkers := max(runtime.NumCPU()/2, 1)
	t.Logf("Using %d workers for %d test inputs", numWorkers, len(inputs))

	results := make([]result, len(inputs))
	work := make(chan int, len(inputs))
	for i := range inputs {
		work <- i
	}
	close(work)

	var wg sync.WaitGroup
	for w := range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Each worker gets its own pipeline (TTS + STT are not thread-safe).
			modelsPath := os.Getenv("SKYEYE_MODELS_PATH")
			if modelsPath == "" {
				modelsPath = "models"
			}

			pocketDir := filepath.Join(modelsPath, pocketmodel.DirName)
			if err := pocketmodel.Verify(pocketDir); err != nil {
				t.Errorf("worker %d: pocket model not available: %v", w, err)
				return
			}

			speaker, err := pocket.New(pocketDir)
			if err != nil {
				t.Errorf("worker %d: failed to create speaker: %v", w, err)
				return
			}
			defer speaker.Close()

			parakeetDir := filepath.Join(modelsPath, parakeetmodel.DirName)
			rec, err := parakeet.NewRecognizer(parakeetDir)
			if err != nil {
				t.Errorf("worker %d: failed to create recognizer: %v", w, err)
				return
			}

			p := parser.New(gciCallsign, true)

			for idx := range work {
				ti := inputs[idx]
				audio, err := speaker.Say(context.Background(), ti.input)
				if err != nil {
					results[idx] = result{input: ti.input, detail: fmt.Sprintf("TTS failed: %v", err)}
					continue
				}

				recognized, err := rec.Recognize(context.Background(), audio, false)
				if err != nil {
					results[idx] = result{input: ti.input, detail: fmt.Sprintf("STT failed: %v", err)}
					continue
				}
				t.Logf("Input:      %q", ti.input)
				t.Logf("Recognized: %q", recognized)

				request := p.Parse(recognized)
				r := result{input: ti.input, recognized: recognized}

				if request == nil {
					r.detail = "parse returned nil"
					results[idx] = r
					continue
				}

				bogeyDope, ok := request.(*brevity.BogeyDopeRequest)
				if !ok {
					r.detail = fmt.Sprintf("wrong type: %T", request)
					results[idx] = r
					continue
				}

				snapped, err := fuzz.FuzzySearchThreshold(
					bogeyDope.Callsign, allCandidates,
					radar.CallsignSimilarityThreshold, fuzz.Levenshtein,
				)
				if err != nil || snapped == "" {
					r.detail = fmt.Sprintf("callsign %q did not snap to any candidate", bogeyDope.Callsign)
					results[idx] = r
					continue
				}

				if snapped != ti.expectedCallsign {
					r.detail = fmt.Sprintf("callsign %q snapped to %q, expected %q", bogeyDope.Callsign, snapped, ti.expectedCallsign)
					results[idx] = r
					continue
				}

				r.success = true
				results[idx] = r
			}
		}()
	}
	wg.Wait()

	total := len(results)
	failures := 0
	for _, r := range results {
		if !r.success {
			failures++
			t.Logf("FAIL: input=%q recognized=%q reason=%s", r.input, r.recognized, r.detail)
		}
	}
	successRate := float64(total-failures) / float64(total)
	t.Logf("Results: %d/%d passed (%.1f%% success rate)", total-failures, total, successRate*100)
	if successRate < 0.99 {
		t.Errorf("Success rate %.1f%% is below 99%% threshold (%d failures out of %d tests)", successRate*100, failures, total)
	}
}

// FuzzRoundTrip verifies the TTS→STT→parser pipeline does not panic on arbitrary input.
// It synthesizes fuzz-generated text, recognizes it, and parses the result.
// The test passes as long as no step panics — the parser is expected to return nil for nonsense input.
func FuzzRoundTrip(f *testing.F) {
	modelsPath := os.Getenv("SKYEYE_MODELS_PATH")
	if modelsPath == "" {
		modelsPath = "models"
	}

	pocketDir := filepath.Join(modelsPath, pocketmodel.DirName)
	if err := pocketmodel.Verify(pocketDir); err != nil {
		f.Skipf("Pocket TTS model not available: %v", err)
	}

	parakeetDir := filepath.Join(modelsPath, parakeetmodel.DirName)
	if err := parakeetmodel.Verify(parakeetDir); err != nil {
		f.Skipf("Parakeet model not available: %v", err)
	}

	speaker, err := pocket.New(pocketDir)
	require.NoError(f, err)
	defer speaker.Close()

	rec, err := parakeet.NewRecognizer(parakeetDir)
	require.NoError(f, err)

	p := parser.New(gciCallsign, false)

	// Seed corpus with realistic GCI requests.
	seeds := []string{
		"Magic, Falcon 2 1, radio check",
		"Magic, Viper 3 1, alpha check",
		"Magic, Hornet 4 1, bogey dope",
		"Magic, Eagle 2 1, picture",
		"Magic, Cobra 3 1, spiked, one eight zero",
		"Magic, Raptor 1 2, bogey dope fighters",
		"Magic, Thunder 5 1, declare, bullseye zero nine zero, forty, twenty thousand",
		"Hello world, this is a test of the text to speech system",
		"Anyface, Mobius 1, radio check",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	ctx := context.Background()
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) == 0 || len(input) > 200 {
			t.Skip()
		}

		audio, err := speaker.Say(ctx, input)
		if err != nil {
			// TTS may legitimately fail on bizarre input; that's fine.
			t.Skipf("TTS failed: %v", err)
		}
		if len(audio) == 0 {
			t.Skip("TTS produced empty audio")
		}

		recognized, err := rec.Recognize(ctx, audio, false)
		if err != nil {
			t.Skipf("STT failed: %v", err)
		}

		// Parser should never panic regardless of input.
		_ = p.Parse(recognized)
	})
}
