package parser

import (
	"testing"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/require"
)

func TestParserPicture(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "anyface, intruder 1-1 request picture",
			expected: &brevity.PictureRequest{
				Callsign: "intruder 1 1",
				Radius:   conf.DefaultPictureRadius,
			},
		},
		{
			text: "anyface, intruder 1-1 picture 30",
			expected: &brevity.PictureRequest{
				Callsign: "intruder 1 1",
				Radius:   30 * unit.NauticalMile,
			},
		},
	}
	runParserTestCases(t, New(TestCallsign), testCases, func(t *testing.T, test parserTestCase, request any) {
		expected := test.expected.(*brevity.PictureRequest)
		actual := request.(*brevity.PictureRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
		require.Equal(t, expected.Radius, actual.Radius)
	})
}
