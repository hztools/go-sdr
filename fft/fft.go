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

// Package fft contains a common interface to perform FFTs between frequency
// and time-Series complex data.
package fft

import (
	"hz.tools/sdr"
)

// Direction indicates if this is either a Forward or Backward FFT.
type Direction bool

var (
	// Forward will read the time-series 'iq' buffer, and write the computed
	// 'frequency' data to the frequency slice.
	Forward Direction = true

	// Backward will read the 'frequency' slice, and write the generated
	// IQ data to the 'iq' buffer.
	Backward Direction = false
)

// Planner will compute an FFT plan for the provided time-series and frequency
// buffers, and compute either a FFT or inverse FFT depending on the provided
// Direction object.
type Planner func(
	iq sdr.SamplesC64, frequency []complex64,
	direction Direction,
) (Plan, error)

// Plan is used to perform an FFT over the IQ or Time Series data, writing
// to the specified target.
type Plan interface {

	// Transform will execute the generated plan, preforming an FFT.
	Transform() error

	// Close will free any allocated resources or opened handles.
	Close() error
}

// TransformOnce will perform either a time-to-frequency or frequency-to-time
// domain transformation once. If this is called multiple times, significant
// overhead can be reduced by using the Planner interface.
func TransformOnce(
	planner Planner,
	iq sdr.SamplesC64,
	frequency []complex64,
	direction Direction,
) error {
	plan, err := planner(iq, frequency, direction)
	if err != nil {
		return err
	}
	return plan.Transform()
}

// vim: foldmethod=marker
