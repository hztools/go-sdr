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
	"hz.tools/sdr"
)

// DownsampleReader will take an oversampled input stream, and create a
// downsampled output stream that increases the ENOB (effective number of
// bits).
//
// For every 4 samples downsampled, you gain the equivalent of one bit of
// precision. For instance, for an 8-bit ADC at 3 Megasamples per secnd
// being downsampled, the following table outlines the ENOB and sample rate
// depending on the factor.
//
// +---------+-------------+------+
// | factor  | sample rate | ENOB |
// +---------+-------------+------+
// |      4  |      750000 |    9 |
// +---------+-------------+------+
// |      16 |      187500 |   10 |
// +---------+-------------+------+
// |      64 |       46875 |   11 |
// +---------+-------------+------+
// |     246 |       11718 |   12 |
// +---------+-------------+------+
//
func DownsampleReader(in sdr.Reader, factor uint) (sdr.Reader, error) {
	var offset = 0

	return ReadTransformer(in, ReadTransformerConfig{
		InputBufferLength:  32 * 1024,
		OutputBufferLength: 32 * 1024,
		OutputSampleRate:   in.SampleRate() / factor,
		OutputSampleFormat: sdr.SampleFormatC64,
		Proc: func(inBuf sdr.Samples, outBuf sdr.Samples) (int, error) {
			n, err := DownsampleBuffer(outBuf, inBuf, factor, offset)
			offset += inBuf.Length()
			return n, err
		},
	})
}

// DownsampleBuffer will take an oversampled input buffer, and write the
// downsampled output samples to the target buffer.
func DownsampleBuffer(to, from sdr.Samples, factor uint, offset int) (int, error) {
	if to.Format() != sdr.SampleFormatC64 {
		return 0, sdr.ErrSampleFormatMismatch
	}

	dFactor := int(factor)
	toLength := to.Length()
	fromLength := from.Length()

	if toLength < fromLength/dFactor {
		return 0, sdr.ErrDstTooSmall
	}

	// TOMBSTONE FOR FUTURE HACKERS
	//
	// Here we don't use the generic sdr.Iq Interface because we need to
	// both get and set at index offsets. THe Iq interface has enough for
	// most copy/io operations (for both the sdr library and users), but
	// in this case, we need to be doing some fairly detailed manipulation
	// of the IQ data.
	//
	// This means if you add a new sample format, this particular code
	// will need to become aware on how to get/set specific indexes.

	var (
		i   int
		out = to.(sdr.SamplesC64)
	)

	for i = 0; i < fromLength/dFactor; i++ {
		var (
			start                  = i * dFactor
			end                    = (i + 1) * dFactor
			samples sdr.SamplesC64 = make(sdr.SamplesC64, dFactor)
			sample  complex64
		)

		switch from := from.(type) {
		case sdr.SamplesU8:
			from[start:end].ToC64(samples)
		case sdr.SamplesI16:
			from[start:end].ToC64(samples)
		case sdr.SamplesC64:
			samples = from[start:end]
		default:
			return 0, sdr.ErrSampleFormatUnknown
		}

		for j := range samples {
			sample += samples[j]
		}

		out[i] = complex64(complex(
			real(sample)/float32(dFactor),
			imag(sample)/float32(dFactor),
		))
	}

	return int(i), nil
}

// vim: foldmethod=marker
