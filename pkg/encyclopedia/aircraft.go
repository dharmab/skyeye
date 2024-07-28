package encyclopedia

import "github.com/dharmab/skyeye/pkg/brevity"

type Aircraft struct {
	// ACMIShortName is the Name proeprty used in ACMI telemetry.
	ACMIShortName string
	// Category is the [brevity.ContactCategory] of the aircraft - fixed wing or rotary wing.
	Category brevity.ContactCategory
	// PlatformDesignation is the official platform designation of the aircraft.
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

var a10Data = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "A-10",
	OfficialName:        "Thunderbolt",
	Nickname:            "Warthog",
}

var c101Data = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "C-101",
	OfficialName:        "Aviojet",
}

var f4Data = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "F-4",
	OfficialName:        "Phantom",
}

var f5Data = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "F-5",
	OfficialName:        "Tiger",
}

var f14Data = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "F-14",
	OfficialName:        "Tomcat",
}

var f15Data = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "F-15",
	// Use "Eagle" for Strike Eagle because radar cannot distinguish between the two
	OfficialName: "Eagle",
}

var f16Data = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "F-16",
	OfficialName:        "Falcon",
	Nickname:            "Viper",
}

var fa18Data = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "F/A-18",
	OfficialName:        "Hornet",
}

var mirageF1Data = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "Mirage F1",
	OfficialName:        "Mirage F1",
}

var fencerData = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "Su-24",
	NATOReportingName:   "Fencer",
}

var foxbatData = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "MiG-25",
	NATOReportingName:   "Foxbat",
}

var fulcrumData = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "MiG-29",
	NATOReportingName:   "Fulcrum",
}

var frogfootData = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "Su-25",
	NATOReportingName:   "Frogfoot",
}

var flankerData = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "Su-27",
	NATOReportingName:   "Flanker",
}

var l39Data = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "L-39",
	OfficialName:        "Albatros",
}

var mb339Data = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "MB-339",
}

var tornadoData = Aircraft{
	Category:            brevity.FixedWing,
	PlatformDesignation: "Tornado",
	OfficialName:        "Tornado",
}

var aircraftData = append([]Aircraft{
	{
		ACMIShortName:       "A-10A",
		Category:            a10Data.Category,
		PlatformDesignation: a10Data.PlatformDesignation,
		TypeDesignation:     "A-10A",
		OfficialName:        a10Data.OfficialName,
		Nickname:            a10Data.Nickname,
	},
	{
		ACMIShortName:       "A-10C",
		Category:            a10Data.Category,
		PlatformDesignation: a10Data.PlatformDesignation,
		TypeDesignation:     "A-10C",
		OfficialName:        a10Data.OfficialName,
		Nickname:            a10Data.Nickname,
	},
	{
		ACMIShortName:       "A-4E",
		Category:            brevity.FixedWing,
		PlatformDesignation: "A-4",
		TypeDesignation:     "A-4E",
		OfficialName:        "Skyhawk",
		Nickname:            "Scooter",
	},
	{
		ACMIShortName:       "A-50",
		Category:            brevity.FixedWing,
		PlatformDesignation: "A-50",
		TypeDesignation:     "A-50",
		NATOReportingName:   "Mainstay",
	},
	{
		ACMIShortName:       "AJS 37",
		Category:            brevity.FixedWing,
		PlatformDesignation: "AJS37",
		OfficialName:        "Viggen",
	},
	{
		ACMIShortName:       "AV-8BNA",
		Category:            brevity.FixedWing,
		PlatformDesignation: "AV-8",
		TypeDesignation:     "AV-8B",
		OfficialName:        "Harrier",
	},
	{
		ACMIShortName:       "B-1B",
		Category:            brevity.FixedWing,
		PlatformDesignation: "B-1",
		TypeDesignation:     "B-1B",
		OfficialName:        "Lancer",
		Nickname:            "Bone",
	},
	{
		ACMIShortName:       "B-52H",
		Category:            brevity.FixedWing,
		PlatformDesignation: "B-52",
		TypeDesignation:     "B-52H",
		OfficialName:        "Stratofortress",
		Nickname:            "Buff",
	},
	{
		ACMIShortName:       "C-101CC",
		Category:            c101Data.Category,
		PlatformDesignation: c101Data.PlatformDesignation,
		TypeDesignation:     "C-101CC",
		OfficialName:        c101Data.OfficialName,
	},
	{
		ACMIShortName:       "C-101EB",
		Category:            c101Data.Category,
		PlatformDesignation: c101Data.PlatformDesignation,
		TypeDesignation:     "C-101EB",
		OfficialName:        c101Data.OfficialName,
	},
	{
		ACMIShortName:       "C-130",
		Category:            brevity.FixedWing,
		PlatformDesignation: "C-130",
		TypeDesignation:     "C-130",
		OfficialName:        "Hercules",
		Nickname:            "Herc",
	},
	{
		ACMIShortName:       "C-17A",
		Category:            brevity.FixedWing,
		PlatformDesignation: "C-17",
		TypeDesignation:     "C-17A",
		OfficialName:        "Globemaster",
	},
	{
		ACMIShortName:       "C-47",
		Category:            brevity.FixedWing,
		PlatformDesignation: "C-47",
		OfficialName:        "Skytrain",
	},
	{
		ACMIShortName:       "E-2C",
		Category:            brevity.FixedWing,
		PlatformDesignation: "E-2",
		TypeDesignation:     "E-2C",
		OfficialName:        "Hawkeye",
	},
	{
		ACMIShortName:       "E-3A",
		Category:            brevity.FixedWing,
		PlatformDesignation: "E-3",
		TypeDesignation:     "E-3A",
		OfficialName:        "Sentry",
	},
	{
		ACMIShortName:       "F-86F",
		Category:            brevity.FixedWing,
		PlatformDesignation: "F-86",
		TypeDesignation:     "F-86F",
		OfficialName:        "Sabre",
	},
	{
		ACMIShortName:       "F-117A",
		Category:            brevity.FixedWing,
		PlatformDesignation: "F-117",
		TypeDesignation:     "F-117A",
		OfficialName:        "Nighthawk",
		Nickname:            "Goblin",
	},
	{
		ACMIShortName:       "F-4E-45MC",
		Category:            f4Data.Category,
		PlatformDesignation: f4Data.PlatformDesignation,
		TypeDesignation:     "F-4E",
		OfficialName:        f4Data.OfficialName,
	},
	{
		ACMIShortName:       "F-4E",
		Category:            f4Data.Category,
		PlatformDesignation: f4Data.PlatformDesignation,
		TypeDesignation:     "F-4E",
		OfficialName:        f4Data.OfficialName,
	},
	{
		ACMIShortName:       "F-5E",
		Category:            f5Data.Category,
		PlatformDesignation: f5Data.PlatformDesignation,
		TypeDesignation:     "F-5E",
		OfficialName:        f5Data.OfficialName,
	},
	{
		ACMIShortName:       "F-5E-3",
		Category:            f5Data.Category,
		PlatformDesignation: f5Data.PlatformDesignation,
		TypeDesignation:     "F-5E",
		OfficialName:        f5Data.OfficialName,
	},
	{
		ACMIShortName:       "F-14A-135-GR",
		Category:            f14Data.Category,
		PlatformDesignation: f14Data.PlatformDesignation,
		TypeDesignation:     "F-14A",
		OfficialName:        f14Data.OfficialName,
	},
	{
		ACMIShortName:       "F-14A",
		Category:            f14Data.Category,
		PlatformDesignation: f14Data.PlatformDesignation,
		TypeDesignation:     "F-14A",
		OfficialName:        f14Data.OfficialName,
	},
	{
		ACMIShortName:       "F-14B",
		Category:            f14Data.Category,
		PlatformDesignation: f14Data.PlatformDesignation,
		TypeDesignation:     "F-14B",
		OfficialName:        f14Data.OfficialName,
	},
	{
		ACMIShortName:       "F-15C",
		Category:            f15Data.Category,
		PlatformDesignation: f15Data.PlatformDesignation,
		TypeDesignation:     "F-15C",
		OfficialName:        f15Data.OfficialName,
	},
	{
		ACMIShortName:       "F-15E",
		Category:            f15Data.Category,
		PlatformDesignation: f15Data.PlatformDesignation,
		TypeDesignation:     "F-15E",
		OfficialName:        "Strike Eagle",
		Nickname:            "Mudhen",
	},
	{
		ACMIShortName:       "F-16A",
		Category:            f16Data.Category,
		PlatformDesignation: f16Data.PlatformDesignation,
		TypeDesignation:     "F-16A",
		OfficialName:        f16Data.OfficialName,
		Nickname:            f16Data.Nickname,
	},
	{
		ACMIShortName:       "F-16C",
		Category:            f16Data.Category,
		PlatformDesignation: f16Data.PlatformDesignation,
		TypeDesignation:     "F-16C",
		OfficialName:        f16Data.OfficialName,
		Nickname:            f16Data.Nickname,
	},
	{
		ACMIShortName:       "F-16C_50",
		Category:            f16Data.Category,
		PlatformDesignation: f16Data.PlatformDesignation,
		TypeDesignation:     "F-16C",
		OfficialName:        f16Data.OfficialName,
		Nickname:            f16Data.Nickname,
	},
	{
		ACMIShortName:       "F/A-18A",
		Category:            fa18Data.Category,
		PlatformDesignation: fa18Data.PlatformDesignation,
		TypeDesignation:     "F/A-18A",
		OfficialName:        fa18Data.OfficialName,
	},
	{
		ACMIShortName:       "F/A-18C",
		Category:            fa18Data.Category,
		PlatformDesignation: fa18Data.PlatformDesignation,
		TypeDesignation:     "F/A-18C",
		OfficialName:        fa18Data.OfficialName,
	},
	{
		ACMIShortName:       "FA-18C_hornet",
		Category:            fa18Data.Category,
		PlatformDesignation: fa18Data.PlatformDesignation,
		TypeDesignation:     "F/A-18C",
		OfficialName:        fa18Data.OfficialName,
	},
	{
		ACMIShortName:       "H-6J",
		Category:            brevity.FixedWing,
		PlatformDesignation: "Tu-16",
		TypeDesignation:     "H-6J",
		NATOReportingName:   "Badger",
	},
	{
		ACMIShortName:       "IL-76MD",
		Category:            brevity.FixedWing,
		PlatformDesignation: "Il-76",
		TypeDesignation:     "Il-76MD",
		NATOReportingName:   "Candid",
	},
	{
		ACMIShortName:       "IL-78M",
		Category:            brevity.FixedWing,
		PlatformDesignation: "Il-78",
		TypeDesignation:     "Il-78M",
		NATOReportingName:   "Midas",
	},
	{
		ACMIShortName:       "J-11A",
		Category:            flankerData.Category,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "J-11A",
		NATOReportingName:   flankerData.NATOReportingName,
	},
	{
		ACMIShortName:       "JF-17",
		Category:            brevity.FixedWing,
		PlatformDesignation: "JF-17",
		TypeDesignation:     "JF-17",
		OfficialName:        "Thunder",
	},
	{
		ACMIShortName:       "KC-130",
		Category:            brevity.FixedWing,
		PlatformDesignation: "KC-130",
		TypeDesignation:     "KC-130",
		OfficialName:        "Hercules",
		Nickname:            "Herc",
	},
	{
		ACMIShortName:       "KC-135",
		Category:            brevity.FixedWing,
		PlatformDesignation: "KC-135",
		TypeDesignation:     "KC-135",
		OfficialName:        "Stratotanker",
	},
	{
		ACMIShortName:       "KJ-2000",
		Category:            brevity.FixedWing,
		PlatformDesignation: "KJ-2000",
		TypeDesignation:     "KJ-2000",
		OfficialName:        "Mainring",
	},
	{
		ACMIShortName:       "L-39C",
		Category:            brevity.FixedWing,
		PlatformDesignation: l39Data.PlatformDesignation,
		TypeDesignation:     "L-39C",
		OfficialName:        l39Data.OfficialName,
	},
	{
		ACMIShortName:       "L-39ZA",
		Category:            l39Data.Category,
		PlatformDesignation: l39Data.PlatformDesignation,
		TypeDesignation:     "L-39ZA",
		OfficialName:        l39Data.OfficialName,
	},
	{
		ACMIShortName:       "M2000C",
		Category:            brevity.FixedWing,
		PlatformDesignation: "Mirage 2000",
		TypeDesignation:     "Mirage 2000C",
		OfficialName:        "Mirage 2000",
	},
	{
		ACMIShortName:       "MB-339A",
		Category:            mb339Data.Category,
		PlatformDesignation: mb339Data.PlatformDesignation,
		TypeDesignation:     "MB-339A",
	},
	{
		ACMIShortName:       "MB-339APAN",
		Category:            mb339Data.Category,
		PlatformDesignation: mb339Data.PlatformDesignation,
		TypeDesignation:     "MB-339A",
	},
	{
		ACMIShortName:       "Mi-24P",
		Category:            brevity.RotaryWing,
		PlatformDesignation: "Mi-24",
		TypeDesignation:     "Mi-24P",
		NATOReportingName:   "Hind",
	},
	{
		ACMIShortName: "Mi-26",
		Category:      brevity.RotaryWing,
		PlatformDesignation: "Mi-26",
		TypeDesignation:     "Mi-26",
		NATOReportingName:   "Hip",
	},
	{
		ACMIShortName:       "MiG-15bis",
		Category:            brevity.FixedWing,
		PlatformDesignation: "MiG-15",
		TypeDesignation:     "MiG-15bis",
		NATOReportingName:   mig15NATOReportingName,
	},
	{
		ACMIShortName:       "MiG-19P",
		Category:            brevity.FixedWing,
		PlatformDesignation: "MiG-19",
		TypeDesignation:     "MiG-19P",
		NATOReportingName:   "Farmer",
	},
	{
		ACMIShortName:       "MiG-21Bis",
		Category:            brevity.FixedWing,
		PlatformDesignation: "MiG-21",
		TypeDesignation:     "MiG-21Bis",
		NATOReportingName:   "Fishbed",
	},
	{
		ACMIShortName:       "MiG-23MLD",
		Category:            brevity.FixedWing,
		PlatformDesignation: "MiG-23",
		TypeDesignation:     "MiG-23MLD",
		NATOReportingName:   "Flogger",
	},
	{
		ACMIShortName:       "MiG-25PD",
		Category:            foxbatData.Category,
		PlatformDesignation: foxbatData.PlatformDesignation,
		TypeDesignation:     "MiG-25PD",
		NATOReportingName:   foxbatData.NATOReportingName,
	},
	{
		ACMIShortName:       "MiG-25RBT",
		Category:            foxbatData.Category,
		PlatformDesignation: foxbatData.PlatformDesignation,
		TypeDesignation:     "MiG-25RBT",
		NATOReportingName:   foxbatData.NATOReportingName,
	},
	{
		ACMIShortName:       "MiG-27K",
		Category:            brevity.FixedWing,
		PlatformDesignation: "MiG-27",
		TypeDesignation:     "MiG-27K",
		NATOReportingName:   "Flogger",
	},
	{
		ACMIShortName:       "MiG-29A",
		Category:            fulcrumData.Category,
		PlatformDesignation: fulcrumData.PlatformDesignation,
		TypeDesignation:     "MiG-29A",
		NATOReportingName:   fulcrumData.NATOReportingName,
	},
	{
		ACMIShortName:       "MiG-29S",
		Category:            fulcrumData.Category,
		PlatformDesignation: fulcrumData.PlatformDesignation,
		TypeDesignation:     "MiG-29S",
		NATOReportingName:   fulcrumData.NATOReportingName,
	},
	{
		ACMIShortName:       "MiG-31",
		Category:            brevity.FixedWing,
		PlatformDesignation: "MiG-31",
		TypeDesignation:     "MiG-31",
		NATOReportingName:   "Foxhound",
	},
	{
		ACMIShortName:       "M2000-5",
		Category:            brevity.FixedWing,
		PlatformDesignation: "Mirage 2000",
		TypeDesignation:     "Mirage 2000-5",
		OfficialName:        "Mirage 2000",
	},
	{
		ACMIShortName:       "MQ-1",
		Category:            brevity.FixedWing,
		PlatformDesignation: "MQ-1",
		TypeDesignation:     "MQ-1A",
		OfficialName:        "Predator",
	},
	{
		ACMIShortName:       "MQ-9",
		Category:            brevity.FixedWing,
		PlatformDesignation: "MQ-9",
		TypeDesignation:     "MQ-9",
		OfficialName:        "Reaper",
	},
	{
		ACMIShortName:       "S-3B",
		Category:            brevity.FixedWing,
		PlatformDesignation: "S-3",
		TypeDesignation:     "S-3B",
		OfficialName:        "Viking",
	},
	{
		ACMIShortName:       "Su-17M4",
		Category:            brevity.FixedWing,
		PlatformDesignation: "Su-17",
		TypeDesignation:     "Su-17M4",
		NATOReportingName:   "Fitter",
	},
	{
		ACMIShortName:       "Su-24M",
		Category:            fencerData.Category,
		PlatformDesignation: fencerData.PlatformDesignation,
		TypeDesignation:     "Su-24M",
		NATOReportingName:   fencerData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-24MR",
		Category:            fencerData.Category,
		PlatformDesignation: fencerData.PlatformDesignation,
		TypeDesignation:     "Su-24MR",
		NATOReportingName:   fencerData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-25",
		Category:            frogfootData.Category,
		PlatformDesignation: frogfootData.PlatformDesignation,
		TypeDesignation:     "Su-25",
		NATOReportingName:   frogfootData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-25T",
		Category:            frogfootData.Category,
		PlatformDesignation: frogfootData.PlatformDesignation,
		TypeDesignation:     "Su-25T",
		NATOReportingName:   frogfootData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-25TM",
		Category:            frogfootData.Category,
		PlatformDesignation: frogfootData.PlatformDesignation,
		TypeDesignation:     "Su-25TM",
		NATOReportingName:   frogfootData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-27",
		Category:            flankerData.Category,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-27",
		NATOReportingName:   flankerData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-30",
		Category:            flankerData.Category,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-30",
		NATOReportingName:   flankerData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-33",
		Category:            flankerData.Category,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-33",
		NATOReportingName:   flankerData.NATOReportingName,
	},
	{
		ACMIShortName:       "Su-34",
		Category:            brevity.FixedWing,
		PlatformDesignation: "Su-34",
		TypeDesignation:     "Su-34",
		OfficialName:        "Fullback",
	},
	{
		ACMIShortName:       "Tornado",
		Category:            tornadoData.Category,
		PlatformDesignation: tornadoData.PlatformDesignation,
		TypeDesignation:     "Tornado",
		OfficialName:        tornadoData.OfficialName,
	},
	{
		ACMIShortName:       "Tu-22M3",
		Category:            brevity.FixedWing,
		PlatformDesignation: "Tu-22",
		TypeDesignation:     "Tu-22M",
		OfficialName:        "Backfire",
	},
	{
		ACMIShortName:       "Tu-95MS",
		Category:            brevity.FixedWing,
		PlatformDesignation: "Tu-95",
		TypeDesignation:     "Tu-95MS",
		OfficialName:        "Bear",
	},
	{
		ACMIShortName:       "Tu-142",
		Category:            brevity.FixedWing,
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
			Category:            mirageF1Data.Category,
			PlatformDesignation: mirageF1Data.PlatformDesignation,
			TypeDesignation:     "Mirage F1" + v,
			OfficialName:        mirageF1Data.OfficialName,
		})
	}
	return variants
}
