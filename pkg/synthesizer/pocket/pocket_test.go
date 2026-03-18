//go:build integration

package pocket_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/parser"
	"github.com/dharmab/skyeye/pkg/recognizer/parakeet"
	parakeetmodel "github.com/dharmab/skyeye/pkg/recognizer/parakeet/model"
	"github.com/dharmab/skyeye/pkg/synthesizer/pocket"
	pocketmodel "github.com/dharmab/skyeye/pkg/synthesizer/pocket/model"
	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const gciCallsign = "Magic"

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
	assert.Equal(t, "eagle 2 1", actual.Callsign)
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
