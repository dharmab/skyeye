package types

import (
	"strconv"
	"testing"

	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// p converts a value to a pointer.
func p[T any](v T) *T { return &v }

func TestObjectGetCoordinates(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		reference orb.Point
		id        uint64
		lines     []string
		want      *Coordinates
		wantErr   bool
	}{
		{
			reference: orb.Point{30, 30},
			id:        0xa03,
			lines: []string{
				"a03,T=5.4309806|6.9989461|60.47||0.1|322.8|-34817.79|220847.14|325,Type=Air+FixedWing,Name=F-4E-45MC,Pilot=Phantom 1-1 | User,Group=Incirlik | 163rd TFG 1-1,Color=Blue,Coalition=Enemies,Country=xb",
			},
			want: NewCoordinates(
				orb.Point{35.4309806, 36.9989461},
				true, true,
				p(60.47*unit.Meter),
				p(-34817.79), p(220847.14),
				nil, p(0.1*unit.Degree), p(322.8*unit.Degree),
				p(325*unit.Degree),
			),
			wantErr: false,
		},
	}

	for i, testCase := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			o := NewObject(testCase.id)
			for _, line := range testCase.lines {
				update, err := ParseObjectUpdate(line)
				require.NoError(t, err)
				for k, v := range update.Properties {
					o.SetProperty(k, v)
				}
			}

			c, err := o.GetCoordinates(testCase.reference)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, testCase.want.ValidLon, c.ValidLon)
				if testCase.want.ValidLon {
					assert.InDelta(t, testCase.want.Location.Lon(), c.Location.Lon(), 0.01)
				}
				assert.Equal(t, testCase.want.ValidLat, c.ValidLat)
				if testCase.want.ValidLat {
					assert.InDelta(t, testCase.want.Location.Lat(), c.Location.Lat(), 0.01)
				}
				if testCase.want.Altitude != nil {
					assert.InDelta(t, testCase.want.Altitude.Meters(), c.Altitude.Meters(), 1)
				} else {
					assert.Nil(t, c.Altitude)
				}
				if testCase.want.Roll != nil {
					assert.InDelta(t, testCase.want.Roll.Degrees(), c.Roll.Degrees(), 0.1)
				} else {
					assert.Nil(t, c.Roll)
				}
				if testCase.want.Pitch != nil {
					assert.InDelta(t, testCase.want.Pitch.Degrees(), c.Pitch.Degrees(), 0.1)
				} else {
					assert.Nil(t, c.Pitch)
				}
				if testCase.want.Yaw != nil {
					assert.InDelta(t, testCase.want.Yaw.Degrees(), c.Yaw.Degrees(), 0.1)
				} else {
					assert.Nil(t, c.Yaw)
				}
				if testCase.want.X != nil {
					assert.NotNil(t, c.X)
					assert.InDelta(t, *testCase.want.X, *c.X, 1)
				} else {
					assert.Nil(t, c.X)
				}
				if testCase.want.Y != nil {
					assert.NotNil(t, c.Y)
					assert.InDelta(t, *testCase.want.Y, *c.Y, 1)
				} else {
					assert.Nil(t, c.Y)
				}
				if testCase.want.Heading != nil {
					assert.InDelta(t, testCase.want.Heading.Degrees(), c.Heading.Degrees(), 0.1)
				} else {
					assert.Nil(t, c.Heading)
				}
			}
		})
	}
}
