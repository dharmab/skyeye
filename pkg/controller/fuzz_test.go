package controller

import (
	"math/rand/v2"
	"testing"
	"testing/synctest"
	"time"

	"github.com/dharmab/skyeye/pkg/bearings"
	"github.com/dharmab/skyeye/pkg/brevity"
	"github.com/dharmab/skyeye/pkg/coalitions"
	"github.com/dharmab/skyeye/pkg/locations"
	"github.com/martinlindhe/unit"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

type fuzzWorld struct {
	nBlue int
	nRed  int
}

func randomPoint(rng *rand.Rand) orb.Point {
	return orb.Point{
		29.0 + rng.Float64()*2.0,
		39.0 + rng.Float64()*2.0,
	}
}

func randomBearing(rng *rand.Rand) bearings.Bearing {
	return bearings.NewMagneticBearing(unit.Angle(rng.Float64()*360) * unit.Degree)
}

func randomRange(rng *rand.Rand) unit.Length {
	return unit.Length(5+rng.Float64()*100) * unit.NauticalMile
}

func randomAltitude(rng *rand.Rand) unit.Length {
	return unit.Length(rng.IntN(50000)) * unit.Foot
}

func randomTrack(rng *rand.Rand) brevity.Track {
	tracks := []brevity.Track{
		brevity.UnknownDirection,
		brevity.North, brevity.Northeast, brevity.East, brevity.Southeast,
		brevity.South, brevity.Southwest, brevity.West, brevity.Northwest,
	}
	return tracks[rng.IntN(len(tracks))]
}

var fuzzCallsigns = []string{"eagle 1", "viper 1", "hornet 1", "tomcat 1", "warthog 1", "ghost 1"}

func randomCallsign(rng *rand.Rand) string {
	return fuzzCallsigns[rng.IntN(len(fuzzCallsigns))]
}

var blueACMI = []string{acmiF15C, acmiF16C, acmiFA18C, acmiF14B, acmiA10C}
var redACMI = []string{acmiSu27, acmiMiG29A, acmiSu25T, acmiMiG21, acmiJ11A}

func setupFuzzWorld(t *testing.T, h *controllerTestHarness, rng *rand.Rand) fuzzWorld {
	t.Helper()
	var w fuzzWorld

	w.nBlue = 1 + rng.IntN(4)
	for i := range w.nBlue {
		acmi := blueACMI[rng.IntN(len(blueACMI))]
		name := []string{"Eagle", "Viper", "Hornet", "Tomcat", "Warthog"}[i%5]
		h.insertAircraft(t, name+" 1 Reaper", acmi, coalitions.Blue, randomPoint(rng))
	}

	w.nRed = rng.IntN(5)
	for i := range w.nRed {
		acmi := redACMI[rng.IntN(len(redACMI))]
		name := []string{"Bandit", "Hostile", "Bogey", "Threat", "Target"}[i%5]
		h.insertAircraft(t, name+" 1", acmi, coalitions.Red, randomPoint(rng))
	}

	return w
}

func runFuzz(t *testing.T, seed uint64, locs []locations.Location, fn func(*testing.T, *controllerTestHarness, *rand.Rand)) {
	t.Helper()
	rng := rand.New(rand.NewPCG(seed, 0))
	for range 50 {
		h := newControllerTestHarness(t, locs)
		setupFuzzWorld(t, h, rng)
		fn(t, h, rng)
	}
}

func TestFuzz_HandleRadioCheck(t *testing.T) {
	t.Parallel()
	runFuzz(t, 42, nil, func(t *testing.T, h *controllerTestHarness, rng *rand.Rand) {
		t.Helper()
		h.ctrl.HandleRadioCheck(h.ctx, &brevity.RadioCheckRequest{Callsign: randomCallsign(rng)})
		got := h.expectResponse(t)
		switch got.(type) {
		case brevity.RadioCheckResponse:
		default:
			t.Fatalf("unexpected response type %T", got)
		}
	})
}

func TestFuzz_HandleCheckIn(t *testing.T) {
	t.Parallel()
	runFuzz(t, 43, nil, func(t *testing.T, h *controllerTestHarness, rng *rand.Rand) {
		t.Helper()
		h.ctrl.HandleCheckIn(h.ctx, &brevity.CheckInRequest{Callsign: randomCallsign(rng)})
		got := h.expectResponse(t)
		switch got.(type) {
		case brevity.CheckInResponse, brevity.NegativeRadarContactResponse:
		default:
			t.Fatalf("unexpected response type %T", got)
		}
	})
}

func TestFuzz_HandleShopping(t *testing.T) {
	t.Parallel()
	runFuzz(t, 44, nil, func(t *testing.T, h *controllerTestHarness, rng *rand.Rand) {
		t.Helper()
		h.ctrl.HandleShopping(h.ctx, &brevity.ShoppingRequest{Callsign: randomCallsign(rng)})
		got := h.expectResponse(t)
		switch got.(type) {
		case brevity.ShoppingResponse, brevity.NegativeRadarContactResponse:
		default:
			t.Fatalf("unexpected response type %T", got)
		}
	})
}

func TestFuzz_HandleTripwire(t *testing.T) {
	t.Parallel()
	runFuzz(t, 45, nil, func(t *testing.T, h *controllerTestHarness, rng *rand.Rand) {
		t.Helper()
		h.ctrl.HandleTripwire(h.ctx, &brevity.TripwireRequest{Callsign: randomCallsign(rng)})
		got := h.expectResponse(t)
		switch got.(type) {
		case brevity.TripwireResponse, brevity.NegativeRadarContactResponse:
		default:
			t.Fatalf("unexpected response type %T", got)
		}
	})
}

func TestFuzz_HandleUnableToUnderstand(t *testing.T) {
	t.Parallel()
	callsigns := append(fuzzCallsigns, "")
	runFuzz(t, 46, nil, func(t *testing.T, h *controllerTestHarness, rng *rand.Rand) {
		t.Helper()
		cs := callsigns[rng.IntN(len(callsigns))]
		h.ctrl.HandleUnableToUnderstand(h.ctx, &brevity.UnableToUnderstandRequest{Callsign: cs})
		got := h.expectResponse(t)
		resp, ok := got.(brevity.SayAgainResponse)
		if !ok {
			t.Fatalf("unexpected response type %T", got)
		}
		assert.NotEmpty(t, resp.Callsign)
	})
}

func TestFuzz_HandleAlphaCheck(t *testing.T) {
	t.Parallel()
	runFuzz(t, 47, nil, func(t *testing.T, h *controllerTestHarness, rng *rand.Rand) {
		t.Helper()
		h.ctrl.HandleAlphaCheck(h.ctx, &brevity.AlphaCheckRequest{Callsign: randomCallsign(rng)})
		got := h.expectResponse(t)
		resp, ok := got.(brevity.AlphaCheckResponse)
		if !ok {
			t.Fatalf("unexpected response type %T", got)
		}
		if resp.Status {
			assert.NotNil(t, resp.Location, "Status=true requires Location")
		}
	})
}

func TestFuzz_HandleBogeyDope(t *testing.T) {
	t.Parallel()
	filters := []brevity.ContactCategory{brevity.Aircraft, brevity.FixedWing, brevity.RotaryWing}
	runFuzz(t, 48, nil, func(t *testing.T, h *controllerTestHarness, rng *rand.Rand) {
		t.Helper()
		filter := filters[rng.IntN(len(filters))]
		h.ctrl.HandleBogeyDope(h.ctx, &brevity.BogeyDopeRequest{
			Callsign: randomCallsign(rng),
			Filter:   filter,
		})
		got := h.expectResponse(t)
		switch resp := got.(type) {
		case brevity.BogeyDopeResponse:
			if resp.Group != nil {
				assert.Equal(t, brevity.Hostile, resp.Group.Declaration())
				assert.NotNil(t, resp.Group.BRAA())
			}
		case brevity.NegativeRadarContactResponse:
		default:
			t.Fatalf("unexpected response type %T", got)
		}
	})
}

func TestFuzz_HandlePicture(t *testing.T) {
	t.Parallel()
	callsigns := append(fuzzCallsigns, "")
	rng := rand.New(rand.NewPCG(49, 0))
	for range 50 {
		cs := callsigns[rng.IntN(len(callsigns))]
		synctest.Test(t, func(t *testing.T) {
			h := newControllerTestHarness(t, nil)
			w := setupFuzzWorld(t, h, rng)
			time.Sleep(6 * time.Second)
			synctest.Wait()
			h.ctrl.HandlePicture(h.ctx, &brevity.PictureRequest{Callsign: cs})
			got := h.expectResponse(t)
			resp, ok := got.(brevity.PictureResponse)
			if !ok {
				t.Fatalf("unexpected response type %T", got)
			}
			if w.nRed == 0 {
				assert.Equal(t, 0, resp.Count)
			}
			for _, g := range resp.Groups {
				assert.Equal(t, brevity.Hostile, g.Declaration())
			}
		})
	}
}

func TestFuzz_HandleSpiked(t *testing.T) {
	t.Parallel()
	runFuzz(t, 50, nil, func(t *testing.T, h *controllerTestHarness, rng *rand.Rand) {
		t.Helper()
		h.ctrl.HandleSpiked(h.ctx, &brevity.SpikedRequest{
			Callsign: randomCallsign(rng),
			Bearing:  randomBearing(rng),
		})
		got := h.expectResponse(t)
		switch resp := got.(type) {
		case brevity.SpikedResponseV2:
			if resp.Status {
				assert.NotNil(t, resp.Group, "Status=true requires Group")
			}
		case brevity.NegativeRadarContactResponse:
		default:
			t.Fatalf("unexpected response type %T", got)
		}
	})
}

func TestFuzz_HandleStrobe(t *testing.T) {
	t.Parallel()
	runFuzz(t, 51, nil, func(t *testing.T, h *controllerTestHarness, rng *rand.Rand) {
		t.Helper()
		h.ctrl.HandleStrobe(h.ctx, &brevity.StrobeRequest{
			Callsign: randomCallsign(rng),
			Bearing:  randomBearing(rng),
		})
		got := h.expectResponse(t)
		switch resp := got.(type) {
		case brevity.StrobeResponse:
			if resp.Status {
				assert.NotNil(t, resp.Group, "Status=true requires Group")
			}
		case brevity.NegativeRadarContactResponse:
		default:
			t.Fatalf("unexpected response type %T", got)
		}
	})
}

func TestFuzz_HandleSnaplock(t *testing.T) {
	t.Parallel()
	runFuzz(t, 52, nil, func(t *testing.T, h *controllerTestHarness, rng *rand.Rand) {
		t.Helper()
		h.ctrl.HandleSnaplock(h.ctx, &brevity.SnaplockRequest{
			Callsign: randomCallsign(rng),
			BRA:      brevity.NewBRA(randomBearing(rng), randomRange(rng), randomAltitude(rng)),
		})
		got := h.expectResponse(t)
		switch resp := got.(type) {
		case brevity.SnaplockResponse:
			switch resp.Declaration {
			case brevity.Clean:
				assert.Nil(t, resp.Group, "Clean requires nil Group")
			case brevity.Friendly:
				assert.NotNil(t, resp.Group, "Friendly requires Group")
			case brevity.Hostile, brevity.Bogey, brevity.Bandit:
				assert.NotNil(t, resp.Group, "%s requires Group", resp.Declaration)
			case brevity.Furball, brevity.Unable, brevity.Neutral:
			}
		case brevity.NegativeRadarContactResponse:
		default:
			t.Fatalf("unexpected response type %T", got)
		}
	})
}

func TestFuzz_HandleDeclare(t *testing.T) {
	t.Parallel()
	runFuzz(t, 53, nil, func(t *testing.T, h *controllerTestHarness, rng *rand.Rand) {
		t.Helper()
		var req brevity.DeclareRequest
		req.Callsign = randomCallsign(rng)
		switch rng.IntN(3) {
		case 0:
			req.Sour = true
		case 1:
			req.IsBRAA = true
			req.Bearing = randomBearing(rng)
			req.Range = randomRange(rng)
			req.Altitude = randomAltitude(rng)
			req.Track = randomTrack(rng)
		case 2:
			req.IsBRAA = false
			req.IsAmbiguous = rng.IntN(2) == 0
			req.Bullseye = brevity.NewBullseye(randomBearing(rng), randomRange(rng))
			req.Altitude = randomAltitude(rng)
			req.Track = randomTrack(rng)
		}

		h.ctrl.HandleDeclare(h.ctx, &req)
		got := h.expectResponse(t)
		switch resp := got.(type) {
		case brevity.DeclareResponse:
			switch resp.Declaration {
			case brevity.Clean:
				assert.Nil(t, resp.Group, "Clean requires nil Group")
			case brevity.Friendly:
				assert.NotNil(t, resp.Group, "Friendly requires Group")
			case brevity.Hostile, brevity.Bogey, brevity.Bandit:
				assert.NotNil(t, resp.Group, "%s requires Group", resp.Declaration)
			case brevity.Furball, brevity.Neutral:
			case brevity.Unable:
				assert.True(t, resp.Sour || !req.IsBRAA, "Unable only for Sour or nil bullseye")
			}
			if req.Sour {
				assert.Equal(t, brevity.Unable, resp.Declaration)
			}
		case brevity.NegativeRadarContactResponse:
		default:
			t.Fatalf("unexpected response type %T", got)
		}
	})
}

func TestFuzz_HandleVector(t *testing.T) {
	t.Parallel()
	locs := []locations.Location{
		{Names: []string{"home plate"}, Longitude: 30.0, Latitude: 40.0},
		{Names: []string{"divert", "alternate"}, Longitude: 31.0, Latitude: 40.5},
	}
	locNames := []string{"home plate", "divert", "alternate", "atlantis", brevity.LocationTanker}
	runFuzz(t, 54, locs, func(t *testing.T, h *controllerTestHarness, rng *rand.Rand) {
		t.Helper()
		loc := locNames[rng.IntN(len(locNames))]
		h.ctrl.HandleVector(h.ctx, &brevity.VectorRequest{
			Callsign: randomCallsign(rng),
			Location: loc,
		})
		got := h.expectResponse(t)
		resp, ok := got.(brevity.VectorResponse)
		if !ok {
			t.Fatalf("unexpected response type %T", got)
		}
		if resp.Contact && resp.Status {
			if loc == brevity.LocationTanker {
				assert.NotNil(t, resp.BRA, "tanker vector requires BRA")
			} else {
				assert.NotNil(t, resp.Vector, "location vector requires Vector")
			}
		}
	})
}
