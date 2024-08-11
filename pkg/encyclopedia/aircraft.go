// package encyclopedia is a database of aircraft data
package encyclopedia

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
)

// Data sources:
// https://github.com/Quaggles/dcs-lua-datamine/tree/master/_G/db/Units/Planes/Plane
// https://github.com/Quaggles/dcs-lua-datamine/tree/master/_G/db/Units/Helicopters/Helicopter

type AircraftTag int

const (
	FixedWing AircraftTag = iota
	RotaryWing
	Unarmed
	Fighter
	Attack
)

type Aircraft struct {
	// ACMIShortName is the Name proeprty used in ACMI telemetry.
	ACMIShortName string
	// tags categorize the aircraft.
	tags map[AircraftTag]bool
	// e.g. F-15, Su-27
	PlatformDesignation string
	// TypeDesignation is the official type designation of the aircraft.
	// e.g. F-15C, F-15E, Su-27
	TypeDesignation string
	// NATOReportingName is the NATO reporting name of the aircraft. Not all aircraft have a NATO reporting name.
	// e.g. Flanker, Bear
	NATOReportingName string
	// OfficialName is the official name of the aircraft. Not all aircraft have an official name.
	// e.g. Thunderbolt, Falcon, Eagle
	OfficialName string
	// Nickname is a common nickname for the aircraft. Not all aircraft have a nickname.
	// e.g. Warthog, Viper, Mudhen
	Nickname     string
	threatRadius unit.Length
}

var DefaultThreatRadius = 35 * unit.NauticalMile

func (a Aircraft) Category() brevity.ContactCategory {
	if _, ok := a.tags[FixedWing]; ok {
		return brevity.FixedWing
	} else if _, ok := a.tags[RotaryWing]; ok {
		return brevity.RotaryWing
	}
	return brevity.Aircraft
}

func (a Aircraft) Tags() []AircraftTag {
	tags := []AircraftTag{}
	for t := range a.tags {
		tags = append(tags, t)
	}
	return tags
}

func (a Aircraft) HasTag(tag AircraftTag) bool {
	_, ok := a.tags[tag]
	return ok
}

func (a Aircraft) HasAnyTag(tags ...AircraftTag) bool {
	for _, tag := range tags {
		if a.HasTag(tag) {
			return true
		}
	}
	return false
}

func (a Aircraft) ThreatRadius() unit.Length {
	if a.threatRadius != 0 || a.HasTag(Unarmed) {
		return a.threatRadius
	}
	if a.HasTag(RotaryWing) {
		return 3 * unit.NauticalMile
	}
	if a.HasTag(Attack) {
		return 20 * unit.NauticalMile
	}
	if a.HasTag(Fighter) {
		return 50 * unit.NauticalMile
	}
	return 35 * unit.NauticalMile
}

func variants(data Aircraft, naming map[string]string) []Aircraft {
	variants := []Aircraft{}
	for nameSuffix, designationSuffx := range naming {
		aircraft := Aircraft{
			tags:                data.tags,
			PlatformDesignation: data.PlatformDesignation,
			TypeDesignation:     data.TypeDesignation + designationSuffx,
			NATOReportingName:   data.NATOReportingName,
			OfficialName:        data.OfficialName,
			Nickname:            data.Nickname,
		}
		if data.ACMIShortName != "" {
			aircraft.ACMIShortName = data.ACMIShortName + nameSuffix
		} else {
			aircraft.ACMIShortName = data.PlatformDesignation + nameSuffix
		}
		variants = append(variants, aircraft)
	}
	return variants
}

var a10Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Attack:    true,
	},
	PlatformDesignation: "A-10",
	OfficialName:        "Thunderbolt",
	Nickname:            "Warthog",
}

func a10Variants() []Aircraft {
	return variants(
		a10Data,
		map[string]string{
			"A":   "A",
			"C":   "C",
			"C_2": "C",
		},
	)
}

var ah64Data = Aircraft{
	tags: map[AircraftTag]bool{
		RotaryWing: true,
		Attack:     true,
	},
	PlatformDesignation: "AH-64",
	OfficialName:        "Apache",
}

func ah64Variants() []Aircraft {
	return variants(
		ah64Data,
		map[string]string{
			"A":        "A",
			"D":        "D",
			"D_BLK_II": "D",
		},
	)
}

var c101Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "C-101",
	OfficialName:        "Aviojet",
}

func c101Variants() []Aircraft {
	return variants(
		c101Data,
		map[string]string{
			"CC": "CC",
			"EB": "EB",
		},
	)
}

var f86Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "F-86",
	OfficialName:        "Sabre",
	threatRadius:        15 * unit.NauticalMile,
}

func f86Variants() []Aircraft {
	return variants(
		f86Data,
		map[string]string{
			"F":    "F",
			"F FC": "F",
		},
	)
}

var f4Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "F-4",
	OfficialName:        "Phantom",
}

func f4Variants() []Aircraft {
	return variants(
		f4Data,
		map[string]string{
			"E-45MC": "E",
			"E":      "E",
		},
	)
}

var f5Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "F-5",
	OfficialName:        "Tiger",
}

func f5Variants() []Aircraft {
	return variants(
		f5Data,
		map[string]string{
			"E":      "E",
			"E-3":    "E",
			"E-3 FC": "E",
		},
	)
}

var f14Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "F-14",
	OfficialName:        "Tomcat",
	threatRadius:        55 * unit.NauticalMile,
}

func f14Variants() []Aircraft {
	return variants(
		f14Data,
		map[string]string{
			"A-135-GR": "A",
			"A":        "A",
			"B":        "B",
		},
	)
}

var f15Data = Aircraft{
	PlatformDesignation: "F-15",
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	// Use "Eagle" for Strike Eagle because radar cannot distinguish between the two
	OfficialName: "Eagle",
	threatRadius: 65 * unit.NauticalMile,
}

func f15Variants() []Aircraft {
	return variants(
		f15Data,
		map[string]string{
			"C":   "C",
			"ESE": "E",
		},
	)
}

var f16Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "F-16",
	OfficialName:        "Falcon",
	Nickname:            "Viper",
}

func f16Variants() []Aircraft {
	return variants(
		f16Data,
		map[string]string{
			"A":       "A",
			"A MLU":   "A",
			"C_50":    "C",
			"C bl.50": "C",
			"C bl.52": "C",
		},
	)
}

var fa18Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "FA-18",
	OfficialName:        "Hornet",
}

func fa18Variants() []Aircraft {
	return variants(
		fa18Data,
		map[string]string{
			"A":        "A",
			"C":        "C",
			"C_hornet": "C",
		},
	)
}

var ka50Data = Aircraft{
	tags: map[AircraftTag]bool{
		RotaryWing: true,
		Attack:     true,
	},
	PlatformDesignation: "Ka-50",
	OfficialName:        "Black Shark",
	NATOReportingName:   "Hokum",
}

func ka50Variants() []Aircraft {
	return variants(
		ka50Data,
		map[string]string{
			"":   "",
			"_3": "",
		},
	)
}

var mi24Data = Aircraft{
	tags: map[AircraftTag]bool{
		RotaryWing: true,
		Attack:     true,
		Unarmed:    true,
	},
	PlatformDesignation: "Mi-24",
	NATOReportingName:   "Hind",
}

func mi24Variants() []Aircraft {
	return variants(
		mi24Data,
		map[string]string{
			"P": "P",
			"V": "V",
		},
	)
}

var mirageF1Data = Aircraft{
	ACMIShortName: "Mirage-F1",
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "Mirage F1",
	OfficialName:        "Mirage F1",
}

func mirageF1Variants() []Aircraft {
	return variants(
		mirageF1Data,
		map[string]string{
			"B":     "B",
			"BD":    "BD",
			"BE":    "BE",
			"BQ":    "BQ",
			"C-200": "C-200",
			"C":     "C",
			"CE":    "CE",
			"CG":    "CG",
			"CH":    "CH",
			"CJ":    "CJ",
			"CK":    "CK",
			"CR":    "CR",
			"CT":    "CT",
			"CZ":    "CZ",
			"DDA":   "DDA",
			"ED":    "ED",
			"EDA":   "EDA",
			"EE":    "EE",
			"EH":    "EH",
			"EQ":    "EQ",
			"JA":    "JA",
			"M-CE":  "M-CE",
			"M-EE":  "M-EE",
		},
	)
}

var oh58Data = Aircraft{
	tags: map[AircraftTag]bool{
		RotaryWing: true,
		Unarmed:    true,
	},
	PlatformDesignation: "OH-58",
	OfficialName:        "Kiowa",
}

var sa342Data = Aircraft{
	tags: map[AircraftTag]bool{
		RotaryWing: true,
		Unarmed:    true,
	},
	PlatformDesignation: "SA 342",
	OfficialName:        "Gazelle",
}

func sa342Variants() []Aircraft {
	vars := []Aircraft{}
	for _, variant := range []string{"L", "M", "Minigun", "Mistral"} {
		vars = append(vars, Aircraft{
			ACMIShortName:       "SA342" + variant,
			tags:                sa342Data.tags,
			PlatformDesignation: sa342Data.PlatformDesignation,
			OfficialName:        sa342Data.OfficialName,
		})
	}
	return vars
}

var ftData = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
	},
	PlatformDesignation: "MiG-15",
	NATOReportingName:   mig15NATOReportingName,
}

func ftVariants() []Aircraft {
	return variants(
		ftData,
		map[string]string{
			"bis":    "bis",
			"bis FC": "bis",
		},
	)
}

var fencerData = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "Su-24",
	NATOReportingName:   "Fencer",
}

func fencerVariants() []Aircraft {
	return variants(
		fencerData,
		map[string]string{
			"M":  "M",
			"MR": "MR",
		},
	)
}

var foxbatData = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "MiG-25",
	NATOReportingName:   "Foxbat",
	threatRadius:        80 * unit.NauticalMile,
}

func foxbatVariants() []Aircraft {
	return variants(
		foxbatData,
		map[string]string{
			"PD":  "PD",
			"RBT": "RBT",
		},
	)
}

var fulcrumData = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "MiG-29",
	NATOReportingName:   "Fulcrum",
}

func fulcrumVariants() []Aircraft {
	return variants(
		fulcrumData,
		map[string]string{
			"A": "A",
			"G": "A",
			"S": "S",
		},
	)
}

var frogfootData = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Attack:    true,
	},
	PlatformDesignation: "Su-25",
	NATOReportingName:   "Frogfoot",
}

func frogfootVariants() []Aircraft {
	return variants(
		frogfootData,
		map[string]string{
			"":   "A",
			"T":  "T",
			"TM": "TM",
		},
	)
}

var flankerData = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "Su-27",
	NATOReportingName:   "Flanker",
	threatRadius:        80 * unit.NauticalMile,
}

var kc135Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Unarmed:   true,
	},
	PlatformDesignation: "KC-135",
	OfficialName:        "Stratotanker",
}

func kc135Variants() []Aircraft {
	return []Aircraft{
		{
			ACMIShortName:       "KC-135",
			tags:                kc135Data.tags,
			PlatformDesignation: kc135Data.PlatformDesignation,
			OfficialName:        kc135Data.OfficialName,
		},
		{
			ACMIShortName:       "KC135MPRS",
			tags:                kc135Data.tags,
			PlatformDesignation: kc135Data.PlatformDesignation,
			OfficialName:        kc135Data.OfficialName,
		},
	}
}

var l39Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "L-39",
	OfficialName:        "Albatros",
}

func l39Variants() []Aircraft {
	return variants(
		l39Data,
		map[string]string{
			"C":  "C",
			"ZA": "ZA",
		},
	)
}

var mb339Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "MB-339",
	threatRadius:        15 * unit.NauticalMile,
}

func mb339Variants() []Aircraft {
	return variants(
		mb339Data,
		map[string]string{
			"A":     "A",
			"A/PAN": "A",
		},
	)
}

var s3Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Unarmed:   true,
	},
	PlatformDesignation: "S-3",
	OfficialName:        "Viking",
}

func s3Variants() []Aircraft {
	return variants(
		s3Data,
		map[string]string{
			"B":        "B",
			"B Tanker": "B",
		},
	)
}

var tornadoData = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		Fighter:   true,
	},
	PlatformDesignation: "Tornado",
	OfficialName:        "Tornado",
}

func tornadoVariants() []Aircraft {
	return variants(
		tornadoData,
		map[string]string{
			" IDS": "IDS",
			" GR4": "GR4",
		},
	)
}

var aircraftData = []Aircraft{
	{
		ACMIShortName: "A-4E-C",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Attack:    true,
		},
		PlatformDesignation: "A-4",
		TypeDesignation:     "A-4E",
		OfficialName:        "Skyhawk",
		Nickname:            "Scooter",
	},
	{
		ACMIShortName: "A-20G",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "A-20",
		TypeDesignation:     "A-20G",
		OfficialName:        "Havoc",
	},
	{
		ACMIShortName: "A-50",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "A-50",
		TypeDesignation:     "A-50",
		NATOReportingName:   "Mainstay",
	},
	{
		ACMIShortName: "AH-1W",
		tags: map[AircraftTag]bool{
			RotaryWing: true,
			Attack:     true,
		},
		PlatformDesignation: "AH-1",
		TypeDesignation:     "AH-1W",
		OfficialName:        "SuperCobra",
	},
	{
		ACMIShortName: "AJS37",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Attack:    true,
		},
		PlatformDesignation: "AJS37",
		OfficialName:        "Viggen",
	},
	{
		ACMIShortName: "AV8BNA",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Attack:    true,
		},
		PlatformDesignation: "AV-8",
		TypeDesignation:     "AV-8B",
		OfficialName:        "Harrier",
	},
	{
		ACMIShortName: "An-26B",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "An-26",
		TypeDesignation:     "An-26B",
		NATOReportingName:   "Curl",
	},
	{
		ACMIShortName: "An-30M",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "An-30",
		TypeDesignation:     "An-30M",
		NATOReportingName:   "Clank",
	},
	{
		ACMIShortName: "B-17G",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "B-17",
		TypeDesignation:     "B-17G",
		OfficialName:        "Flying Fortress",
	},
	{
		ACMIShortName: "B-52H",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "B-52",
		TypeDesignation:     "B-52H",
		OfficialName:        "Stratofortress",
		Nickname:            "Buff",
	},
	{
		ACMIShortName: "B-1B",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "B-1",
		TypeDesignation:     "B-1B",
		OfficialName:        "Lancer",
		Nickname:            "Bone",
	},
	{
		ACMIShortName: "C-17A",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "C-17",
		TypeDesignation:     "C-17A",
		OfficialName:        "Globemaster",
	},
	{
		ACMIShortName: "C-47",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "C-47",
		OfficialName:        "Skytrain",
	},
	{
		ACMIShortName: "C-130",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "C-130",
		TypeDesignation:     "C-130",
		OfficialName:        "Hercules",
		Nickname:            "Herc",
	},
	{
		ACMIShortName: "CH-47D",
		tags: map[AircraftTag]bool{
			RotaryWing: true,
			Unarmed:    true,
		},
		PlatformDesignation: "CH-47",
		TypeDesignation:     "CH-47D",
		OfficialName:        "Chinook",
	},
	{
		ACMIShortName: "CH-53E",
		tags: map[AircraftTag]bool{
			RotaryWing: true,
			Unarmed:    true,
		},
		PlatformDesignation: "CH-53",
		TypeDesignation:     "CH-53E",
		OfficialName:        "Super Stallion",
	},
	{
		ACMIShortName: "E-2C",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "E-2",
		TypeDesignation:     "E-2C",
		OfficialName:        "Hawkeye",
	},
	{
		ACMIShortName: "E-3A",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "E-3",
		TypeDesignation:     "E-3A",
		OfficialName:        "Sentry",
	},
	{
		ACMIShortName: "F-117A",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "F-117",
		TypeDesignation:     "F-117A",
		OfficialName:        "Nighthawk",
		Nickname:            "Goblin",
	},
	{
		ACMIShortName: "H-6J",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "Tu-16",
		TypeDesignation:     "H-6J",
		NATOReportingName:   "Badger",
	},
	{
		ACMIShortName: "IL-76MD",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "Il-76",
		TypeDesignation:     "Il-76MD",
		NATOReportingName:   "Candid",
	},
	{
		ACMIShortName: "IL-78M",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "Il-78",
		TypeDesignation:     "Il-78M",
		NATOReportingName:   "Midas",
	},
	{
		ACMIShortName:       "J-11A",
		tags:                flankerData.tags,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "J-11A",
		NATOReportingName:   flankerData.NATOReportingName,
		threatRadius:        flankerData.threatRadius,
	},
	{
		ACMIShortName: "JF-17",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Fighter:   true,
		},
		PlatformDesignation: "JF-17",
		TypeDesignation:     "JF-17",
		OfficialName:        "Thunder",
	},
	{
		ACMIShortName: "KA-27",
		tags: map[AircraftTag]bool{
			RotaryWing: true,
			Unarmed:    true,
		},
		PlatformDesignation: "Ka-27",
		TypeDesignation:     "Ka-27",
		NATOReportingName:   "Helix",
	},
	{
		ACMIShortName: "KC130",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "KC-130",
		TypeDesignation:     "KC-130",
		OfficialName:        "Hercules",
		Nickname:            "Herc",
	},
	{
		ACMIShortName: "KJ-2000",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "KJ-2000",
		TypeDesignation:     "KJ-2000",
		OfficialName:        "Mainring",
	},
	{
		ACMIShortName: "M-2000C",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Fighter:   true,
		},
		PlatformDesignation: "Mirage 2000",
		TypeDesignation:     "Mirage 2000C",
		OfficialName:        "Mirage 2000",
	},
	{
		ACMIShortName: "Mi-8MT",
		tags: map[AircraftTag]bool{
			RotaryWing: true,
			Unarmed:    true,
		},
		PlatformDesignation: "Mi-8",
		TypeDesignation:     "Mi-8MT",
		NATOReportingName:   "Hip",
	},
	{
		ACMIShortName: "Mi-26",
		tags: map[AircraftTag]bool{
			RotaryWing: true,
			Unarmed:    true,
		},
		PlatformDesignation: "Mi-26",
		TypeDesignation:     "Mi-26",
		NATOReportingName:   "Halo",
	},
	{
		ACMIShortName: "Mi-28N",
		tags: map[AircraftTag]bool{
			RotaryWing: true,
			Attack:     true,
		},
		PlatformDesignation: "Mi-28",
		TypeDesignation:     "Mi-28N",
		OfficialName:        "Havoc",
	},
	{
		ACMIShortName: "MiG-19P",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Fighter:   true,
		},
		PlatformDesignation: "MiG-19",
		TypeDesignation:     "MiG-19P",
		NATOReportingName:   "Farmer",
		threatRadius:        20 * unit.NauticalMile,
	},
	{
		ACMIShortName: "MiG-21Bis",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Fighter:   true,
		},
		PlatformDesignation: "MiG-21",
		TypeDesignation:     "MiG-21bis",
		NATOReportingName:   "Fishbed",
	},
	{
		ACMIShortName: "MiG-23MLD",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Fighter:   true,
		},
		PlatformDesignation: "MiG-23",
		TypeDesignation:     "MiG-23MLD",
		NATOReportingName:   "Flogger",
	},
	{
		ACMIShortName: "MiG-27K",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Fighter:   true,
		},
		PlatformDesignation: "MiG-27",
		TypeDesignation:     "MiG-27K",
		NATOReportingName:   "Flogger",
	},
	{
		ACMIShortName: "MiG-31",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Fighter:   true,
		},
		PlatformDesignation: "MiG-31",
		TypeDesignation:     "MiG-31",
		NATOReportingName:   "Foxhound",
		threatRadius:        80 * unit.NauticalMile,
	},
	{
		ACMIShortName: "M2000-5",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Fighter:   true,
		},
		PlatformDesignation: "Mirage 2000",
		TypeDesignation:     "Mirage 2000-5",
		OfficialName:        "Mirage 2000",
	},
	{
		ACMIShortName: "MQ-1",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "MQ-1",
		TypeDesignation:     "MQ-1A",
		OfficialName:        "Predator",
	},
	{
		ACMIShortName: "MQ-9",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "MQ-9",
		TypeDesignation:     "MQ-9",
		OfficialName:        "Reaper",
	},
	{
		ACMIShortName:       "OH-58D",
		tags:                oh58Data.tags,
		PlatformDesignation: oh58Data.PlatformDesignation,
		TypeDesignation:     "OH-58D",
		OfficialName:        oh58Data.OfficialName,
	},
	{
		ACMIShortName:       "OH58D",
		tags:                oh58Data.tags,
		PlatformDesignation: oh58Data.PlatformDesignation,
		TypeDesignation:     "OH-58D",
		OfficialName:        oh58Data.OfficialName,
	},
	{
		ACMIShortName: "SH-3W",
		tags: map[AircraftTag]bool{
			RotaryWing: true,
			Unarmed:    true,
		},
		PlatformDesignation: "SH-3",
		TypeDesignation:     "SH-3W",
		OfficialName:        "Sea King",
	},
	{
		ACMIShortName: "SH-60B",
		tags: map[AircraftTag]bool{
			RotaryWing: true,
			Unarmed:    true,
		},
		PlatformDesignation: "SH-60",
		TypeDesignation:     "SH-60B",
		OfficialName:        "Seahawk",
	},
	{
		ACMIShortName: "Su-17M4",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Fighter:   true,
		},
		PlatformDesignation: "Su-17",
		TypeDesignation:     "Su-17M4",
		NATOReportingName:   "Fitter",
		threatRadius:        20 * unit.NauticalMile,
	},
	{
		ACMIShortName:       "Su-27",
		tags:                flankerData.tags,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-27",
		NATOReportingName:   flankerData.NATOReportingName,
		threatRadius:        flankerData.threatRadius,
	},
	{
		ACMIShortName:       "Su-30",
		tags:                flankerData.tags,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-30",
		NATOReportingName:   flankerData.NATOReportingName,
		threatRadius:        flankerData.threatRadius,
	},
	{
		ACMIShortName:       "Su-33",
		tags:                flankerData.tags,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-33",
		NATOReportingName:   flankerData.NATOReportingName,
		threatRadius:        flankerData.threatRadius,
	},
	{
		ACMIShortName: "Su-34",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Fighter:   true,
		},
		PlatformDesignation: "Su-34",
		TypeDesignation:     "Su-34",
		OfficialName:        "Fullback",
		threatRadius:        flankerData.threatRadius,
	},
	{
		ACMIShortName: "Tu-22M3",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "Tu-22",
		TypeDesignation:     "Tu-22M",
		OfficialName:        "Backfire",
	},
	{
		ACMIShortName: "Tu-95MS",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "Tu-95",
		TypeDesignation:     "Tu-95MS",
		OfficialName:        "Bear",
	},
	{
		ACMIShortName: "Tu-142",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			Unarmed:   true,
		},
		PlatformDesignation: "Tu-142",
		TypeDesignation:     "Tu-142",
		OfficialName:        "Bear",
	},
	{
		ACMIShortName: "UH-1H",
		tags: map[AircraftTag]bool{
			RotaryWing: true,
			Unarmed:    true,
		},
		PlatformDesignation: "UH-1",
		TypeDesignation:     "UH-1H",
		OfficialName:        "Iroquois",
		Nickname:            "Huey",
	},
	{
		ACMIShortName: "UH-60A",
		tags: map[AircraftTag]bool{
			RotaryWing: true,
			Unarmed:    true,
		},
		PlatformDesignation: "UH-60",
		TypeDesignation:     "UH-60A",
		OfficialName:        "Black Hawk",
	},
}

// aircraftDataLUT maps the name exported in ACMI data to aircraft data
var aircraftDataLUT map[string]Aircraft

func init() {
	aircraftDataLUT = make(map[string]Aircraft)
	for _, vars := range [][]Aircraft{
		aircraftData,
		a10Variants(),
		ah64Variants(),
		c101Variants(),
		f86Variants(),
		f4Variants(),
		f5Variants(),
		f14Variants(),
		f15Variants(),
		f16Variants(),
		fa18Variants(),
		ka50Variants(),
		mi24Variants(),
		mirageF1Variants(),
		sa342Variants(),
		ftVariants(),
		fencerVariants(),
		foxbatVariants(),
		fulcrumVariants(),
		frogfootVariants(),
		kc135Variants(),
		l39Variants(),
		mb339Variants(),
		s3Variants(),
		tornadoVariants(),
	} {
		for _, data := range vars {
			aircraftDataLUT[data.ACMIShortName] = data
		}
	}
}

// GetAircraftData returns the aircraft data for the given name, if it exists.
// The name should be the Name property of an ACMI object.
// The second return value is false if the data does not exist.
func GetAircraftData(name string) (Aircraft, bool) {
	data, ok := aircraftDataLUT[name]
	if !ok {
		log.Warn().Str("name", name).Msg("Aircraft missing from encyclopedia")
	}
	return data, ok
}
