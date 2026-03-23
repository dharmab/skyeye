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
		names   []string
		wantErr bool
	}{
		{names: []string{"home plate"}, wantErr: false},
		{names: []string{"rock", "base"}, wantErr: false},
		{names: nil, wantErr: false},
		{names: []string{}, wantErr: false},
		{names: []string{"tanker"}, wantErr: true},
		{names: []string{"Tanker"}, wantErr: true},
		{names: []string{"TANKER"}, wantErr: true},
		{names: []string{"bullseye"}, wantErr: true},
		{names: []string{"Bullseye"}, wantErr: true},
		{names: []string{"BULLSEYE"}, wantErr: true},
		{names: []string{"home", "tanker"}, wantErr: true},
		{names: []string{"bullseye", "home"}, wantErr: true},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()
			loc := Location{Names: test.names}
			err := loc.Validate()
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
