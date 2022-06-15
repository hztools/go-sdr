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

package fft

import (
	"fmt"

	"hz.tools/sdr"
)

// convolve
func convolve(
	planner Planner,
	dst sdr.Samples,
	iq1 sdr.Samples,
	iq2 sdr.Samples,
	op func(sdr.Samples, sdr.Samples, sdr.Samples, []complex64, []complex64) error,
) (func() error, error) {
	if iq1.Length() != iq2.Length() || iq1.Length() != dst.Length() {
		// TODO(paultag): This isn't strictly right, we should perhaps check that
		// they're the same power of two and zero-pad but we're very lazy so
		// let's make the user explicitly do that.
		return nil, fmt.Errorf("sdr/fft: IQ/Dest buffer lengths do not match exactly")
	}

	if iq1.Format() != iq2.Format() || iq1.Format() != dst.Format() {
		// TODO(paultag): Again; we may want to make this magic at some point
		// but for now let's be explicit.
		return nil, sdr.ErrSampleFormatMismatch
	}

	switch iq1.Format() {
	case sdr.SampleFormatC64:
		break
	default:
		return nil, sdr.ErrSampleFormatUnknown
	}

	freq1 := make([]complex64, iq1.Length())
	freq2 := make([]complex64, iq2.Length())

	planForward1, err := planner(iq1.(sdr.SamplesC64), freq1, Forward)
	if err != nil {
		return nil, err
	}
	planForward2, err := planner(iq2.(sdr.SamplesC64), freq2, Forward)
	if err != nil {
		return nil, err
	}

	planBackward, err := planner(dst.(sdr.SamplesC64), freq1, Backward)
	if err != nil {
		return nil, err
	}

	return func() error {
		if err := planForward1.Transform(); err != nil {
			return err
		}
		if err := planForward2.Transform(); err != nil {
			return err
		}

		if err := op(dst, iq1, iq2, freq1, freq2); err != nil {
			return err
		}
		return planBackward.Transform()
	}, nil
}

// Convolve will plan a convolution of two time-series IQ samples,
// returning a function to repeatedly convolve the two provided IQ buffers.
// The output of the convolution will be written to the dst Samples. The
// dst argument may safely be one of iq1 or iq2.
//
// Under the hood this will use the provided FFT Planner to multiply the
// samples in the frequency domain, which winds up a lot faster than having
// to perform the convolution in the time domain.
func Convolve(
	planner Planner,
	dst sdr.Samples,
	iq1 sdr.Samples,
	iq2 sdr.Samples,
) (func() error, error) {
	return convolve(planner, dst, iq1, iq2,
		func(dst sdr.Samples,
			iq1 sdr.Samples, iq2 sdr.Samples,
			freq1 []complex64, freq2 []complex64,
		) error {
			for i := range freq1 {
				freq1[i] = freq1[i] * freq2[i]
			}
			return nil
		},
	)
}

// CrossCorrelate will plan a cross correlation in the frequency domain
// between two iq buffers to determine what offset (if any) contain
// a correlation with the exemplar.
func CrossCorrelate(
	planner Planner,
	dst sdr.Samples,
	iq1 sdr.Samples,
	iq2 sdr.Samples,
) (func() error, error) {
	return convolve(planner, dst, iq1, iq2,
		func(dst sdr.Samples,
			iq1 sdr.Samples, iq2 sdr.Samples,
			freq1 []complex64, freq2 []complex64,
		) error {
			for i := range freq1 {
				freq1[i] = freq1[i] * complex(
					real(freq2[i]),
					-imag(freq2[i]),
				)
			}
			return nil
		},
	)
}

// ConvolveFreq will plan a convolution of frequency-domain complex
// numbers against time-series iq data in the frequency domain,
// returning a function to repeatedly convolve the two provided IQ buffers.
// The output of the convolution will be written to the dst Samples. The
// dst argument may safely be one of iq1 or iq2.
//
// Under the hood this will use the provided FFT Planner to multiply the
// samples in the frequency domain, which winds up a lot faster than having
// to perform the convolution in the time domain.
func ConvolveFreq(
	planner Planner,
	dst sdr.Samples,
	src sdr.Samples,
	freq []complex64,
) (func() error, error) {
	if src.Length() != dst.Length() || src.Length() != len(freq) {
		return nil, fmt.Errorf("sdr/fft.Convolve: Lengths do not match exactly")
	}

	if src.Format() != dst.Format() {
		return nil, sdr.ErrSampleFormatMismatch
	}

	switch src.Format() {
	case sdr.SampleFormatC64:
		break
	default:
		return nil, sdr.ErrSampleFormatUnknown
	}

	freq1 := make([]complex64, src.Length())

	planForward, err := planner(src.(sdr.SamplesC64), freq1, Forward)
	if err != nil {
		return nil, err
	}

	planBackward, err := planner(dst.(sdr.SamplesC64), freq1, Backward)
	if err != nil {
		return nil, err
	}

	return func() error {
		if err := planForward.Transform(); err != nil {
			return err
		}
		for i := range freq1 {
			freq1[i] = freq1[i] * freq[i]
		}
		return planBackward.Transform()
	}, nil
}

// ConvolveOnce will perform a one-off convolution of two time series iq streams
// in the frequency domain, writing the results to the dst samples. The dst
// argument may safely be one of iq1 or iq2.
//
// If this function is invoked multiple times, consider using the Convolve
// function to plan the convolution, to save on setup time from the fft planner.
func ConvolveOnce(
	planner Planner,
	dst sdr.Samples,
	iq1 sdr.Samples,
	iq2 sdr.Samples,
) error {
	conv, err := Convolve(planner, dst, iq1, iq2)
	if err != nil {
		return err
	}
	return conv()
}

// vim: foldmethod=marker
