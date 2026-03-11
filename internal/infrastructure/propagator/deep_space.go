package propagator

import "math"

// IsDeepSpace returns true if the satellite is in a deep-space orbit
// (orbital period > 225 minutes, equivalent to mean motion < 6.4 rev/day).
func IsDeepSpace(meanMotion float64) bool {
	return meanMotion < 6.4
}

// deepSpaceCorrections holds precomputed lunar-solar perturbation coefficients.
type deepSpaceCorrections struct {
	// Lunar perturbation terms
	lunarRAAN     float64 // lunar ascending node rate
	lunarInclCorr float64 // lunar inclination correction amplitude

	// Solar perturbation terms
	solarRAAN     float64 // solar ascending node rate
	solarInclCorr float64 // solar inclination correction amplitude

	// Secular rates (rad/min)
	raanDotDS float64 // deep-space RAAN rate
	wDotDS    float64 // deep-space arg perigee rate
	mDotDS    float64 // deep-space mean anomaly rate
}

// Lunar-solar perturbation constants.
const (
	// Lunar ascending node rate: ~-0.00338 rad/day converted to rad/min
	lunarNodeRate = -0.00338 / MinutesPerDay

	// Solar argument rate (Earth's orbital rate): ~0.01720 rad/day converted to rad/min
	solarArgRate = 0.01720 / MinutesPerDay
)

// initDeepSpace computes lunar-solar perturbation coefficients from the
// satellite's orbital elements. This is a simplified model suitable for
// visualization that captures the dominant secular effects.
func initDeepSpace(inclination, eccentricity, raan, argPerigee, meanMotion float64) *deepSpaceCorrections {
	ds := &deepSpaceCorrections{}

	cosI := math.Cos(inclination)
	sinI := math.Sin(inclination)

	// Perturbation amplitudes scale with the satellite's orbital geometry.
	// For near-equatorial orbits the inclination coupling is small; for
	// near-circular orbits the eccentricity coupling is small.
	e2 := eccentricity * eccentricity
	beta := math.Sqrt(1.0 - e2)

	// Lunar perturbation amplitude: proportional to cos(inclination)
	// and inversely proportional to the mean motion (larger orbit → slower → more affected).
	ds.lunarRAAN = lunarNodeRate * cosI / beta
	ds.lunarInclCorr = 0.5 * lunarNodeRate * sinI / beta

	// Solar perturbation amplitude: similar structure
	ds.solarRAAN = solarArgRate * cosI / beta
	ds.solarInclCorr = 0.5 * solarArgRate * sinI / beta

	// Secular drift rates for the three angular elements due to
	// combined third-body effects. These add to the J2 secular rates
	// already computed in the standard SGP4 initialization.
	//
	// RAAN precession from lunar-solar gravity
	ds.raanDotDS = ds.lunarRAAN + ds.solarRAAN

	// Argument of perigee rate contribution
	// Solar perturbation dominates for near-equatorial deep-space orbits
	ds.wDotDS = -ds.solarRAAN * (1.0 - 2.5*sinI*sinI)

	// Mean anomaly rate correction (very small)
	ds.mDotDS = -0.5 * ds.wDotDS * eccentricity

	return ds
}

// applyDeepSpace applies deep-space secular corrections to the orbital
// elements. It returns the adjusted RAAN, argument of perigee, and mean anomaly.
func applyDeepSpace(ds *deepSpaceCorrections, raan, w, M, tsince float64) (float64, float64, float64) {
	raanAdj := raan + ds.raanDotDS*tsince
	wAdj := w + ds.wDotDS*tsince
	mAdj := M + ds.mDotDS*tsince
	return raanAdj, wAdj, mAdj
}
