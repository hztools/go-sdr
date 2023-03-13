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

package sdr

import (
	"fmt"
)

var (
	// ErrSampleFormatMismatch will be returned when there's a mismatch between
	// sample formats.
	ErrSampleFormatMismatch = fmt.Errorf("sdr: iq sample formats do not match")

	// ErrSampleFormatUnknown will be returned when a specific iq format is not
	// implemented.
	ErrSampleFormatUnknown = fmt.Errorf("sdr: iq sample format is not understood")

	// ErrDstTooSmall will be returned when attempting to perform an operation
	// and the target buffer is too small to use.
	ErrDstTooSmall = fmt.Errorf("sdr: destination sample buffer is too small")
)

// Samples represents a vector of IQ data.
//
// This type is an interface_ and not a struct or typedef to complex64
// because this allows the generic IQ helpers in this package to operate on
// the native format of the SDR without requiring expensive conversions to
// other types.
//
// This package contains 4 Samples implementations:
//
//   - SamplesU8  - interleaved uint8 values
//   - SamplesI8  - interleaved int8 values
//   - SamplesI16 - interleaved int16 values
//   - SamplesC64 - vector of complex64 values (interleaved float32 values)
//
// This should cover most common SDRs, but if you're handing a type of IQ data
// that is not supported, you may either implement the Samples type yourself
// along with the required interface points, convert to a format this library
// supports, or send a PR to this library adding support to this format.
type Samples interface {
	// Format returns the type of this vector, as exported by the SampleFormat
	// enum.
	Format() SampleFormat

	// Size will return the size of this sdr.Samples in *bytes*. This is used
	// when your code needs to be aware of the underlying storage size. This
	// should usually only be used at i/o boundaries.
	Size() int

	// Length will return the number of IQ samples in this vector of Samples.
	//
	// This is the count of real and imaginary pairs, so in the case
	// of the U8 type, this will be half the size of the vector.
	//
	// This function is usually the correct one to use when processing
	// sample information.
	Length() int

	// Slice will return a slice of the sample buffer from the provided
	// starting position until the ending position. The returned value is
	// assumed to be a slice, which is to say, mutations of the returned
	// Samples will modify the slice from whence it came.
	//
	// samples.Slice(0, 10) is assumed to be the same as samples[:10], except
	// it does not require the typecast to the concrete type implementing
	// this interface.
	Slice(int, int) Samples
}

// SampleFormat is an ID used throughout go-rf to uniquely identify what type
// the IQ samples are in. This allows code to quickly compare to see if two
// types are talking about the same type or not, without resoring to expensive
// read operations and type casting.
type SampleFormat uint8

// Size will return the number of bytes that are needed to represent a single
// phasor, both real and imaginary.
func (sf SampleFormat) Size() int {
	switch sf {
	case SampleFormatU8, SampleFormatI8:
		return 2
	case SampleFormatI16:
		return 4
	case SampleFormatC64:
		return 8
	default:
		return 0
	}
}

const (
	// SampleFormatC64 indicates that SamplesC64 will be handled. See
	// sdr.SamplesC64 for more information.
	SampleFormatC64 SampleFormat = 1

	// SampleFormatU8 indicates that SamplesU8 will be handled. See
	// sdr.SamplesU8 for more information.
	SampleFormatU8 SampleFormat = 2

	// SampleFormatI16 indicates that SamplesI16 will be handled. See
	// sdr.SamplesI16 for more information.
	SampleFormatI16 SampleFormat = 3

	// SampleFormatI8 indicates that SamplesI8 will be handled. See
	// sdr.SamplesI8 for more information.
	SampleFormatI8 SampleFormat = 4
)

// MakeSamples will create a buffer of a specified size and type. This will
// return a newly allocated slice of Samples. This function is used when the
// code processing Samples is fairly generic, to avoid switches on type, and
// to reduce the cost of adding new sample formats.
//
// If your code is specific to a given Sample format, for instance if your code
// only supports SamplesC64, it's fine to not use this function at all.
func MakeSamples(sampleFormat SampleFormat, sampleSize int) (Samples, error) {
	switch sampleFormat {
	case SampleFormatU8:
		return make(SamplesU8, sampleSize), nil
	case SampleFormatI8:
		return make(SamplesI8, sampleSize), nil
	case SampleFormatI16:
		return make(SamplesI16, sampleSize), nil
	case SampleFormatC64:
		return make(SamplesC64, sampleSize), nil
	default:
		return nil, ErrSampleFormatUnknown
	}
}

// String returns the format name as a human readable String.
func (sf SampleFormat) String() string {
	switch sf {
	case SampleFormatI8:
		return "interleaved int8"
	case SampleFormatU8:
		return "interleaved uint8"
	case SampleFormatI16:
		return "interleaved int16"
	case SampleFormatC64:
		return "complex64"
	default:
		return "unknown"
	}
}

// vim: foldmethod=marker
