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

	"github.com/stretchr/testify/assert"
	"testing"

	"hz.tools/sdr"
	"hz.tools/sdr/stream"
)

func TestDecimateBufferU8(t *testing.T) {
	// we're matching size to detect any over or underruns. The output buffer
	// should be 1000*8/10, but here we'll just make it match to try a raw copy.
	inputBuffer := make(sdr.SamplesU8, 1000*8)
	outputBuffer := make(sdr.SamplesU8, 1000*8)

	n, err := stream.DecimateBuffer(outputBuffer, inputBuffer, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(inputBuffer)/10, n)

	assert.NoError(t, err)
}

func TestDecimateBufferI16(t *testing.T) {
	// we're matching size to detect any over or underruns. The output buffer
	// should be 1000*8/10, but here we'll just make it match to try a raw copy.
	inputBuffer := make(sdr.SamplesI16, 1000*8)
	outputBuffer := make(sdr.SamplesI16, 1000*8)

	n, err := stream.DecimateBuffer(outputBuffer, inputBuffer, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(inputBuffer)/10, n)

	assert.NoError(t, err)
}

func TestDecimateBufferC64(t *testing.T) {
	// we're matching size to detect any over or underruns. The output buffer
	// should be 1000*8/10, but here we'll just make it match to try a raw copy.
	inputBuffer := make(sdr.SamplesC64, 1000*8)
	outputBuffer := make(sdr.SamplesC64, 1000*8)

	n, err := stream.DecimateBuffer(outputBuffer, inputBuffer, 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, len(inputBuffer)/10, n)

	assert.NoError(t, err)
}

func TestDecimateBufferMismatch(t *testing.T) {
	// we're matching size to detect any over or underruns. The output buffer
	// should be 1000*8/10, but here we'll just make it match to try a raw copy.
	inputBuffer := make(sdr.SamplesC64, 1000*8)
	outputBuffer := make(sdr.SamplesU8, 1000*8)

	n, err := stream.DecimateBuffer(outputBuffer, inputBuffer, 10, 0)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)
	assert.Equal(t, 0, n)
}

func TestDecimateBufferShort(t *testing.T) {
	// we're matching size to detect any over or underruns. The output buffer
	// should be 1000*8/10, but here we'll just make it match to try a raw copy.
	inputBuffer := make(sdr.SamplesU8, 1000*8)
	outputBuffer := make(sdr.SamplesU8, 10)

	n, err := stream.DecimateBuffer(outputBuffer, inputBuffer, 10, 0)
	assert.Error(t, err)
	assert.Equal(t, 0, n)
}

func TestDecimateRateFormat(t *testing.T) {
	pipeReader, _ := sdr.Pipe(10000, sdr.SampleFormatU8)
	decReader, err := stream.DecimateReader(pipeReader, 10)
	assert.NoError(t, err)

	assert.Equal(t, uint32(10000/10), decReader.SampleRate())
	assert.Equal(t, pipeReader.SampleFormat(), decReader.SampleFormat())
}

func TestDecimateCount(t *testing.T) {
	pipeReader, pipeWriter := sdr.Pipe(10000, sdr.SampleFormatU8)

	decReader, err := stream.DecimateReader(pipeReader, 10)
	assert.NoError(t, err)

	inputBuffer := make(sdr.SamplesU8, 1000*8)
	wg := sync.WaitGroup{}
	go func() {
		defer wg.Done()
		outputBuffer := make(sdr.SamplesU8, 1000*8)
		n, err := sdr.ReadFull(decReader, outputBuffer)
		// error here is ok since we're using readfull
		assert.Error(t, err)
		// inputBuffer's length divided by the factor passed to Decimate
		assert.Equal(t, (1000*8)/10, n)
	}()
	wg.Add(1)

	n, err := pipeWriter.Write(inputBuffer)
	assert.NoError(t, err)
	assert.Equal(t, 1000*8, n)
	pipeWriter.Close()

	wg.Wait()
}

func TestDecimateSkippyboi(t *testing.T) {
	pipeReader, pipeWriter := sdr.Pipe(10000, sdr.SampleFormatU8)

	decReader, err := stream.DecimateReader(pipeReader, 10)
	assert.NoError(t, err)

	inputBuffer := make(sdr.SamplesU8, 1000*8)
	for i := 0; i < len(inputBuffer); i++ {
		z := uint8(i % 10)
		inputBuffer[i] = [2]uint8{z, z}
	}

	wg := sync.WaitGroup{}
	go func() {
		defer wg.Done()

		outputBuffer := make(sdr.SamplesU8, 100*8)
		n, err := sdr.ReadFull(decReader, outputBuffer)
		// error here is ok since we're using readfull
		assert.NoError(t, err)
		// inputBuffer's length divided by the factor passed to Decimate
		assert.Equal(t, (1000*8)/10, n)

		for j := 0; j < n; j++ {
			assert.Equal(t, [2]uint8{0, 0}, outputBuffer[j])
		}
	}()
	wg.Add(1)

	n, err := pipeWriter.Write(inputBuffer)
	assert.NoError(t, err)
	assert.Equal(t, 1000*8, n)
	pipeWriter.Close()

	wg.Wait()
}

// vim: foldmethod=marker
