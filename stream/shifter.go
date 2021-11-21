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

	"hz.tools/rf"
	"hz.tools/sdr"
)

type shiftReader struct {
	r     sdr.Reader
	shift rf.Hz
	fn    func(rf.Hz, sdr.SamplesC64)
}

func (sr *shiftReader) SampleFormat() sdr.SampleFormat {
	return sr.r.SampleFormat()
}

func (sr *shiftReader) SampleRate() uint {
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
	sr.fn(sr.shift, sC64)
	return n, nil
}

// ShiftBuffer will return a function that will track phase to shift consecutive
// buffers.
func ShiftBuffer(sampleRate uint) func(rf.Hz, sdr.SamplesC64) {
	var (
		ts  float64
		inc float64 = (1 / float64(sampleRate))
		tau         = math.Pi * 2
	)

	return func(freq rf.Hz, buf sdr.SamplesC64) {
		shift := float64(freq)
		for j := range buf {
			ts += inc
			if ts > tau {
				ts -= tau
			}

			im, rl := math.Sincos(tau * shift * ts)
			buf[j] = buf[j] * complex64(complex(rl, im))
		}
	}
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
		r:     r,
		shift: shift,
		fn:    ShiftBuffer(r.SampleRate()),
	}, nil
}

// vim: foldmethod=marker
