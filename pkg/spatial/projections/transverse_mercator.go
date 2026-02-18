package projections

import (
	"math"

	"github.com/paulmach/orb"
)

// WGS84 ellipsoid parameters.
//
// The Earth is not a perfect sphere. WGS84 approximates its shape as an
// oblate ellipsoid — slightly flattened at the poles and wider at the equator.
// This is a better approximation than a sphere, though the real Earth has
// further irregularities (gravity anomalies, terrain) that the ellipsoid
// ignores. The ellipsoid is defined by two values: the semi-major axis
// (equatorial radius, ~6378 km) and the flattening ratio (~1/298). All
// coordinates in this package reference this ellipsoid.
const (
	// wgs84SemiMajorAxis is the equatorial radius of the WGS84 ellipsoid in meters.
	wgs84SemiMajorAxis = 6378137.0
	// wgs84Flattening is the flattening ratio (a-b)/a, describing how much
	// the poles are compressed relative to the equator.
	wgs84Flattening = 1.0 / 298.257223563
)

// Derived ellipsoid geometry. The semi-minor axis is the polar radius.
// Eccentricity measures how much the ellipse deviates from a circle — zero
// for a perfect sphere, approaching one for a very elongated ellipse. The
// "second eccentricity" is the same concept but referenced to the minor axis;
// it appears in the projection formulas because the curvature corrections are
// most naturally expressed relative to the polar radius.
var (
	// wgs84SemiMinorAxis is the polar radius of the WGS84 ellipsoid in meters.
	wgs84SemiMinorAxis = wgs84SemiMajorAxis * (1 - wgs84Flattening)
	// wgs84EccentricitySquared measures how much the ellipsoid deviates from a
	// sphere (zero for a perfect sphere, approaching one for a very elongated
	// ellipse), referenced to the semi-major (equatorial) axis.
	wgs84EccentricitySquared = (square(wgs84SemiMajorAxis) - square(wgs84SemiMinorAxis)) / square(wgs84SemiMajorAxis)
	// wgs84SecondEccentricitySquared is the same concept as eccentricity but
	// referenced to the semi-minor (polar) axis. It appears in the projection
	// formulas because curvature corrections are most naturally expressed
	// relative to the polar radius.
	wgs84SecondEccentricitySquared = (square(wgs84SemiMajorAxis) - square(wgs84SemiMinorAxis)) / square(wgs84SemiMinorAxis)
)

// TransverseMercator converts between WGS84 geographic coordinates
// (longitude/latitude in degrees) and a flat Transverse Mercator coordinate
// system (easting/northing in meters).
//
// A Transverse Mercator projection works by conceptually wrapping a cylinder
// around the Earth so that it touches along a chosen line of longitude (the
// central meridian) instead of along the equator. Points on the ellipsoid are
// then projected onto this cylinder and unrolled into a flat surface. Near the
// central meridian the distortion is very small, making it a good local
// coordinate system for distance and angle calculations. Farther from the
// central meridian, distortion grows — which is why each DCS terrain uses its
// own projection centered on that terrain's longitude.
type TransverseMercator struct {
	// centralMeridian is the longitude (in radians) where the projection
	// cylinder touches the Earth.
	centralMeridian float64
	// originLatitude is the latitude (in radians) used as the northing origin
	// (usually 0°).
	originLatitude float64
	// scaleFactor is a slight shrink applied at the central meridian (typically
	// 0.9996) so that distortion is spread more evenly across the projection
	// zone.
	scaleFactor float64
	// falseEasting is a constant offset (in meters) added to easting
	// coordinates so that all values in the useful area are positive.
	falseEasting float64
	// falseNorthing is a constant offset (in meters) added to northing
	// coordinates so that all values in the useful area are positive.
	falseNorthing float64
}

// NewTransverseMercator creates a Transverse Mercator projection with the
// given options. The default scale factor is 1.0; all other parameters default
// to 0.
func NewTransverseMercator(opts ...Option) *TransverseMercator {
	tm := &TransverseMercator{
		scaleFactor: 1.0,
	}
	for _, opt := range opts {
		opt(tm)
	}
	return tm
}

// ToProjected converts a WGS84 point (orb.Point{longitude, latitude} in
// degrees) to flat projected coordinates (orb.Point{easting, northing} in
// meters).
//
// The implementation uses the series expansion from "Map Projections: A Working
// Manual" (https://pubs.usgs.gov/pp/1395/report.pdf), Chapter 8: Transverse
// Mercator projection.
func (p *TransverseMercator) ToProjected(point orb.Point) orb.Point {
	// "For the ellipsoidal form, the most practical form of the equations is a set
	// of series approximations which converge rapidly to the correct centimeter or
	// less at full scale in a zone extending 3° to 4° of longitude from the central
	// meridian. Beyond this, the forward series as given here is accurate to about
	// a centimeter at 7° longitude, but the inverse series does not have sufficient
	// terms for this accuracy. The forward series may be used with meter accuracy
	// to 10° of longitude."

	// TODO: Find a more accurate series expansion in US Army manuals.
	lon := point.Lon() * math.Pi / 180.0
	lat := point.Lat() * math.Pi / 180.0

	e2 := wgs84EccentricitySquared
	ep2 := wgs84SecondEccentricitySquared
	a := wgs84SemiMajorAxis
	k0 := p.scaleFactor

	// N is the radius of curvature in the prime vertical — the distance from
	// the surface to the minor axis along a line perpendicular to the ellipsoid.
	// It varies with latitude because the ellipsoid is not a sphere.
	N := a / math.Sqrt(1-e2*square(math.Sin(lat)))

	// T, C, A are intermediate terms used in the series expansion.

	// T captures latitude-dependent curvature
	T := square(math.Tan(lat))
	// C captures the second eccentricity's effect
	C := ep2 * square(math.Cos(lat))
	// A is the longitude offset scaled by cosine of latitude (i.e., the
	// convergence of meridians toward the poles).
	deltaLon := lon - p.centralMeridian
	A := deltaLon * math.Cos(lat)

	M := meridionalArc(lat)
	M0 := meridionalArc(p.originLatitude)

	A2 := A * A
	A3 := A2 * A
	A4 := A3 * A
	A5 := A4 * A
	A6 := A5 * A

	easting := k0 * N * (A +
		(1-T+C)*A3/6 +
		(5-18*T+square(T)+72*C-58*ep2)*A5/120)

	northing := k0 * ((M - M0) +
		N*math.Tan(lat)*(A2/2+
			(5-T+9*C+4*square(C))*A4/24+
			(61-58*T+square(T)+600*C-330*ep2)*A6/720))

	return orb.Point{
		easting + p.falseEasting,
		northing + p.falseNorthing,
	}
}

// ToWGS84 converts projected coordinates (orb.Point{easting, northing} in
// meters) back to WGS84 geographic coordinates (orb.Point{longitude, latitude}
// in degrees).
//
// The inverse projection is not a simple algebraic inversion of the forward
// formulas. Instead it uses the "footpoint latitude" approach (Snyder eq. 8-18
// through 8-23): first, the northing is converted back to a meridional arc
// distance, then a series expansion recovers an approximate latitude (the
// footpoint), and finally correction terms refine the latitude and recover
// the longitude.
func (p *TransverseMercator) ToWGS84(projected orb.Point) orb.Point {
	easting := projected[0] - p.falseEasting
	northing := projected[1] - p.falseNorthing

	e2 := wgs84EccentricitySquared
	ep2 := wgs84SecondEccentricitySquared
	a := wgs84SemiMajorAxis
	k0 := p.scaleFactor

	M0 := meridionalArc(p.originLatitude)

	// Recover meridional arc distance from the northing, then compute mu,
	// the "rectifying latitude" — latitude on an equal-arc sphere.
	M := M0 + northing/k0
	mu := M / (a * (1 - e2/4 - 3*square(e2)/64 - 5*cube(e2)/256))

	// Footpoint latitude: series expansion that converts from the rectifying
	// latitude back to geodetic latitude on the ellipsoid. e1 is a small
	// quantity derived from eccentricity that controls convergence.
	e1 := (1 - math.Sqrt(1-e2)) / (1 + math.Sqrt(1-e2))
	phi1 := mu +
		(3*e1/2-27*cube(e1)/32)*math.Sin(2*mu) +
		(21*square(e1)/16-55*math.Pow(e1, 4)/32)*math.Sin(4*mu) +
		(151*cube(e1)/96)*math.Sin(6*mu) +
		(1097*math.Pow(e1, 4)/512)*math.Sin(8*mu)

	sinPhi1 := math.Sin(phi1)
	cosPhi1 := math.Cos(phi1)
	tanPhi1 := math.Tan(phi1)

	N1 := a / math.Sqrt(1-e2*square(sinPhi1))
	// R1 is the meridional radius of curvature — the radius of the north-south
	// cross-section of the ellipsoid at the footpoint latitude.
	R1 := a * (1 - e2) / math.Pow(1-e2*square(sinPhi1), 1.5)
	T1 := square(tanPhi1)
	C1 := ep2 * square(cosPhi1)
	D := easting / (N1 * k0)

	D2 := D * D
	D3 := D2 * D
	D4 := D3 * D
	D5 := D4 * D
	D6 := D5 * D

	lat := phi1 - (N1*tanPhi1/R1)*(D2/2-
		(5+3*T1+10*C1-4*square(C1)-9*ep2)*D4/24+
		(61+90*T1+298*C1+45*square(T1)-252*ep2-3*square(C1))*D6/720)

	lon := p.centralMeridian + (D-
		(1+2*T1+C1)*D3/6+
		(5-2*C1+28*T1-3*square(C1)+8*ep2+24*square(T1))*D5/120)/cosPhi1

	return orb.Point{
		lon * 180.0 / math.Pi,
		lat * 180.0 / math.Pi,
	}
}

// meridionalArc computes the distance along the surface of the ellipsoid from
// the equator to the given latitude (in radians), measured along the central
// meridian. On a sphere this would be simply radius*latitude, but the
// ellipsoid's varying curvature requires a series expansion to integrate the
// arc length accurately. The coefficients A0–A6 come from expanding the
// elliptic integral in powers of eccentricity.
func meridionalArc(lat float64) float64 {
	e2 := wgs84EccentricitySquared
	a := wgs84SemiMajorAxis

	e4 := square(e2)
	e6 := cube(e2)

	A0 := 1 - e2/4 - 3*e4/64 - 5*e6/256
	A2 := 3.0 / 8.0 * (e2 + e4/4 + 15*e6/128)
	A4 := 15.0 / 256.0 * (e4 + 3*e6/4)
	A6 := 35 * e6 / 3072

	return a * (A0*lat -
		A2*math.Sin(2*lat) +
		A4*math.Sin(4*lat) -
		A6*math.Sin(6*lat))
}
