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
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/rf"
	"hz.tools/sdr"
	"hz.tools/sdr/stream"
	"hz.tools/sdr/testutils"
)

func TestShifter(t *testing.T) {
	cw := make(sdr.SamplesC64, 1024*60)
	testutils.CW(cw, rf.Hz(1), 1.8e6, 0)

	pipeReader, pipeWriter := sdr.Pipe(1.8e6, sdr.SampleFormatC64)

	shiftHigh, err := stream.ShiftReader(pipeReader, rf.KHz)
	assert.NoError(t, err)

	// TODO(paultag): Check that it actually, well, shifted to 1 KHz, but
	// that requires an external dependency to a specific fft backend, or
	// something like goertzel, which I'm not ready to do from hz.tools/sdr

	shiftLow, err := stream.ShiftReader(shiftHigh, -rf.KHz)
	assert.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		_, err := pipeWriter.Write(cw)
		assert.NoError(t, err)
	}()

	buf := make(sdr.SamplesC64, 1024*60)
	_, err = sdr.ReadFull(shiftLow, buf)
	assert.NoError(t, err)

	var epsilon float64 = 0.0001

	for i := range cw {
		assert.InEpsilon(t, 1+real(cw[i]), 1+real(buf[i]), epsilon)
		assert.InEpsilon(t, 1+imag(cw[i]), 1+imag(buf[i]), epsilon)
	}

	wg.Wait()
}

// vim: foldmethod=marker
