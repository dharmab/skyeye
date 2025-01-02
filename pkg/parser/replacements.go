package parser

import (
	"slices"
)

// replacementLUT is a map of alternate forms of request words. These are used
// to provide aliases for certain commands and to deal with quality issues in
// speech-to-text. Do not use this variable directly, instead use replacements
// so that more specific forms are matched before more general forms.
var replacementLUT = map[string]string{
	"alphacheck":         alphaCheck,
	"alphachek":          alphaCheck,
	"alphajack":          alphaCheck,
	"arfachek":           alphaCheck,
	"arfatcheck":         alphaCheck,
	"bobbiedope":         bogeyDope,
	"bobby doke":         bogeyDope,
	"bobbydo":            bogeyDope,
	"bobido":             bogeyDope,
	"boby dope":          bogeyDope,
	"bobydo":             bogeyDope,
	"bog it":             bogeyDope,
	"bogado":             bogeyDope,
	"bogeido":            bogeyDope,
	"bogeied":            bogeyDope,
	"bogey-doke":         bogeyDope,
	"bogeydoke":          bogeyDope,
	"bogeydome":          bogeyDope,
	"bogeydope":          bogeyDope,
	"bogeydote":          bogeyDope,
	"bogeydough":         bogeyDope,
	"bogeydove":          bogeyDope,
	"bogeyedope":         bogeyDope,
	"boggeto":            bogeyDope,
	"boggid":             bogeyDope,
	"boggido":            bogeyDope,
	"boggy":              bogeyDope,
	"bogi":               bogeyDope,
	"bogido":             bogeyDope,
	"bogidope":           bogeyDope,
	"bogie":              bogeyDope,
	"bogit":              bogeyDope,
	"bogota":             bogeyDope,
	"bogue":              bogeyDope,
	"bogueed":            bogeyDope,
	"bogueto":            bogeyDope,
	"boguetto":           bogeyDope,
	"boguido":            bogeyDope,
	"bogy":               bogeyDope,
	"boido":              bogeyDope,
	"boji":               bogeyDope,
	"bokeh":              bogeyDope,
	"bokeido":            bogeyDope,
	"bokey":              bogeyDope,
	"bokeydope":          bogeyDope,
	"boki":               bogeyDope,
	"booby dop":          bogeyDope,
	"boobydope":          bogeyDope,
	"boog it":            bogeyDope,
	"boogie":             bogeyDope,
	"boogiedope":         bogeyDope,
	"boogitope":          bogeyDope,
	"book it out":        bogeyDope,
	"book it up":         bogeyDope,
	"bovido":             bogeyDope,
	"bowido":             bogeyDope,
	"bowie dope":         bogeyDope,
	"boy dope":           bogeyDope,
	"bubby dope":         bogeyDope,
	"bubbydo":            bogeyDope,
	"bucket up":          bogeyDope,
	"bug it out":         bogeyDope,
	"bug it":             bogeyDope,
	"bugadobe":           bogeyDope,
	"bugadope":           bogeyDope,
	"bugga":              bogeyDope,
	"bugged up":          bogeyDope,
	"buggetoo":           bogeyDope,
	"buggetto":           bogeyDope,
	"buggettope":         bogeyDope,
	"buggidop":           bogeyDope,
	"buggie":             bogeyDope,
	"buggy":              bogeyDope,
	"buggydoke":          bogeyDope,
	"buggydope":          bogeyDope,
	"bugito":             bogeyDope,
	"bulgie":             bogeyDope,
	"checking in":        checkIn,
	"com check":          radioCheck,
	"comcheck":           radioCheck,
	"comes check":        radioCheck,
	"comm":               radioCheck,
	"comms":              radioCheck,
	"commscheck":         radioCheck,
	"commshack":          radioCheck,
	"comp check":         radioCheck,
	"comps check":        radioCheck,
	"coms":               radioCheck,
	"comsjack":           radioCheck,
	"declared":           declare,
	"fog it up":          bogeyDope,
	"fogey":              bogeyDope,
	"fogeyed":            bogeyDope,
	"foggy":              bogeyDope,
	"foggydope":          bogeyDope,
	"fogy dope":          bogeyDope,
	"fogy":               bogeyDope,
	"go be dope":         bogeyDope,
	"go geeto":           bogeyDope,
	"log it up":          bogeyDope,
	"lucky dope":         bogeyDope,
	"mic check":          radioCheck,
	"mike check":         radioCheck,
	"mogito":             bogeyDope,
	"odi":                bogeyDope,
	"ogi doke":           bogeyDope,
	"ogi dop":            bogeyDope,
	"ogi dope":           bogeyDope,
	"ogidope":            bogeyDope,
	"okey":               bogeyDope,
	"oogie":              bogeyDope,
	"ovido":              bogeyDope,
	"perimeter":          tripwire,
	"pogado":             bogeyDope,
	"pogadope":           bogeyDope,
	"pogdedo":            bogeyDope,
	"pogeto":             bogeyDope,
	"poggy dope":         bogeyDope,
	"pogido":             bogeyDope,
	"pogidop":            bogeyDope,
	"pogito":             bogeyDope,
	"pogy":               bogeyDope,
	"poke it open":       bogeyDope,
	"poke it up":         bogeyDope,
	"pokedo":             bogeyDope,
	"pokedome":           bogeyDope,
	"pokido":             bogeyDope,
	"puggy dope":         bogeyDope,
	"radiocheck":         radioCheck,
	"radiochick":         radioCheck,
	"read a check":       radioCheck,
	"read it check":      radioCheck,
	"set threat radius":  tripwire,
	"set threat range":   tripwire,
	"set warning radius": tripwire,
	"set warning range":  tripwire,
	"set warning":        tripwire,
	"snap lock":          snaplock,
	"trip wire":          tripwire,
	"vogadope":           bogeyDope,
	"vogidobe":           bogeyDope,
	"vogue":              bogeyDope,
	"voki":               bogeyDope,
	"warn me":            tripwire,
}

type replacement struct {
	Original string
	Normal   string
}

// replacements is a slice of replacment forms of request words. replacements
// is ordered so that more specific forms are matched before more general
// forms.
var replacements = []replacement{}

func init() {
	for k, v := range replacementLUT {
		replacements = append(replacements, replacement{k, v})
	}
	slices.SortFunc(replacements, func(a, b replacement) int {
		// Longer original strings should be matched first.
		if len(a.Original) > len(b.Original) {
			return -1
		}
		if len(a.Original) < len(b.Original) {
			return 1
		}
		return 0
	})
}
