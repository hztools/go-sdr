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
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
)

func TestSamplesPool(t *testing.T) {
	pool, err := sdr.NewSamplesPool(sdr.SampleFormatC64, 1024*32)
	assert.NoError(t, err)
	assert.NotNil(t, pool)

	buf := pool.Get()
	assert.NotNil(t, buf)
	assert.Equal(t, 1024*32, buf.Length())
	buf.(sdr.SamplesC64)[0] = 1 + 1i

	buf1 := pool.Get()
	assert.NotNil(t, buf1)
	assert.Equal(t, 1024*32, buf1.Length())
	buf1.(sdr.SamplesC64)[0] = 2 + 2i

	// TODO(paultag): This behavior is not actually something we can depend
	// on, but I have no other way to check allocations. I'm going to hope
	// that this behavior doens't change, and since we only use this for
	// testing it's maybe better.

	// Do *NOT* copy this for real code, or depend on ordering. The worst that
	// happens here is a test failure. The worst that happens in real life is
	// an FCC fine :)

	pool.Put(buf)
	buf = pool.Get()
	assert.Equal(t, complex64(1+1i), buf.(sdr.SamplesC64)[0])

	pool.Put(buf1)
	buf1 = pool.Get()
	assert.Equal(t, complex64(2+2i), buf1.(sdr.SamplesC64)[0])
}

func TestSamplesPoolTypes(t *testing.T) {
	for _, sampleFormat := range []sdr.SampleFormat{
		sdr.SampleFormatC64,
		sdr.SampleFormatI16,
		sdr.SampleFormatI8,
		sdr.SampleFormatU8,
	} {
		t.Run(sampleFormat.String(), func(t *testing.T) {
			pool, err := sdr.NewSamplesPool(sampleFormat, 1024*32)
			assert.NoError(t, err)
			assert.NotNil(t, pool)
			buf := pool.Get()
			assert.NotNil(t, buf)
			assert.Equal(t, 1024*32, buf.Length())
		})
	}
}

// vim: foldmethod=marker
