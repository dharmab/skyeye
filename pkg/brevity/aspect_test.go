package brevity

import (
	"fmt"
	"testing"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/assert"
)

func TestAspectFromAngle(t *testing.T) {
	t.Parallel()
	n := bearings.NewMagneticBearing(0 * unit.Degree)
	nne := bearings.NewMagneticBearing(22.5 * unit.Degree)
	ne := bearings.NewMagneticBearing(45 * unit.Degree)
	ene := bearings.NewMagneticBearing(67.5 * unit.Degree)
	e := bearings.NewMagneticBearing(90 * unit.Degree)
	ese := bearings.NewMagneticBearing(112.5 * unit.Degree)
	se := bearings.NewMagneticBearing(135 * unit.Degree)
	sse := bearings.NewMagneticBearing(157.5 * unit.Degree)
	s := bearings.NewMagneticBearing(180 * unit.Degree)
	ssw := bearings.NewMagneticBearing(202.5 * unit.Degree)
	sw := bearings.NewMagneticBearing(225 * unit.Degree)
	wsw := bearings.NewMagneticBearing(247.5 * unit.Degree)
	w := bearings.NewMagneticBearing(270 * unit.Degree)
	wnw := bearings.NewMagneticBearing(292.5 * unit.Degree)
	nw := bearings.NewMagneticBearing(315 * unit.Degree)
	nnw := bearings.NewMagneticBearing(337.5 * unit.Degree)

	testCases := []struct {
		bearing  bearings.Bearing
		track    bearings.Bearing
		expected Aspect
	}{
		// ⇧
		// ⬆
		{
			bearing:  n,
			track:    n,
			expected: Drag,
		},
		// ⬀
		// ⬆
		{
			bearing:  n,
			track:    nne,
			expected: Drag,
		},
		{
			bearing:  n,
			track:    ne,
			expected: Drag,
		},
		{
			bearing:  n,
			track:    ene,
			expected: Beam,
		},
		// ⇨
		// ⬆
		{
			bearing:  n,
			track:    e,
			expected: Beam,
		},
		// ⬂
		// ⬆
		{
			bearing:  n,
			track:    ese,
			expected: Flank,
		},
		{
			bearing:  n,
			track:    se,
			expected: Flank,
		},
		{
			bearing:  n,
			track:    sse,
			expected: Hot,
		},
		// ⇩
		// ⬆
		{
			bearing:  n,
			track:    s,
			expected: Hot,
		},
		// ⬃
		// ⬆
		{
			bearing:  n,
			track:    ssw,
			expected: Hot,
		},
		{
			bearing:  n,
			track:    sw,
			expected: Flank,
		},
		{
			bearing:  n,
			track:    wsw,
			expected: Flank,
		},
		// ⇦
		// ⬆
		{
			bearing:  n,
			track:    w,
			expected: Beam,
		},
		// ⬁
		// ⬆
		{
			bearing:  n,
			track:    wnw,
			expected: Beam,
		},
		{
			bearing:  n,
			track:    nw,
			expected: Drag,
		},
		{
			bearing:  n,
			track:    nnw,
			expected: Drag,
		},
		// ⮕⇧
		{
			bearing:  e,
			track:    n,
			expected: Beam,
		},
		// ⮕⬀
		{
			bearing:  e,
			track:    nne,
			expected: Beam,
		},
		{
			bearing:  e,
			track:    ne,
			expected: Drag,
		},

		{
			bearing:  e,
			track:    ene,
			expected: Drag,
		},
		// ⮕⇨
		{
			bearing:  e,
			track:    e,
			expected: Drag,
		},
		// ⮕⬂
		{
			bearing:  e,
			track:    ese,
			expected: Drag,
		},
		{
			bearing:  e,
			track:    se,
			expected: Drag,
		},
		{
			bearing:  e,
			track:    sse,
			expected: Beam,
		},
		// ⮕⇩
		{
			bearing:  e,
			track:    s,
			expected: Beam,
		},
		// ⮕⬃
		{
			bearing:  e,
			track:    ssw,
			expected: Flank,
		},
		{
			bearing:  e,
			track:    sw,
			expected: Flank,
		},
		{
			bearing:  e,
			track:    wsw,
			expected: Hot,
		},
		// ⮕⇦
		{
			bearing:  e,
			track:    w,
			expected: Hot,
		},
		// ⮕⬁
		{
			bearing:  e,
			track:    wnw,
			expected: Hot,
		},
		{
			bearing:  e,
			track:    nw,
			expected: Flank,
		},
		{
			bearing:  e,
			track:    nnw,
			expected: Flank,
		},
		// ⬇
		// ⇧
		{
			bearing:  s,
			track:    n,
			expected: Hot,
		},
		// ⬇
		// ⬀
		{
			bearing:  s,
			track:    nne,
			expected: Hot,
		},
		{
			bearing:  s,
			track:    ne,
			expected: Flank,
		},
		{
			bearing:  s,
			track:    ene,
			expected: Flank,
		},
		// ⬇
		// ⇨
		{
			bearing:  s,
			track:    e,
			expected: Beam,
		},
		// ⬇
		// ⬂
		{
			bearing:  s,
			track:    ese,
			expected: Beam,
		},
		{
			bearing:  s,
			track:    se,
			expected: Drag,
		},
		{
			bearing:  s,
			track:    sse,
			expected: Drag,
		},
		// ⬇
		// ⇩
		{
			bearing:  s,
			track:    s,
			expected: Drag,
		},
		// ⬇
		// ⬃
		{
			bearing:  s,
			track:    ssw,
			expected: Drag,
		},
		{
			bearing:  s,
			track:    sw,
			expected: Drag,
		},
		{
			bearing:  s,
			track:    wsw,
			expected: Beam,
		},
		// ⬇
		// ⇦
		{
			bearing:  s,
			track:    w,
			expected: Beam,
		},
		// ⬇
		// ⬁
		{
			bearing:  s,
			track:    wnw,
			expected: Flank,
		},
		{
			bearing:  s,
			track:    nw,
			expected: Flank,
		},
		{
			bearing:  s,
			track:    nnw,
			expected: Hot,
		},
		// ⇧⬅
		{
			bearing:  w,
			track:    n,
			expected: Beam,
		},
		// ⬀⬅
		{
			bearing:  w,
			track:    nne,
			expected: Flank,
		},
		{
			bearing:  w,
			track:    ne,
			expected: Flank,
		},
		{
			bearing:  w,
			track:    ene,
			expected: Hot,
		},
		// ⇨⬅
		{
			bearing:  w,
			track:    e,
			expected: Hot,
		},
		// ⬂⬅
		{
			bearing:  w,
			track:    ese,
			expected: Hot,
		},
		{
			bearing:  w,
			track:    se,
			expected: Flank,
		},
		{
			bearing:  w,
			track:    sse,
			expected: Flank,
		},
		// ⇩⬅

		{
			bearing:  w,
			track:    s,
			expected: Beam,
		},
		// ⬃⬅
		{
			bearing:  w,
			track:    ssw,
			expected: Beam,
		},
		{
			bearing:  w,
			track:    sw,
			expected: Drag,
		},
		{
			bearing:  w,
			track:    wsw,
			expected: Drag,
		},
		// ⇦⬅
		{
			bearing:  w,
			track:    w,
			expected: Drag,
		},
		// ⬁⬅
		{
			bearing:  w,
			track:    wnw,
			expected: Drag,
		},
		{
			bearing:  w,
			track:    nw,
			expected: Drag,
		},
		{
			bearing:  w,
			track:    nnw,
			expected: Beam,
		},
	}

	for _, test := range testCases {
		t.Run(fmt.Sprintf("bearing %.1f track %.1f", test.bearing.Degrees(), test.track.Degrees()), func(t *testing.T) {
			t.Parallel()
			actual := AspectFromAngle(test.bearing, test.track)
			assert.Equal(t, test.expected, actual)
		})
	}
}
