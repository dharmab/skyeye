package encyclopedia

import "github.com/dharmab/skyeye/pkg/brevity"

type Aircraft struct {
	EditorType          string
	Category            brevity.ContactCategory
	PlatformDesignation string
	TypeDesignation     string
	NATOReportingName   string
	OfficialName        string
	Nickname            string
}

var a10Data = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "A-10",
	OfficialName:        "Thunderbolt",
	Nickname:            "Warthog",
}

var c101Data = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "C-101",
	OfficialName:        "Aviojet",
}

var f5Data = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "F-5",
	OfficialName:        "Tiger",
}

var f14Data = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "F-14",
	OfficialName:        "Tomcat",
}

var f15Data = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "F-15",
	// Use "Eagle" for Strike Eagle because radar cannot distinguish between the two
	OfficialName: "Eagle",
}

var f16Data = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "F-16",
	OfficialName:        "Falcon",
	Nickname:            "Viper",
}

var fa18Data = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "F/A-18",
	OfficialName:        "Hornet",
}

var mirageF1Data = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "Mirage F1",
	OfficialName:        "Mirage F1",
}

var fencerData = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "Su-24",
	NATOReportingName:   "Fencer",
}

var foxbatData = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "MiG-25",
	NATOReportingName:   "Foxbat",
}

var fulcrumData = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "MiG-29",
	NATOReportingName:   "Fulcrum",
}

var frogfootData = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "Su-25",
	NATOReportingName:   "Frogfoot",
}

var flankerData = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "Su-27",
	NATOReportingName:   "Flanker",
}

var l39Data = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "L-39",
	OfficialName:        "Albatros",
}

var mb339Data = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "MB-339",
}

var tornadoData = Aircraft{
	Category:            brevity.Airplanes,
	PlatformDesignation: "Tornado",
	OfficialName:        "Tornado",
}

var aircraftData = append([]Aircraft{
	{
		EditorType:          "A-10A",
		Category:            a10Data.Category,
		PlatformDesignation: a10Data.PlatformDesignation,
		TypeDesignation:     "A-10A",
		OfficialName:        a10Data.OfficialName,
		Nickname:            a10Data.Nickname,
	},
	{
		EditorType:          "A-10C",
		Category:            a10Data.Category,
		PlatformDesignation: a10Data.PlatformDesignation,
		TypeDesignation:     "A-10C",
		OfficialName:        a10Data.OfficialName,
		Nickname:            a10Data.Nickname,
	},
	{
		EditorType:          "A-10C_2",
		Category:            a10Data.Category,
		PlatformDesignation: a10Data.PlatformDesignation,
		TypeDesignation:     "A-10C",
		OfficialName:        a10Data.OfficialName,
		Nickname:            a10Data.Nickname,
	},
	{
		EditorType:          "A-4E-C",
		Category:            brevity.Airplanes,
		PlatformDesignation: "A-4",
		TypeDesignation:     "A-4E",
		OfficialName:        "Skyhawk",
		Nickname:            "Scooter",
	},
	{
		EditorType:          "A-50",
		Category:            brevity.Airplanes,
		PlatformDesignation: "A-50",
		TypeDesignation:     "A-50",
		NATOReportingName:   "Mainstay",
	},
	{
		EditorType:          "AJS37",
		Category:            brevity.Airplanes,
		PlatformDesignation: "AJS37",
		OfficialName:        "Viggen",
	},
	{
		EditorType:          "AV8BNA",
		Category:            brevity.Airplanes,
		PlatformDesignation: "AV-8",
		TypeDesignation:     "AV-8B",
		OfficialName:        "Harrier",
	},
	{
		EditorType:          "B-1B",
		Category:            brevity.Airplanes,
		PlatformDesignation: "B-1",
		TypeDesignation:     "B-1B",
		OfficialName:        "Lancer",
		Nickname:            "Bone",
	},
	{
		EditorType:          "B-52H",
		Category:            brevity.Airplanes,
		PlatformDesignation: "B-52",
		TypeDesignation:     "B-52H",
		OfficialName:        "Stratofortress",
		Nickname:            "Buff",
	},
	{
		EditorType:          "C-101CC",
		Category:            c101Data.Category,
		PlatformDesignation: c101Data.PlatformDesignation,
		TypeDesignation:     "C-101CC",
		OfficialName:        c101Data.OfficialName,
	},
	{
		EditorType:          "C-101EB",
		Category:            c101Data.Category,
		PlatformDesignation: c101Data.PlatformDesignation,
		TypeDesignation:     "C-101EB",
		OfficialName:        c101Data.OfficialName,
	},
	{
		EditorType:          "C-130",
		Category:            brevity.Airplanes,
		PlatformDesignation: "C-130",
		TypeDesignation:     "C-130",
		OfficialName:        "Hercules",
		Nickname:            "Herc",
	},
	{
		EditorType:          "C-17A",
		Category:            brevity.Airplanes,
		PlatformDesignation: "C-17",
		TypeDesignation:     "C-17A",
		OfficialName:        "Globemaster",
	},
	{
		EditorType:          "C-47",
		Category:            brevity.Airplanes,
		PlatformDesignation: "C-47",
		OfficialName:        "Skytrain",
	},
	{
		EditorType:          "E-2C",
		Category:            brevity.Airplanes,
		PlatformDesignation: "E-2",
		TypeDesignation:     "E-2C",
		OfficialName:        "Hawkeye",
	},
	{
		EditorType:          "E-3A",
		Category:            brevity.Airplanes,
		PlatformDesignation: "E-3",
		TypeDesignation:     "E-3A",
		OfficialName:        "Sentry",
	},
	{
		EditorType:          "F-86F Sabre",
		Category:            brevity.Airplanes,
		PlatformDesignation: "F-86",
		TypeDesignation:     "F-86F",
		OfficialName:        "Sabre",
	},
	{
		EditorType:          "F-117A",
		Category:            brevity.Airplanes,
		PlatformDesignation: "F-117",
		TypeDesignation:     "F-117A",
		OfficialName:        "Nighthawk",
		Nickname:            "Goblin",
	},
	{
		EditorType:          "F-4E",
		Category:            brevity.Airplanes,
		PlatformDesignation: "F-4",
		TypeDesignation:     "F-4E",
		OfficialName:        "Phantom",
	},
	{
		EditorType:          "F-5E",
		Category:            f5Data.Category,
		PlatformDesignation: f5Data.PlatformDesignation,
		TypeDesignation:     "F-5E",
		OfficialName:        f5Data.OfficialName,
	},
	{
		EditorType:          "F-5E-3",
		Category:            f5Data.Category,
		PlatformDesignation: f5Data.PlatformDesignation,
		TypeDesignation:     "F-5E",
		OfficialName:        f5Data.OfficialName,
	},
	{
		EditorType:          "F-14A-135-GR",
		Category:            f14Data.Category,
		PlatformDesignation: f14Data.PlatformDesignation,
		TypeDesignation:     "F-14A",
		OfficialName:        f14Data.OfficialName,
	},
	{
		EditorType:          "F-14A",
		Category:            f14Data.Category,
		PlatformDesignation: f14Data.PlatformDesignation,
		TypeDesignation:     "F-14A",
		OfficialName:        f14Data.OfficialName,
	},
	{
		EditorType:          "F-14B",
		Category:            f14Data.Category,
		PlatformDesignation: f14Data.PlatformDesignation,
		TypeDesignation:     "F-14B",
		OfficialName:        f14Data.OfficialName,
	},
	{
		EditorType:          "F-15C",
		Category:            f15Data.Category,
		PlatformDesignation: f15Data.PlatformDesignation,
		TypeDesignation:     "F-15C",
		OfficialName:        f15Data.OfficialName,
	},
	{
		EditorType:          "F-15E",
		Category:            f15Data.Category,
		PlatformDesignation: f15Data.PlatformDesignation,
		TypeDesignation:     "F-15E",
		OfficialName:        f15Data.OfficialName,
	},
	{
		EditorType:          "F-15ESE",
		Category:            f15Data.Category,
		PlatformDesignation: f15Data.PlatformDesignation,
		TypeDesignation:     "F-15E",
		OfficialName:        f15Data.OfficialName,
	},
	{
		EditorType:          "F-16A",
		Category:            f16Data.Category,
		PlatformDesignation: f16Data.PlatformDesignation,
		TypeDesignation:     "F-16A",
		OfficialName:        f16Data.OfficialName,
		Nickname:            f16Data.Nickname,
	},
	{
		EditorType:          "F-16A MLU",
		Category:            f16Data.Category,
		PlatformDesignation: f16Data.PlatformDesignation,
		TypeDesignation:     "F-16A",
		OfficialName:        f16Data.OfficialName,
		Nickname:            f16Data.Nickname,
	},
	{
		EditorType:          "F-16C_50",
		Category:            f16Data.Category,
		PlatformDesignation: f16Data.PlatformDesignation,
		TypeDesignation:     "F-16C",
		OfficialName:        f16Data.OfficialName,
		Nickname:            f16Data.Nickname,
	},
	{
		EditorType:          "F-16C bl.50",
		Category:            f16Data.Category,
		PlatformDesignation: f16Data.PlatformDesignation,
		TypeDesignation:     "F-16C",
		OfficialName:        f16Data.OfficialName,
		Nickname:            f16Data.Nickname,
	},
	{
		EditorType:          "F-16C bl.52d",
		Category:            f16Data.Category,
		PlatformDesignation: f16Data.PlatformDesignation,
		TypeDesignation:     "F-16C",
		OfficialName:        f16Data.OfficialName,
		Nickname:            f16Data.Nickname,
	},
	{
		EditorType:          "FA-18A",
		Category:            fa18Data.Category,
		PlatformDesignation: fa18Data.PlatformDesignation,
		TypeDesignation:     "F/A-18A",
		OfficialName:        fa18Data.OfficialName,
	},
	{
		EditorType:          "FA-18C",
		Category:            fa18Data.Category,
		PlatformDesignation: fa18Data.PlatformDesignation,
		TypeDesignation:     "F/A-18C",
		OfficialName:        fa18Data.OfficialName,
	},
	{
		EditorType:          "FA-18C_hornet",
		Category:            fa18Data.Category,
		PlatformDesignation: fa18Data.PlatformDesignation,
		TypeDesignation:     "F/A-18C",
		OfficialName:        fa18Data.OfficialName,
	},
	{
		EditorType:          "H-6J",
		Category:            brevity.Airplanes,
		PlatformDesignation: "Tu-16",
		TypeDesignation:     "H-6J",
		NATOReportingName:   "Badger",
	},
	{
		EditorType:          "IL-76MD",
		Category:            brevity.Airplanes,
		PlatformDesignation: "Il-76",
		TypeDesignation:     "Il-76MD",
		NATOReportingName:   "Candid",
	},
	{
		EditorType:          "IL-78M",
		Category:            brevity.Airplanes,
		PlatformDesignation: "Il-78",
		TypeDesignation:     "Il-78M",
		NATOReportingName:   "Midas",
	},
	{
		EditorType:          "J-11A",
		Category:            flankerData.Category,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "J-11A",
		NATOReportingName:   flankerData.NATOReportingName,
	},
	{
		EditorType:          "JF-17",
		Category:            brevity.Airplanes,
		PlatformDesignation: "JF-17",
		TypeDesignation:     "JF-17",
		OfficialName:        "Thunder",
	},
	{
		EditorType:          "KC-130",
		Category:            brevity.Airplanes,
		PlatformDesignation: "KC-130",
		TypeDesignation:     "KC-130",
		OfficialName:        "Hercules",
		Nickname:            "Herc",
	},
	{
		EditorType:          "KC135MPRS",
		Category:            brevity.Airplanes,
		PlatformDesignation: "KC-135",
		TypeDesignation:     "KC-135",
		OfficialName:        "Stratotanker",
	},
	{
		EditorType:          "KJ-2000",
		Category:            brevity.Airplanes,
		PlatformDesignation: "KJ-2000",
		TypeDesignation:     "KJ-2000",
		OfficialName:        "Mainring",
	},
	{
		EditorType:          "L-39C",
		Category:            brevity.Airplanes,
		PlatformDesignation: l39Data.PlatformDesignation,
		TypeDesignation:     "L-39C",
		OfficialName:        l39Data.OfficialName,
	},
	{
		EditorType:          "L-39ZA",
		Category:            l39Data.Category,
		PlatformDesignation: l39Data.PlatformDesignation,
		TypeDesignation:     "L-39ZA",
		OfficialName:        l39Data.OfficialName,
	},
	{
		EditorType:          "M-2000C",
		Category:            brevity.Airplanes,
		PlatformDesignation: "Mirage 2000",
		TypeDesignation:     "Mirage 2000C",
		OfficialName:        "Mirage 2000",
	},
	{
		EditorType:          "MB-339A",
		Category:            mb339Data.Category,
		PlatformDesignation: mb339Data.PlatformDesignation,
		TypeDesignation:     "MB-339A",
	},
	{
		EditorType:          "MB-339APAN",
		Category:            mb339Data.Category,
		PlatformDesignation: mb339Data.PlatformDesignation,
		TypeDesignation:     "MB-339A",
	},
	{
		EditorType:          "MQ-9",
		Category:            brevity.Airplanes,
		PlatformDesignation: "MQ-9",
		TypeDesignation:     "MQ-9",
		OfficialName:        "Reaper",
	},
	{
		EditorType:          "MiG-15bis",
		Category:            brevity.Airplanes,
		PlatformDesignation: "MiG-15",
		TypeDesignation:     "MiG-15bis",
		NATOReportingName:   mig15NATOReportingName,
	},
	{
		EditorType:          "MiG-19P",
		Category:            brevity.Airplanes,
		PlatformDesignation: "MiG-19",
		TypeDesignation:     "MiG-19P",
		NATOReportingName:   "Farmer",
	},
	{
		EditorType:          "MiG-23MLB",
		Category:            brevity.Airplanes,
		PlatformDesignation: "MiG-23",
		TypeDesignation:     "MiG-23MLB",
		NATOReportingName:   "Flogger",
	},
	{
		EditorType:          "MiG-25PD",
		Category:            foxbatData.Category,
		PlatformDesignation: foxbatData.PlatformDesignation,
		TypeDesignation:     "MiG-25PD",
		NATOReportingName:   foxbatData.NATOReportingName,
	},
	{
		EditorType:          "MiG-25RBT",
		Category:            foxbatData.Category,
		PlatformDesignation: foxbatData.PlatformDesignation,
		TypeDesignation:     "MiG-25RBT",
		NATOReportingName:   foxbatData.NATOReportingName,
	},
	{
		EditorType:          "MiG-27K",
		Category:            brevity.Airplanes,
		PlatformDesignation: "MiG-27",
		TypeDesignation:     "MiG-27K",
		NATOReportingName:   "Flogger",
	},
	{
		EditorType:          "MiG-29A",
		Category:            fulcrumData.Category,
		PlatformDesignation: fulcrumData.PlatformDesignation,
		TypeDesignation:     "MiG-29A",
		NATOReportingName:   fulcrumData.NATOReportingName,
	},
	{
		EditorType:          "MiG-29G",
		Category:            fulcrumData.Category,
		PlatformDesignation: fulcrumData.PlatformDesignation,
		TypeDesignation:     "MiG-29G",
		NATOReportingName:   fulcrumData.NATOReportingName,
	},
	{
		EditorType:          "MiG-29S",
		Category:            fulcrumData.Category,
		PlatformDesignation: fulcrumData.PlatformDesignation,
		TypeDesignation:     "MiG-29S",
		NATOReportingName:   fulcrumData.NATOReportingName,
	},
	{
		EditorType:          "MiG-31",
		Category:            brevity.Airplanes,
		PlatformDesignation: "MiG-31",
		TypeDesignation:     "MiG-31",
		NATOReportingName:   "Foxhound",
	},
	{
		EditorType:          "Mirage 2000-5",
		Category:            brevity.Airplanes,
		PlatformDesignation: "Mirage 2000",
		TypeDesignation:     "Mirage 2000-5",
		OfficialName:        "Mirage 2000",
	},
	{
		EditorType:          "RQ-1A",
		Category:            brevity.Airplanes,
		PlatformDesignation: "RQ-1",
		TypeDesignation:     "RQ-1A",
		OfficialName:        "Predator",
	},
	{
		EditorType:          "S-3B",
		Category:            brevity.Airplanes,
		PlatformDesignation: "S-3",
		TypeDesignation:     "S-3B",
		OfficialName:        "Viking",
	},
	{
		EditorType:          "S-3B Tanker",
		Category:            brevity.Airplanes,
		PlatformDesignation: "S-3",
		TypeDesignation:     "S-3B",
		OfficialName:        "Viking",
	},
	{
		EditorType:          "Su-17M4",
		Category:            brevity.Airplanes,
		PlatformDesignation: "Su-17",
		TypeDesignation:     "Su-17M4",
		NATOReportingName:   "Fitter",
	},
	{
		EditorType:          "Su-24M",
		Category:            fencerData.Category,
		PlatformDesignation: fencerData.PlatformDesignation,
		TypeDesignation:     "Su-24M",
		NATOReportingName:   fencerData.NATOReportingName,
	},
	{
		EditorType:          "Su-24MR",
		Category:            fencerData.Category,
		PlatformDesignation: fencerData.PlatformDesignation,
		TypeDesignation:     "Su-24MR",
		NATOReportingName:   fencerData.NATOReportingName,
	},
	{
		EditorType:          "Su-25T",
		Category:            frogfootData.Category,
		PlatformDesignation: frogfootData.PlatformDesignation,
		TypeDesignation:     "Su-25T",
		NATOReportingName:   frogfootData.NATOReportingName,
	},
	{
		EditorType:          "Su-25TM",
		Category:            frogfootData.Category,
		PlatformDesignation: frogfootData.PlatformDesignation,
		TypeDesignation:     "Su-25TM",
		NATOReportingName:   frogfootData.NATOReportingName,
	},
	{
		EditorType:          "Su-27",
		Category:            flankerData.Category,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-27",
		NATOReportingName:   flankerData.NATOReportingName,
	},
	{
		EditorType:          "Su-30",
		Category:            flankerData.Category,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-30",
		NATOReportingName:   flankerData.NATOReportingName,
	},
	{
		EditorType:          "Su-33",
		Category:            flankerData.Category,
		PlatformDesignation: flankerData.PlatformDesignation,
		TypeDesignation:     "Su-33",
		NATOReportingName:   flankerData.NATOReportingName,
	},
	{
		EditorType:          "Su-34",
		Category:            brevity.Airplanes,
		PlatformDesignation: "Su-34",
		TypeDesignation:     "Su-34",
		OfficialName:        "Fullback",
	},
	{
		EditorType:          "Tornado GR4",
		Category:            tornadoData.Category,
		PlatformDesignation: tornadoData.PlatformDesignation,
		TypeDesignation:     "Tornado GR4",
		OfficialName:        tornadoData.OfficialName,
	},
	{
		EditorType:          "Tornado IDS",
		Category:            tornadoData.Category,
		PlatformDesignation: tornadoData.PlatformDesignation,
		TypeDesignation:     "Tornado IDS",
		OfficialName:        tornadoData.OfficialName,
	},

	{
		EditorType:          "Tu-22M3",
		Category:            brevity.Airplanes,
		PlatformDesignation: "Tu-22",
		TypeDesignation:     "Tu-22M",
		OfficialName:        "Backfire",
	},
	{

		EditorType:          "Tu-95MS",
		Category:            brevity.Airplanes,
		PlatformDesignation: "Tu-95",
		TypeDesignation:     "Tu-95MS",
		OfficialName:        "Bear",
	},
	{
		EditorType:          "Tu-142",
		Category:            brevity.Airplanes,
		PlatformDesignation: "Tu-142",
		TypeDesignation:     "Tu-142",
		OfficialName:        "Bear",
	},
	{
		EditorType:          "Tu-160",
		Category:            brevity.Airplanes,
		PlatformDesignation: "Tu-160",
		TypeDesignation:     "Tu-160",
		OfficialName:        "Blackjack",
	},
}, mirageF1Variants()...)

func mirageF1Variants() []Aircraft {
	variants := []Aircraft{}
	for _, v := range []string{"B", "BD", "BE", "BQ", "C-200", "C", "CE", "CG", "CH", "CJ", "CK", "CR", "CT", "CZ", "DDA", "ED", "EDA", "EE", "EH", "EQ", "JA", "M-CE", "M-EE"} {
		variants = append(variants, Aircraft{
			EditorType:          "Mirage F1" + v,
			Category:            mirageF1Data.Category,
			PlatformDesignation: mirageF1Data.PlatformDesignation,
			TypeDesignation:     "Mirage F1" + v,
			OfficialName:        mirageF1Data.OfficialName,
		})
	}
	return variants
}
