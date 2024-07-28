package encyclopedia

import "github.com/martinlindhe/unit"

// ThreatClass is a broad classification of how dangerous an aircraft is based on the air-to-air weapons it may carry
type ThreatClass int

// BVR timeline sourced from CVW-8's testing of the DCS World Hornet

const (
	// UnspecifiedThreat is the default threat class, used for aircraft not in the encyclopedia
	Unspecified ThreatClass = iota
	// NoFactor means the aircraft is unarmed
	NoFactor
	// Guns means guns only. Only guns useful against fixed-wing aircraft are considered
	Guns
	// SAR1OrIR means Semi-Active Radar (class 1) or Infrared missiles
	SAR1OrIR
	// SAR2OrAR1 ismeans Semi-Active Radar (class 2) or Active Radar missiles
	SAR2OrAR1
)

// NoLaterThanCommitRange (NLT COMMIT) is the minimum range to accept a commit and still be on timeline
func NoLaterThanCommitRange(tc ThreatClass) unit.Length {
	if tc == SAR2OrAR1 {
		return unit.Length(65) * unit.NauticalMile
	}
	return unit.Length(55) * unit.NauticalMile
}

// TacticalRange (TAC RANGE) is an informative range 30 seconds prior to meld
func TacticalRange(tc ThreatClass) unit.Length {
	if tc == SAR2OrAR1 {
		return unit.Length(55) * unit.NauticalMile
	}
	return unit.Length(42) * unit.NauticalMile
}

// MeldAndSortRange (MELD & SORT) is the range at which meld and sort should commence
func MeldAndSortRange(tc ThreatClass) unit.Length {
	if tc == SAR2OrAR1 {
		return unit.Length(45) * unit.NauticalMile
	}
	return unit.Length(32) * unit.NauticalMile
}

// NoLaterThanShotRange (NLT SR) is the range by which the first shot must be off the rail
func NoLaterThanShotRange(tc ThreatClass) unit.Length {
	if tc == SAR2OrAR1 {
		return unit.Length(35) * unit.NauticalMile
	}
	return unit.Length(25) * unit.NauticalMile
}

// MinimumOutRange (MOR) is the range where the out maneuver will defeat any weapons in flight or still on the jet and preserve enough distance to re-attack the same group
func MinimumOutRange(tc ThreatClass) unit.Length {
	if tc == SAR2OrAR1 {
		return unit.Length(23) * unit.NauticalMile
	}
	return unit.Length(13) * unit.NauticalMile
}

// MinimumReattackRange (MRAR) is the range where a dragging fighter may turn in for a re-attack, support a missile to active and still abort within NLT MAR
func MinimumReattackRange(tc ThreatClass) unit.Length {
	if tc == SAR2OrAR1 {
		return unit.Length(16) * unit.NauticalMile
	}
	return unit.Length(11) * unit.NauticalMile
}

// LastShotRange (LSR) is range where a fighter may support a missile to active and still abort within NLT MAR
func LastShotRange(tc ThreatClass) unit.Length {
	if tc == SAR2OrAR1 {
		return unit.Length(14) * unit.NauticalMile
	}
	return unit.Length(9) * unit.NauticalMile
}

// MinimumAbortRange (MAR) is the range where a an abort maneuver will defeat any adversary weapons and the fighter will momentarily remain outside the adversary WEZ
func MinimumAbortRange(tc ThreatClass) unit.Length {
	if tc == SAR2OrAR1 {
		return unit.Length(12) * unit.NauticalMile
	}
	return unit.Length(8) * unit.NauticalMile
}

// SternWeaponEngagementZone (STERN WEZ) is the range where an adversary's missiles cannot be kinematically defeated
func SternWeaponEngagementZone(tc ThreatClass) unit.Length {
	if tc == SAR2OrAR1 {
		return unit.Length(7) * unit.NauticalMile
	}
	return unit.Length(4) * unit.NauticalMile
}

// OffensiveFlowRange (OFW) is the range where a fighter may clean up a merge and still be on MELD timeline for a follow-on group
func OffensiveFlowRange(tc ThreatClass) unit.Length {
	if tc == SAR2OrAR1 {
		return unit.Length(42) * unit.NauticalMile
	}
	return unit.Length(32) * unit.NauticalMile
}

// DefensiveFlowRange (DFW) is the range where a fighter may clean up a merge and defend against follow-on group
func DefensiveFlowRange(tc ThreatClass) unit.Length {
	if tc == SAR2OrAR1 {
		return unit.Length(28) * unit.NauticalMile
	}
	return unit.Length(18) * unit.NauticalMile
}
