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

// Add will take any number of Readers, and mix all those Readers into a single
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
func Add(readers ...sdr.Reader) (sdr.Reader, error) {

	switch len(readers) {
	case 0:
		return nil, fmt.Errorf("stream.Add: No readers passed")
	case 1:
		return readers[0], nil
	}

	var (
		sampleFormat = readers[0].SampleFormat()
		sampleRate   = readers[0].SampleRate()
	)

	switch sampleFormat {
	case sdr.SampleFormatC64:
	case sdr.SampleFormatI16:
	case sdr.SampleFormatI8:
	default:
		return nil, sdr.ErrSampleFormatUnknown
	}

	for _, reader := range readers {
		if reader.SampleFormat() != sampleFormat {
			return nil, fmt.Errorf("stream.Add: Readers are not all the same format")
		}

		if reader.SampleRate() != sampleRate {
			return nil, fmt.Errorf("stream.Add: Readers are not all the same rate")
		}
	}

	return &addReader{
		sampleFormat: sampleFormat,
		sampleRate:   sampleRate,
		readers:      readers,
	}, nil
}

type addReader struct {
	sampleFormat sdr.SampleFormat
	sampleRate   uint
	readers      []sdr.Reader
	err          error
}

func (ar *addReader) SampleFormat() sdr.SampleFormat {
	return ar.sampleFormat
}

func (ar *addReader) SampleRate() uint {
	return ar.sampleRate
}

func (ar *addReader) AddI8(out sdr.SamplesI8, buffers ...sdr.Samples) {
	for _, bufS := range buffers {
		buf := bufS.(sdr.SamplesI8)
		for idx, sample := range buf {
			out[idx][0] += sample[0]
			out[idx][1] += sample[1]
		}
	}
}

func (ar *addReader) AddI16(out sdr.SamplesI16, buffers ...sdr.Samples) {
	for _, bufS := range buffers {
		buf := bufS.(sdr.SamplesI16)
		for idx, sample := range buf {
			out[idx][0] += sample[0]
			out[idx][1] += sample[1]
		}
	}
}

func (ar *addReader) AddC64(out sdr.SamplesC64, buffers ...sdr.Samples) {
	for _, buf := range buffers {
		simd.AddComplex(out, buf.(sdr.SamplesC64), out)
	}
}

func (ar *addReader) Read(s sdr.Samples) (int, error) {
	// TODO(paultag): SIMD moderate hanging fruit

	if ar.err != nil {
		return 0, ar.err
	}

	switch s.Format() {
	case sdr.SampleFormatC64:
	case sdr.SampleFormatI16:
	case sdr.SampleFormatI8:
	default:
		return 0, sdr.ErrSampleFormatUnknown
	}

	buffers := make([]sdr.Samples, len(ar.readers))

	for i, reader := range ar.readers {
		var err error
		buffers[i], err = sdr.MakeSamples(s.Format(), s.Length())
		// See note below on error handling.
		if err != nil {
			ar.err = err
			return 0, err
		}

		_, err = sdr.ReadFull(reader, buffers[i])
		// here we know the number of phasors read must match the buffer
		// length, or we get an error. as a result, we can avoid storing
		// the number of samples read.

		if err != nil {
			ar.err = err
			// TODO(paultag): This should likely not be 0, perhaps we need
			// to take the number read from any buffer? Max read from any?
			// Least read from any? This is going to cause alignment
			// headaches regardless.
			return 0, err
		}
	}

	switch samples := s.(type) {
	case sdr.SamplesC64:
		for si := range samples {
			samples[si] = complex(0, 0)
		}
		ar.AddC64(samples, buffers...)
		return len(samples), nil
	case sdr.SamplesI16:
		for si := range samples {
			samples[si] = [2]int16{0, 0}
		}
		ar.AddI16(samples, buffers...)
		return len(samples), nil
	case sdr.SamplesI8:
		for si := range samples {
			samples[si] = [2]int8{0, 0}
		}
		ar.AddI8(samples, buffers...)
		return len(samples), nil
	default:
		// Egads this is bad. I have no earthly idea how we've wound up
		// in this state sensibly. I have half a mind to throw a panic, but
		// maybe this isn't a practical concern. I'm going to live to regret
		// that thought...
		return 0, sdr.ErrSampleFormatUnknown
	}
}

// vim: foldmethod=marker
