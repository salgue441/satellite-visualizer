package propagator

import (
	"math"
	"time"

	"satellite-visualizer/internal/domain"
)

// CK2 through CK4 are the un-normalized zonal harmonic values used in SGP4.
const (
	ck2  = 0.5 * J2 * AE * AE
	ck4  = -0.375 * J4 * AE * AE * AE * AE
	qoms24 = QOMS2T
	s4     = S
)

// sgp4State holds precomputed constants for a specific set of orbital elements.
type sgp4State struct {
	// Recovered orbital elements
	a0dp   float64 // recovered semi-major axis (earth radii)
	n0dp   float64 // recovered mean motion (rad/min)
	e0     float64 // initial eccentricity
	i0     float64 // initial inclination (rad)
	w0     float64 // initial argument of perigee (rad)
	raan0  float64 // initial RAAN (rad)
	M0     float64 // initial mean anomaly (rad)
	bstar  float64 // B* drag coefficient
	epoch  float64 // Julian date of epoch

	// Precomputed values
	cosio  float64
	sinio  float64
	eta    float64
	c1     float64
	c2     float64
	c3     float64
	c4     float64
	c5     float64
	d2     float64
	d3     float64
	d4     float64
	t2cof  float64
	t3cof  float64
	t4cof  float64
	t5cof  float64
	Mdot   float64
	wdot   float64
	raandot float64
	omgcof float64
	xmcof  float64
	xnodcf float64
	delM0  float64
	sinM0  float64
	isimp  bool // simple drag model flag
}

// SGP4Propagator implements the OrbitPropagator interface using the SGP4 algorithm.
type SGP4Propagator struct{}

// initSGP4 initializes the SGP4 state from orbital elements.
func initSGP4(elems domain.OrbitalElements) sgp4State {
	var st sgp4State

	st.epoch = elems.Epoch
	st.e0 = elems.Eccentricity
	st.i0 = elems.Inclination
	st.w0 = elems.ArgPerigee
	st.raan0 = elems.RAAN
	st.M0 = elems.MeanAnomaly
	st.bstar = elems.BStar

	// Convert mean motion from rev/day to rad/min
	n0 := elems.MeanMotion * TwoPi / MinutesPerDay

	cosio := math.Cos(st.i0)
	sinio := math.Sin(st.i0)
	st.cosio = cosio
	st.sinio = sinio

	cosio2 := cosio * cosio
	x3thm1 := 3.0*cosio2 - 1.0
	e0sq := st.e0 * st.e0
	betao2 := 1.0 - e0sq
	betao := math.Sqrt(betao2)

	// Recover original mean motion and semi-major axis from input elements
	a1 := math.Pow(XKE/n0, 2.0/3.0)
	d1 := 0.75 * J2 * (3.0*cosio2 - 1.0) / (betao * betao2) / (a1 * a1)
	a0 := a1 * (1.0 - d1/3.0 - d1*d1 - 134.0/81.0*d1*d1*d1)
	d0 := 0.75 * J2 * (3.0*cosio2 - 1.0) / (betao * betao2) / (a0 * a0)

	n0dp := n0 / (1.0 + d0)
	a0dp := a0 / (1.0 - d0)

	st.n0dp = n0dp
	st.a0dp = a0dp

	// Determine perigee height for atmospheric model selection
	perige := (a0dp*(1.0-st.e0) - AE) * EarthRadius

	// Adjust s and qoms2t for low-perigee satellites
	s4local := s4
	qoms24local := qoms24
	if perige < 156.0 {
		s4km := perige
		if perige < 98.0 {
			s4km = 20.0
		}
		s4local = s4km/EarthRadius + AE
		qoms24local = math.Pow((120.0-s4km)/EarthRadius, 4.0)
	}

	tsi := 1.0 / (a0dp - s4local)
	eta := a0dp * st.e0 * tsi
	st.eta = eta
	etasq := eta * eta
	eeta := st.e0 * eta
	psisq := math.Abs(1.0 - etasq)
	coef := qoms24local * math.Pow(tsi, 4.0)
	coef1 := coef / math.Pow(psisq, 3.5)

	c2 := coef1 * n0dp * (a0dp*(1.0+1.5*etasq+eeta*(4.0+etasq)) +
		0.75*ck2*tsi/psisq*x3thm1*(8.0+3.0*etasq*(8.0+etasq)))
	st.c2 = c2
	st.c1 = st.bstar * c2

	c3 := 0.0
	if st.e0 > 1.0e-4 {
		c3 = coef * tsi * J3 / J2 * n0dp * AE * sinio / st.e0
	}
	st.c3 = c3

	c4 := 2.0 * n0dp * coef1 * a0dp * betao2 *
		(eta*(2.0+0.5*etasq) + st.e0*(0.5+2.0*etasq) -
			2.0*ck2*tsi/(a0dp*psisq)*
				(-3.0*x3thm1*(1.0-2.0*eeta+etasq*(1.5-0.5*eeta))+
					0.75*(1.0-cosio2)*(2.0*etasq-eeta*(1.0+etasq))*math.Cos(2.0*st.w0)))
	st.c4 = c4

	c5 := 2.0 * coef1 * a0dp * betao2 * (1.0 + 2.75*(etasq+eeta) + eeta*etasq)
	st.c5 = c5

	theta4 := cosio2 * cosio2
	temp1 := 3.0 * ck2 * n0dp / (betao2 * betao)
	temp2 := temp1 * ck2 / betao2
	temp3 := 1.25 * ck4 * n0dp / (betao2 * betao2 * betao)

	Mdot := n0dp + 0.5*temp1*betao*x3thm1 +
		0.0625*temp2*betao*(13.0-78.0*cosio2+137.0*theta4)
	st.Mdot = Mdot

	x1m5th := 1.0 - 5.0*cosio2
	wdot := -0.5*temp1*x1m5th +
		0.0625*temp2*(7.0-114.0*cosio2+395.0*theta4) +
		temp3*(3.0-36.0*cosio2+49.0*theta4)
	st.wdot = wdot

	xhdot1 := -temp1 * cosio
	raandot := xhdot1 + (0.5*temp2*(4.0-19.0*cosio2)+2.0*temp3*(3.0-7.0*cosio2))*cosio
	st.raandot = raandot

	// Check if we should use simplified drag model
	st.isimp = (a0dp*(1.0-st.e0) < (220.0/EarthRadius + AE))

	if !st.isimp {
		c1sq := st.c1 * st.c1
		st.d2 = 4.0 * a0dp * tsi * c1sq
		temp := st.d2 * tsi * st.c1 / 3.0
		st.d3 = (17.0*a0dp + s4local) * temp
		st.d4 = 0.5 * temp * a0dp * tsi * (221.0*a0dp + 31.0*s4local) * st.c1
		st.t3cof = st.d2 + 2.0*c1sq
		st.t4cof = 0.25 * (3.0*st.d3 + st.c1*(12.0*st.d2+10.0*c1sq))
		st.t5cof = 0.2 * (3.0*st.d4 + 12.0*st.c1*st.d3 + 6.0*st.d2*st.d2 + 15.0*c1sq*(2.0*st.d2+c1sq))
	}

	st.sinM0 = math.Sin(st.M0)
	if st.e0 > 1.0e-4 {
		st.omgcof = st.bstar * c3 * math.Cos(st.w0)
		st.xmcof = -2.0 / 3.0 * coef * st.bstar * AE / eeta
		st.delM0 = math.Pow(1.0+eta*math.Cos(st.M0), 3.0)
	}
	st.xnodcf = 3.5 * betao2 * xhdot1 * st.c1
	st.t2cof = 1.5 * st.c1

	return st
}

// Propagate calculates the ECI position and velocity of a satellite
// at the given time using the SGP4 algorithm.
func (p *SGP4Propagator) Propagate(
	elements domain.OrbitalElements,
	t time.Time,
) (domain.Position, domain.Velocity, error) {
	st := initSGP4(elements)

	// Compute minutes since epoch
	targetJD := JulianDate(t)
	tsince := MinutesSinceEpoch(st.epoch, targetJD)

	// Secular updates
	xmdf := st.M0 + st.Mdot*tsince
	wdf := st.w0 + st.wdot*tsince
	xnoddf := st.raan0 + st.raandot*tsince

	// Apply deep-space lunar-solar corrections if applicable
	if IsDeepSpace(elements.MeanMotion) {
		ds := initDeepSpace(elements.Inclination, elements.Eccentricity,
			elements.RAAN, elements.ArgPerigee, elements.MeanMotion)
		xnoddf, wdf, xmdf = applyDeepSpace(ds, xnoddf, wdf, xmdf, tsince)
	}

	omega := wdf
	xmp := xmdf

	tsq := tsince * tsince
	xnode := xnoddf + st.xnodcf*tsq
	tempa := 1.0 - st.c1*tsince
	tempe := st.bstar * st.c4 * tsince
	templ := st.t2cof * tsq

	if !st.isimp {
		tcube := tsq * tsince
		t4 := tcube * tsince

		delomg := st.omgcof * tsince
		delm := 0.0
		if st.e0 > 1.0e-4 {
			delm = st.xmcof * (math.Pow(1.0+st.eta*math.Cos(xmdf), 3.0) - st.delM0)
		}
		temp := delomg + delm
		xmp = xmdf + temp
		omega = wdf - temp

		tempa = tempa - st.d2*tsq - st.d3*tcube - st.d4*t4
		tempe = tempe + st.bstar*st.c5*(math.Sin(xmp)-st.sinM0)
		templ = templ + st.t3cof*tcube + t4*(st.t4cof+tsince*st.t5cof)
	}

	// Update for secular gravity and atmospheric drag
	a := st.a0dp * tempa * tempa
	e := st.e0 - tempe
	xl := xmp + omega + xnode + st.n0dp*templ

	// Clamp eccentricity
	if e < 1.0e-6 {
		e = 1.0e-6
	}
	if e >= 1.0 {
		return domain.Position{}, domain.Velocity{}, domain.ErrCalculationFailed
	}

	beta := math.Sqrt(1.0 - e*e)
	xn := XKE / math.Pow(a, 1.5) // mean motion

	// Short-period periodics from J2
	axn := e * math.Cos(omega)
	temp := 1.0 / (a * beta * beta)
	aynl := temp * ck2 * 0.5 * st.sinio * st.cosio
	ayn := e*math.Sin(omega) + aynl

	// Mean anomaly + argument of perigee for Kepler solve
	capu := WrapTwoPi(xl - xnode)

	// Solve modified Kepler equation: epw = capu + axn*sin(epw) - ayn*cos(epw)
	// Using Newton-Raphson iteration
	epw := capu
	for i := 0; i < 10; i++ {
		sinepw := math.Sin(epw)
		cosepw := math.Cos(epw)
		f := epw - axn*sinepw + ayn*cosepw - capu
		fd := 1.0 - axn*cosepw - ayn*sinepw
		delta := f / fd
		epw -= delta
		if math.Abs(delta) < 1.0e-12 {
			break
		}
	}

	// Short period preliminary quantities
	sinepw := math.Sin(epw)
	cosepw := math.Cos(epw)

	ecose := axn*cosepw + ayn*sinepw
	esine := axn*sinepw - ayn*cosepw
	elsq := axn*axn + ayn*ayn
	temp1 := 1.0 - elsq
	pl := a * temp1
	if pl < 0.0 {
		return domain.Position{}, domain.Velocity{}, domain.ErrCalculationFailed
	}
	r := a * (1.0 - ecose)
	rdot := XKE * math.Sqrt(a) * esine / r
	rfdot := XKE * math.Sqrt(pl) / r

	temp2 := 1.0 / pl
	betal := math.Sqrt(temp1)
	cosu := a / r * (cosepw - axn + ayn*esine*temp2/(1.0+betal))
	sinu := a / r * (sinepw - ayn - axn*esine*temp2/(1.0+betal))
	u := math.Atan2(sinu, cosu)
	sin2u := 2.0 * sinu * cosu
	cos2u := 2.0*cosu*cosu - 1.0

	// Short-period corrections
	temp3 := temp2 * ck2
	rk := r*(1.0-1.5*temp3*betal*(3.0*st.cosio*st.cosio-1.0)) + 0.5*temp3*(1.0-st.cosio*st.cosio)*cos2u
	uk := u - 0.25*temp3*(7.0*st.cosio*st.cosio-1.0)*sin2u
	xnodek := xnode + 1.5*temp3*st.cosio*sin2u
	xinck := st.i0 + 1.5*temp3*st.cosio*st.sinio*cos2u
	rdotk := rdot - xn*temp3*(1.0-st.cosio*st.cosio)*sin2u
	rfdotk := rfdot + xn*temp3*((1.0-st.cosio*st.cosio)*cos2u+1.5*(3.0*st.cosio*st.cosio-1.0))

	// Orientation vectors
	sinuk := math.Sin(uk)
	cosuk := math.Cos(uk)
	sinik := math.Sin(xinck)
	cosik := math.Cos(xinck)
	sinnok := math.Sin(xnodek)
	cosnok := math.Cos(xnodek)

	xmx := -sinnok * cosik
	xmy := cosnok * cosik

	ux := xmx*sinuk + cosnok*cosuk
	uy := xmy*sinuk + sinnok*cosuk
	uz := sinik * sinuk

	vx := xmx*cosuk - cosnok*sinuk
	vy := xmy*cosuk - sinnok*sinuk
	vz := sinik * cosuk

	// Position in km and velocity in km/s
	posX := rk * ux * EarthRadius
	posY := rk * uy * EarthRadius
	posZ := rk * uz * EarthRadius

	velX := (rdotk*ux + rfdotk*vx) * EarthRadius / 60.0
	velY := (rdotk*uy + rfdotk*vy) * EarthRadius / 60.0
	velZ := (rdotk*uz + rfdotk*vz) * EarthRadius / 60.0

	return domain.Position{X: posX, Y: posY, Z: posZ},
		domain.Velocity{X: velX, Y: velY, Z: velZ},
		nil
}
