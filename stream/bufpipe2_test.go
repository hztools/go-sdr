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
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
	"hz.tools/sdr/stream"
)

func TestBufPipe2Basic(t *testing.T) {
	pipe, err := stream.NewBufPipe2(1, 0, sdr.SampleFormatU8)
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

func TestBufPipe2Close(t *testing.T) {
	pipe, err := stream.NewBufPipe2(1, 0, sdr.SampleFormatU8)
	assert.NoError(t, err)

	b1 := make(sdr.SamplesU8, 1024)
	for i := range b1 {
		b1[i][0] = 0xFF
		b1[i][1] = 0xFF
	}

	i, err := pipe.Write(b1)
	assert.NoError(t, err)
	assert.Equal(t, 1024, i)

	// Close the pipe
	pipe.Close()

	b2 := make(sdr.SamplesU8, 1024)
	i, err = sdr.ReadFull(pipe, b2)
	assert.NoError(t, err)
	assert.Equal(t, 1024, i)

	for i := range b2 {
		assert.Equal(t, uint8(0xFF), b2[i][0])
		assert.Equal(t, uint8(0xFF), b2[i][1])
	}

	_, err = pipe.Write(b1)
	assert.Error(t, err)
}

func BenchmarkBufPipe2(b *testing.B) {
	for _, i := range []int{0, 1, 2, 4, 8, 16, 32, 64, 128} {
		b.Run(fmt.Sprintf("Cap-%d", i), func(b *testing.B) {
			pipe, err := stream.NewBufPipe2(i, 0, sdr.SampleFormatC64)
			assert.NoError(b, err)

			wg := sync.WaitGroup{}

			rb := make(sdr.SamplesC64, 1024)
			go func(r sdr.Reader) {
				defer wg.Done()
				for {
					_, err := r.Read(rb)
					if err != nil {
						return
					}
				}
			}(pipe)
			wg.Add(1)

			wb := make(sdr.SamplesC64, 1024)

			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				pipe.Write(wb)
			}
			b.StopTimer()
			wg.Wait()
		})
	}
}

// vim: foldmethod=marker
