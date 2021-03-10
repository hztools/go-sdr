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

// DecimateReader will take even Nth sample (where N is the `factor` argument)
// from an sdr.Reader, and provide the downsampled or compressed iq stream
// through the returned Reader.
//
// This will reduce the sample rate by the provided factor (so if the input
// Reader is at 18 Msps, and we apply a factor of 10 Decimation, we'll get
// an output Reader of 1.8 Msps.
func DecimateReader(in sdr.Reader, factor uint) (sdr.Reader, error) {
	var offset = 0

	return ReadTransformer(in, ReadTransformerConfig{
		InputBufferLength:  32 * 1024,
		OutputBufferLength: 32 * 1024,
		OutputSampleRate:   in.SampleRate() / factor,
		OutputSampleFormat: in.SampleFormat(),
		Proc: func(inBuf sdr.Samples, outBuf sdr.Samples) (int, error) {
			n, err := DecimateBuffer(outBuf, inBuf, factor, offset)
			offset += inBuf.Length()
			return n, err
		},
	})
}

// DecimateBuffer will take every Nth sample, reducing the number of samples per
// second on the other end by the same factor.
//
// This is sometimes also called "Downsamping" or "Compression", but a lot
// of other tools use the term decimation, even though it's not always
// a downsample of a factor of 100.
func DecimateBuffer(to, from sdr.Samples, factor uint, offset int) (int, error) {
	if from.Format() != to.Format() {
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

	var i int
	for i = 0; i < fromLength/dFactor; i++ {
		switch from := from.(type) {
		case sdr.SamplesU8:
			to := to.(sdr.SamplesU8)
			to[i] = from[dFactor*i]
		case sdr.SamplesI16:
			to := to.(sdr.SamplesI16)
			to[i] = from[dFactor*i]
		case sdr.SamplesC64:
			to := to.(sdr.SamplesC64)
			to[i] = from[dFactor*i]
		default:
			return 0, sdr.ErrSampleFormatUnknown
		}
	}

	return int(i), nil
}

// vim: foldmethod=marker
