package parser

import (
	"testing"

	"github.com/dharmab/skyeye/internal/conf"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/require"
)

type parserTestCase struct {
	text            string
	expectedRequest any
	expectedOk      bool
}

func runParserTestCases(t *testing.T, p Parser, testCases []parserTestCase) {
	for _, test := range testCases {
		t.Run(test.text, func(t *testing.T) {
			request, ok := p.Parse(test.text)
			require.EqualValuesf(t, test.expectedRequest, request, "parser.Parse() request: expected = %v, actual %v", test.expectedRequest, request)
			require.Equal(t, test.expectedOk, ok, "parser.Parse() ok: expected = %v, actual %v", test.expectedOk, ok)
		})
	}
}
func TestParserSadPaths(t *testing.T) {
	testCases := []parserTestCase{
		{
			text:            "anyface",
			expectedRequest: nil,
			expectedOk:      false,
		},
		{
			text:            "anyface radio check",
			expectedRequest: nil,
			expectedOk:      false,
		},
	}
	runParserTestCases(t, New(), testCases)
}

func TestParserAlphaCheck(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "ANYFACE, HORNET 1, MISSION NUMBER 5-1-1-1, CHECKING IN AS FRAGGED, REQUEST ALPHA CHECK DEPOT",
			expectedRequest: &alphaCheckRequest{
				callsign: "hornet 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 alpha check",
			expectedRequest: &alphaCheckRequest{
				callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 mission number 5111, checking in as fragged, request alpha check bullseye",
			expectedRequest: &alphaCheckRequest{
				callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
	}
	runParserTestCases(t, New(), testCases)
}

func TestParserBogeyDope(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "ANYFACE, EAGLE 1 BOGEY DOPE",
			expectedRequest: &bogeyDopeRequest{
				callsign: "eagle 1",
				filter:   brevity.Everything,
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 bogey dope fighters",
			expectedRequest: &bogeyDopeRequest{
				callsign: "intruder 1 1",
				filter:   brevity.Airplanes,
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 bogey dope just helos",
			expectedRequest: &bogeyDopeRequest{
				callsign: "intruder 1 1",
				filter:   brevity.Helicopters,
			},
			expectedOk: true,
		},
	}
	runParserTestCases(t, New(), testCases)
}

func TestParserDeclare(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "ANYFACE, EAGLE 1 DECLARE BULLSEYE 230/12, TWELVE THOUSAND",
			expectedRequest: &declareRequest{
				callsign: "eagle 1",
				bullseye: brevity.NewBullseye(
					230*unit.Degree,
					12*unit.NauticalMile,
				),
				altitude: 12000 * unit.Foot,
				track:    brevity.UnknownDirection,
			},
			expectedOk: true,
		},
	}
	runParserTestCases(t, New(), testCases)
}

func TestParserPicture(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "anyface, intruder 1-1 request picture",
			expectedRequest: &pictureRequest{
				callsign: "intruder 1 1",
				radius:   conf.DefaultPictureRadius,
			},
			expectedOk: true,
		},
		{
			text: "anyface, intruder 1-1 picture radius 30",
			expectedRequest: &pictureRequest{
				callsign: "intruder 1 1",
				radius:   30 * unit.NauticalMile,
			},
			expectedOk: true,
		},
	}
	runParserTestCases(t, New(), testCases)
}

func TestParserRadioCheck(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "anyface intruder 11 radio check",
			expectedRequest: &radioCheckRequest{
				callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 1-1 radio check",
			expectedRequest: &radioCheckRequest{
				callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder fife one radio check",
			expectedRequest: &radioCheckRequest{
				callsign: "intruder 5 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 request radio check",
			expectedRequest: &radioCheckRequest{
				callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 radio check 133 point zero",
			expectedRequest: &radioCheckRequest{
				callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 radio check on button five",
			expectedRequest: &radioCheckRequest{
				callsign: "intruder 1 1",
			},
			expectedOk: true,
		},
	}
	runParserTestCases(t, New(), testCases)
}

func TestParserSpiked(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "ANYFACE, EAGLE 1 SPIKED 2-7-0",
			expectedRequest: &spikedRequest{
				callsign:       "eagle 1",
				bearingDegrees: 270,
			},
			expectedOk: true,
		},
	}
	runParserTestCases(t, New(), testCases)

}

func TestParserSnaplock(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "ANYFACE, FREEDOM 31, SNAPLOCK 125/10, EIGHT THOUSAND",
			expectedRequest: &snaplockRequest{
				callsign: "freedom 3 1",
				bra: brevity.NewBRA(
					125*unit.Degree,
					10*unit.NauticalMile,
					8000*unit.Foot,
				),
			},
			expectedOk: true,
		},
	}
	runParserTestCases(t, New(), testCases)
}
