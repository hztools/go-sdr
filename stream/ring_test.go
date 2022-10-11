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
	"io"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
	"hz.tools/sdr/stream"
)

func TestRingBuffer(t *testing.T) {
	rb, err := stream.NewRingBuffer(0, sdr.SampleFormatC64, stream.RingBufferOptions{
		Slots:      1024,
		SlotLength: 1024 * 32,
	})
	assert.NoError(t, err)
	assert.NotNil(t, rb)

	b := make(sdr.SamplesC64, 1024*32)
	for i := range b {
		b[i] = 1
	}

	_, err = rb.Write(b)
	assert.NoError(t, err)

	for i := range b {
		b[i] = 0
	}

	n, err := rb.Read(b)
	assert.NoError(t, err)
	assert.Equal(t, len(b), n)

	for i := range b {
		assert.Equal(t, complex64(complex(1, 0)), b[i])
	}
}

func TestRingBufferDstLength(t *testing.T) {
	b := make(sdr.SamplesC64, 1024*32)

	rb, err := stream.NewRingBuffer(0, sdr.SampleFormatC64, stream.RingBufferOptions{
		Slots:      10,
		SlotLength: 1024,
	})
	assert.NoError(t, err)
	assert.NotNil(t, rb)

	// Attempt to write a buffer that's too large
	_, err = rb.Write(b)
	assert.Error(t, err)

	// Attempt to read to a buffer that's too small
	_, err = rb.Read(b[:10])
	assert.Error(t, err)
}

func TestRingBufferOverrunBlocking(t *testing.T) {
	rb, err := stream.NewRingBuffer(0, sdr.SampleFormatC64, stream.RingBufferOptions{
		Slots:      10,
		SlotLength: 1024,
		BlockReads: true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, rb)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		br := make(sdr.SamplesC64, 1024)
		for i := 1; i < 5; i++ {
			_, err := rb.Read(br)
			assert.NoError(t, err)
			assert.Equal(t, complex64(complex(float32(i), 0)), br[0])
		}
	}()

	b := make(sdr.SamplesC64, 1024)
	b[0] = 1
	_, err = rb.Write(b)
	assert.NoError(t, err)

	b[0] = 2
	_, err = rb.Write(b)
	assert.NoError(t, err)

	b[0] = 3
	_, err = rb.Write(b)
	assert.NoError(t, err)

	b[0] = 4
	_, err = rb.Write(b)
	assert.NoError(t, err)

	wg.Wait()
}

func TestRingBufferOverrun(t *testing.T) {
	b := make(sdr.SamplesC64, 1024)

	rb, err := stream.NewRingBuffer(0, sdr.SampleFormatC64, stream.RingBufferOptions{
		Slots:      4,
		SlotLength: 1024,
	})
	assert.NoError(t, err)
	assert.NotNil(t, rb)

	b[0] = 1
	_, err = rb.Write(b)
	assert.NoError(t, err)

	b[0] = 2
	_, err = rb.Write(b)
	assert.NoError(t, err)

	b[0] = 3
	_, err = rb.Write(b)
	assert.NoError(t, err)

	b[0] = 4
	_, err = rb.Write(b)
	assert.NoError(t, err)

	// We've written four times to a 3 slot buffer, this will
	// override the first slot with the fourth buffer. Let's
	// check that we get sensible returns.

	b[0] = 0
	_, err = rb.Read(b)
	assert.NoError(t, err)
	assert.Equal(t, complex64(2), b[0])

	b[0] = 0
	_, err = rb.Read(b)
	assert.NoError(t, err)
	assert.Equal(t, complex64(3), b[0])

	b[0] = 0
	_, err = rb.Read(b)
	assert.NoError(t, err)
	assert.Equal(t, complex64(4), b[0])

	_, err = rb.Read(b)
	assert.Error(t, err)

}

func BenchmarkRing(b *testing.B) {
	rb, err := stream.NewRingBuffer(0, sdr.SampleFormatC64, stream.RingBufferOptions{
		Slots:      32,
		SlotLength: 1024,
	})
	assert.NoError(b, err)
	assert.NotNil(b, rb)

	wg := sync.WaitGroup{}

	br := make(sdr.SamplesC64, 1024)
	go func(r sdr.Reader) {
		defer wg.Done()
		for {
			_, err := r.Read(br)
			if err != nil {
				return
			}
		}
	}(rb)
	wg.Add(1)

	wb := make(sdr.SamplesC64, 1024)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		rb.Write(wb)
	}
}

func TestRingBufferCloseWithData(t *testing.T) {
	rb, err := stream.NewRingBuffer(0, sdr.SampleFormatC64, stream.RingBufferOptions{
		Slots:      10,
		SlotLength: 1024,
		BlockReads: true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, rb)

	b := make(sdr.SamplesC64, 1024)
	b[0] = 1
	_, err = rb.Write(b)
	assert.NoError(t, err)

	b[0] = 2
	_, err = rb.Write(b)
	assert.NoError(t, err)

	b[0] = 3
	_, err = rb.Write(b)
	assert.NoError(t, err)

	b[0] = 4
	_, err = rb.Write(b)
	assert.NoError(t, err)

	rb.(interface {
		CloseWithError(error) error
	}).CloseWithError(io.EOF)

	// Check that the write failed and that we got an Error.
	_, err = rb.Write(b)
	assert.Equal(t, io.EOF, err)

	for i := 1; i < 5; i++ {
		_, err := rb.Read(b)
		assert.NoError(t, err)
		assert.Equal(t, complex64(complex(float32(i), 0)), b[0])
	}

	_, err = rb.Read(b)
	assert.Equal(t, io.EOF, err)
}

// vim: foldmethod=marker
