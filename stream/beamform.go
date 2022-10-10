// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2022
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE. }}}

package stream

import (
	"fmt"
	"math"

	"hz.tools/rf"
	"hz.tools/sdr"
)

// Beamform will combine a set of sdr.Readers into a single sdr.Reader using
// the provided phase angles to stear the beam.
type Beamform struct {
	sdr.Reader

	readers sdr.Readers
	config  BeamformConfig
}

// computeDistance will compute the distance between two points
func computeDistance(p1, p2 [2]float64) float64 {
	var (
		xd = p1[0] - p2[0]
		xy = p1[1] - p2[1]
	)
	return math.Sqrt((xd * xd) + (xy * xy))
}

// BeamformAngles2D will determine what the phase offset required by
// the Beamform Reader.
//
//   - frequency is the *rf* frequency (not IF).
//   - angle is in *degrees*
//   - ref is the X/Y coordinate in meters of the 'center' (can be anywhere).
//   - antennas are X/Y coordinates in *meters*
func BeamformAngles2D(
	frequency rf.Hz,
	angle float64,
	center [2]float64,
	antennas [][2]float64,
) []complex64 {
	var (
		// radians = degrees * π/180 (really, (tau/360, but...)
		angleR   float64 = angle * (math.Pi / 180)
		angleSin float64 = math.Sin(angleR)

		ret = make([]complex64, len(antennas))
	)

	for i, antenna := range antennas {
		var (
			distance = computeDistance(center, antenna)

			// phaseShift is in Degrees
			phaseShift  = (360 * distance * angleSin) / frequency.Wavelength()
			phaseShiftR = phaseShift * (math.Pi / 180)
		)

		// Right, so we have degrees, let's work out the complex value here.
		// We know the magnitude is "1" (unit square, no gain), but we need
		// to rotate the 1+0j by the number of degrees.
		//
		// Going back to trig, we have the hypotenuse, and the angle, and
		// we need to work out the opposite and adjacent lengths of the
		// right triangle.
		//
		//         /+ <-- cmplx here
		//        / |
		//       /  |
		//    1 /   |
		//     /    | <--- "Opposite" (Imag)
		//    /     |
		//   /      |
		//  +-------+
		//  ^      \_______ "Adjacent" (Real)
		//   0+0i
		//

		ret[i] = complex(
			float32(math.Cos(phaseShiftR)), // "Adjacent"
			float32(math.Sin(phaseShiftR)), // "Opposite"
		)
	}

	return ret
}

// BeamformAngles will determine what the phase offset required by
// the Beamform Reader.
//
//   - frequency is the *rf* frequency (not IF).
//   - angle is in *degrees*
//   - distances is in *meters*
func BeamformAngles(
	frequency rf.Hz,
	angle float64,
	distances []float64,
) []complex64 {
	// Let's first take in the data we have and convert
	// it as required.

	var (
		// radians = degrees * π/180 (really, (tau/360, but...)
		angleR   float64 = angle * (math.Pi / 180)
		angleSin float64 = math.Sin(angleR)

		ret = make([]complex64, len(distances))
	)

	for i, distance := range distances {
		var (
			// phaseShift is in Degrees
			phaseShift  = (360 * distance * angleSin) / frequency.Wavelength()
			phaseShiftR = phaseShift * (math.Pi / 180)
		)

		// Right, so we have degrees, let's work out the complex value here.
		// We know the magnitude is "1" (unit square, no gain), but we need
		// to rotate the 1+0j by the number of degrees.
		//
		// Going back to trig, we have the hypotenuse, and the angle, and
		// we need to work out the opposite and adjacent lengths of the
		// right triangle.
		//
		//         /+ <-- cmplx here
		//        / |
		//       /  |
		//    1 /   |
		//     /    | <--- "Opposite" (Imag)
		//    /     |
		//   /      |
		//  +-------+
		//  ^      \_______ "Adjacent" (Real)
		//   0+0i
		//

		ret[i] = complex(
			float32(math.Cos(phaseShiftR)), // "Adjacent"
			float32(math.Sin(phaseShiftR)), // "Opposite"
		)
	}

	return ret
}

// SetPhaseAngles will set the phase angle to shift every stream by.
func (b *Beamform) SetPhaseAngles(angles []complex64) error {
	if len(angles) != len(b.readers) {
		return fmt.Errorf("Beamform.SetPhaseAngles: angles must match the reader length")
	}
	for i, reader := range b.readers {
		reader.(*multiplyReader).SetMultiplier(angles[i])
	}
	return nil
}

// BeamformConfig contains configuration for the combined samples.
type BeamformConfig struct {
	Angles []complex64
}

// ReadBeamform will create a new sdr.Reader from a series of coherent
// sdr.Readers using the provided phase angles.
func ReadBeamform(rs sdr.Readers, cfg BeamformConfig) (*Beamform, error) {
	multReaders := make(sdr.Readers, len(rs))
	for i := range rs {
		reader, err := ConvertReader(rs[i], sdr.SampleFormatC64)
		if err != nil {
			return nil, err
		}
		multReaders[i], err = Multiply(reader, 1)
		if err != nil {
			return nil, err
		}
	}
	addReader, err := Add(multReaders...)
	if err != nil {
		return nil, err
	}
	b := &Beamform{
		Reader:  addReader,
		readers: multReaders,
		config:  cfg,
	}
	b.SetPhaseAngles(cfg.Angles)
	return b, nil
}

// vim: foldmethod=marker
