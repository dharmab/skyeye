package encyclopedia

import (
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/rs/zerolog/log"
)

// Data source: https://github.com/Quaggles/dcs-lua-datamine/tree/master/_G/db/Units/Planes/Plane

type AircraftTag int

const (
	AnyAircraft AircraftTag = iota
	FixedWing
	RotaryWing
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
	Nickname string
	// threatFactor is a weight roughly indicating how dangerous the aircraft is.
	// 0 is unarmed. Threat factor increases with both airframe and weapon capabilities.
	threatFactor float64
}

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

func (a Aircraft) ThreatFactor() float64 {
	return a.threatFactor
}

const (
	Unarmed               float64 = 0
	Helicopter            float64 = 0.01
	AttackerWithGuns      float64 = 0.02
	AttackerWithIR        float64 = 0.05
	FighterWithGuns       float64 = 0.10
	FighterWithIR         float64 = 0.50
	DefaultThreat         float64 = 0.50
	FighterWithSAR        float64 = 0.75
	FighterWithAR         float64 = 0.90
	AirSuperiorityFighter float64 = 1.00
)

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
			threatFactor:        data.threatFactor,
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
	},
	PlatformDesignation: "A-10",
	OfficialName:        "Thunderbolt",
	Nickname:            "Warthog",
	threatFactor:        AttackerWithIR,
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

var c101Data = Aircraft{
	tags:                map[AircraftTag]bool{FixedWing: true},
	PlatformDesignation: "C-101",
	OfficialName:        "Aviojet",
	threatFactor:        FighterWithIR,
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
	tags:                map[AircraftTag]bool{FixedWing: true},
	PlatformDesignation: "F-86",
	OfficialName:        "Sabre",
	threatFactor:        FighterWithGuns,
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
	tags:                map[AircraftTag]bool{FixedWing: true},
	PlatformDesignation: "F-4",
	OfficialName:        "Phantom",
	threatFactor:        FighterWithSAR,
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
	},
	PlatformDesignation: "F-5",
	OfficialName:        "Tiger",
	threatFactor:        FighterWithGuns,
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
	},
	PlatformDesignation: "F-14",
	OfficialName:        "Tomcat",
	threatFactor:        AirSuperiorityFighter,
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
	},
	// Use "Eagle" for Strike Eagle because radar cannot distinguish between the two
	OfficialName: "Eagle",
	threatFactor: AirSuperiorityFighter,
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
	},
	PlatformDesignation: "F-16",
	OfficialName:        "Falcon",
	Nickname:            "Viper",
	threatFactor:        FighterWithAR,
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
	},
	PlatformDesignation: "FA-18",
	OfficialName:        "Hornet",
	threatFactor:        FighterWithAR,
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

var mirageF1Data = Aircraft{
	ACMIShortName: "Mirage-F1",
	tags: map[AircraftTag]bool{
		FixedWing: true,
	},
	PlatformDesignation: "Mirage F1",
	OfficialName:        "Mirage F1",
	threatFactor:        FighterWithSAR,
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

var ftData = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
	},
	PlatformDesignation: "MiG-15",
	NATOReportingName:   mig15NATOReportingName,
	threatFactor:        FighterWithSAR,
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
	},
	PlatformDesignation: "Su-24",
	NATOReportingName:   "Fencer",
	threatFactor:        FighterWithSAR,
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
	},
	PlatformDesignation: "MiG-25",
	NATOReportingName:   "Foxbat",
	threatFactor:        FighterWithSAR,
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
	},
	PlatformDesignation: "MiG-29",
	NATOReportingName:   "Fulcrum",
	threatFactor:        FighterWithAR,
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
	},
	PlatformDesignation: "Su-25",
	NATOReportingName:   "Frogfoot",
	threatFactor:        AttackerWithIR,
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
	},
	PlatformDesignation: "Su-27",
	NATOReportingName:   "Flanker",
	threatFactor:        AirSuperiorityFighter,
}

var kc135Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
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
	tags:                map[AircraftTag]bool{FixedWing: true},
	PlatformDesignation: "L-39",
	OfficialName:        "Albatros",
	threatFactor:        FighterWithIR,
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
	tags:                map[AircraftTag]bool{FixedWing: true},
	PlatformDesignation: "MB-339",
	threatFactor:        FighterWithGuns,
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
	tags:                map[AircraftTag]bool{FixedWing: true},
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
	tags:                map[AircraftTag]bool{FixedWing: true},
	PlatformDesignation: "Tornado",
	OfficialName:        "Tornado",
	threatFactor:        FighterWithSAR,
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
		ACMIShortName: "A-4E",
		tags: map[AircraftTag]bool{
			FixedWing: true,
		},
		PlatformDesignation: "A-4",
		TypeDesignation:     "A-4E",
		OfficialName:        "Skyhawk",
		Nickname:            "Scooter",
		threatFactor:        AttackerWithIR,
	},
	{
		ACMIShortName:       "A-20G",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "A-20",
		TypeDesignation:     "A-20G",
		OfficialName:        "Havoc",
		threatFactor:        AttackerWithGuns,
	},
	{
		ACMIShortName:       "A-50",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "A-50",
		TypeDesignation:     "A-50",
		NATOReportingName:   "Mainstay",
	},
	{
		ACMIShortName: "AJS 37",
		tags: map[AircraftTag]bool{
			FixedWing: true,
		},
		PlatformDesignation: "AJS37",
		OfficialName:        "Viggen",
		threatFactor:        AttackerWithIR,
	},
	{
		ACMIShortName: "AV8BNA",
		tags: map[AircraftTag]bool{
			FixedWing: true,
		},
		PlatformDesignation: "AV-8",
		TypeDesignation:     "AV-8B",
		OfficialName:        "Harrier",
		threatFactor:        FighterWithIR,
	},
	{
		ACMIShortName:       "An-26B",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "An-26",
		TypeDesignation:     "An-26B",
		NATOReportingName:   "Curl",
	},
	{
		ACMIShortName:       "An-30M",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "An-30",
		TypeDesignation:     "An-30M",
		NATOReportingName:   "Clank",
	},
	{
		ACMIShortName:       "B-17G",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "B-17",
		TypeDesignation:     "B-17G",
		OfficialName:        "Flying Fortress",
	},
	{
		ACMIShortName:       "B-52H",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "B-52",
		TypeDesignation:     "B-52H",
		OfficialName:        "Stratofortress",
		Nickname:            "Buff",
	},
	{
		ACMIShortName:       "B-1B",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "B-1",
		TypeDesignation:     "B-1B",
		OfficialName:        "Lancer",
		Nickname:            "Bone",
	},
	{
		ACMIShortName:       "C-17A",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "C-17",
		TypeDesignation:     "C-17A",
		OfficialName:        "Globemaster",
	},
	{
		ACMIShortName:       "C-47",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "C-47",
		OfficialName:        "Skytrain",
	},
	{
		ACMIShortName:       "C-130",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "C-130",
		TypeDesignation:     "C-130",
		OfficialName:        "Hercules",
		Nickname:            "Herc",
	},
	{
		ACMIShortName:       "E-2C",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "E-2",
		TypeDesignation:     "E-2C",
		OfficialName:        "Hawkeye",
	},
	{
		ACMIShortName:       "E-3A",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "E-3",
		TypeDesignation:     "E-3A",
		OfficialName:        "Sentry",
	},
	{
		ACMIShortName:       "F-117A",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "F-117",
		TypeDesignation:     "F-117A",
		OfficialName:        "Nighthawk",
		Nickname:            "Goblin",
	},
	{
		ACMIShortName:       "H-6J",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "Tu-16",
		TypeDesignation:     "H-6J",
		NATOReportingName:   "Badger",
	},
	{
		ACMIShortName:       "IL-76MD",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "Il-76",
		TypeDesignation:     "Il-76MD",
		NATOReportingName:   "Candid",
	},
	{
		ACMIShortName:       "IL-78M",
		tags:                map[AircraftTag]bool{FixedWing: true},
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
		threatFactor:        flankerData.threatFactor,
	},
	{
		ACMIShortName:       "JF-17",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "JF-17",
		TypeDesignation:     "JF-17",
		OfficialName:        "Thunder",
		threatFactor:        FighterWithAR,
	},
	{
		ACMIShortName:       "KC130",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "KC-130",
		TypeDesignation:     "KC-130",
		OfficialName:        "Hercules",
		Nickname:            "Herc",
	},
	{
		ACMIShortName:       "KJ-2000",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "KJ-2000",
		TypeDesignation:     "KJ-2000",
		OfficialName:        "Mainring",
	},
	{
		ACMIShortName: "M-2000C",
		tags: map[AircraftTag]bool{
			FixedWing: true,
		},
		PlatformDesignation: "Mirage 2000",
		TypeDesignation:     "Mirage 2000C",
		OfficialName:        "Mirage 2000",
		threatFactor:        FighterWithSAR,
	},
	{
		ACMIShortName: "Mi-24P",
		tags: map[AircraftTag]bool{
			RotaryWing: true,
		},
		PlatformDesignation: "Mi-24",
		TypeDesignation:     "Mi-24P",
		NATOReportingName:   "Hind",
		threatFactor:        Helicopter,
	},
	{
		ACMIShortName:       "Mi-26",
		tags:                map[AircraftTag]bool{RotaryWing: true},
		PlatformDesignation: "Mi-26",
		TypeDesignation:     "Mi-26",
		NATOReportingName:   "Hip",
	},
	{
		ACMIShortName: "MiG-19P",
		tags: map[AircraftTag]bool{
			FixedWing: true,
		},
		PlatformDesignation: "MiG-19",
		TypeDesignation:     "MiG-19P",
		NATOReportingName:   "Farmer",
		threatFactor:        FighterWithGuns,
	},
	{
		ACMIShortName: "MiG-21Bis",
		tags: map[AircraftTag]bool{
			FixedWing: true,
		},
		PlatformDesignation: "MiG-21",
		TypeDesignation:     "MiG-21bis",
		NATOReportingName:   "Fishbed",
		threatFactor:        FighterWithSAR,
	},
	{
		ACMIShortName: "MiG-23MLD",
		tags: map[AircraftTag]bool{
			FixedWing: true,
		},
		PlatformDesignation: "MiG-23",
		TypeDesignation:     "MiG-23MLD",
		NATOReportingName:   "Flogger",
		threatFactor:        FighterWithSAR,
	},
	{
		ACMIShortName: "MiG-27K",
		tags: map[AircraftTag]bool{
			FixedWing: true,
		},
		PlatformDesignation: "MiG-27",
		TypeDesignation:     "MiG-27K",
		NATOReportingName:   "Flogger",
		threatFactor:        FighterWithSAR,
	},
	{
		ACMIShortName: "MiG-31",
		tags: map[AircraftTag]bool{
			FixedWing: true,
		},
		PlatformDesignation: "MiG-31",
		TypeDesignation:     "MiG-31",
		NATOReportingName:   "Foxhound",
		threatFactor:        FighterWithSAR,
	},
	{
		ACMIShortName: "M2000-5",
		tags: map[AircraftTag]bool{
			FixedWing: true,
		},
		PlatformDesignation: "Mirage 2000",
		TypeDesignation:     "Mirage 2000-5",
		OfficialName:        "Mirage 2000",
		threatFactor:        FighterWithSAR,
	},
	{
		ACMIShortName:       "MQ-1",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "MQ-1",
		TypeDesignation:     "MQ-1A",
		OfficialName:        "Predator",
	},
	{
		ACMIShortName:       "MQ-9",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "MQ-9",
		TypeDesignation:     "MQ-9",
		OfficialName:        "Reaper",
	},
	{
		ACMIShortName: "Su-17M4",
		tags: map[AircraftTag]bool{
			FixedWing: true,
		},
		PlatformDesignation: "Su-17",
		TypeDesignation:     "Su-17M4",
		NATOReportingName:   "Fitter",
		threatFactor:        FighterWithSAR,
	},
	{
		ACMIShortName:       "Su-27",
		tags:                flankerData.tags,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-27",
		NATOReportingName:   flankerData.NATOReportingName,
		threatFactor:        flankerData.threatFactor,
	},
	{
		ACMIShortName:       "Su-30",
		tags:                flankerData.tags,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-30",
		NATOReportingName:   flankerData.NATOReportingName,
		threatFactor:        flankerData.threatFactor,
	},
	{
		ACMIShortName:       "Su-33",
		tags:                flankerData.tags,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-33",
		NATOReportingName:   flankerData.NATOReportingName,
		threatFactor:        flankerData.threatFactor,
	},
	{
		ACMIShortName: "Su-34",
		tags: map[AircraftTag]bool{
			FixedWing: true,
		},
		PlatformDesignation: "Su-34",
		TypeDesignation:     "Su-34",
		OfficialName:        "Fullback",
		threatFactor:        flankerData.threatFactor,
	},
	{
		ACMIShortName:       "Tu-22M3",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "Tu-22",
		TypeDesignation:     "Tu-22M",
		OfficialName:        "Backfire",
	},
	{
		ACMIShortName:       "Tu-95MS",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "Tu-95",
		TypeDesignation:     "Tu-95MS",
		OfficialName:        "Bear",
	},
	{
		ACMIShortName:       "Tu-142",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "Tu-142",
		TypeDesignation:     "Tu-142",
		OfficialName:        "Bear",
	},
	{
		ACMIShortName:       "UH-1H",
		tags:                map[AircraftTag]bool{RotaryWing: true},
		PlatformDesignation: "UH-1",
		TypeDesignation:     "UH-1H",
		OfficialName:        "Iroquois",
		Nickname:            "Huey",
	},
}

// aircraftDataLUT maps the name exported in ACMI data to aircraft data
var aircraftDataLUT map[string]Aircraft

func init() {
	aircraftDataLUT = make(map[string]Aircraft)
	for _, vars := range [][]Aircraft{
		aircraftData,
		a10Variants(),
		c101Variants(),
		f86Variants(),
		f4Variants(),
		f5Variants(),
		f14Variants(),
		f15Variants(),
		f16Variants(),
		fa18Variants(),
		mirageF1Variants(),
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
