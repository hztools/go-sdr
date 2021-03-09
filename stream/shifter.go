// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2021
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
	"math"
	"math/cmplx"

	"hz.tools/rf"
	"hz.tools/sdr"
)

type shiftReader struct {
	inc   float64
	ts    float64
	shift float64
	r     sdr.Reader
}

func (sr *shiftReader) SampleFormat() sdr.SampleFormat {
	return sr.r.SampleFormat()
}

func (sr *shiftReader) SampleRate() uint32 {
	return sr.r.SampleRate()
}

func (sr *shiftReader) Read(s sdr.Samples) (int, error) {
	switch s.Format() {
	case sdr.SampleFormatC64:
		break
	default:
		return 0, sdr.ErrSampleFormatUnknown
	}

	n, err := sr.r.Read(s)
	if err != nil {
		return n, err
	}

	// TODO(paultag): Fix this to be safe when the above format checks
	// grow.
	sC64 := s.Slice(0, n).(sdr.SamplesC64)
	tau := math.Pi * 2

	for j := range sC64 {
		sr.ts += sr.inc
		if sr.ts > tau {
			sr.ts -= tau
		}
		sC64[j] = sC64[j] * complex64(cmplx.Exp(complex(0, tau*sr.shift*sr.ts)))
	}

	return n, nil
}

// ShiftReader will shift the iq samples by the target frequency. So a carrier
// at the provided shift frequency offset will be read through at DC.
func ShiftReader(r sdr.Reader, shift rf.Hz) (sdr.Reader, error) {
	switch r.SampleFormat() {
	case sdr.SampleFormatC64:
		break
	default:
		return nil, sdr.ErrSampleFormatUnknown
	}

	return &shiftReader{
		ts:    0,
		inc:   (1 / float64(r.SampleRate())),
		shift: float64(shift),
		r:     r,
	}, nil
}

// vim: foldmethod=marker
