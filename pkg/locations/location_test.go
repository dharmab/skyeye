package locations

import (
	"strconv"
	"testing"

	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
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
