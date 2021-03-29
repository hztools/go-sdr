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
	"hz.tools/sdr/testutils"
)

func TestConvertReaderAPI(t *testing.T) {
	pipeReader, pipeWriter := sdr.Pipe(0, sdr.SampleFormatU8)
	c64pipeReader, err := stream.ConvertReader(pipeReader, sdr.SampleFormatC64)
	assert.NoError(t, err)
	testutils.TestReader(t, "ConvertReader-U8-C64", c64pipeReader)
	testutils.TestReadWriteSamples(t, "ConvertReader-ReadWrite-U8-C64", c64pipeReader, pipeWriter)
}

func TestConvertWriterAPI(t *testing.T) {
	_, pipeWriter := sdr.Pipe(0, sdr.SampleFormatU8)

	c64writer, err := stream.ConvertWriter(pipeWriter, sdr.SampleFormatC64)
	assert.NoError(t, err)
	testutils.TestWriter(t, "ConvertWriter-U8-C64", c64writer)
}

func TestConvertReaderBufferU8C64(t *testing.T) {
	// we're matching size to detect any over or underruns. The output buffer
	// should be 1000*8/10, but here we'll just make it match to try a raw copy.
	inputBuffer := make(sdr.SamplesU8, 1024*32)
	outputBuffer := make(sdr.SamplesC64, 1024*32)

	for i := 0; i < len(inputBuffer); i++ {
		inputBuffer[i] = [2]uint8{0xFF, 0xFF}
	}

	pipeReader, pipeWriter := sdr.Pipe(0, sdr.SampleFormatU8)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		n, err := pipeWriter.Write(inputBuffer)
		assert.Equal(t, 1024*32, n)
		assert.NoError(t, err)
	}()

	c64pipeReader, err := stream.ConvertReader(pipeReader, sdr.SampleFormatC64)
	assert.NoError(t, err)

	n, err := sdr.ReadFull(c64pipeReader, outputBuffer)
	assert.Equal(t, 1024*32, n)
	assert.NoError(t, err)

	wg.Wait()
}

func TestConvertWriterBufferU8C64(t *testing.T) {
	inputBuffer := make(sdr.SamplesC64, 1000*8)
	outputBuffer := make(sdr.SamplesU8, 1000*8)

	pipeReader, pipeWriter := sdr.Pipe(1337, sdr.SampleFormatU8)

	convWriter, err := stream.ConvertWriter(pipeWriter, sdr.SampleFormatC64)
	assert.NoError(t, err)

	wg := sync.WaitGroup{}
	go func() {
		defer wg.Done()
		n, err := convWriter.Write(inputBuffer)
		assert.Equal(t, 1000*8, n)
		assert.NoError(t, err)
	}()
	wg.Add(1)

	n, err := sdr.ReadFull(pipeReader, outputBuffer)
	assert.Equal(t, 1000*8, n)
	assert.NoError(t, err)

	wg.Wait()
}

// vim: foldmethod=marker
