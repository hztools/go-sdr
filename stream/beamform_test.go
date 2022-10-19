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

package stream_test

import (
	"math"
	"math/cmplx"
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/rf"
	"hz.tools/sdr/stream"
)

func TestBeamformMath(t *testing.T) {
	rotations := stream.BeamformAngles(
		900*rf.MHz,
		0,
		[]float64{0, 1},
	)

	assert.InEpsilon(t, 1, real(rotations[0]), 1e-10)
	assert.InEpsilon(t, 1, real(rotations[1]), 1e-10)

	assert.InEpsilon(t, 1, 1+imag(rotations[0]), 1e-10)
	assert.InEpsilon(t, 1, 1+imag(rotations[1]), 1e-10)

	// rough numbers below taken from the math at
	// https://www.radartutorial.eu/06.antennas/Phased%20Array%20Antenna.en.html
	// as an independent check on the beamform computation

	freq10cm := rf.GHz * 2.997925
	d15cm := 0.15

	rotations = stream.BeamformAngles(
		freq10cm,
		40,
		[]float64{0, d15cm},
	)

	phase := 347.1 * math.Pi / 180
	assert.InEpsilon(t, phase, (2*math.Pi)+cmplx.Phase(cmplx.Conj(complex128(rotations[1]))), 1e-4)
}

func TestBeamformShortInput(t *testing.T) {
	rotations := stream.BeamformAngles(
		900*rf.MHz,
		0,
		[]float64{},
	)
	assert.Nil(t, rotations)

	rotations = stream.BeamformAngles2D(
		900*rf.MHz,
		0,
		[2]float64{0, 10},
		[][2]float64{},
	)
	assert.Nil(t, rotations)
}

func TestBeamformMath2D(t *testing.T) {
	rotations := stream.BeamformAngles2D(
		900*rf.MHz,
		0,
		[2]float64{0, 10},
		[][2]float64{
			[2]float64{0, 10},
			[2]float64{1, 10},
		},
	)

	assert.InEpsilon(t, 1, real(rotations[0]), 1e-10)
	assert.InEpsilon(t, 1, real(rotations[1]), 1e-10)

	assert.InEpsilon(t, 1, 1+imag(rotations[0]), 1e-10)
	assert.InEpsilon(t, 1, 1+imag(rotations[1]), 1e-10)

	// rough numbers below taken from the math at
	// https://www.radartutorial.eu/06.antennas/Phased%20Array%20Antenna.en.html
	// as an independent check on the beamform computation

	freq10cm := rf.GHz * 2.997925
	d15cm := 0.15

	rotations = stream.BeamformAngles(
		freq10cm,
		40,
		[]float64{0, d15cm},
	)

	phase := 347.1 * math.Pi / 180
	assert.InEpsilon(t, phase, (2*math.Pi)+cmplx.Phase(cmplx.Conj(complex128(rotations[1]))), 1e-4)
}

func TestBeamformMath2DWavelength(t *testing.T) {
	rotations := stream.BeamformAngles2D(
		299.792*rf.MHz,   // mostly just about 1m wavelength
		0,                // no beamform angle
		[2]float64{0, 0}, // synthetic center is at 0,0
		[][2]float64{
			[2]float64{0, 0},
			[2]float64{0, 10},
		},
	)

	assert.InEpsilon(t, 1, real(rotations[0]), 1e-10)
	assert.InEpsilon(t, 1, real(rotations[1]), 1e-10)

	rotations = stream.BeamformAngles2D(
		299.792*rf.MHz,   // mostly just about 1m wavelength
		0,                // no beamform angle
		[2]float64{0, 0}, // synthetic center is at 0,0
		[][2]float64{
			[2]float64{0, 0},
			[2]float64{10, 0},
		},
	)
	assert.InEpsilon(t, 1, real(rotations[0]), 1e-10)
	assert.InEpsilon(t, 1, real(rotations[1]), 1e-10)

	rotations = stream.BeamformAngles2D(
		299.792*rf.MHz,   // mostly just about 1m wavelength
		0,                // no beamform angle
		[2]float64{0, 0}, // synthetic center is at 0,0
		[][2]float64{
			[2]float64{0, 0},
			[2]float64{0, 0.5}, // we're 180 degrees out of phase
		},
	)
	assert.InEpsilon(t, math.Pi, cmplx.Phase(cmplx.Conj(complex128(rotations[1]))), 1e-4)

	rotations = stream.BeamformAngles2D(
		299.792*rf.MHz,   // mostly just about 1m wavelength
		0,                // no beamform angle
		[2]float64{0, 0}, // synthetic center is at 0,0
		[][2]float64{
			[2]float64{0, 0},
			[2]float64{0, 0.75}, // we're 270 degrees out of phase
		},
	)
	assert.InEpsilon(t, -(math.Pi / 2), cmplx.Phase(cmplx.Conj(complex128(rotations[1]))), 1e-4)
}

// vim: foldmethod=marker
