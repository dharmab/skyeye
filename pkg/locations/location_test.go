package locations

import (
	"strconv"
	"testing"

	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocationValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		location Location
		wantErr  bool
	}{
		{name: "simple valid", location: Location{Names: []string{"home plate"}}, wantErr: false},
		{name: "multi name valid", location: Location{Names: []string{"rock", "base"}}, wantErr: false},
		{name: "nil names rejected", location: Location{Names: nil}, wantErr: true},
		{name: "empty names rejected", location: Location{Names: []string{}}, wantErr: true},
		{name: "empty string name rejected", location: Location{Names: []string{""}}, wantErr: true},
		{name: "whitespace name rejected", location: Location{Names: []string{"   "}}, wantErr: true},
		{name: "one empty name among valid rejected", location: Location{Names: []string{"home", ""}}, wantErr: true},
		{name: "reserved tanker", location: Location{Names: []string{"tanker"}}, wantErr: true},
		{name: "reserved Tanker", location: Location{Names: []string{"Tanker"}}, wantErr: true},
		{name: "reserved TANKER", location: Location{Names: []string{"TANKER"}}, wantErr: true},
		{name: "reserved bullseye", location: Location{Names: []string{"bullseye"}}, wantErr: true},
		{name: "reserved Bullseye", location: Location{Names: []string{"Bullseye"}}, wantErr: true},
		{name: "reserved BULLSEYE", location: Location{Names: []string{"BULLSEYE"}}, wantErr: true},
		{name: "mixed with reserved tanker", location: Location{Names: []string{"home", "tanker"}}, wantErr: true},
		{name: "mixed with reserved bullseye", location: Location{Names: []string{"bullseye", "home"}}, wantErr: true},
		{name: "latitude max boundary valid", location: Location{Names: []string{"np"}, Latitude: 90, Longitude: 0}, wantErr: false},
		{name: "latitude min boundary valid", location: Location{Names: []string{"sp"}, Latitude: -90, Longitude: 0}, wantErr: false},
		{name: "latitude above range", location: Location{Names: []string{"bad"}, Latitude: 90.1}, wantErr: true},
		{name: "latitude below range", location: Location{Names: []string{"bad"}, Latitude: -90.1}, wantErr: true},
		{name: "longitude max boundary valid", location: Location{Names: []string{"east"}, Longitude: 180}, wantErr: false},
		{name: "longitude min boundary valid", location: Location{Names: []string{"west"}, Longitude: -180}, wantErr: false},
		{name: "longitude above range", location: Location{Names: []string{"bad"}, Longitude: 180.1}, wantErr: true},
		{name: "longitude below range", location: Location{Names: []string{"bad"}, Longitude: -180.1}, wantErr: true},
	}
	for i, test := range tests {
		name := test.name
		if name == "" {
			name = strconv.Itoa(i)
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			err := test.location.Validate()
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadLocations(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		data    string
		want    []Location
		wantErr bool
	}{
		{
			name: "json",
			data: `[{"names":["Incirlik","Home plate"],"latitude":37.0,"longitude":35.4}]`,
			want: []Location{{Names: []string{"Incirlik", "Home plate"}, Latitude: 37.0, Longitude: 35.4}},
		},
		{
			name: "yaml",
			data: "- names:\n  - Incirlik\n  - Home plate\n  latitude: 37.0\n  longitude: 35.4\n",
			want: []Location{{Names: []string{"Incirlik", "Home plate"}, Latitude: 37.0, Longitude: 35.4}},
		},
		{
			name:    "invalid",
			data:    "not valid json or yaml [[[",
			wantErr: true,
		},
		{
			name:    "reserved name",
			data:    `[{"names":["tanker"],"latitude":0,"longitude":0}]`,
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := LoadLocations([]byte(test.data))
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want, got)
			}
		})
	}
}

func TestLocationPoint(t *testing.T) {
	t.Parallel()
	loc := Location{
		Names:     []string{"test"},
		Longitude: -117.5,
		Latitude:  34.0,
	}
	expected := orb.Point{-117.5, 34.0}
	assert.Equal(t, expected, loc.Point())
}
