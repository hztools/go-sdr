// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2020-2021
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

package sdr

import (
	"math"
	"unsafe"

	"hz.tools/sdr/internal/simd"
)

// SamplesC64 indicates that the samples are in a complex64
// number, which is itself two interleaved float32 numbers, the
// i and q value. In memory, this is the same thing as interleaving i and q
// values in a float array.
//
// This is the format that is most useful to process iq data in from a
// mathematical perspective, and is going to be the most common type to
// work with when writing signal processing code.
type SamplesC64 []complex64

// Format returns the type of this vector, as exported by the SampleFormat
// enum.
func (s SamplesC64) Format() SampleFormat {
	return SampleFormatC64
}

// Size will return the size of this sdr.Samples in *bytes*. This is used
// when your code needs to be aware of the underlying storage size. This
// should usually only be used at i/o boundaries.
func (s SamplesC64) Size() int {
	return int(unsafe.Sizeof(complex64(0))) * len(s)
}

// Length will return the number of IQ samples in this vector of Samples.
//
// This is the count of real and imaginary pairs, so in the case
// of the U8 type, this will be half the size of the vector.
//
// This function is usually the correct one to use when processing
// sample information.
func (s SamplesC64) Length() int {
	return len(s)
}

// Slice will return a slice of the sample buffer from the provided
// starting position until the ending position. The returned value is
// assumed to be a slice, which is to say, mutations of the returned
// Samples will modify the slice from whence it came.
//
// samples.Slice(0, 10) is assumed to be the same as samples[:10], except
// it does not require the typecast to the concrete type implementing
// this interface.
func (s SamplesC64) Slice(start, end int) Samples {
	return s[start:end]
}

// ToU8 will convert the Complex data to a vector of interleaved uint8s.
func (s SamplesC64) ToU8(out SamplesU8) (int, error) {
	if s.Length() > out.Length() {
		return 0, ErrDstTooSmall
	}
	for i, sample := range s {
		sampleReal := (real(sample) * 127.5) + 127.5
		sampleImag := (imag(sample) * 127.5) + 127.5
		// TODO(paultag): Check for over/underflow and cap the values.
		out[i][0] = uint8(sampleReal)
		out[i][1] = uint8(sampleImag)
	}
	return s.Length(), nil
}

// ToI16 will convert the complex64 data to int16 data.
func (s SamplesC64) ToI16(out SamplesI16) (int, error) {
	if s.Length() > out.Length() {
		return 0, ErrDstTooSmall
	}
	for i := range s {
		out[i] = [2]int16{
			int16(real(s[i]) * math.MaxInt16),
			int16(imag(s[i]) * math.MaxInt16),
		}
	}
	return s.Length(), nil
}

// ToI8 will convert the complex64 data to int8 data.
func (s SamplesC64) ToI8(out SamplesI8) (int, error) {
	if s.Length() > out.Length() {
		return 0, ErrDstTooSmall
	}
	for i := range s {
		out[i] = [2]int8{
			int8(real(s[i]) * math.MaxInt8),
			int8(imag(s[i]) * math.MaxInt8),
		}
	}
	return s.Length(), nil
}

// Scale will multiply each I and Q value by the provided real value 'r'. This
// will *not* do a complex multiplication, this is the same as if each phasor
// had their real and imag parts multiplied by the provided real value 'r'.
func (s SamplesC64) Scale(r float32) {
	simd.ScaleComplex(r, s)
}

// Multiply will conduct a complex multiplication of each phasor in this buffer
// by a provided complex number 'c'.
func (s SamplesC64) Multiply(c complex64) {
	simd.RotateComplex(c, s)
}

// Add will conduct a complex addition of each phasor in this buffer
// by a provided complex number 'c', writing results to 'dst'.
func (s SamplesC64) Add(c []complex64) error {
	return simd.AddComplex(s, c, s)
}

// vim: foldmethod=marker
