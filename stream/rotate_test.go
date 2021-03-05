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

package stream_test

import (
	"math"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
	"hz.tools/sdr/stream"
)

var Tau = math.Pi * 2

func generateCw(buf sdr.SamplesC64, sampleRate int, phase float64) {
	var (
		carrierFreq float64 = 10
	)

	for i := range buf {
		now := float64(i) / float64(sampleRate)
		buf[i] = complex64(complex(
			math.Cos(Tau*carrierFreq*now+phase),
			math.Sin(Tau*carrierFreq*now+phase),
		))
	}
}

func TestRotate(t *testing.T) {
	cwPhase0 := make(sdr.SamplesC64, 1024*60)
	cwPhase90 := make(sdr.SamplesC64, 1024*60)

	generateCw(cwPhase0, 1.8e6, 0)
	generateCw(cwPhase90, 1.8e6, math.Pi/2)

	pipeReader, pipeWriter := sdr.Pipe(1.8e6, sdr.SampleFormatC64)

	rotateReader, err := stream.Multiply(pipeReader, 0-1i)
	assert.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		_, err := pipeWriter.Write(cwPhase90)
		assert.NoError(t, err)
	}()

	buf := make(sdr.SamplesC64, 1024*60)
	_, err = sdr.ReadFull(rotateReader, buf)
	assert.NoError(t, err)

	var epsilon float64 = 0.0001

	for i := range cwPhase0 {
		assert.InEpsilon(t, 1+real(cwPhase0[i]), 1+real(buf[i]), epsilon)
		assert.InEpsilon(t, 1+imag(cwPhase0[i]), 1+imag(buf[i]), epsilon)
	}

	wg.Wait()
}

// vim: foldmethod=marker
