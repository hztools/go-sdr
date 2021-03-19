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
)

// SamplesI16 indicates that the samples are being sent as a vector
// of interleaved int16 integers. The values range from +32767 to -32768.
// 0 remains, well 0.
//
// There are a few hazards for people working with this type at an IO boundary
// with an SDR (if you're consuming this data, you don't really have to care
// much - it's a number! numbers are great! Enjoy!).
//
// Firstly, this type isn't really any particular device's native format,
// which may cause a bit of confusion, but it's close enough to be useful when
// reading from ADCs that have 12 or 14 bits of precision, since you don't want
// to align to non-8-bit boundaries while also maintaining sanity. As a result,
// working with this type can be a bit awkward when doing IO with an SDR,
// since you either have to align to the MSB or LSB. Some SDRs (I'm looking
// at you, PlutoSDR) expect their samples to be in both formats depending on
// direction.
//
// If you are at the IO boundary with an SDR the correct format for this type is
// MSB aligned data. As soon as you convert the byte stream into this type, be
// sure it's immediately MSB aligned. If your SDR is giving you LSB algined data,
// you may need to call ShiftLSBToMSBBits with the number of bits your data is
// in (for instance, `12` for a 12 bit ADC) to MSB align.
type SamplesI16 [][2]int16

// Format returns the type of this vector, as exported by the SampleFormat
// enum.
func (s SamplesI16) Format() SampleFormat {
	return SampleFormatI16
}

// Size will return the size of this sdr.Samples in *bytes*. This is used
// when your code needs to be aware of the underlying storage size. This
// should usually only be used at i/o boundaries.
func (s SamplesI16) Size() int {
	return int(unsafe.Sizeof([2]int16{})) * len(s)
}

// Length will return the number of IQ samples in this vector of Samples.
//
// This is the count of real and imaginary pairs, so in the case
// of the U8 type, this will be half the size of the vector.
//
// This function is usually the correct one to use when processing
// sample information.
func (s SamplesI16) Length() int {
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
func (s SamplesI16) Slice(start, end int) Samples {
	return s[start:end]
}

// ShiftLSBToMSBBits is a helper function to be used when the input data
// is not actually 16 bits. This is usually fine if the data is MSB aligned,
// since the range is still roughly the int16 max / min (just like, 8 off).
// However, if the data is LSB aligned, this is a major shitshow, since the
// max is no longer 2**16, the max is 2**12 or 2**14. Rather than make
// multiple sdr.Sample types for each ADC bit count, the SDR code should
// call ShiftLSBToMSBBits at the boundary to shift the data from LSB to MSB
// algined to get full-range values.
//
// The value `bits` is the number of bits the ADC sends. This will result in
// a bitshift of 16 - bits. So, if you have a 12 bit ADC, and this is invoked,
// each I and Q sample will be shifted left 4 bits, bringing the range from
// +/- 2047 to +/-32768.
//
// This will mutate the buffer in place.
func (s SamplesI16) ShiftLSBToMSBBits(bits int) {
	// TODO(paultag): This would be pretty straightforward to implement in
	// ASM / SIMD; may be a good task for down the road.
	shift := 16 - bits
	for i := range s {
		s[i][0] = s[i][0] << shift
		s[i][1] = s[i][1] << shift
	}
}

// ToU8 will convert the int16 data to interleaved uint8 bit samples.
// This looks a lot like a (weirdly) simplified version of c64 -> u8
// since both have to deal with shifting from negative.
func (s SamplesI16) ToU8(out SamplesU8) (int, error) {
	if s.Length() > out.Length() {
		return 0, ErrDstTooSmall
	}

	for i, sample := range s {
		out[i] = [2]uint8{
			// This line is very confusing.
			//
			// Given ints in the range from +/-32768, we want values
			// between 0 and 2**16-1. What we do is convert the int16 to
			// an int32, add 32768, to shift the minimum to 0, cast to a uint16,
			// then drop the lower byte by casting to a uint8.
			uint8(uint16(int32(sample[0])+32768) >> 8),
			uint8(uint16(int32(sample[1])+32768) >> 8),
		}
	}
	return s.Length(), nil
}

// ToC64 will convert the int16 data to a vector of complex64 numbers.
func (s SamplesI16) ToC64(out SamplesC64) (int, error) {
	if s.Length() > out.Length() {
		return 0, ErrDstTooSmall
	}
	for i := range s {
		cI := float32(s[i][0]) / math.MaxInt16
		cQ := float32(s[i][1]) / math.MaxInt16
		out[i] = complex(cI, cQ)
	}
	return s.Length(), nil
}

// ToI8 will convert the int16 data to interleaved int8 bit samples.
func (s SamplesI16) ToI8(out SamplesI8) (int, error) {
	if s.Length() > out.Length() {
		return 0, ErrDstTooSmall
	}

	for i, sample := range s {
		out[i] = [2]int8{
			int8(sample[0] >> 8),
			int8(sample[1] >> 8),
		}
	}
	return s.Length(), nil
}

// vim: foldmethod=marker
