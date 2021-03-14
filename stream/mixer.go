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
	"fmt"

	"hz.tools/sdr"
	"hz.tools/sdr/internal/simd"
)

// Mix will take any number of Readers, and mix all those Readers into a single
// Reader. This is likely not generally useful for generating data to transmit,
// but could be very useful for testing.
//
// Each of the readers must be the same SampleFormat and SampleRate or
// an error will be returned.
//
// The samples, when added must not exceed +1 or go below -1. This causes
// clipping and clipping is bad. If you're adding two waveforms, be absoluely
// sure you've adjusted the gain correctly on the readers with the stream.Gain
// reader.
func Mix(readers ...sdr.Reader) (sdr.Reader, error) {

	switch len(readers) {
	case 0:
		return nil, fmt.Errorf("stream.Mix: No readers passed")
	case 1:
		return readers[0], nil
	}

	var (
		sampleFormat = readers[0].SampleFormat()
		sampleRate   = readers[0].SampleRate()
	)

	if sampleFormat != sdr.SampleFormatC64 {
		return nil, sdr.ErrSampleFormatUnknown
	}

	for _, reader := range readers {
		if reader.SampleFormat() != sampleFormat {
			return nil, fmt.Errorf("stream.Mix: Readers are not all the same format")
		}

		if reader.SampleRate() != sampleRate {
			return nil, fmt.Errorf("stream.Mix: Readers are not all the same rate")
		}
	}

	return &mixerReader{
		sampleFormat: sampleFormat,
		sampleRate:   sampleRate,
		readers:      readers,
	}, nil
}

type mixerReader struct {
	sampleFormat sdr.SampleFormat
	sampleRate   uint
	readers      []sdr.Reader
	err          error
}

func (mr *mixerReader) SampleFormat() sdr.SampleFormat {
	return mr.sampleFormat
}

func (mr *mixerReader) SampleRate() uint {
	return mr.sampleRate
}

func (mr *mixerReader) MixC64(out sdr.SamplesC64, buffers ...sdr.SamplesC64) {
	for _, buf := range buffers {
		simd.AddComplex(out, buf, out)
	}
}

func (mr *mixerReader) Read(s sdr.Samples) (int, error) {
	// TODO(paultag): SIMD moderate hanging fruit

	if mr.err != nil {
		return 0, mr.err
	}

	if s.Format() != sdr.SampleFormatC64 {
		return 0, sdr.ErrSampleFormatUnknown
	}

	samples := s.(sdr.SamplesC64)
	samplesLength := samples.Length()

	buffers := make([]sdr.SamplesC64, len(mr.readers))

	for i, reader := range mr.readers {
		buffers[i] = make(sdr.SamplesC64, samplesLength)
		_, err := sdr.ReadFull(reader, buffers[i])
		// here we know the number of phasors read must match the buffer
		// length, or we get an error. as a result, we can avoid storing
		// the number of samples read.

		if err != nil {
			mr.err = err
			// TODO(paultag): This should likely not be 0, perhaps we need
			// to take the number read from any buffer? Max read from any?
			// Least read from any? This is going to cause alignment
			// headaches regardless.
			return 0, err
		}
	}

	// now let's mix the samples.
	//
	// in the future this should allow for gain whilst mixing too.
	for si := range samples {
		samples[si] = complex(0, 0)
	}

	mr.MixC64(samples, buffers...)

	return len(samples), nil
}

//
func clampRealToRange(v, min, max float32) float32 {
	switch {
	case v < min:
		return min
	case v > max:
		return max
	default:
		return v
	}
}

//
func clampSampleToRange(iq complex64, min, max float32) complex64 {
	return complex(
		clampRealToRange(real(iq), min, max),
		clampRealToRange(imag(iq), min, max),
	)
}

// vim: foldmethod=marker
