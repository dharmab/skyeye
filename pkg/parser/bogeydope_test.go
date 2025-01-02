package parser

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/stretchr/testify/require"
)

func TestParserBogeyDope(t *testing.T) {
	t.Parallel()
	testCases := []parserTestCase{
		{
			text: "ANYFACE, EAGLE 1 BOGEY DOPE",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "eagle 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: "anyface intruder 11 bogey dope fighters",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "intruder 1 1",
				Filter:   brevity.FixedWing,
			},
		},
		{
			text: "anyface intruder 11 bogey dope just helos",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "intruder 1 1",
				Filter:   brevity.RotaryWing,
			},
		},
		{
			text: "Anyface_hogger41, boogie dope",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "hogger 4 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: "Anyface, ugly tutu, bogeydope",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "ugly 2 2",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", Gunfighter 2-1, request 'BOGIDOPE",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "gunfighter 2 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", REVA 1-3, POGGY DOPE.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "reva 1 3",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", MAKO, POGY",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "mako",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", MAKO 1-1, request POGGY DOPE",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "mako 1 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + " Viper11 BuggyDoke.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "viper 1 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", HORNET, 1, 2, BOGGID, 2.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "hornet 1 2",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + " ugly one-one, POKIDO.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "ugly 1 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + "ugly one-one, buggy two.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "ugly 1 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", Viking31, request POGIDOP.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "viking 3 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", Waking Free 1, request to log it up.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "waking 3 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + " HUD 13, PUGGY DOPE.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "hud 1 3",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + " Mage 1-2, Bugga Dope.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "mage 1 2",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + " serpent, 6/8, BUBBYDO.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "serpent 6 8",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + " REBEL 1-1, POGADO.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "rebel 1 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + " REBEL 1-1, POGY-DO.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "rebel 1 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + " REBEL 1-1, POGIDO.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "rebel 1 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", Glimmer, Buggetto.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "glimmer",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", Hood, 1-3, BOWIDO.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "hood 1 3",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: "\"" + TestCallsign + " \"HOOD 1-3 BOBBY DOKE\"",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "hood 1 3",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", serpent, 6/8, BOBY DOPE.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "serpent 6 8",
				Filter:   brevity.Aircraft,
			},
		},
		{
			// Yes, this is a real transcription from the wild!
			// Whisper was trained on YouTube videos and it seems to have
			// picked up this rapper's name...
			text: TestCallsign + ", serpent, 6ix9ine, Bogeydough.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "serpent 6 9",
				Filter:   brevity.Aircraft,
			},
		},
	}
	runParserTestCases(t, New(TestCallsign, true), testCases, func(t *testing.T, test parserTestCase, request any) {
		t.Helper()
		expected := test.expected.(*brevity.BogeyDopeRequest)
		actual := request.(*brevity.BogeyDopeRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
		require.Equal(t, expected.Filter, actual.Filter)
	})
}
