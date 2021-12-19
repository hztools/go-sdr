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

func TestAddReader(t *testing.T) {
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

	mix, err := stream.Add(pipeReader1, pipeReader2, pipeReader3)
	assert.NoError(t, err)

	_, err = sdr.ReadFull(mix, outBuf)
	assert.NoError(t, err)

	for i := range outBuf {
		assert.Equal(t, complex64(complex(30, 60)), outBuf[i])
	}

	wg.Wait()
}

func TestAddReaderI8(t *testing.T) {
	pipeReader1, pipeWriter1 := sdr.Pipe(0, sdr.SampleFormatI8)
	pipeReader2, pipeWriter2 := sdr.Pipe(0, sdr.SampleFormatI8)
	out := make(sdr.SamplesI8, 1024*32)
	in := make(sdr.SamplesI8, 1024*32)
	for i := range in {
		in[i] = [2]int8{10, 10}
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := pipeWriter1.Write(in)
		assert.NoError(t, err)
	}()
	go func() {
		defer wg.Done()
		_, err := pipeWriter2.Write(in)
		assert.NoError(t, err)
	}()
	mix, err := stream.Add(pipeReader1, pipeReader2)
	assert.NoError(t, err)
	_, err = sdr.ReadFull(mix, out)
	assert.NoError(t, err)
	for i := range out {
		assert.Equal(t, [2]int8{20, 20}, out[i])
	}
	wg.Wait()
}

func TestAddReaderI16(t *testing.T) {
	pipeReader1, pipeWriter1 := sdr.Pipe(0, sdr.SampleFormatI16)
	pipeReader2, pipeWriter2 := sdr.Pipe(0, sdr.SampleFormatI16)
	out := make(sdr.SamplesI16, 1024*32)
	in := make(sdr.SamplesI16, 1024*32)
	for i := range in {
		in[i] = [2]int16{10, 10}
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		_, err := pipeWriter1.Write(in)
		assert.NoError(t, err)
	}()
	go func() {
		defer wg.Done()
		_, err := pipeWriter2.Write(in)
		assert.NoError(t, err)
	}()
	mix, err := stream.Add(pipeReader1, pipeReader2)
	assert.NoError(t, err)
	_, err = sdr.ReadFull(mix, out)
	assert.NoError(t, err)
	for i := range out {
		assert.Equal(t, [2]int16{20, 20}, out[i])
	}
	wg.Wait()
}

func BenchmarkAddComplex2(b *testing.B) {
	pipeReader, _ := sdr.Pipe(10000, sdr.SampleFormatC64)
	mixReader, err := stream.Add(pipeReader, pipeReader, pipeReader)
	if err != nil {
		panic(err)
	}

	buf := make(sdr.SamplesC64, 1024*8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mixReader.(interface {
			AddC64(sdr.SamplesC64, ...sdr.SamplesC64)
		}).AddC64(buf, buf)
	}
}

func BenchmarkAddComplex4(b *testing.B) {
	pipeReader, _ := sdr.Pipe(10000, sdr.SampleFormatC64)
	mixReader, err := stream.Add(pipeReader, pipeReader, pipeReader)
	if err != nil {
		panic(err)
	}

	buf := make(sdr.SamplesC64, 1024*8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mixReader.(interface {
			AddC64(sdr.SamplesC64, ...sdr.SamplesC64)
		}).AddC64(buf, buf, buf, buf)
	}
}

func BenchmarkAddComplex16(b *testing.B) {
	pipeReader, _ := sdr.Pipe(10000, sdr.SampleFormatC64)
	mixReader, err := stream.Add(pipeReader, pipeReader, pipeReader)
	if err != nil {
		panic(err)
	}

	buf := make(sdr.SamplesC64, 1024*8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mixReader.(interface {
			AddC64(sdr.SamplesC64, ...sdr.SamplesC64)
		}).AddC64(
			buf, buf, buf, buf,
			buf, buf, buf, buf,
			buf, buf, buf, buf,
			buf, buf, buf, buf,
		)
	}
}

// vim: foldmethod=marker
