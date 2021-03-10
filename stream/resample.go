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
	"fmt"

	"hz.tools/sdr"
	"hz.tools/sdr/fft"
)

func copyFreq(order fft.Order, outFreq, inFreq []complex64) error {

	switch order {
	case fft.ZeroFirst:
		minLen := len(inFreq)
		if minLen > len(outFreq) {
			minLen = len(outFreq)
		}
		minLen = minLen / 2

		// 0 1 2 3 4 -4 -3 -2 -1
		// 0 1 2 3 -3 -2 -1

		copy(outFreq[:minLen], inFreq[:minLen])
		copy(outFreq[len(outFreq)-minLen:], inFreq[len(inFreq)-minLen:])

		return nil

		//	case fft.NegativeFirst:
		// this one is likely easier to implement, it's a matter of copying
		// around the center between the two buffers

		// -4 -3 -2 -1 0 1 2 3 4
		// -3 -2 -1 0 1 2 3

	default:
		return fmt.Errorf("Unknown fft order")
	}

}

// ResampleReader will use an FFT to up or down sample the stream to another
// specific sample rate that is not evenly divisable by a decimator, for
// example.
func ResampleReader(
	r sdr.Reader,
	planner fft.Planner,
	outputSampleRate uint,
) (sdr.Reader, error) {
	switch r.SampleFormat() {
	case sdr.SampleFormatC64:
	default:
		return nil, sdr.ErrSampleFormatUnknown
	}

	var (
		inWindowSize  = r.SampleRate() / 10
		outWindowSize = outputSampleRate / 10

		inFreq = make([]complex64, inWindowSize)
		inIq   = make(sdr.SamplesC64, inWindowSize)

		outFreq = make([]complex64, outWindowSize)
		outIq   = make(sdr.SamplesC64, outputSampleRate)
	)

	forward, err := planner(inIq, inFreq, fft.Forward, nil)
	if err != nil {
		return nil, err
	}

	backward, err := planner(outIq, outFreq, fft.Backward, nil)
	if err != nil {
		return nil, err
	}

	return ReadTransformer(r, ReadTransformerConfig{
		InputBufferLength:  int(inWindowSize),
		OutputBufferLength: int(outWindowSize),
		OutputSampleFormat: sdr.SampleFormatC64,
		OutputSampleRate:   outputSampleRate,
		Proc: func(inI sdr.Samples, outI sdr.Samples) (int, error) {
			in, ok := inI.(sdr.SamplesC64)
			if !ok {
				return 0, sdr.ErrSampleFormatUnknown
			}
			out, ok := outI.(sdr.SamplesC64)
			if !ok {
				return 0, sdr.ErrSampleFormatUnknown
			}

			copy(inIq, in)

			if err := forward.Transform(); err != nil {
				return 0, err
			}

			if err := copyFreq(fft.ZeroFirst, outFreq, inFreq); err != nil {
				return 0, err
			}

			if err := backward.Transform(); err != nil {
				return 0, err
			}

			copy(out, outIq)

			return out.Length(), nil
		},
	})
}

// vim: foldmethod=marker
