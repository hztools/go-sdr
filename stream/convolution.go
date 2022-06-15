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
	"hz.tools/sdr/fft"
)

// ConvolutionReader will perform an fft against a window of samples,
// and multiply those sampls in frequency-space against the provided window.
//
// This can do things like apply a filter, etc. The fun really is endless.
//
// The `filter` slice is expected to be in the frequency domain, not time
// domain. This should *not* be a sdr.SamplesC64, it will yield absurd
// results.
func ConvolutionReader(
	r sdr.Reader,
	planner fft.Planner,
	filter []complex64,
) (sdr.Reader, error) {
	switch r.SampleFormat() {
	case sdr.SampleFormatC64:
	default:
		return nil, sdr.ErrSampleFormatUnknown
	}

	var (
		fftLength = len(filter)
		iq        = make(sdr.SamplesC64, fftLength)
	)

	conv, err := fft.ConvolveFreq(planner, iq, iq, filter)
	if err != nil {
		return nil, err
	}

	return ReadTransformer(r, ReadTransformerConfig{
		InputBufferLength:  fftLength,
		OutputBufferLength: fftLength,
		OutputSampleFormat: sdr.SampleFormatC64,
		OutputSampleRate:   r.SampleRate(),
		Proc: func(inI sdr.Samples, outI sdr.Samples) (int, error) {
			in, ok := inI.(sdr.SamplesC64)
			if !ok {
				return 0, sdr.ErrSampleFormatUnknown
			}
			out, ok := outI.(sdr.SamplesC64)
			if !ok {
				return 0, sdr.ErrSampleFormatUnknown
			}
			out = out[:in.Length()]
			copy(iq, in)

			if err := conv(); err != nil {
				return 0, err
			}

			copy(out, iq)
			return in.Length(), nil
		},
	})
}

// vim: foldmethod=marker
