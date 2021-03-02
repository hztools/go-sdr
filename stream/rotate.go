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
)

type multiplyReader struct {
	m complex64
	r sdr.Reader
}

func (mr *multiplyReader) SampleFormat() sdr.SampleFormat {
	return mr.r.SampleFormat()
}

func (mr *multiplyReader) SampleRate() uint32 {
	return mr.r.SampleRate()
}

func (mr *multiplyReader) Read(s sdr.Samples) (int, error) {
	switch s.Format() {
	case sdr.SampleFormatC64:
		break
	default:
		return 0, sdr.ErrSampleFormatUnknown
	}

	i, err := mr.r.Read(s)
	if err != nil {
		return i, err
	}

	// TODO(paultag): Fix this to be safe when the above format checks
	// grow.
	sC64 := s.(sdr.SamplesC64)
	sC64.Multiply(mr.m)
	// vfor j := range sC64 {
	// v	sC64[j] = sC64[j] * mr.m
	// v}

	return i, nil
}

// Multiply will multiply each iq sample by the value m. This will 'rotate'
// each sample by the defined amount.
func Multiply(r sdr.Reader, m complex64) (sdr.Reader, error) {
	switch r.SampleFormat() {
	case sdr.SampleFormatC64:
		break
	default:
		return nil, sdr.ErrSampleFormatUnknown
	}

	return &multiplyReader{r: r, m: m}, nil
}

// vim: foldmethod=marker
