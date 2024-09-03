package composer

import (
	"fmt"
	"math/rand/v2"

	"github.com/dharmab/skyeye/pkg/brevity"
)

func (c *composer) ComposeTripwireResponse(response brevity.TripwireResponse) NaturalLanguageResponse {
	replies1 := []string{
		"%s, I've got my copy of MULTI-SERVICE TACTICS TECHNIQUES AND PROCEDURES for Air Control Communication right here, and I don't see anything in here about a so-called TRIPWIRE.",
		"%s, I'm not sure what you mean by TRIPWIRE. I don't see that term in MULTI-SERVICE TACTICS TECHNIQUES AND PROCEDURES for Air Control Communication.",
		"%s, TRIPWIRE is not a term we use in Air Battle Management.",
		"%s, I'm not sure what you mean by TRIPWIRE.",
		"%s, I don't see anything about a TRIPWIRE in MULTI-SERVICE TACTICS TECHNIQUES AND PROCEDURES for Air Control Communication.",
		"%s, I'm not sure what you mean by TRIPWIRE. I don't see that term in MULTI-SERVICE TACTICS TECHNIQUES AND PROCEDURES for Air Control Communication.",
		"%s, give me a second, I'm just searching my copy of MULTI-SERVICE TACTICS TECHNIQUES AND PROCEDURES for Air Control Communication for what a TRIPWIRE is. Nope, I couldn't find it in there.",
		"%s, I have no idea what a TRIPWIRE is. Frankly, I don't want to know.",
		"%s, I think you have me confused with someone else.",
		"%s, did you know how many times the word TRIPWIRE appears in MULTI-SERVICE TACTICS TECHNIQUES AND PROCEDURES for Air Control Communication? I'll give you a hint: it's less than once. ",
		"%s, TRIPWIRE ain't no brevity I ever heard of!",
		"%s, please refer to MULTI-SERVICE TACTICS TECHNIQUES AND PROCEDURES for Air Control Communication. You will find that it does not contain any so-called TRIPWIRE.",
	}
	variation1 := replies1[rand.IntN(len(replies1))]

	replies2 := []string{
		"Look, I'm watching you on the radar, and I'll let you know if I see any threats, okay?",
		"Look, I'm watching you on the radar, and I'll let you know if I see any threats.",
		"I'll let you know with a THREAT call if I see anything that could be a danger to you, and you can ask me for an updated PICTURE at any time.",
		"I'm watching the radar for threats, and I'll let you know if I see anything that could be a danger to you.",
		"I'll keep watching you on the radar and let you know if I see anything that could be a threat.",
		"I'm monitoring you on my radar scope, and will let you know about any threats.",
		"I am monitoring you on the radar and will automatically inform you about any threats.",
		"Why don't you just focus on flying, and I'll focus on watching the radar for threats?",
		"Let's keep it simple: you fly the plane, I watch the radar for threats, and I'll let you know if I see anything.",
		"I'm watching for threats on the radar, and I'll inform you if anything needs your attention.",
		"I am following you on the radar and will tell you about any threats.",
		"I'm monitoring you on the radar. I will inform you if I see any threats, and you can ask me for an updated PICTURE at any time.",
	}
	variation2 := replies2[rand.IntN(len(replies2))]

	reply := fmt.Sprintf(
		fmt.Sprintf("%s %s", variation1, variation2),
		response.Callsign,
	)
	return NaturalLanguageResponse{
		Subtitle: reply,
		Speech:   reply,
	}
}
