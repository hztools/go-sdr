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

// +build sdr.experimental

package stream_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
	"hz.tools/sdr/stream"
)

func TestWindowWriter(t *testing.T) {
	pipeReader, pipeWriter := sdr.Pipe(0, sdr.SampleFormatC64)
	windowWriter, err := stream.WindowWriter(pipeWriter)
	assert.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		buf := make(sdr.SamplesC64, 1024*64)
		for i := range buf {
			buf[i] = 1 + 1i
		}

		_, err := windowWriter.Write(buf)
		assert.NoError(t, err)
	}()

	buf := make(sdr.SamplesC64, 1024*64)
	_, err = sdr.ReadFull(pipeReader, buf)
	assert.NoError(t, err)

	// TODO(paultag): Check that the output is a Blackman window.

	wg.Wait()
}

// vim: foldmethod=marker
