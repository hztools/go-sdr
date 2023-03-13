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

package testutils

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
)

// TestReadWriteSamples will test that writing a specific number of samples
// comes out the Reader on the other end.
func TestReadWriteSamples(t *testing.T, name string, r sdr.Reader, w sdr.Writer) {
	t.Run(name, func(t *testing.T) {
		var (
			sampleChunk  = 1024
			sampleChunks = 32
			wg           = sync.WaitGroup{}
		)

		go func() {
			defer wg.Done()
			wb, err := sdr.MakeSamples(w.SampleFormat(), sampleChunk)
			assert.NoError(t, err)

			for i := 0; i < sampleChunks; i++ {
				i, err := w.Write(wb)
				assert.NoError(t, err)
				assert.Equal(t, sampleChunk, i)
			}
		}()
		wg.Add(1)

		rb, err := sdr.MakeSamples(r.SampleFormat(), sampleChunk*sampleChunks)
		assert.NoError(t, err)
		i, err := sdr.ReadFull(r, rb)
		assert.NoError(t, err)
		assert.Equal(t, sampleChunk*sampleChunks, i)

		wg.Wait()
	})
}

// TestReader will check that sample mismatches trigger the correct SDR Errors.
func TestReader(t *testing.T, name string, r sdr.Reader) {
	t.Run(name, func(t *testing.T) {
		t.Run("SampleFormatU8", func(t *testing.T) {
			testReaderSampleFormat(t, sdr.SampleFormatU8, r)
		})
		t.Run("SampleFormatI8", func(t *testing.T) {
			testReaderSampleFormat(t, sdr.SampleFormatI8, r)
		})
		t.Run("SampleFormatI16", func(t *testing.T) {
			testReaderSampleFormat(t, sdr.SampleFormatI16, r)
		})
		t.Run("SampleFormatC64", func(t *testing.T) {
			testReaderSampleFormat(t, sdr.SampleFormatC64, r)
		})
		t.Run("SampleRate", func(t *testing.T) {
			// We're just invoking this to ensure we don't panic.
			r.SampleRate()
		})
	})
}

func testReaderSampleFormat(t *testing.T, sf sdr.SampleFormat, r sdr.Reader) {
	if sf == r.SampleFormat() {
		t.Skip()
		return
	}

	s, err := sdr.MakeSamples(sf, 128)
	assert.NoError(t, err)
	_, err = r.Read(s)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)
}

// vim: foldmethod=marker
