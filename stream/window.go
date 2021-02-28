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

// +build sdr.experimental

package stream

import (
	"math"

	"hz.tools/sdr"
	"hz.tools/sdr/internal/warning"
)

type windowWriter struct {
	sdr.Writer
	cachedWindows map[int][]float32
}

func (wr *windowWriter) generateWindow(size int) []float32 {
	var (
		buf         = make([]float32, size)
		a0  float64 = 0.42
		a1  float64 = 0.5
		a2  float64 = 0.08
	)

	for i := range buf {
		buf[i] = float32(a0 -
			(a1 * math.Cos((sdr.Tau*float64(i))/float64(size))) +
			(a2 * math.Cos((sdr.Tau*2*float64(i))/float64(size))))
	}

	return buf
}

func (wr *windowWriter) getCachedWindow(size int) []float32 {
	if buf, ok := wr.cachedWindows[size]; ok {
		return buf
	}

	buf := wr.generateWindow(size)
	wr.cachedWindows[size] = buf
	return buf
}

func (wr *windowWriter) writeC64(s sdr.SamplesC64) (int, error) {
	window := wr.getCachedWindow(s.Length())

	for i := range s {
		s[i] = complex(
			real(s[i])*window[i],
			imag(s[i])*window[i],
		)
	}

	return wr.Writer.Write(s)
}

func (wr *windowWriter) Write(s sdr.Samples) (int, error) {
	switch s := s.(type) {
	case sdr.SamplesC64:
		return wr.writeC64(s)
	default:
		return 0, sdr.ErrSampleFormatUnknown
	}
}

// WindowWriter is an *EXPERIMENTAL* Function
//
// WindowWriter will run all IQ samples through the Blackman windowing function.
func WindowWriter(w sdr.Writer) (sdr.Writer, error) {
	warning.Experimental("WindowWriter")

	if w.SampleFormat() != sdr.SampleFormatC64 {
		return nil, sdr.ErrSampleFormatUnknown
	}

	return &windowWriter{
		Writer:        w,
		cachedWindows: map[int][]float32{},
	}, nil
}

// vim: foldmethod=marker
