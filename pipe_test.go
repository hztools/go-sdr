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

package sdr_test

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
	"hz.tools/sdr/testutils"
)

func TestPipeStd(t *testing.T) {
	for n, sf := range map[string]sdr.SampleFormat{
		"C64": sdr.SampleFormatC64,
		"U8":  sdr.SampleFormatU8,
		"I16": sdr.SampleFormatI16,
	} {
		pipeReader, pipeWriter := sdr.Pipe(0, sf)
		testutils.TestReader(t, fmt.Sprintf("Read-Pipe-%s", n), pipeReader)
		testutils.TestWriter(t, fmt.Sprintf("Write-Pipe-%s", n), pipeWriter)
	}
}

func TestPipe(t *testing.T) {
	pipeReader, pipeWriter := sdr.Pipe(0, sdr.SampleFormatC64)

	wg := sync.WaitGroup{}
	go func(w sdr.Writer) {
		defer wg.Done()
		for i := 0; i < 10; i++ {
			wb := make(sdr.SamplesC64, 1024)
			wb[10] = complex64(20 + 10i)
			i, err := w.Write(wb)
			assert.NoError(t, err)
			assert.Equal(t, 1024, i)
		}
	}(pipeWriter)
	wg.Add(1)

	buf := make(sdr.SamplesC64, 1024*10)
	i, err := sdr.ReadFull(pipeReader, buf)
	assert.NoError(t, err)
	assert.Equal(t, 1024*10, i)

	for i := 0; i < 10; i++ {
		tbuf := buf[i*1024:]

		assert.Equal(t, complex64(0+0i), tbuf[0])
		assert.Equal(t, complex64(0+0i), tbuf[200])
		assert.Equal(t, complex64(0+0i), tbuf[1000])

		assert.Equal(t, complex64(20+10i), tbuf[10])
	}

	wg.Wait()
}

func TestPipeReadMismatchedWrite(t *testing.T) {
	_, pipeWriter := sdr.Pipe(0, sdr.SampleFormatU8)
	buf := make(sdr.SamplesC64, 1024)
	i, err := pipeWriter.Write(buf)
	assert.Equal(t, 0, i)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)
}

func TestPipeReadMismatchedRead(t *testing.T) {
	pipeReader, _ := sdr.Pipe(0, sdr.SampleFormatU8)
	buf := make(sdr.SamplesC64, 1024)
	i, err := pipeReader.Read(buf)
	assert.Equal(t, 0, i)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)
}

func TestPipeReadClose(t *testing.T) {
	pipeReader, pipeWriter := sdr.Pipe(0, sdr.SampleFormatU8)

	wg := sync.WaitGroup{}
	go func(w sdr.Writer) {
		defer wg.Done()
		wb := make(sdr.SamplesU8, 1024)
		wb[10] = [2]uint8{20, 10}
		i, err := w.Write(wb)
		assert.Equal(t, sdr.ErrPipeClosed, err)
		assert.Equal(t, 255, i)
	}(pipeWriter)
	wg.Add(1)

	buf := make(sdr.SamplesU8, 255)
	i, err := sdr.ReadFull(pipeReader, buf)
	assert.NoError(t, err)
	assert.Equal(t, 255, i)

	assert.Equal(t, [2]uint8{20, 10}, buf[10])

	assert.NoError(t, pipeReader.Close())
	wg.Wait()
}

func TestPipeWriteClose(t *testing.T) {
	pipeReader, pipeWriter := sdr.Pipe(0, sdr.SampleFormatU8)

	rb := make(sdr.SamplesU8, 255)
	wg := sync.WaitGroup{}
	go func(w sdr.Reader) {
		defer wg.Done()
		i, err := sdr.ReadFull(pipeReader, rb)
		assert.NoError(t, err)
		assert.Equal(t, 255, i)
		assert.Equal(t, [2]uint8{20, 10}, rb[10])

	}(pipeReader)
	wg.Add(1)

	go func(w sdr.Writer) {
		defer wg.Done()
		wb := make(sdr.SamplesU8, 1024)
		wb[10] = [2]uint8{20, 10}

		i, err := w.Write(wb)
		assert.Equal(t, 255, i)
		assert.Equal(t, sdr.ErrPipeClosed, err)
	}(pipeWriter)
	wg.Add(1)

	time.Sleep(time.Second / 5)
	assert.NoError(t, pipeWriter.Close())

	i, err := sdr.ReadFull(pipeReader, rb)
	assert.Equal(t, sdr.ErrPipeClosed, err)
	assert.Equal(t, 0, i)

	wg.Wait()
}

func TestPipeExternalCancel(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	pipeReader, _ := sdr.PipeWithContext(ctx, 0, sdr.SampleFormatU8)
	cancel()
	buf := make(sdr.SamplesU8, 1024)
	i, err := pipeReader.Read(buf)
	assert.Equal(t, 0, i)
	assert.Equal(t, sdr.ErrPipeClosed, err)
}

func TestPipeReadCustomError(t *testing.T) {
	ctx := context.Background()
	pipeReader, _ := sdr.PipeWithContext(ctx, 0, sdr.SampleFormatU8)
	pipeReader.CloseWithError(io.EOF)

	buf := make(sdr.SamplesU8, 1024)
	i, err := pipeReader.Read(buf)
	assert.Equal(t, 0, i)
	assert.Equal(t, io.EOF, err)
}

func TestPipeWriteCustomError(t *testing.T) {
	ctx := context.Background()
	pipeReader, pipeWriter := sdr.PipeWithContext(ctx, 0, sdr.SampleFormatU8)
	pipeReader.CloseWithError(io.EOF)

	buf := make(sdr.SamplesU8, 1024)
	i, err := pipeWriter.Write(buf)
	assert.Equal(t, 0, i)
	assert.Equal(t, io.EOF, err)
}

func TestPipeParts(t *testing.T) {
	pipeReader, pipeWriter := sdr.Pipe(0, sdr.SampleFormatC64)

	wg := sync.WaitGroup{}
	go func(w sdr.Writer) {
		defer wg.Done()
		defer pipeReader.Close()
		wb := make(sdr.SamplesC64, 1024)
		wb[10] = complex64(20 + 10i)
		wb[512] = complex64(20 + 10i)
		i, err := w.Write(wb)
		assert.NoError(t, err)
		assert.Equal(t, 1024, i)
	}(pipeWriter)
	wg.Add(1)

	buf := make(sdr.SamplesC64, 128)
	i, err := sdr.ReadFull(pipeReader, buf)
	assert.NoError(t, err)
	assert.Equal(t, 128, i)
	assert.Equal(t, complex64(20+10i), buf[10])
	buf = make(sdr.SamplesC64, 1024)
	i, err = sdr.ReadFull(pipeReader, buf)
	assert.Equal(t, sdr.ErrPipeClosed, err)
	assert.Equal(t, 1024-128, i)
	assert.Equal(t, complex64(20+10i), buf[512-128])
	wg.Wait()
}

// vim: foldmethod=marker
