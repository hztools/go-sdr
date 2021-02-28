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

func TestMixReader(t *testing.T) {
	pipeReader1, pipeWriter1 := sdr.Pipe(10000, sdr.SampleFormatC64)
	pipeReader2, pipeWriter2 := sdr.Pipe(10000, sdr.SampleFormatC64)
	pipeReader3, pipeWriter3 := sdr.Pipe(10000, sdr.SampleFormatC64)

	buf := make(sdr.SamplesC64, 1000)
	for i := range buf {
		buf[i] = complex64(complex(10, 20))
	}

	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		_, err := pipeWriter1.Write(buf)
		assert.NoError(t, err)
	}()
	go func() {
		defer wg.Done()
		_, err := pipeWriter2.Write(buf)
		assert.NoError(t, err)
	}()
	go func() {
		defer wg.Done()
		_, err := pipeWriter3.Write(buf)
		assert.NoError(t, err)
	}()

	outBuf := make(sdr.SamplesC64, 1000)

	mix, err := stream.Mix(pipeReader1, pipeReader2, pipeReader3)
	assert.NoError(t, err)

	_, err = sdr.ReadFull(mix, outBuf)
	assert.NoError(t, err)

	for i := range outBuf {
		assert.Equal(t, complex64(complex(30, 60)), outBuf[i])
	}

	wg.Wait()
}

func BenchmarkMixComplex2(b *testing.B) {
	pipeReader, _ := sdr.Pipe(10000, sdr.SampleFormatC64)
	mixReader, err := stream.Mix(pipeReader, pipeReader, pipeReader)
	if err != nil {
		panic(err)
	}

	buf := make(sdr.SamplesC64, 1024*8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mixReader.(interface {
			MixC64(sdr.SamplesC64, ...sdr.SamplesC64)
		}).MixC64(buf, buf)
	}
}

func BenchmarkMixComplex4(b *testing.B) {
	pipeReader, _ := sdr.Pipe(10000, sdr.SampleFormatC64)
	mixReader, err := stream.Mix(pipeReader, pipeReader, pipeReader)
	if err != nil {
		panic(err)
	}

	buf := make(sdr.SamplesC64, 1024*8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mixReader.(interface {
			MixC64(sdr.SamplesC64, ...sdr.SamplesC64)
		}).MixC64(buf, buf, buf, buf)
	}
}

func BenchmarkMixComplex16(b *testing.B) {
	pipeReader, _ := sdr.Pipe(10000, sdr.SampleFormatC64)
	mixReader, err := stream.Mix(pipeReader, pipeReader, pipeReader)
	if err != nil {
		panic(err)
	}

	buf := make(sdr.SamplesC64, 1024*8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mixReader.(interface {
			MixC64(sdr.SamplesC64, ...sdr.SamplesC64)
		}).MixC64(
			buf, buf, buf, buf,
			buf, buf, buf, buf,
			buf, buf, buf, buf,
			buf, buf, buf, buf,
		)
	}
}

// vim: foldmethod=marker
