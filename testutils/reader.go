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
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
)

func TestReader(t *testing.T, name string, r sdr.Reader) {
	sf := r.SampleFormat()

	t.Run(name, func(t *testing.T) {
		if sf != sdr.SampleFormatU8 {
			t.Run("SampleFormatU8", func(t *testing.T) {
				testReaderSampleFormat(t, sdr.SampleFormatU8, r)
			})
		}
		if sf != sdr.SampleFormatI16 {
			t.Run("SampleFormatI16", func(t *testing.T) {
				testReaderSampleFormat(t, sdr.SampleFormatI16, r)
			})
		}
		if sf != sdr.SampleFormatC64 {
			t.Run("SampleFormatC64", func(t *testing.T) {
				testReaderSampleFormat(t, sdr.SampleFormatC64, r)
			})
		}

		t.Run("SampleRate", func(t *testing.T) {
			// We're just invoking this to ensure we don't panic.
			r.SampleRate()
		})
	})
}

func testReaderSampleFormat(t *testing.T, sf sdr.SampleFormat, r sdr.Reader) {
	s, err := sdr.MakeSamples(sf, 128)
	assert.NoError(t, err)
	_, err = r.Read(s)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)
}

// vim: foldmethod=marker
