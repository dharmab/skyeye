package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/stretchr/testify/assert"
)

func TestParserPicture(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "anyface, intruder 1-1 request picture",
			expected: &brevity.PictureRequest{
				Callsign: "intruder 1 1",
			},
		},
		{
			text: "anyface, intruder 1-1 picture 30",
			expected: &brevity.PictureRequest{
				Callsign: "intruder 1 1",
			},
		},
	}
	runParserTestCases(t, New(TestCallsign), testCases, func(t *testing.T, test parserTestCase, request any) {
		expected := test.expected.(*brevity.PictureRequest)
		actual := request.(*brevity.PictureRequest)
		assert.Equal(t, expected.Callsign, actual.Callsign)
	})
}
