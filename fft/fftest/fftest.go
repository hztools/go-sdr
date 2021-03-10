// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2020
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

package fftest

import (
	"math"
	"math/cmplx"
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/rf"
	"hz.tools/sdr"
	"hz.tools/sdr/fft"
)

func generateCw(buf sdr.SamplesC64, freq rf.Hz, sampleRate int, phase float64) {
	var (
		carrierFreq float64 = float64(freq)
		tau                 = math.Pi * 2
	)

	for i := range buf {
		now := float64(i) / float64(sampleRate)
		buf[i] = complex64(complex(
			math.Cos(tau*carrierFreq*now+phase),
			math.Sin(tau*carrierFreq*now+phase),
		))
	}
}

type testFrequencies struct {
	Frequency rf.Hz
	Index     int
}

// Run will run the standard FFT tests against the provided Planner.
func Run(t *testing.T, planner fft.Planner) {
	t.Run("ForwardFFT", func(t *testing.T) {
		testForwardFFT(t, planner)
	})

	t.Run("BackwardFFT", func(t *testing.T) {
		testBackwardFFT(t, planner)
	})

	t.Run("MismatchedSamples", func(t *testing.T) {
		testMismatchDstFFT(t, planner)
	})
}

func testForwardFFT(t *testing.T, planner fft.Planner) {
	cwPhase0 := make(sdr.SamplesC64, 1024)
	out := make([]complex64, 1024)

	for _, tfreq := range []testFrequencies{
		testFrequencies{Frequency: rf.Hz(10), Index: 0},
		testFrequencies{Frequency: rf.Hz(900000), Index: 512},
		testFrequencies{Frequency: rf.Hz(450000), Index: 256},
		testFrequencies{Frequency: rf.Hz(225000), Index: 128},
	} {
		generateCw(cwPhase0, tfreq.Frequency, 1.8e6, 0)

		plan, err := planner(cwPhase0, out, fft.Forward)
		assert.NoError(t, err)
		assert.NoError(t, plan.Transform())
		assert.NoError(t, plan.Close())

		var (
			powerMax float64 = 0
			powerI   int     = -1
		)
		power := make([]float64, cwPhase0.Length())
		for i := range power {
			power[i] = cmplx.Abs(complex128(out[i]))
			if power[i] > powerMax {
				powerMax = power[i]
				powerI = i
			}
		}
		assert.Equal(t, tfreq.Index, powerI)
	}
}

func testBackwardFFT(t *testing.T, planner fft.Planner) {
	for _, bin := range []int{
		5, 10, 127, 522, 242, 415, 825,
	} {
		var err error

		iq := make(sdr.SamplesC64, 1024)
		freq := make([]complex64, 1024)

		freq[bin] = 1 + 1i

		plan, err := planner(iq, freq, fft.Backward)
		assert.NoError(t, err)
		assert.NoError(t, plan.Transform())
		assert.NoError(t, plan.Close())

		freq[bin] = 0 + 0i

		plan, err = planner(iq, freq, fft.Forward)
		assert.NoError(t, err)
		assert.NoError(t, plan.Transform())
		assert.NoError(t, plan.Close())

		var (
			powerMax float64 = 0
			powerI   int     = -1
		)
		power := make([]float64, len(freq))
		for i := range power {
			power[i] = cmplx.Abs(complex128(freq[i]))
			if power[i] > powerMax {
				powerMax = power[i]
				powerI = i
			}
		}

		assert.Equal(t, bin, powerI)
	}
}

func testMismatchDstFFT(t *testing.T, planner fft.Planner) {
	iq := make(sdr.SamplesC64, 1024)
	freq := make([]complex64, 128)
	_, err := planner(iq, freq, fft.Forward)
	assert.Equal(t, sdr.ErrDstTooSmall, err)

	iq = make(sdr.SamplesC64, 128)
	freq = make([]complex64, 1024)
	_, err = planner(iq, freq, fft.Backward)
	assert.Equal(t, sdr.ErrDstTooSmall, err)

}

// vim: foldmethod=marker
