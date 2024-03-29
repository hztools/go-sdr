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

func TestGainReaderAPI(t *testing.T) {
	pipeReader, pipeWriter := sdr.Pipe(1024, sdr.SampleFormatC64)
	gainReader := stream.Gain(pipeReader, 0.5)
	testutils.TestReader(t, "GainReader-C64", gainReader)
	testutils.TestReadWriteSamples(t, "GainReader-ReadWrite-C64", gainReader, pipeWriter)

	pipeReader, pipeWriter = sdr.Pipe(1024, sdr.SampleFormatI16)
	gainReader = stream.Gain(pipeReader, 0.5)
	testutils.TestReader(t, "GainReader-I16", gainReader)
	// testutils.TestReadWriteSamples(t, "GainReader-ReadWrite-I16", gainReader, pipeWriter)

	pipeReader, pipeWriter = sdr.Pipe(1024, sdr.SampleFormatU8)
	gainReader = stream.Gain(pipeReader, 0.5)
	testutils.TestReader(t, "GainReader-U8", gainReader)
	// testutils.TestReadWriteSamples(t, "GainReader-ReadWrite-U8", gainReader, pipeWriter)
}

func TestGainBufferC64(t *testing.T) {
	pipeReader, pipeWriter := sdr.Pipe(1024, sdr.SampleFormatC64)

	wg := sync.WaitGroup{}

	go func() {
		defer wg.Done()

		b := make(sdr.SamplesC64, 1024)
		for i := range b {
			b[i] = complex(1, 1)
		}

		_, err := pipeWriter.Write(b)
		assert.NoError(t, err)
	}()
	wg.Add(1)

	gainReader := stream.Gain(pipeReader, 0.5)

	b := make(sdr.SamplesC64, 1024)
	_, err := sdr.ReadFull(gainReader, b)
	assert.NoError(t, err)

	assert.Equal(t, float32(0.5), real(b[10]))
	assert.Equal(t, float32(0.5), imag(b[10]))

	wg.Wait()
}

func BenchmarkGainReader(b *testing.B) {
	buf := make(sdr.SamplesC64, 1024)
	for i := range buf {
		buf[i] = complex64(complex(5, 5))
	}

	pipeReader, _ := sdr.Pipe(1024, sdr.SampleFormatC64)
	gainReader := stream.Gain(pipeReader, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gainReader.(interface {
			Scale(sdr.Samples) error
		}).Scale(buf)
	}
}

// vim: foldmethod=marker
