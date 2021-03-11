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
	"sync"

	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
)

func TestCopySamplesU8(t *testing.T) {
	src := make(sdr.SamplesU8, 10)
	dst := make(sdr.SamplesU8, 10)

	src[1] = [2]uint8{10, 20}

	i, err := sdr.CopySamples(dst, src)
	assert.NoError(t, err)
	assert.Equal(t, 10, i)

	assert.Equal(t, [2]uint8{10, 20}, dst[1])
}

func TestCopySamplesC64(t *testing.T) {
	src := make(sdr.SamplesC64, 10)
	dst := make(sdr.SamplesC64, 10)

	src[1] = complex64(10 + 20i)

	i, err := sdr.CopySamples(dst, src)
	assert.NoError(t, err)
	assert.Equal(t, 10, i)

	assert.Equal(t, complex64(10+20i), dst[1])
}

func TestCopySamplesMismatch(t *testing.T) {
	src := make(sdr.SamplesC64, 10)
	dst := make(sdr.SamplesU8, 10)

	_, err := sdr.CopySamples(dst, src)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)
}

func TestCopyMismatch(t *testing.T) {
	pipeReader1, _ := sdr.Pipe(0, sdr.SampleFormatU8)
	_, pipeWriter2 := sdr.Pipe(0, sdr.SampleFormatC64)

	_, err := sdr.Copy(pipeWriter2, pipeReader1)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)
}

func TestCopyBufferMismatch(t *testing.T) {
	pipeReader1, _ := sdr.Pipe(0, sdr.SampleFormatC64)
	_, pipeWriter2 := sdr.Pipe(0, sdr.SampleFormatC64)

	buf, err := sdr.MakeSamples(sdr.SampleFormatU8, 128)
	assert.NoError(t, err)

	_, err = sdr.CopyBuffer(pipeWriter2, pipeReader1, buf)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)

	pipeReader1, _ = sdr.Pipe(0, sdr.SampleFormatU8)
	_, err = sdr.CopyBuffer(pipeWriter2, pipeReader1, buf)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)
}

func TestCopyU8(t *testing.T) {
	pipeReader1, pipeWriter1 := sdr.Pipe(0, sdr.SampleFormatU8)
	pipeReader2, pipeWriter2 := sdr.Pipe(0, sdr.SampleFormatU8)

	wg := sync.WaitGroup{}
	go func() {
		defer wg.Done()
		buf := make(sdr.SamplesU8, 1024)
		buf[10][0] = 0x24
		_, err := pipeWriter1.Write(buf)
		assert.NoError(t, err)
		assert.NoError(t, pipeWriter1.Close())
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		i, err := sdr.Copy(pipeWriter2, pipeReader1)
		assert.Equal(t, int64(1024), i)
		assert.Equal(t, sdr.ErrPipeClosed, err)
	}()
	wg.Add(1)

	buf := make(sdr.SamplesU8, 1024)
	sdr.ReadFull(pipeReader2, buf)
	assert.Equal(t, uint8(0x24), buf[10][0])

	wg.Wait()
}

func TestCopyBufferU8(t *testing.T) {
	pipeReader1, pipeWriter1 := sdr.Pipe(0, sdr.SampleFormatU8)
	pipeReader2, pipeWriter2 := sdr.Pipe(0, sdr.SampleFormatU8)

	wg := sync.WaitGroup{}
	go func() {
		defer wg.Done()
		buf := make(sdr.SamplesU8, 1024)
		buf[10][0] = 0x24
		_, err := pipeWriter1.Write(buf)
		assert.NoError(t, err)
		assert.NoError(t, pipeWriter1.Close())
	}()
	wg.Add(1)

	go func() {
		defer wg.Done()
		buf, err := sdr.MakeSamples(sdr.SampleFormatU8, 128)
		assert.NoError(t, err)

		i, err := sdr.CopyBuffer(pipeWriter2, pipeReader1, buf)
		assert.Equal(t, int64(1024), i)
		assert.Equal(t, sdr.ErrPipeClosed, err)
	}()
	wg.Add(1)

	buf := make(sdr.SamplesU8, 1024)
	sdr.ReadFull(pipeReader2, buf)
	assert.Equal(t, uint8(0x24), buf[10][0])

	wg.Wait()
}

// vim: foldmethod=marker
