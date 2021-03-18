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

// TestWriter will check that sample format mismatches will trigger the correct
// sdr errors.
func TestWriter(t *testing.T, name string, w sdr.Writer) {
	t.Run(name, func(t *testing.T) {
		t.Run("SampleFormatU8", func(t *testing.T) {
			testWriterSampleFormat(t, sdr.SampleFormatU8, w)
		})
		t.Run("SampleFormatI16", func(t *testing.T) {
			testWriterSampleFormat(t, sdr.SampleFormatI16, w)
		})
		t.Run("SampleFormatC64", func(t *testing.T) {
			testWriterSampleFormat(t, sdr.SampleFormatC64, w)
		})
		t.Run("SampleRate", func(t *testing.T) {
			// We're just invoking this to ensure we don't panic.
			w.SampleRate()
		})
	})
}

func testWriterSampleFormat(t *testing.T, sf sdr.SampleFormat, w sdr.Writer) {
	if sf == w.SampleFormat() {
		t.Skip()
		return
	}

	s, err := sdr.MakeSamples(sf, 128)
	assert.NoError(t, err)
	_, err = w.Write(s)
	assert.Equal(t, sdr.ErrSampleFormatMismatch, err)
}

// vim: foldmethod=marker
