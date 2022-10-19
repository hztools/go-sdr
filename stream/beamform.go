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
	"math/cmplx"

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
	if len(antennas) == 0 {
		return nil
	}

	var (
		ret = make([]complex64, len(antennas))
	)

	for i, antenna := range antennas {
		var (
			// "natural" triangle
			nDistance = computeDistance(antenna, center)
		)

		// before we go any further, if the distance between the synthetic
		// point and the antenna point is 0, it's not worth going on, since
		// the trig will go whacky anyway.
		if nDistance == 0 {
			ret[i] = 1 // really 1+0i
			continue
		}

		var (
			angleR = angle * (math.Pi / 180)

			nOpposite = (antenna[1] - center[1])
			nThetaR   = math.Asin(nOpposite / nDistance)

			// "phase" triangle
			pDistance = nDistance
			pThetaR   = (nThetaR + angleR)
			pOpposite = math.Sin(pThetaR) * pDistance

			phaseShift  = (pOpposite / frequency.Wavelength()) * 360
			phaseShiftR = phaseShift * (math.Pi / 180)
		)

		ret[i] = complex64(cmplx.Conj(complex(
			math.Cos(phaseShiftR), // "Adjacent"
			math.Sin(phaseShiftR), // "Opposite"
		)))
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
	if len(distances) == 0 {
		return nil
	}
	antennas := make([][2]float64, len(distances))
	for i := range antennas {
		antennas[i] = [2]float64{distances[i], 0}
	}
	return BeamformAngles2D(frequency, angle, antennas[0], antennas)
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
