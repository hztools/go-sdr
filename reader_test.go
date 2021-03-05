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
	"io"
	"sync"

	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
)

func TestMultiReaderNone(t *testing.T) {
	_, err := sdr.MultiReader()
	assert.Error(t, err)
}

func TestMultiReaderOne(t *testing.T) {
	pipeReader1, _ := sdr.Pipe(0, sdr.SampleFormatU8)

	multiReader, err := sdr.MultiReader(pipeReader1)
	assert.NoError(t, err)
	assert.Equal(t, pipeReader1, multiReader)
}

func TestMultiReaderU8(t *testing.T) {
	pipeReader1, pipeWriter1 := sdr.Pipe(0, sdr.SampleFormatU8)
	pipeReader2, pipeWriter2 := sdr.Pipe(0, sdr.SampleFormatU8)

	wg := sync.WaitGroup{}
	writeVals := func(t *testing.T, w sdr.PipeWriter, val uint8) {
		defer w.CloseWithError(io.EOF)
		defer wg.Done()
		buf := make(sdr.SamplesU8, 1024)
		for i := range buf {
			buf[i] = [2]uint8{val, 0}
		}
		i, err := w.Write(buf)
		assert.NoError(t, err)
		assert.Equal(t, 1024, i)
	}
	go writeVals(t, pipeWriter1, 1)
	go writeVals(t, pipeWriter2, 2)
	wg.Add(2)

	multiReader, err := sdr.MultiReader(pipeReader1, pipeReader2)
	assert.NoError(t, err)

	buf := make(sdr.SamplesU8, 1024/2)
	for _, val := range []uint8{1, 1, 2, 2} {
		i, err := sdr.ReadFull(multiReader, buf)
		assert.NoError(t, err)
		assert.Equal(t, 1024/2, i)
		for _, el := range buf {
			assert.Equal(t, val, el[0])
		}
	}
	wg.Wait()
}

func TestMultiReaderError(t *testing.T) {
	pipeReader1, pipeWriter1 := sdr.Pipe(0, sdr.SampleFormatU8)
	pipeReader2, pipeWriter2 := sdr.Pipe(0, sdr.SampleFormatU8)

	wg := sync.WaitGroup{}
	writeVals := func(t *testing.T, w sdr.PipeWriter, val uint8) {
		defer wg.Done()

		if val == 2 {
			w.CloseWithError(sdr.ErrShortBuffer)
			return
		}

		defer w.CloseWithError(io.EOF)
		buf := make(sdr.SamplesU8, 1024)
		for i := range buf {
			buf[i] = [2]uint8{val, 0}
		}
		i, err := w.Write(buf)
		assert.NoError(t, err)
		assert.Equal(t, 1024, i)
	}
	go writeVals(t, pipeWriter1, 1)
	go writeVals(t, pipeWriter2, 2)
	wg.Add(2)

	multiReader, err := sdr.MultiReader(pipeReader1, pipeReader2)
	assert.NoError(t, err)

	buf := make(sdr.SamplesU8, 1024)

	_, err = sdr.ReadFull(multiReader, buf)
	assert.NoError(t, err)

	_, err = sdr.ReadFull(multiReader, buf)
	assert.Equal(t, sdr.ErrShortBuffer, err)
	wg.Wait()
}

func TestMultiReaderSFM(t *testing.T) {
	pipeReader1, _ := sdr.Pipe(0, sdr.SampleFormatU8)
	pipeReader2, _ := sdr.Pipe(0, sdr.SampleFormatC64)

	_, err := sdr.MultiReader(pipeReader1, pipeReader2)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)

	_, err = sdr.MultiReader(pipeReader2, pipeReader1)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)

	_, err = sdr.MultiReader(pipeReader2, pipeReader2, pipeReader1, pipeReader2)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)
}

func TestMultiReaderSampleRateMismatch(t *testing.T) {
	pipeReader1, _ := sdr.Pipe(1, sdr.SampleFormatU8)
	pipeReader2, _ := sdr.Pipe(2, sdr.SampleFormatU8)

	_, err := sdr.MultiReader(pipeReader1, pipeReader2)
	assert.Error(t, err)

	_, err = sdr.MultiReader(pipeReader2, pipeReader1)
	assert.Error(t, err)

	_, err = sdr.MultiReader(pipeReader2, pipeReader2, pipeReader1, pipeReader2)
	assert.Error(t, err)
}

// vim: foldmethod=marker
