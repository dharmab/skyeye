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
	runParserTestCases(t, New(), testCases)
}

func TestParserBogeyDope(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "ANYFACE, EAGLE 1 BOGEY DOPE",
			expectedRequest: &brevity.BogeyDopeRequest{
				Callsign: "eagle 1",
				Filter:   brevity.Everything,
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 bogey dope fighters",
			expectedRequest: &brevity.BogeyDopeRequest{
				Callsign: "intruder 1 1",
				Filter:   brevity.Airplanes,
			},
			expectedOk: true,
		},
		{
			text: "anyface intruder 11 bogey dope just helos",
			expectedRequest: &brevity.BogeyDopeRequest{
				Callsign: "intruder 1 1",
				Filter:   brevity.Helicopters,
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
			expectedRequest: &brevity.DeclareRequest{
				Callsign: "eagle 1",
				Location: *brevity.NewBullseye(
					unit.Angle(230)*unit.Degree,
					unit.Length(12)*unit.NauticalMile,
				),
				Altitude: 12000 * unit.Foot,
				Track:    brevity.UnknownDirection,
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
	runParserTestCases(t, New(), testCases)
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
	runParserTestCases(t, New(), testCases)
}

func TestParserSpiked(t *testing.T) {
	testCases := []parserTestCase{
		{
			text: "ANYFACE, EAGLE 1 SPIKED 2-7-0",
			expectedRequest: &brevity.SpikedRequest{
				Callsign: "eagle 1",
				Bearing:  unit.Angle(270) * unit.Degree,
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
			expectedRequest: &brevity.SnaplockRequest{
				Callsign: "freedom 3 1",
				BRA: brevity.NewBRA(
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
