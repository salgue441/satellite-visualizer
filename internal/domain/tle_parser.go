package domain

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ParseTLE parses two TLE lines and returns the extracted orbital elements.
// All angular values are converted to radians. Epoch is returned as a Julian Date.
func ParseTLE(line1, line2 string) (OrbitalElements, error) {
	var elems OrbitalElements

	// Line length validation
	if len(line1) != 69 || len(line2) != 69 {
		return elems, ErrInvalidTle
	}

	// Line number validation
	if line1[0] != '1' || line2[0] != '2' {
		return elems, ErrInvalidTle
	}

	// Checksum validation
	if !validChecksum(line1) || !validChecksum(line2) {
		return elems, ErrInvalidTle
	}

	// Parse Line 1
	catNo, err := strconv.Atoi(strings.TrimSpace(line1[2:7]))
	if err != nil {
		return elems, fmt.Errorf("%w: catalog number: %v", ErrInvalidTle, err)
	}
	elems.NoradCatNo = catNo

	// Epoch year + day (cols 19-32, 0-indexed: 18:32)
	epochYear, err := strconv.Atoi(strings.TrimSpace(line1[18:20]))
	if err != nil {
		return elems, fmt.Errorf("%w: epoch year: %v", ErrInvalidTle, err)
	}
	epochDay, err := strconv.ParseFloat(strings.TrimSpace(line1[20:32]), 64)
	if err != nil {
		return elems, fmt.Errorf("%w: epoch day: %v", ErrInvalidTle, err)
	}
	elems.Epoch = epochToJulian(epochYear, epochDay)

	// BStar (cols 54-61, 0-indexed: 53:61)
	bstar, err := parseBStar(line1[53:61])
	if err != nil {
		return elems, fmt.Errorf("%w: bstar: %v", ErrInvalidTle, err)
	}
	elems.BStar = bstar

	// Element set number (cols 65-68, 0-indexed: 64:68)
	elemSetNo, err := strconv.Atoi(strings.TrimSpace(line1[64:68]))
	if err != nil {
		return elems, fmt.Errorf("%w: element set no: %v", ErrInvalidTle, err)
	}
	elems.ElementSetNo = elemSetNo

	// Parse Line 2
	deg2rad := math.Pi / 180.0

	inclination, err := strconv.ParseFloat(strings.TrimSpace(line2[8:16]), 64)
	if err != nil {
		return elems, fmt.Errorf("%w: inclination: %v", ErrInvalidTle, err)
	}
	elems.Inclination = inclination * deg2rad

	raan, err := strconv.ParseFloat(strings.TrimSpace(line2[17:25]), 64)
	if err != nil {
		return elems, fmt.Errorf("%w: raan: %v", ErrInvalidTle, err)
	}
	elems.RAAN = raan * deg2rad

	// Eccentricity (cols 27-33, 0-indexed: 26:33) — implied leading "0."
	eccStr := "0." + strings.TrimSpace(line2[26:33])
	ecc, err := strconv.ParseFloat(eccStr, 64)
	if err != nil {
		return elems, fmt.Errorf("%w: eccentricity: %v", ErrInvalidTle, err)
	}
	elems.Eccentricity = ecc

	argPerigee, err := strconv.ParseFloat(strings.TrimSpace(line2[34:42]), 64)
	if err != nil {
		return elems, fmt.Errorf("%w: arg perigee: %v", ErrInvalidTle, err)
	}
	elems.ArgPerigee = argPerigee * deg2rad

	meanAnomaly, err := strconv.ParseFloat(strings.TrimSpace(line2[43:51]), 64)
	if err != nil {
		return elems, fmt.Errorf("%w: mean anomaly: %v", ErrInvalidTle, err)
	}
	elems.MeanAnomaly = meanAnomaly * deg2rad

	meanMotion, err := strconv.ParseFloat(strings.TrimSpace(line2[52:63]), 64)
	if err != nil {
		return elems, fmt.Errorf("%w: mean motion: %v", ErrInvalidTle, err)
	}
	elems.MeanMotion = meanMotion

	revNo, err := strconv.Atoi(strings.TrimSpace(line2[63:68]))
	if err != nil {
		return elems, fmt.Errorf("%w: revolution no: %v", ErrInvalidTle, err)
	}
	elems.RevolutionNo = revNo

	return elems, nil
}

// validChecksum verifies the TLE line checksum.
// Sum all digits; '-' counts as 1; all other non-digit chars count as 0.
// Result mod 10 must equal the last character.
func validChecksum(line string) bool {
	if len(line) < 1 {
		return false
	}

	sum := 0
	for i := 0; i < len(line)-1; i++ {
		c := line[i]
		if c >= '0' && c <= '9' {
			sum += int(c - '0')
		} else if c == '-' {
			sum += 1
		}
	}

	expected := byte('0' + sum%10)
	return line[len(line)-1] == expected
}

// parseBStar parses the BStar drag term from TLE format.
// Format example: " 25302-4" means 0.25302e-4.
func parseBStar(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" || s == "00000-0" || s == "00000+0" {
		return 0.0, nil
	}

	// Find the position of the exponent sign (last '+' or '-')
	expIdx := -1
	for i := len(s) - 1; i >= 1; i-- {
		if s[i] == '+' || s[i] == '-' {
			expIdx = i
			break
		}
	}

	if expIdx == -1 {
		// No exponent, treat as implied decimal: 0.NNNNN
		val, err := strconv.ParseFloat("0."+s, 64)
		return val, err
	}

	mantissa := s[:expIdx]
	exponent := s[expIdx:]

	// Handle leading sign on mantissa
	sign := 1.0
	if mantissa[0] == '-' {
		sign = -1.0
		mantissa = mantissa[1:]
	} else if mantissa[0] == '+' {
		mantissa = mantissa[1:]
	}

	mantissaVal, err := strconv.ParseFloat("0."+mantissa, 64)
	if err != nil {
		return 0, err
	}

	expVal, err := strconv.Atoi(exponent)
	if err != nil {
		return 0, err
	}

	return sign * mantissaVal * math.Pow(10, float64(expVal)), nil
}

// epochToJulian converts a TLE epoch (2-digit year + fractional day) to Julian Date.
func epochToJulian(twoDigitYear int, fractionalDay float64) float64 {
	var year int
	if twoDigitYear <= 56 {
		year = 2000 + twoDigitYear
	} else {
		year = 1900 + twoDigitYear
	}

	// Calculate Julian Date of January 1 of the given year.
	// Using the standard formula for JD at 0h UT on Jan 1.
	jdJan1 := julianDateOfJan1(year)

	return jdJan1 + fractionalDay - 1.0
}

// julianDateOfJan1 returns the Julian Date at 0h UT on January 1 of the given year.
// Uses the Meeus algorithm (ch. 7) for Gregorian calendar to JD conversion.
// For January, month <= 2, so we use Y=year-1, M=13 in the formula.
func julianDateOfJan1(year int) float64 {
	y := float64(year - 1)
	a := math.Floor(y / 100.0)
	b := 2.0 - a + math.Floor(a/4.0)
	return math.Floor(365.25*(y+4716.0)) + math.Floor(30.6001*14.0) + 1.0 + b - 1524.5
}
