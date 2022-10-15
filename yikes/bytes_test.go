// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2022
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

package yikes_test

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
	"hz.tools/sdr/yikes"
)

func TestSamplesC64(t *testing.T) {
	buf := make(sdr.SamplesC64, 1024*32)
	buf[2] = 1 + 1i
	base := uintptr(unsafe.Pointer(&buf[0]))
	l := len(buf)
	buf2, err := yikes.Samples(base, l, sdr.SampleFormatC64)
	assert.NoError(t, err)
	assert.Equal(t, complex64(1+1i), buf2.(sdr.SamplesC64)[2])
	assert.Equal(t, len(buf), len(buf2.(sdr.SamplesC64)))
}

func TestSamplesUnknown(t *testing.T) {
	buf := make(sdr.SamplesC64, 1024*32)
	buf[2] = 1 + 1i
	base := uintptr(unsafe.Pointer(&buf[0]))
	l := len(buf)
	_, err := yikes.Samples(base, l, sdr.SampleFormat(254))
	assert.Equal(t, sdr.ErrSampleFormatUnknown, err)
}

// vim: foldmethod=marker
