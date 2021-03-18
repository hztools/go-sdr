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

package testutils

import (
	"math"

	"hz.tools/rf"
	"hz.tools/sdr"
)

// CW will generate a Carrier Wave at a specific frequency.
func CW(buf sdr.SamplesC64, freq rf.Hz, sampleRate int, phase float64) {
	var (
		carrierFreq float64 = float64(freq)
		tau                 = math.Pi * 2
	)

	for i := range buf {
		now := float64(i) / float64(sampleRate)
		buf[i] = complex64(complex(
			math.Cos(tau*carrierFreq*now+phase),
			math.Sin(tau*carrierFreq*now+phase),
		))
	}
}

// vim: foldmethod=marker
