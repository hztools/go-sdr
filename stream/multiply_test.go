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

	"hz.tools/rf"
	"hz.tools/sdr"
	"hz.tools/sdr/stream"
	"hz.tools/sdr/testutils"
)

func TestRotate(t *testing.T) {
	cwPhase0 := make(sdr.SamplesC64, 1024*60)
	cwPhase90 := make(sdr.SamplesC64, 1024*60)

	testutils.CW(cwPhase0, rf.Hz(10), 1.8e6, 0)
	testutils.CW(cwPhase90, rf.Hz(10), 1.8e6, math.Pi/2)

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

func TestRotateU8(t *testing.T) {
	var (
		valsU8  = make(sdr.SamplesU8, 1024*60)
		valsC64 = make(sdr.SamplesC64, 1024*60)
		refU8   = make(sdr.SamplesU8, 1024*60)
	)
	var counter uint16
	for i := range valsU8 {
		valsU8[i] = [2]uint8{
			uint8(counter & 0xFF),
			uint8((counter & 0xFF00 >> 8)),
		}
		counter++
	}
	sdr.ConvertBuffer(valsC64, valsU8)
	valsC64.Multiply(0 - 1i)
	sdr.ConvertBuffer(refU8, valsC64)

	pipeReader, pipeWriter := sdr.Pipe(1.8e6, sdr.SampleFormatU8)

	rotateReader, err := stream.Multiply(pipeReader, 0-1i)
	assert.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		_, err := pipeWriter.Write(valsU8)
		assert.NoError(t, err)
	}()

	buf := make(sdr.SamplesU8, 1024*60)
	_, err = sdr.ReadFull(rotateReader, buf)
	assert.NoError(t, err)

	for i := range buf {
		assert.Equal(t, refU8[i], buf[i])
	}

	wg.Wait()
}

func BenchmarkMultiplyReaderU8(b *testing.B) {
	buf := make(sdr.SamplesU8, 1024*64)
	for i := range buf {
		buf[i] = [2]uint8{128, 200}
	}

	pipeReader, pipeWriter := sdr.Pipe(1024, sdr.SampleFormatC64)
	multReader, err := stream.Multiply(pipeReader, 0+1i)
	assert.NoError(b, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		pipeWriter.Write(buf)
		pipeWriter.Close()
	}()

	bufout := make(sdr.SamplesU8, 1024*64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		multReader.Read(bufout)
	}
}

func BenchmarkMultiplyReaderI8(b *testing.B) {
	buf := make(sdr.SamplesI8, 1024*64)
	for i := range buf {
		buf[i] = [2]int8{100, -5}
	}

	pipeReader, pipeWriter := sdr.Pipe(1024, sdr.SampleFormatC64)
	multReader, err := stream.Multiply(pipeReader, 0+1i)
	assert.NoError(b, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		pipeWriter.Write(buf)
		pipeWriter.Close()
	}()

	bufout := make(sdr.SamplesI8, 1024*64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		multReader.Read(bufout)
	}
}

func BenchmarkMultiplyReaderC64(b *testing.B) {
	buf := make(sdr.SamplesC64, 1024*64)
	for i := range buf {
		buf[i] = complex64(complex(5, 5))
	}

	pipeReader, pipeWriter := sdr.Pipe(1024, sdr.SampleFormatC64)
	multReader, err := stream.Multiply(pipeReader, 0+1i)
	assert.NoError(b, err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		pipeWriter.Write(buf)
		pipeWriter.Close()
	}()

	bufout := make(sdr.SamplesC64, 1024*64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		multReader.Read(bufout)
	}
}

func TestRotateI8(t *testing.T) {
	var (
		valsI8  = make(sdr.SamplesI8, 1024*60)
		valsC64 = make(sdr.SamplesC64, 1024*60)
		refI8   = make(sdr.SamplesI8, 1024*60)
	)
	var counter uint16
	for i := range valsI8 {
		valsI8[i] = [2]int8{
			int8(counter & 0xFF),
			int8(int(counter&0xFF00>>8) - 127),
		}
		counter++
	}
	sdr.ConvertBuffer(valsC64, valsI8)
	valsC64.Multiply(0 - 1i)
	sdr.ConvertBuffer(refI8, valsC64)

	pipeReader, pipeWriter := sdr.Pipe(1.8e6, sdr.SampleFormatI8)

	rotateReader, err := stream.Multiply(pipeReader, 0-1i)
	assert.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		_, err := pipeWriter.Write(valsI8)
		assert.NoError(t, err)
	}()

	buf := make(sdr.SamplesI8, 1024*60)
	_, err = sdr.ReadFull(rotateReader, buf)
	assert.NoError(t, err)

	for i := range buf {
		assert.Equal(t, refI8[i], buf[i])
	}

	wg.Wait()
}

func TestRotateStd(t *testing.T) {
	pipeReader, pipeWriter := sdr.Pipe(1.8e6, sdr.SampleFormatC64)
	rotateReader, err := stream.Multiply(pipeReader, 0-1i)
	assert.NoError(t, err)

	testutils.TestReader(t, "C64-Read-Rotate", rotateReader)
	testutils.TestReadWriteSamples(t, "C64-ReadWrite-Rotate", rotateReader, pipeWriter)

	pipeReader, pipeWriter = sdr.Pipe(1.8e6, sdr.SampleFormatU8)
	rotateReader, err = stream.Multiply(pipeReader, 0-1i)
	assert.NoError(t, err)

	testutils.TestReader(t, "U8-Read-Rotate", rotateReader)
	testutils.TestReadWriteSamples(t, "U8-ReadWrite-Rotate", rotateReader, pipeWriter)

	pipeReader, pipeWriter = sdr.Pipe(1.8e6, sdr.SampleFormatI8)
	rotateReader, err = stream.Multiply(pipeReader, 0-1i)
	assert.NoError(t, err)

	testutils.TestReader(t, "I8-Read-Rotate", rotateReader)
	testutils.TestReadWriteSamples(t, "I8-ReadWrite-Rotate", rotateReader, pipeWriter)
}

// vim: foldmethod=marker
