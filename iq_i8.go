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

package sdr

import (
	"unsafe"
)

// SamplesI8 indicates that the samples are being sent as a vector
// of interleaved int8 numbers, where 0 is -1, and 1 is 0xFF.
//
// This type is very hard to process, since the 0 value is 127.5, which
// is not representable, but it's *very* effective to send data over
// a connection, since it's the most compact represntation.
//
// This is the native format of the rtl-sdr.
type SamplesI8 [][2]int8

// Format returns the type of this vector, as exported by the SampleFormat
// enum.
func (s SamplesI8) Format() SampleFormat {
	return SampleFormatI8
}

// Size will return the size of this sdr.Samples in *bytes*. This is used
// when your code needs to be aware of the underlying storage size. This
// should usually only be used at i/o boundaries.
func (s SamplesI8) Size() int {
	return int(unsafe.Sizeof([2]int8{})) * len(s)
}

// Length will return the number of IQ samples in this vector of Samples.
//
// This is the count of real and imaginary pairs, so in the case
// of the I8 type, this will be half the size of the vector.
//
// This function is usually the correct one to use when processing
// sample information.
func (s SamplesI8) Length() int {
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
func (s SamplesI8) Slice(start, end int) Samples {
	return s[start:end]
}

// ToI16 will convert the int8 data to a vector of interleaved int16
// values.
func (s SamplesI8) ToI16(out SamplesI16) (int, error) {
	if s.Length() > out.Length() {
		return 0, ErrDstTooSmall
	}
	for i := range s {
		out[i] = [2]int16{
			int16(s[i][0]) << 8,
			int16(s[i][1]) << 8,
		}
	}
	return s.Length(), nil
}

// ToU8 will convert the int8 data to a vector of interleaved uint8
func (s SamplesI8) ToU8(out SamplesU8) (int, error) {
	if s.Length() > out.Length() {
		return 0, ErrDstTooSmall
	}
	for i := range s {
		out[i] = [2]uint8{
			uint8(int16(s[i][0]) + 128),
			uint8(int16(s[i][1]) + 128),
		}
	}
	return s.Length(), nil
}

// ToC64 will convert the int8 data to a vector of complex64 numbers.
func (s SamplesI8) ToC64(out SamplesC64) (int, error) {
	if s.Length() > out.Length() {
		return 0, ErrDstTooSmall
	}
	convI8ToC64Native(s, out)
	return s.Length(), nil
}

func convI8ToC64Native(s1 SamplesI8, s2 SamplesC64) {
	for i := range s1 {
		rel := s1[i][0]
		img := s1[i][1]

		// we use 128 not 127 to avoid going under -1 for min values,
		// but this means +1 is never fully +1
		s2[i] = complex(
			(float32(rel))/128,
			(float32(img))/128,
		)
	}
}

// vim: foldmethod=marker
