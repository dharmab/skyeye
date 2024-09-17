package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/stretchr/testify/assert"
)

func TestParserPicture(t *testing.T) {
	t.Parallel()
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
		{
			text: "anyface, picture",
			expected: &brevity.PictureRequest{
				Callsign: "NULL",
			},
		},
	}
	runParserTestCases(t, New(TestCallsign, true), testCases, func(t *testing.T, test parserTestCase, request any) {
		t.Helper()
		expected := test.expected.(*brevity.PictureRequest)
		actual := request.(*brevity.PictureRequest)
		assert.Equal(t, expected.Callsign, actual.Callsign)
	})
}
