package composer

import (
	"testing"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/martinlindhe/unit"
	"github.com/stretchr/testify/assert"
)

func TestComposeVectorResponse_NoContact(t *testing.T) {
	t.Parallel()
	c := &Composer{Callsign: "Anyface"}
	resp := c.ComposeVectorResponse(brevity.VectorResponse{
		Callsign: "eagle 1",
		Location: "home plate",
		Contact:  false,
	})
	assert.Equal(t, "eagle 1, negative contact", resp.Subtitle)
	assert.Equal(t, "eagle 1, negative contact", resp.Speech)
}

func TestComposeVectorResponse_UnableNamedLocation(t *testing.T) {
	t.Parallel()
	c := &Composer{Callsign: "Anyface"}
	resp := c.ComposeVectorResponse(brevity.VectorResponse{
		Callsign: "eagle 1",
		Location: "home plate",
		Contact:  true,
		Status:   false,
	})
	assert.Contains(t, resp.Subtitle, "unable to provide vector to home plate")
}

func TestComposeVectorResponse_UnableTanker(t *testing.T) {
	t.Parallel()
	c := &Composer{Callsign: "Anyface"}
	resp := c.ComposeVectorResponse(brevity.VectorResponse{
		Callsign: "eagle 1",
		Location: brevity.LocationTanker,
		Contact:  true,
		Status:   false,
	})
	assert.Contains(t, resp.Subtitle, "no compatible tankers available")
	assert.Contains(t, resp.Speech, "no compatible tankers available")
}

func TestComposeVectorResponse_RegularVector(t *testing.T) {
	t.Parallel()
	c := &Composer{Callsign: "Anyface"}
	bearing := bearings.NewMagneticBearing(90 * unit.Degree)
	resp := c.ComposeVectorResponse(brevity.VectorResponse{
		Callsign: "eagle 1",
		Location: "home plate",
		Contact:  true,
		Status:   true,
		Vector:   brevity.NewVector(bearing, 42*unit.NauticalMile),
	})
	// Subtitle uses bearing string form / slash-separated distance.
	assert.Contains(t, resp.Subtitle, "vector to home plate")
	assert.Contains(t, resp.Subtitle, "/42")
	assert.Contains(t, resp.Speech, "vector to home plate")
}

func TestComposeVectorResponse_TankerWithTrack(t *testing.T) {
	t.Parallel()
	c := &Composer{Callsign: "Anyface"}
	bearing := bearings.NewMagneticBearing(180 * unit.Degree)
	resp := c.ComposeVectorResponse(brevity.VectorResponse{
		Callsign: "eagle 1",
		Location: "Texaco 1",
		Contact:  true,
		Status:   true,
		BRA:      brevity.NewBRA(bearing, 35*unit.NauticalMile, 20000*unit.Foot),
		Track:    brevity.North,
	})
	assert.Contains(t, resp.Subtitle, "nearest tanker")
	assert.Contains(t, resp.Subtitle, "Texaco 1")
	assert.Contains(t, resp.Subtitle, "/35")
	assert.Contains(t, resp.Subtitle, "track north")
	assert.Contains(t, resp.Speech, "track north")
}

func TestComposeVectorResponse_TankerWithoutTrack(t *testing.T) {
	t.Parallel()
	c := &Composer{Callsign: "Anyface"}
	bearing := bearings.NewMagneticBearing(180 * unit.Degree)
	resp := c.ComposeVectorResponse(brevity.VectorResponse{
		Callsign: "eagle 1",
		Location: "Texaco 1",
		Contact:  true,
		Status:   true,
		BRA:      brevity.NewBRA(bearing, 35*unit.NauticalMile, 20000*unit.Foot),
		Track:    brevity.UnknownDirection,
	})
	assert.Contains(t, resp.Subtitle, "nearest tanker")
	assert.NotContains(t, resp.Subtitle, "track")
	assert.NotContains(t, resp.Speech, "track")
}
