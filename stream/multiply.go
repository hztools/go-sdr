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

// SetMultiplier is an undocumented API to update the complex value
// after the construction of the Reader.
func (mr *multiplyReader) SetMultiplier(m complex64) {
	mr.m = m
}

func (mr *multiplyReader) SampleFormat() sdr.SampleFormat {
	return mr.r.SampleFormat()
}

func (mr *multiplyReader) SampleRate() uint {
	return mr.r.SampleRate()
}

func (mr *multiplyReader) Read(s sdr.Samples) (int, error) {
	switch s.Format() {
	case sdr.SampleFormatC64:
		break
	default:
		return 0, sdr.ErrSampleFormatMismatch
	}

	i, err := mr.r.Read(s)
	if err != nil {
		return i, err
	}

	if mr.m == 1 {
		// Don't bother spending time multiplying if it won't do anything.
		return i, nil
	}

	// TODO(paultag): Fix this to be safe when the above format checks
	// grow.
	sC64 := s.Slice(0, i).(sdr.SamplesC64)
	sC64.Multiply(mr.m)

	return i, nil
}

// Multiply will multiply each iq sample by the value m. This will 'rotate'
// each sample by the defined amount.
func Multiply(r sdr.Reader, m complex64) (sdr.Reader, error) {
	switch r.SampleFormat() {
	case sdr.SampleFormatI8:
		ret := &int8MultiplyReader{r: r}
		ret.SetMultiplier(m)
		return ret, nil
	case sdr.SampleFormatU8:
		ret := &uint8MultiplyReader{r: r}
		ret.SetMultiplier(m)
		return ret, nil
	case sdr.SampleFormatC64:
		return &multiplyReader{r: r, m: m}, nil
	default:
		return nil, sdr.ErrSampleFormatUnknown
	}
}

type uint8MultiplyReader struct {
	m complex64

	// tab is an index for a complex real/imag value, returning the rotated
	// real/imag vaule. This is a memory (and one-time CPU) tradeoff to
	// generate every rotation up-front.
	tab sdr.SamplesU8

	r sdr.Reader
}

func (mr *uint8MultiplyReader) mult(x [2]uint8) [2]uint8 {
	return mr.tab[mr.index(x)]
}

func (mr *uint8MultiplyReader) index(x [2]uint8) int {
	return (int(x[0]) * 255) + int(x[1])
}

func (mr *uint8MultiplyReader) SampleFormat() sdr.SampleFormat {
	return mr.r.SampleFormat()
}

func (mr *uint8MultiplyReader) SampleRate() uint {
	return mr.r.SampleRate()
}

func (mr *uint8MultiplyReader) Read(s sdr.Samples) (int, error) {
	switch s.Format() {
	case sdr.SampleFormatU8:
		break
	default:
		return 0, sdr.ErrSampleFormatMismatch
	}

	i, err := mr.r.Read(s)
	if err != nil {
		return i, err
	}

	// TODO(paultag): Fix this to be safe when the above format checks
	// grow.
	sU8 := s.Slice(0, i).(sdr.SamplesU8)
	for i := range sU8 {
		sU8[i] = mr.mult(sU8[i])
	}
	return i, nil
}

// SetMultiplier is an undocumented API to update the complex value
// after the construction of the Reader. For the uint8 variant, this has
// a one-time CPU hit.
func (mr *uint8MultiplyReader) SetMultiplier(m complex64) {
	var (
		// This is 65535 samples of work to change the multiply const,
		// which, while not 0, is a lot better than the O(n). FWIW, 65535
		// IQ samples at 2 Msps is 0.03 seconds of IQ data, to have all
		// Reads be pure-lookups. This is worth it so long as you're
		// reading more than 65535 IQ samples via this reader before
		// changing the Multiplier again.

		ubuf = make(sdr.SamplesU8, 65535)
		cbuf = make(sdr.SamplesC64, 65535)
	)

	var realv uint16
	for ; realv < 256; realv++ {
		var imagv uint16
		for ; imagv <= 256; imagv++ {
			v := [2]uint8{uint8(realv), uint8(imagv)}
			ubuf[mr.index(v)] = v
		}
	}

	// Here, we'll round trip it through Complex64 once, do a SIMD optimized
	// multiply operation, and return the uint8 buffer as a lookup table.

	sdr.ConvertBuffer(cbuf, ubuf)
	cbuf.Multiply(m)
	sdr.ConvertBuffer(ubuf, cbuf)
	mr.tab = ubuf
}

type int8MultiplyReader struct {
	m complex64

	tab sdr.LookupTable
	r   sdr.Reader
}

func (mr *int8MultiplyReader) SampleFormat() sdr.SampleFormat {
	return mr.r.SampleFormat()
}

func (mr *int8MultiplyReader) SampleRate() uint {
	return mr.r.SampleRate()
}

func (mr *int8MultiplyReader) Read(s sdr.Samples) (int, error) {
	switch s.Format() {
	case sdr.SampleFormatI8:
		break
	default:
		return 0, sdr.ErrSampleFormatMismatch
	}

	i, err := mr.r.Read(s)
	if err != nil {
		return i, err
	}

	// TODO(paultag): Fix this to be safe when the above format checks
	// grow.
	sI8 := s.Slice(0, i).(sdr.SamplesI8)
	mr.tab.Lookup(sI8, sI8)
	return i, nil
}

// SetMultiplier is an undocumented API to update the complex value
// after the construction of the Reader. For the int8 variant, this has
// a one-time CPU hit.
func (mr *int8MultiplyReader) SetMultiplier(m complex64) {
	var (
		// This is 65535 samples of work to change the multiply const,
		// which, while not 0, is a lot better than the O(n). FWIW, 65535
		// IQ samples at 2 Msps is 0.03 seconds of IQ data, to have all
		// Reads be pure-lookups. This is worth it so long as you're
		// reading more than 65535 IQ samples via this reader before
		// changing the Multiplier again.

		ubuf = sdr.LookupTableIdentityI8()
		cbuf = make(sdr.SamplesC64, 65536)
		err  error
	)

	// Here, we'll round trip it through Complex64 once, do a SIMD optimized
	// multiply operation, and return the int8 buffer as a lookup table.

	sdr.ConvertBuffer(cbuf, ubuf)
	cbuf.Multiply(m)
	sdr.ConvertBuffer(ubuf, cbuf)
	mr.tab, err = sdr.NewLookupTable(sdr.SampleFormatI8, ubuf)
	if err != nil {
		// This shouldn't be reachable since we're positive of the construction,
		// although we ought to return something nicer than this.
		panic(err)
	}
}

// vim: foldmethod=marker
