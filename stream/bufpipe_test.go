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
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
	"hz.tools/sdr/stream"
)

func TestBufPipeBasic(t *testing.T) {
	pipe, err := stream.NewBufPipe(1, 0, sdr.SampleFormatU8)
	assert.NoError(t, err)

	b1 := make(sdr.SamplesU8, 1024)
	for i := range b1 {
		b1[i][0] = 0xFF
		b1[i][1] = 0xFF
	}

	i, err := pipe.Write(b1)
	assert.NoError(t, err)
	assert.Equal(t, 1024, i)

	b2 := make(sdr.SamplesU8, 1024)
	i, err = sdr.ReadFull(pipe, b2)
	assert.NoError(t, err)
	assert.Equal(t, 1024, i)

	for i := range b2 {
		assert.Equal(t, uint8(0xFF), b2[i][0])
		assert.Equal(t, uint8(0xFF), b2[i][1])
	}
}

func TestBufPipeDouble(t *testing.T) {
	pipe, err := stream.NewBufPipe(2, 0, sdr.SampleFormatU8)
	assert.NoError(t, err)

	b1 := make(sdr.SamplesU8, 1024)
	for i := range b1 {
		b1[i][0] = 0xFF
		b1[i][1] = 0xFF
	}

	// Write 2 1024 buffers, and check that it queues, and that we can
	// read both out.

	i, err := pipe.Write(b1)
	assert.NoError(t, err)
	assert.Equal(t, 1024, i)

	i, err = pipe.Write(b1)
	assert.NoError(t, err)
	assert.Equal(t, 1024, i)

	b2 := make(sdr.SamplesU8, 1024*2)
	i, err = sdr.ReadFull(pipe, b2)
	assert.NoError(t, err)
	assert.Equal(t, 2048, i)

	for i := range b2 {
		assert.Equal(t, uint8(0xFF), b2[i][0])
		assert.Equal(t, uint8(0xFF), b2[i][1])
	}
}

func TestBufPipeTooManyWrite(t *testing.T) {
	pipe, err := stream.NewBufPipe(2, 0, sdr.SampleFormatU8)
	assert.NoError(t, err)

	b1 := make(sdr.SamplesU8, 1024)

	i, err := pipe.Write(b1)
	assert.NoError(t, err)
	assert.Equal(t, 1024, i)

	i, err = pipe.Write(b1)
	assert.NoError(t, err)
	assert.Equal(t, 1024, i)

	for i := 0; i < 2; i++ {
		// we have to do this due to the interactions with the go scheduler
		// and the `do` function. We'll only try twice, though.
		_, err = pipe.Write(b1)
		if err == nil {
			continue
		}
		assert.Equal(t, stream.ErrBufferOverrun, err)
		return
	}
	t.Fatal("Write did not overflow")
}

func TestBufPipeTooManyWriteThenRead(t *testing.T) {
	pipe, err := stream.NewBufPipe(2, 0, sdr.SampleFormatU8)
	assert.NoError(t, err)

	b1 := make(sdr.SamplesU8, 1024)

	i, err := pipe.Write(b1)
	assert.NoError(t, err)
	assert.Equal(t, 1024, i)

	i, err = pipe.Write(b1)
	assert.NoError(t, err)
	assert.Equal(t, 1024, i)

	for i := 0; i < 2; i++ {
		// we have to do this due to the interactions with the go scheduler
		// and the `do` function. We'll only try twice, though.
		_, err = pipe.Write(b1)
		if err == nil {
			continue
		}
		assert.Equal(t, stream.ErrBufferOverrun, err)
		return
	}
	t.Fatal("Write did not overflow")

	_, err = pipe.Read(b1)
	assert.Equal(t, stream.ErrBufferOverrun, err)
}

func TestBufPipeMismatch(t *testing.T) {
	pipe, err := stream.NewBufPipe(1, 0, sdr.SampleFormatU8)
	assert.NoError(t, err)

	b1 := make(sdr.SamplesC64, 1024)

	i, err := pipe.Write(b1)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)
	assert.Equal(t, 0, i)

	i, err = pipe.Read(b1)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)
	assert.Equal(t, 0, i)
}

func TestBufPipeParts(t *testing.T) {
	pipe, err := stream.NewBufPipe(1, 0, sdr.SampleFormatC64)
	assert.NoError(t, err)

	wb := make(sdr.SamplesC64, 1024)
	wb[10] = complex64(20 + 10i)
	wb[512] = complex64(20 + 10i)
	i, err := pipe.Write(wb)
	assert.NoError(t, err)
	assert.Equal(t, 1024, i)

	buf := make(sdr.SamplesC64, 128)
	i, err = sdr.ReadFull(pipe, buf)
	assert.NoError(t, err)
	assert.Equal(t, 128, i)
	assert.Equal(t, complex64(20+10i), buf[10])
	buf = make(sdr.SamplesC64, 1024-128)
	i, err = sdr.ReadFull(pipe, buf)
	assert.NoError(t, err)
	assert.Equal(t, 1024-128, i)
	assert.Equal(t, complex64(20+10i), buf[512-128])
}

// vim: foldmethod=marker
