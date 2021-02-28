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
	"encoding/binary"
	"io"
	"sync"

	"github.com/stretchr/testify/assert"
	"testing"

	"hz.tools/sdr"
)

func TestBytesIOLE(t *testing.T) {
	ioReader, ioWriter := io.Pipe()

	pipeReader := sdr.ByteReader(ioReader, binary.LittleEndian, 0, sdr.SampleFormatC64)
	pipeWriter := sdr.ByteWriter(ioWriter, binary.LittleEndian, 0, sdr.SampleFormatC64)

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
		assert.Equal(t, float32(10), imag(buf[(i*1024)+10]))
	}
}

func TestBytesIOBE(t *testing.T) {
	ioReader, ioWriter := io.Pipe()

	pipeReader := sdr.ByteReader(ioReader, binary.LittleEndian, 0, sdr.SampleFormatC64)
	pipeWriter := sdr.ByteWriter(ioWriter, binary.LittleEndian, 0, sdr.SampleFormatC64)

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
		assert.Equal(t, float32(10), imag(buf[(i*1024)+10]))
	}
}

// vim: foldmethod=marker
