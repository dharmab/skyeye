package parser

import (
	"testing"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
)

func TestParserPicture(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "anyface, intruder 1-1 request picture",
			expectedRequest: &brevity.PictureRequest{
				Callsign: "intruder 1 1",
				Radius:   conf.DefaultPictureRadius,
			},
			expectedOk: true,
		},
		{
			text: "anyface, intruder 1-1 picture radius 30",
			expectedRequest: &brevity.PictureRequest{
				Callsign: "intruder 1 1",
				Radius:   30 * unit.NauticalMile,
			},
			expectedOk: true,
		},
	}
	runParserTestCases(t, New(TestCallsign), testCases)
}
