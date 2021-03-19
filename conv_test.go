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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
)

type unsupportedFormat [][2]int32

func (uf unsupportedFormat) Slice(s, e int) sdr.Samples {
	return uf[s:e]
}

func (uf unsupportedFormat) Format() sdr.SampleFormat {
	return sdr.SampleFormat(100)
}

func (uf unsupportedFormat) Length() int {
	return len(uf)
}

func (uf unsupportedFormat) Size() int {
	return 8
}

func TestConvertBuffer(t *testing.T) {
	allFormats := map[string]sdr.SampleFormat{
		"U8":  sdr.SampleFormatU8,
		"I8":  sdr.SampleFormatI8,
		"I16": sdr.SampleFormatI16,
		"C64": sdr.SampleFormatC64,
	}

	for inFormatName, inFormat := range allFormats {
		for outFormatName, outFormat := range allFormats {
			t.Run(fmt.Sprintf("Conv-%s-%s", inFormatName, outFormatName), func(t *testing.T) {
				in, err := sdr.MakeSamples(inFormat, 1024)
				assert.NoError(t, err)
				out, err := sdr.MakeSamples(outFormat, 1024)
				assert.NoError(t, err)

				i, err := sdr.ConvertBuffer(out, in)
				assert.NoError(t, err)
				assert.Equal(t, 1024, i)
			})
		}
	}

	unBuffer := make(unsupportedFormat, 1024)

	for formatName, format := range allFormats {
		t.Run(fmt.Sprintf("Conv-UNK-%s", formatName), func(t *testing.T) {
			samples, err := sdr.MakeSamples(format, 1024)
			assert.NoError(t, err)

			_, err = sdr.ConvertBuffer(samples, unBuffer)
			assert.Equal(t, sdr.ErrConversionNotImplemented, err)
		})
		t.Run(fmt.Sprintf("Conv-%s-UNK", formatName), func(t *testing.T) {
			samples, err := sdr.MakeSamples(format, 1024)
			assert.NoError(t, err)

			_, err = sdr.ConvertBuffer(unBuffer, samples)
			assert.Equal(t, sdr.ErrSampleFormatUnknown, err)
		})
	}
}

// vim: foldmethod=marker
