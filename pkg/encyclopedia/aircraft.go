package encyclopedia

import (
	"github.com/dharmab/skyeye/pkg/brevity"
)

type AircraftTag int

const (
	AnyAircraft AircraftTag = iota
	FixedWing
	RotaryWing

	HasActiveRadarMissiles
	HasSemiActiveRadarMissiles
	HasInfraredMissiles
	HasCannon
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

func (a Aircraft) ThreatClass() ThreatClass {
	if a.HasAnyTag(HasActiveRadarMissiles, HasSemiActiveRadarMissiles) {
		return SAR2OrAR1
	} else if a.HasTag(HasInfraredMissiles) {
		return SAR1OrIR
	} else if a.HasTag(HasCannon) {
		return Guns
	}
	return NoFactor
}

var a10Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:           true,
		HasInfraredMissiles: true,
		HasCannon:           true,
	},
	PlatformDesignation: "A-10",
	OfficialName:        "Thunderbolt",
	Nickname:            "Warthog",
}

var c101Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:           true,
		HasInfraredMissiles: true,
		HasCannon:           true,
	},
	PlatformDesignation: "C-101",
	OfficialName:        "Aviojet",
}

var f4Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:                  true,
		HasSemiActiveRadarMissiles: true,
		HasInfraredMissiles:        true,
		HasCannon:                  true,
	},
	PlatformDesignation: "F-4",
	OfficialName:        "Phantom",
}

var f5Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:           true,
		HasInfraredMissiles: true,
		HasCannon:           true,
	},
	PlatformDesignation: "F-5",
	OfficialName:        "Tiger",
}

var f14Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:                  true,
		HasActiveRadarMissiles:     true,
		HasSemiActiveRadarMissiles: true,
		HasInfraredMissiles:        true,
		HasCannon:                  true,
	},
	PlatformDesignation: "F-14",
	OfficialName:        "Tomcat",
}

var f15Data = Aircraft{
	PlatformDesignation: "F-15",
	tags: map[AircraftTag]bool{
		FixedWing:                  true,
		HasActiveRadarMissiles:     true,
		HasSemiActiveRadarMissiles: true,
		HasInfraredMissiles:        true,
		HasCannon:                  true,
	},
	// Use "Eagle" for Strike Eagle because radar cannot distinguish between the two
	OfficialName: "Eagle",
}

var f16Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:              true,
		HasActiveRadarMissiles: true,
		HasInfraredMissiles:    true,
		HasCannon:              true,
	},
	PlatformDesignation: "F-16",
	OfficialName:        "Falcon",
	Nickname:            "Viper",
}

var fa18Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:                  true,
		HasActiveRadarMissiles:     true,
		HasSemiActiveRadarMissiles: true,
		HasInfraredMissiles:        true,
		HasCannon:                  true,
	},
	PlatformDesignation: "F/A-18",
	OfficialName:        "Hornet",
}

var mirageF1Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:                  true,
		HasSemiActiveRadarMissiles: true,
		HasInfraredMissiles:        true,
		HasCannon:                  true,
	},
	PlatformDesignation: "Mirage F1",
	OfficialName:        "Mirage F1",
}

var fencerData = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:                  true,
		HasSemiActiveRadarMissiles: true,
		HasInfraredMissiles:        true,
		HasCannon:                  true,
	},
	PlatformDesignation: "Su-24",
	NATOReportingName:   "Fencer",
}

var foxbatData = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:                  true,
		HasSemiActiveRadarMissiles: true,
		HasInfraredMissiles:        true,
		HasCannon:                  true,
	},
	PlatformDesignation: "MiG-25",
	NATOReportingName:   "Foxbat",
}

var fulcrumData = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:                  true,
		HasSemiActiveRadarMissiles: true,
		HasInfraredMissiles:        true,
		HasCannon:                  true,
	},
	PlatformDesignation: "MiG-29",
	NATOReportingName:   "Fulcrum",
}

var frogfootData = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:           true,
		HasInfraredMissiles: true,
		HasCannon:           true,
	},
	PlatformDesignation: "Su-25",
	NATOReportingName:   "Frogfoot",
}

var flankerData = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:                  true,
		HasActiveRadarMissiles:     true,
		HasSemiActiveRadarMissiles: true,
		HasInfraredMissiles:        true,
		HasCannon:                  true,
	},
	PlatformDesignation: "Su-27",
	NATOReportingName:   "Flanker",
}

var l39Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:           true,
		HasInfraredMissiles: true,
		HasCannon:           true,
	},
	PlatformDesignation: "L-39",
	OfficialName:        "Albatros",
}

var mb339Data = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing: true,
		HasCannon: true,
	},
	PlatformDesignation: "MB-339",
}

var tornadoData = Aircraft{
	tags: map[AircraftTag]bool{
		FixedWing:                  true,
		HasSemiActiveRadarMissiles: true,
		HasInfraredMissiles:        true,
		HasCannon:                  true,
	},
	PlatformDesignation: "Tornado",
	OfficialName:        "Tornado",
}

var aircraftData = append([]Aircraft{
	{
		ACMIShortName:       "A-10A",
		tags:                a10Data.tags,
		PlatformDesignation: a10Data.PlatformDesignation,
		TypeDesignation:     "A-10A",
		OfficialName:        a10Data.OfficialName,
		Nickname:            a10Data.Nickname,
	},
	{
		ACMIShortName:       "A-10C",
		tags:                a10Data.tags,
		PlatformDesignation: a10Data.PlatformDesignation,
		TypeDesignation:     "A-10C",
		OfficialName:        a10Data.OfficialName,
		Nickname:            a10Data.Nickname,
	},
	{
		ACMIShortName: "A-4E",
		tags: map[AircraftTag]bool{
			FixedWing:           true,
			HasInfraredMissiles: true,
			HasCannon:           true,
		},
		PlatformDesignation: "A-4",
		TypeDesignation:     "A-4E",
		OfficialName:        "Skyhawk",
		Nickname:            "Scooter",
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
			FixedWing:           true,
			HasInfraredMissiles: true,
		},
		PlatformDesignation: "AJS37",
		OfficialName:        "Viggen",
	},
	{
		ACMIShortName: "AV-8BNA",
		tags: map[AircraftTag]bool{
			FixedWing:           true,
			HasInfraredMissiles: true,
			HasCannon:           true,
		},
		PlatformDesignation: "AV-8",
		TypeDesignation:     "AV-8B",
		OfficialName:        "Harrier",
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
		ACMIShortName:       "B-52H",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "B-52",
		TypeDesignation:     "B-52H",
		OfficialName:        "Stratofortress",
		Nickname:            "Buff",
	},
	{
		ACMIShortName:       "C-101CC",
		tags:                c101Data.tags,
		PlatformDesignation: c101Data.PlatformDesignation,
		TypeDesignation:     "C-101CC",
		OfficialName:        c101Data.OfficialName,
	},
	{
		ACMIShortName:       "C-101EB",
		tags:                c101Data.tags,
		PlatformDesignation: c101Data.PlatformDesignation,
		TypeDesignation:     "C-101EB",
		OfficialName:        c101Data.OfficialName,
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
		ACMIShortName: "F-86F",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			HasCannon: true,
		},
		PlatformDesignation: "F-86",
		TypeDesignation:     "F-86F",
		OfficialName:        "Sabre",
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
		ACMIShortName:       "F-4E-45MC",
		tags:                f4Data.tags,
		PlatformDesignation: f4Data.PlatformDesignation,
		TypeDesignation:     "F-4E",
		OfficialName:        f4Data.OfficialName,
	},
	{
		ACMIShortName:       "F-4E",
		tags:                f4Data.tags,
		PlatformDesignation: f4Data.PlatformDesignation,
		TypeDesignation:     "F-4E",
		OfficialName:        f4Data.OfficialName,
	},
	{
		ACMIShortName:       "F-5E",
		tags:                f5Data.tags,
		PlatformDesignation: f5Data.PlatformDesignation,
		TypeDesignation:     "F-5E",
		OfficialName:        f5Data.OfficialName,
	},
	{
		ACMIShortName:       "F-5E-3",
		tags:                f5Data.tags,
		PlatformDesignation: f5Data.PlatformDesignation,
		TypeDesignation:     "F-5E",
		OfficialName:        f5Data.OfficialName,
	},
	{
		ACMIShortName:       "F-14A-135-GR",
		tags:                f14Data.tags,
		PlatformDesignation: f14Data.PlatformDesignation,
		TypeDesignation:     "F-14A",
		OfficialName:        f14Data.OfficialName,
	},
	{
		ACMIShortName:       "F-14A",
		tags:                f14Data.tags,
		PlatformDesignation: f14Data.PlatformDesignation,
		TypeDesignation:     "F-14A",
		OfficialName:        f14Data.OfficialName,
	},
	{
		ACMIShortName:       "F-14B",
		tags:                f14Data.tags,
		PlatformDesignation: f14Data.PlatformDesignation,
		TypeDesignation:     "F-14B",
		OfficialName:        f14Data.OfficialName,
	},
	{
		ACMIShortName:       "F-15C",
		tags:                f15Data.tags,
		PlatformDesignation: f15Data.PlatformDesignation,
		TypeDesignation:     "F-15C",
		OfficialName:        f15Data.OfficialName,
	},
	{
		ACMIShortName:       "F-15E",
		tags:                f15Data.tags,
		PlatformDesignation: f15Data.PlatformDesignation,
		TypeDesignation:     "F-15E",
		OfficialName:        "Strike Eagle",
		Nickname:            "Mudhen",
	},
	{
		ACMIShortName:       "F-16A",
		tags:                f16Data.tags,
		PlatformDesignation: f16Data.PlatformDesignation,
		TypeDesignation:     "F-16A",
		OfficialName:        f16Data.OfficialName,
		Nickname:            f16Data.Nickname,
	},
	{
		ACMIShortName:       "F-16C",
		tags:                f16Data.tags,
		PlatformDesignation: f16Data.PlatformDesignation,
		TypeDesignation:     "F-16C",
		OfficialName:        f16Data.OfficialName,
		Nickname:            f16Data.Nickname,
	},
	{
		ACMIShortName:       "F-16C_50",
		tags:                f16Data.tags,
		PlatformDesignation: f16Data.PlatformDesignation,
		TypeDesignation:     "F-16C",
		OfficialName:        f16Data.OfficialName,
		Nickname:            f16Data.Nickname,
	},
	{
		ACMIShortName:       "F/A-18A",
		tags:                fa18Data.tags,
		PlatformDesignation: fa18Data.PlatformDesignation,
		TypeDesignation:     "F/A-18A",
		OfficialName:        fa18Data.OfficialName,
	},
	{
		ACMIShortName:       "F/A-18C",
		tags:                fa18Data.tags,
		PlatformDesignation: fa18Data.PlatformDesignation,
		TypeDesignation:     "F/A-18C",
		OfficialName:        fa18Data.OfficialName,
	},
	{
		ACMIShortName:       "FA-18C_hornet",
		tags:                fa18Data.tags,
		PlatformDesignation: fa18Data.PlatformDesignation,
		TypeDesignation:     "F/A-18C",
		OfficialName:        fa18Data.OfficialName,
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
	},
	{
		ACMIShortName: "JF-17",
		tags: map[AircraftTag]bool{
			FixedWing:              true,
			HasActiveRadarMissiles: true,
			HasInfraredMissiles:    true,
			// Gun is A-G only
		},
		PlatformDesignation: "JF-17",
		TypeDesignation:     "JF-17",
		OfficialName:        "Thunder",
	},
	{
		ACMIShortName:       "KC-130",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "KC-130",
		TypeDesignation:     "KC-130",
		OfficialName:        "Hercules",
		Nickname:            "Herc",
	},
	{
		ACMIShortName:       "KC-135",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "KC-135",
		TypeDesignation:     "KC-135",
		OfficialName:        "Stratotanker",
	},
	{
		ACMIShortName:       "KJ-2000",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "KJ-2000",
		TypeDesignation:     "KJ-2000",
		OfficialName:        "Mainring",
	},
	{
		ACMIShortName:       "L-39C",
		tags:                l39Data.tags,
		PlatformDesignation: l39Data.PlatformDesignation,
		TypeDesignation:     "L-39C",
		OfficialName:        l39Data.OfficialName,
	},
	{
		ACMIShortName:       "L-39ZA",
		tags:                l39Data.tags,
		PlatformDesignation: l39Data.PlatformDesignation,
		TypeDesignation:     "L-39ZA",
		OfficialName:        l39Data.OfficialName,
	},
	{
		ACMIShortName: "M2000C",
		tags: map[AircraftTag]bool{
			FixedWing:                  true,
			HasSemiActiveRadarMissiles: true,
			HasInfraredMissiles:        true,
			HasCannon:                  true,
		},
		PlatformDesignation: "Mirage 2000",
		TypeDesignation:     "Mirage 2000C",
		OfficialName:        "Mirage 2000",
	},
	{
		ACMIShortName:       "MB-339A",
		tags:                mb339Data.tags,
		PlatformDesignation: mb339Data.PlatformDesignation,
		TypeDesignation:     "MB-339A",
	},
	{
		ACMIShortName:       "MB-339APAN",
		tags:                mb339Data.tags,
		PlatformDesignation: mb339Data.PlatformDesignation,
		TypeDesignation:     "MB-339A",
	},
	// TODO Mi-28, IR
	{
		ACMIShortName: "Mi-24P",
		tags: map[AircraftTag]bool{
			RotaryWing:          true,
			HasInfraredMissiles: true,
			HasCannon:           true,
		},
		PlatformDesignation: "Mi-24",
		TypeDesignation:     "Mi-24P",
		NATOReportingName:   "Hind",
	},
	{
		ACMIShortName:       "Mi-26",
		tags:                map[AircraftTag]bool{RotaryWing: true},
		PlatformDesignation: "Mi-26",
		TypeDesignation:     "Mi-26",
		NATOReportingName:   "Hip",
	},
	{
		ACMIShortName: "MiG-15bis",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			HasCannon: true,
		},
		PlatformDesignation: "MiG-15",
		TypeDesignation:     "MiG-15bis",
		NATOReportingName:   mig15NATOReportingName,
	},
	{
		ACMIShortName: "MiG-19P",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			HasCannon: true,
		},
		PlatformDesignation: "MiG-19",
		TypeDesignation:     "MiG-19P",
		NATOReportingName:   "Farmer",
	},
	{
		ACMIShortName: "MiG-21Bis",
		tags: map[AircraftTag]bool{
			FixedWing:                  true,
			HasSemiActiveRadarMissiles: true,
			HasInfraredMissiles:        true,
			HasCannon:                  true,
		},
		PlatformDesignation: "MiG-21",
		TypeDesignation:     "MiG-21Bis",
		NATOReportingName:   "Fishbed",
	},
	{
		ACMIShortName: "MiG-23MLD",
		tags: map[AircraftTag]bool{
			FixedWing:                  true,
			HasSemiActiveRadarMissiles: true,
			HasInfraredMissiles:        true,
			HasCannon:                  true,
		},
		PlatformDesignation: "MiG-23",
		TypeDesignation:     "MiG-23MLD",
		NATOReportingName:   "Flogger",
	},
	{
		ACMIShortName:       "MiG-25PD",
		tags:                foxbatData.tags,
		PlatformDesignation: foxbatData.PlatformDesignation,
		TypeDesignation:     "MiG-25PD",
		NATOReportingName:   foxbatData.NATOReportingName,
	},
	{
		ACMIShortName:       "MiG-25RBT",
		tags:                foxbatData.tags,
		PlatformDesignation: foxbatData.PlatformDesignation,
		TypeDesignation:     "MiG-25RBT",
		NATOReportingName:   foxbatData.NATOReportingName,
	},
	{
		ACMIShortName: "MiG-27K",
		tags: map[AircraftTag]bool{
			FixedWing:                  true,
			HasSemiActiveRadarMissiles: true,
			HasInfraredMissiles:        true,
			HasCannon:                  true,
		},
		PlatformDesignation: "MiG-27",
		TypeDesignation:     "MiG-27K",
		NATOReportingName:   "Flogger",
	},
	{
		ACMIShortName:       "MiG-29A",
		tags:                fulcrumData.tags,
		PlatformDesignation: fulcrumData.PlatformDesignation,
		TypeDesignation:     "MiG-29A",
		NATOReportingName:   fulcrumData.NATOReportingName,
	},
	{
		ACMIShortName:       "MiG-29S",
		tags:                fulcrumData.tags,
		PlatformDesignation: fulcrumData.PlatformDesignation,
		TypeDesignation:     "MiG-29S",
		NATOReportingName:   fulcrumData.NATOReportingName,
	},
	{
		ACMIShortName: "MiG-31",
		tags: map[AircraftTag]bool{
			FixedWing:                  true,
			HasSemiActiveRadarMissiles: true,
			HasInfraredMissiles:        true,
			HasCannon:                  true,
		},
		PlatformDesignation: "MiG-31",
		TypeDesignation:     "MiG-31",
		NATOReportingName:   "Foxhound",
	},
	{
		ACMIShortName: "M2000-5",
		tags: map[AircraftTag]bool{
			FixedWing:                  true,
			HasSemiActiveRadarMissiles: true,
			HasInfraredMissiles:        true,
			HasCannon:                  true,
		},
		PlatformDesignation: "Mirage 2000",
		TypeDesignation:     "Mirage 2000-5",
		OfficialName:        "Mirage 2000",
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
		ACMIShortName:       "S-3B",
		tags:                map[AircraftTag]bool{FixedWing: true},
		PlatformDesignation: "S-3",
		TypeDesignation:     "S-3B",
		OfficialName:        "Viking",
	},
	{
		ACMIShortName: "Su-17M4",
		tags: map[AircraftTag]bool{
			FixedWing: true,
			HasCannon: true,
		},
		PlatformDesignation: "Su-17",
		TypeDesignation:     "Su-17M4",
		NATOReportingName:   "Fitter",
	},
	{
		ACMIShortName:       "Su-24M",
		tags:                fencerData.tags,
		PlatformDesignation: fencerData.PlatformDesignation,
		TypeDesignation:     "Su-24M",
		NATOReportingName:   fencerData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-24MR",
		tags:                fencerData.tags,
		PlatformDesignation: fencerData.PlatformDesignation,
		TypeDesignation:     "Su-24MR",
		NATOReportingName:   fencerData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-25",
		tags:                frogfootData.tags,
		PlatformDesignation: frogfootData.PlatformDesignation,
		TypeDesignation:     "Su-25",
		NATOReportingName:   frogfootData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-25T",
		tags:                frogfootData.tags,
		PlatformDesignation: frogfootData.PlatformDesignation,
		TypeDesignation:     "Su-25T",
		NATOReportingName:   frogfootData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-25TM",
		tags:                frogfootData.tags,
		PlatformDesignation: frogfootData.PlatformDesignation,
		TypeDesignation:     "Su-25TM",
		NATOReportingName:   frogfootData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-27",
		tags:                flankerData.tags,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-27",
		NATOReportingName:   flankerData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-30",
		tags:                flankerData.tags,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-30",
		NATOReportingName:   flankerData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-33",
		tags:                flankerData.tags,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-33",
		NATOReportingName:   flankerData.NATOReportingName,
	},
	{
		ACMIShortName: "Su-34",
		tags: map[AircraftTag]bool{
			FixedWing:                  true,
			HasActiveRadarMissiles:     true,
			HasSemiActiveRadarMissiles: true,
			HasInfraredMissiles:        true,
			HasCannon:                  true,
		},
		PlatformDesignation: "Su-34",
		TypeDesignation:     "Su-34",
		OfficialName:        "Fullback",
	},
	{
		ACMIShortName:       "Tornado",
		tags:                tornadoData.tags,
		PlatformDesignation: tornadoData.PlatformDesignation,
		TypeDesignation:     "Tornado",
		OfficialName:        tornadoData.OfficialName,
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
}, mirageF1Variants()...)

func mirageF1Variants() []Aircraft {
	variants := []Aircraft{}
	for _, v := range []string{"B", "BD", "BE", "BQ", "C-200", "C", "CE", "CG", "CH", "CJ", "CK", "CR", "CT", "CZ", "DDA", "ED", "EDA", "EE", "EH", "EQ", "JA", "M-CE", "M-EE"} {
		variants = append(variants, Aircraft{
			ACMIShortName:       "F1" + v,
			tags:                mirageF1Data.tags,
			PlatformDesignation: mirageF1Data.PlatformDesignation,
			TypeDesignation:     "Mirage F1" + v,
			OfficialName:        mirageF1Data.OfficialName,
		})
	}
	return variants
}

func GetAircraftData(shortName string) (Aircraft, bool) {
	for _, a := range aircraftData {
		if a.ACMIShortName == shortName {
			return a, true
		}
	}
	return Aircraft{}, false
}
