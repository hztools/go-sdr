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

	"hz.tools/sdr"
	"hz.tools/sdr/stream"
)

func TestDownsampleCount(t *testing.T) {
	pipeReader, pipeWriter := sdr.Pipe(10000, sdr.SampleFormatC64)

	decReader, err := stream.DownsampleReader(pipeReader, 4)
	assert.NoError(t, err)

	inputBuffer := make(sdr.SamplesC64, 1024*32)
	wg := sync.WaitGroup{}
	go func() {
		defer wg.Done()
		outputBuffer := make(sdr.SamplesC64, 1024*32)
		n, err := sdr.ReadFull(decReader, outputBuffer)
		assert.Error(t, err)
		assert.Equal(t, (1024*32)/4, n)
	}()
	wg.Add(1)

	n, err := pipeWriter.Write(inputBuffer)
	assert.NoError(t, err)
	assert.Equal(t, 1024*32, n)
	pipeWriter.Close()

	wg.Wait()
}

func TestDownsampleCalc(t *testing.T) {
	pipeReader, pipeWriter := sdr.Pipe(10000, sdr.SampleFormatC64)

	decReader, err := stream.DownsampleReader(pipeReader, 4)
	assert.NoError(t, err)

	inputBuffer := make(sdr.SamplesC64, 1024*32)
	wg := sync.WaitGroup{}
	go func() {
		defer wg.Done()
		outputBuffer := make(sdr.SamplesC64, 1024*32)
		n, err := sdr.ReadFull(decReader, outputBuffer)
		assert.Error(t, err)
		assert.Equal(t, (1024*32)/4, n)

		outputBuffer = outputBuffer[:n]

		for _, el := range outputBuffer {
			assert.Equal(t, complex64(complex(1.5, 1.5)), el)
		}
	}()
	wg.Add(1)

	for i := range inputBuffer {
		e := float32(i % 4)
		inputBuffer[i] = complex64(complex(e, e))
	}

	n, err := pipeWriter.Write(inputBuffer)
	assert.NoError(t, err)
	assert.Equal(t, 1024*32, n)
	pipeWriter.Close()

	wg.Wait()
}

// vim: foldmethod=marker
