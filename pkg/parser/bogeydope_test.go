package parser

import (
	"fmt"
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
			text: TestCallsign + " serpent, 6/8, BUBBYDO.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "serpent 6 8",
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
		{
			text: TestCallsign + ", Hornet one two, Bowie dope.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "hornet 1 2",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", stubs on one, poke it up.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "stubs on 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + "'s far on 1-1. Buggydope.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "s far on 1 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + " who is saying one-on-one request, Buggydope.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "who is saying 1 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", CAT11 request \"BOGI\".",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "cat 1 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", CAT1/1 \"BOGI\"",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "cat 1 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", hurry one, two. Bogeydome.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "hurry 1 2",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + ", this is Bulldog 1-1, request by Vito.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "bulldog 1 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: fmt.Sprintf("- %s, this is TANKAN11, request boat be doped.", TestCallsign),
			expected: &brevity.BogeyDopeRequest{
				Callsign: "tankan 1 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + " and eagle 1-1, BoogieDote.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "and eagle 1 1",
				Filter:   brevity.Aircraft,
			},
		},
		{
			text: TestCallsign + "Cat11, Quest BOGUDO.",
			expected: &brevity.BogeyDopeRequest{
				Callsign: "cat 1 1",
				Filter:   brevity.Aircraft,
			},
		},
	}
	simpleCases := []string{
		"request 'BOGIDOPE",
		"request 'POGGY DOPE.",
		"POGY",
		"request POGGY DOPE",
		"request BuggyDoke.",
		"request POGIDOP.",
		"request to log it up.",
		"PUGGY DOPE.",
		"Bugga Dope.",
		"POGADO.",
		"POGY-DO.",
		"POGIDO.",
		"request BOGUETTO.",
		"request BOGUE.",
		"Bogeydove.",
		"Boogitope.",
		"BOKI NOLA.",
		"boobydope.",
		"VOGUE IT UP!",
		"poke it open.",
		"BOBBYDO.",
		"BOGUIDO, please.",
		"Boogado.",
		"VOGUY-DOPE.",
		"Bugadoop.",
		"Bogeynope.",
		"doggy dope.",
		"Povey-Dope.",
		"Boogitup.",
		"'Bogydope'",
		"BUGGADOPE",
		"BUGGET OAP.",
		"BOGILOPE",
		"bug a dope",
		"buggett ope.",
		"BOBBY DOPE",
		"Spokiedope",
		"Boogity",
		"OGIIDO.",
		"OGYDO",
		"Bokeydoke",
		"PUKIDO",
		"BOGU DOPE",
		"BUGGIT-OPE.",
		"Boguie Dope",
		"request 'Bogydope'",
		"Bugie Dope",
		"Boogito.",
		"Oki-dope.",
		"Boogitov.",
		"Bogeito.",
		"POGGY DILT",
		"request Boogit up",
		"'BOCHY' 'DOPE",
		"BOKE IT UP",
		"BOGDOPE",
		"POGGY DOG",
		"Boogie Dope",
		"Dogito",
		"BOGLOPE",
		"POGGY DOK",
		"WOGITOP",
		"BUBBY DO",
		"VOGEDOPE",
		"Puggido",
		"OGEYDOPE",
		"BOOY DOPE",
		"OGGY DOPE",
		"OG dope",
		"POGGY DOP",
		"OIDO",
		"POGEDO",
		"FOGIDO",
		"POGEDDOPE",
		"BOGIDOKE was 1-3-0.",
		"BOGIDO",
		"BoogieDope",
		"BOBBY DOOM",
		"BOGAN DOE",
		"Bo-I-Doke",
		"for your dope",
		"REQUESTBOADO",
		"REQUESTVOGED HOPE",
		"Bobi Dop",
		"OBEY-DOPE",
		"Bucky Dope",
		"BAKITO",
		"BUGADO",
		"Booby-Doo",
		"Poby-Dope.",
		"BOBBY DO",
		"POKETOP",
		"Bugie",
		"Boogalup",
		"Bovito",
		"POGGY DO",
		"VOGITO",
		"BOMBDO",
		"BOBIDOP",
		"BOGGY DOPE",
		"BOYADOP",
	}
	for _, text := range simpleCases {
		tc := parserTestCase{
			text: fmt.Sprintf("%s, eagle 1-1, %s", TestCallsign, text),
			expected: &brevity.BogeyDopeRequest{
				Callsign: "eagle 1 1",
				Filter:   brevity.Aircraft,
			},
		}
		testCases = append(testCases, tc)
	}
	runParserTestCases(t, New(TestCallsign, true), testCases, func(t *testing.T, test parserTestCase, request any) {
		t.Helper()
		expected := test.expected.(*brevity.BogeyDopeRequest)
		actual := request.(*brevity.BogeyDopeRequest)
		require.Equal(t, expected.Callsign, actual.Callsign)
		require.Equal(t, expected.Filter, actual.Filter)
	})
}
