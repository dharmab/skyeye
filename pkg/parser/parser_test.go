package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/stretchr/testify/require"
)

const TestCallsign = "Skyeye"

type parserTestCase struct {
	text            string
	expectedRequest any
	expectedOk      bool
}

func runParserTestCases(
	t *testing.T,
	p Parser,
	testCases []parserTestCase,
	fn func(*testing.T, parserTestCase, any),
) {
	for _, test := range testCases {
		t.Run(test.text, func(t *testing.T) {
			result, ok := p.Parse(test.text)
			require.Equal(t, test.expectedOk, ok)
			require.IsType(t, test.expectedRequest, result)
			fn(t, test, result)
		})
	}
}
func TestParserSadPaths(t *testing.T) {
	testCases := []parserTestCase{
		{
			text:            "anyface",
			expectedRequest: &brevity.UnableToUnderstandRequest{},
			expectedOk:      true,
		},
		{
			text:            "anyface radio check",
			expectedRequest: &brevity.UnableToUnderstandRequest{},
			expectedOk:      true,
		},
	}
	runParserTestCases(
		t,
		New(TestCallsign),
		testCases,
		func(*testing.T, parserTestCase, any) {},
	)
}

func TestParserAlphaCheck(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "ANYFACE, HORNET 1, CHECKING IN AS FRAGGED, REQUEST ALPHA CHECK DEPOT",
			expectedRequest: &brevity.AlphaCheckRequest{
				Callsign: "hornet 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 alpha check",
			expectedRequest: &brevity.AlphaCheckRequest{
				Callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11, checking in as fragged, request alpha check bullseye",
			expectedRequest: &brevity.AlphaCheckRequest{
				Callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
	}
	runParserTestCases(t, New(TestCallsign), testCases, func(t *testing.T, test parserTestCase, request any) {
		expected := test.expectedRequest.(*brevity.AlphaCheckRequest)
		actual := request.(*brevity.AlphaCheckRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
	})
}

func TestParserRadioCheck(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "Any Face Intruder 1-1 ready a check",
			expectedRequest: &brevity.RadioCheckRequest{
				Callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface Wildcat11 radio check out.",
			expectedRequest: &brevity.RadioCheckRequest{
				Callsign: "wildcat 1 1",
			},
			expectedOk: true,
		},
		{
			text: "Any face, Wildcat11, radio check",
			expectedRequest: &brevity.RadioCheckRequest{
				Callsign: "wildcat 1 1",
			},
			expectedOk: true,
		},
		{
			text: "Any Face In shooter 1-1 Radio Check.",
			expectedRequest: &brevity.RadioCheckRequest{
				Callsign: "inshooter 1 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 radio check",
			expectedRequest: &brevity.RadioCheckRequest{
				Callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 1-1 radio check",
			expectedRequest: &brevity.RadioCheckRequest{
				Callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder fife one radio check",
			expectedRequest: &brevity.RadioCheckRequest{
				Callsign: "intruder 5 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 request radio check",
			expectedRequest: &brevity.RadioCheckRequest{
				Callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 radio check 133 point zero",
			expectedRequest: &brevity.RadioCheckRequest{
				Callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 radio check on button five",
			expectedRequest: &brevity.RadioCheckRequest{
				Callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface work out 2 1 radio check",
			expectedRequest: &brevity.RadioCheckRequest{
				Callsign: "workout 2 1",
			},
			expectedOk: true,
		},
	}
	runParserTestCases(t, New(TestCallsign), testCases, func(t *testing.T, test parserTestCase, request any) {
		expected := test.expectedRequest.(*brevity.RadioCheckRequest)
		actual := request.(*brevity.RadioCheckRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
	})
}
