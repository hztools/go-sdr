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

package stream

import (
	"hz.tools/sdr"
	"hz.tools/sdr/internal/simd"
)

// Gain will apply a gain scaler to all iq samples in the reader, as the values
// are read.
func Gain(r sdr.Reader, v float32) sdr.Reader {
	return &gain{v: v, r: r}
}

type gain struct {
	v float32
	r sdr.Reader
}

func (g *gain) Scale(s sdr.Samples) error {
	// TODO(paultag): SIMD low hanging fruit

	switch s := s.(type) {
	case sdr.SamplesC64:
		simd.ScaleComplex(g.v, s)
		return nil
	// add int16 and uint8
	default:
		return sdr.ErrSampleFormatUnknown
	}
}

func (g *gain) Read(s sdr.Samples) (int, error) {
	i, err := g.r.Read(s)
	if err != nil {
		return i, err
	}
	err = g.Scale(s.Slice(0, i))
	return i, err
}

func (g *gain) SampleFormat() sdr.SampleFormat {
	return g.r.SampleFormat()
}

func (g *gain) SampleRate() uint32 {
	return g.r.SampleRate()
}

// vim: foldmethod=marker
