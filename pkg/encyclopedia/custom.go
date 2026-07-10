package encyclopedia

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dharmab/collections/sets"
	"github.com/martinlindhe/unit"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// serializedAircraft is the on-disk representation of an Aircraft. It exists only to translate
// between the human-friendly string schema and the Aircraft type's unexported enum fields; it is
// never exposed outside this package.
type serializedAircraft struct {
	ACMIShortName       string   `yaml:"acmi_short_name"`
	Tags                []string `yaml:"tags"`
	PlatformDesignation string   `yaml:"platform_designation"`
	TypeDesignation     string   `yaml:"type_designation"`
	NATOReportingName   string   `yaml:"nato_reporting_name"`
	OfficialName        string   `yaml:"official_name"`
	Nickname            string   `yaml:"nickname"`
	ThreatRadiusNM      float64  `yaml:"threat_radius_nm"`
	FuelProvider        string   `yaml:"fuel_provider"`
	FuelReceiver        string   `yaml:"fuel_receiver"`
}

// tagsByName maps the tag names accepted in custom aircraft files to their AircraftTag values.
var tagsByName = map[string]AircraftTag{
	"fixed-wing":  FixedWing,
	"rotary-wing": RotaryWing,
	"unarmed":     Unarmed,
	"fighter":     Fighter,
	"attack":      Attack,
}

// refuelingByName maps the refueling method names accepted in custom aircraft files to their
// AirRefuelingMethod values. An empty string or "none" maps to NoAirRefueling.
var refuelingByName = map[string]AirRefuelingMethod{
	"":                 NoAirRefueling,
	"none":             NoAirRefueling,
	"boom":             FlyingBoom,
	"probe-and-drogue": ProbeAndDrogue,
}

// toAircraft validates the string schema and converts it into an Aircraft.
func (s serializedAircraft) toAircraft() (Aircraft, error) {
	if strings.TrimSpace(s.ACMIShortName) == "" {
		return Aircraft{}, errors.New("aircraft must have a non-empty acmi_short_name")
	}

	// SkyEye derives an aircraft's reporting name from these fields, so at least one is required or
	// the GCI would have nothing to call the contact.
	if strings.TrimSpace(s.NATOReportingName) == "" &&
		strings.TrimSpace(s.Nickname) == "" &&
		strings.TrimSpace(s.OfficialName) == "" &&
		strings.TrimSpace(s.PlatformDesignation) == "" {
		return Aircraft{}, fmt.Errorf("aircraft %q must have at least one of nato_reporting_name, nickname, official_name, or platform_designation", s.ACMIShortName)
	}

	if len(s.Tags) == 0 {
		return Aircraft{}, fmt.Errorf("aircraft %q must have at least one tag", s.ACMIShortName)
	}
	tags := sets.New[AircraftTag]()
	wingCount := 0
	for _, name := range s.Tags {
		tag, ok := tagsByName[strings.ToLower(strings.TrimSpace(name))]
		if !ok {
			return Aircraft{}, fmt.Errorf("aircraft %q has unrecognized tag %q", s.ACMIShortName, name)
		}
		if tag == FixedWing || tag == RotaryWing {
			wingCount++
		}
		sets.Add(tags, tag)
	}
	if wingCount != 1 {
		return Aircraft{}, fmt.Errorf("aircraft %q must have exactly one of the tags fixed-wing or rotary-wing", s.ACMIShortName)
	}

	if s.ThreatRadiusNM < 0 {
		return Aircraft{}, fmt.Errorf("aircraft %q has negative threat_radius_nm %v", s.ACMIShortName, s.ThreatRadiusNM)
	}

	fuelProvider, ok := refuelingByName[strings.ToLower(strings.TrimSpace(s.FuelProvider))]
	if !ok {
		return Aircraft{}, fmt.Errorf("aircraft %q has unrecognized fuel_provider %q", s.ACMIShortName, s.FuelProvider)
	}
	fuelReceiver, ok := refuelingByName[strings.ToLower(strings.TrimSpace(s.FuelReceiver))]
	if !ok {
		return Aircraft{}, fmt.Errorf("aircraft %q has unrecognized fuel_receiver %q", s.ACMIShortName, s.FuelReceiver)
	}

	return Aircraft{
		ACMIShortName:       s.ACMIShortName,
		tags:                tags,
		PlatformDesignation: s.PlatformDesignation,
		TypeDesignation:     s.TypeDesignation,
		NATOReportingName:   s.NATOReportingName,
		OfficialName:        s.OfficialName,
		Nickname:            s.Nickname,
		threatRadius:        unit.Length(s.ThreatRadiusNM) * unit.NauticalMile,
		fuelProvider:        fuelProvider,
		fuelReceiver:        fuelReceiver,
	}, nil
}

// LoadCustomAircraft parses custom aircraft data. YAML is a superset of JSON, so the data may be in
// either format. Each entry is validated during conversion.
func LoadCustomAircraft(data []byte) ([]Aircraft, error) {
	var serialized []serializedAircraft
	if err := yaml.Unmarshal(data, &serialized); err != nil {
		return nil, fmt.Errorf("failed to parse custom aircraft file: %w", err)
	}
	aircraft := make([]Aircraft, 0, len(serialized))
	for _, s := range serialized {
		a, err := s.toAircraft()
		if err != nil {
			return nil, err
		}
		aircraft = append(aircraft, a)
	}
	return aircraft, nil
}

// AddCustomAircraft registers custom aircraft into the encyclopedia, keyed by ACMI short name.
// An entry whose ACMI short name matches a built-in or previously added entry overrides it.
func AddCustomAircraft(aircraft []Aircraft) {
	for _, data := range aircraft {
		event := log.Info().Str("aircraft", data.ACMIShortName)
		if _, exists := aircraftDataLUT[data.ACMIShortName]; exists {
			event = event.Bool("override", true)
		}
		aircraftDataLUT[data.ACMIShortName] = data
		event.Msg("loaded custom aircraft into encyclopedia")
	}
}
