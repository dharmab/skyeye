package encyclopedia

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const a4YAML = `
- acmi_short_name: A-4E-C
  tags: [fixed-wing, attack]
  platform_designation: A-4
  type_designation: A-4E
  official_name: Skyhawk
  nickname: Scooter
  threat_radius_nm: 15
  fuel_receiver: probe-and-drogue
`

const a4JSON = `[
  {
    "acmi_short_name": "A-4E-C",
    "tags": ["fixed-wing", "attack"],
    "platform_designation": "A-4",
    "type_designation": "A-4E",
    "official_name": "Skyhawk",
    "nickname": "Scooter",
    "threat_radius_nm": 15,
    "fuel_receiver": "probe-and-drogue"
  }
]`

// assertIsA4 checks that the parsed aircraft matches the expected A-4 Skyhawk fields.
func assertIsA4(t *testing.T, a Aircraft) {
	t.Helper()
	assert.Equal(t, "A-4E-C", a.ACMIShortName)
	assert.Equal(t, "A-4", a.PlatformDesignation)
	assert.Equal(t, "A-4E", a.TypeDesignation)
	assert.Equal(t, "Skyhawk", a.OfficialName)
	assert.Equal(t, "Scooter", a.Nickname)
	assert.True(t, a.HasTag(FixedWing))
	assert.True(t, a.HasTag(Attack))
	assert.Equal(t, brevity.FixedWing, a.Category())
	assert.InDelta(t, float64(15*unit.NauticalMile), float64(a.ThreatRadius()), 0.001)
	assert.Equal(t, ProbeAndDrogue, a.FuelReceiver())
	assert.Equal(t, NoAirRefueling, a.FuelProvider())
}

func TestLoadCustomAircraftJSON(t *testing.T) {
	t.Parallel()
	aircraft, err := LoadCustomAircraft([]byte(a4JSON))
	require.NoError(t, err)
	require.Len(t, aircraft, 1)
	assertIsA4(t, aircraft[0])
}

func TestLoadCustomAircraftYAML(t *testing.T) {
	t.Parallel()
	aircraft, err := LoadCustomAircraft([]byte(a4YAML))
	require.NoError(t, err)
	require.Len(t, aircraft, 1)
	assertIsA4(t, aircraft[0])
}

func TestLoadCustomAircraftTankerProvider(t *testing.T) {
	t.Parallel()
	data := `[{"acmi_short_name":"KC-Custom","tags":["fixed-wing","unarmed"],"platform_designation":"KC-Custom","fuel_provider":"boom"}]`
	aircraft, err := LoadCustomAircraft([]byte(data))
	require.NoError(t, err)
	require.Len(t, aircraft, 1)
	assert.Equal(t, FlyingBoom, aircraft[0].FuelProvider())
	assert.Equal(t, NoAirRefueling, aircraft[0].FuelReceiver())
}

func TestLoadCustomAircraftThreatRadiusDefault(t *testing.T) {
	t.Parallel()
	// Omitting threat_radius_nm on a fighter falls back to the tag-based default.
	data := `[{"acmi_short_name":"MiG-Custom","tags":["fixed-wing","fighter"],"nato_reporting_name":"Fulcrum"}]`
	aircraft, err := LoadCustomAircraft([]byte(data))
	require.NoError(t, err)
	require.Len(t, aircraft, 1)
	assert.InDelta(t, float64(SAR2AR1Threat), float64(aircraft[0].ThreatRadius()), 0.001)
}

func TestLoadCustomAircraftErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		data string
	}{
		{name: "not json or yaml", data: "not valid [[["},
		{name: "missing acmi short name", data: `[{"tags":["fixed-wing"],"nickname":"X"}]`},
		{name: "no reporting name", data: `[{"acmi_short_name":"X","tags":["fixed-wing"]}]`},
		{name: "no tags", data: `[{"acmi_short_name":"X","nickname":"X"}]`},
		{name: "unrecognized tag", data: `[{"acmi_short_name":"X","nickname":"X","tags":["spaceship"]}]`},
		{name: "no wing tag", data: `[{"acmi_short_name":"X","nickname":"X","tags":["attack"]}]`},
		{name: "two wing tags", data: `[{"acmi_short_name":"X","nickname":"X","tags":["fixed-wing","rotary-wing"]}]`},
		{name: "negative threat radius", data: `[{"acmi_short_name":"X","nickname":"X","tags":["fixed-wing"],"threat_radius_nm":-5}]`},
		{name: "bad fuel provider", data: `[{"acmi_short_name":"X","nickname":"X","tags":["fixed-wing"],"fuel_provider":"laser"}]`},
		{name: "bad fuel receiver", data: `[{"acmi_short_name":"X","nickname":"X","tags":["fixed-wing"],"fuel_receiver":"laser"}]`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := LoadCustomAircraft([]byte(test.data))
			require.Error(t, err)
		})
	}
}

//nolint:paralleltest // mutates the package-level aircraftDataLUT, so it must not run concurrently with map readers
func TestAddCustomAircraftOverride(t *testing.T) {
	name := "TEST-OVERRIDE-JET"
	require.NotContains(t, aircraftDataLUT, name)
	t.Cleanup(func() { delete(aircraftDataLUT, name) })

	data := `[{"acmi_short_name":"TEST-OVERRIDE-JET","tags":["fixed-wing","fighter"],"official_name":"Testbird","threat_radius_nm":40}]`
	aircraft, err := LoadCustomAircraft([]byte(data))
	require.NoError(t, err)
	AddCustomAircraft(aircraft)

	got, ok := GetAircraftData(name)
	require.True(t, ok)
	assert.Equal(t, "Testbird", got.OfficialName)
	assert.InDelta(t, float64(40*unit.NauticalMile), float64(got.ThreatRadius()), 0.001)

	// A second entry with the same ACMI short name overrides the first.
	override := `[{"acmi_short_name":"TEST-OVERRIDE-JET","tags":["fixed-wing","attack"],"official_name":"Override","threat_radius_nm":10}]`
	overrideAircraft, err := LoadCustomAircraft([]byte(override))
	require.NoError(t, err)
	AddCustomAircraft(overrideAircraft)

	got, ok = GetAircraftData(name)
	require.True(t, ok)
	assert.Equal(t, "Override", got.OfficialName)
	assert.InDelta(t, float64(10*unit.NauticalMile), float64(got.ThreatRadius()), 0.001)
}
