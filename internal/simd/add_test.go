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

package simd_test

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"hz.tools/sdr"
	"hz.tools/sdr/internal/simd"
)

func TestAddComplex(t *testing.T) {
	a := make(sdr.SamplesC64, 4)
	b := make(sdr.SamplesC64, 4)
	c := make(sdr.SamplesC64, 4)

	v := complex(float32(10), float32(10))

	for i := range a {
		a[i] = v
		b[i] = v
	}

	assert.NoError(t, simd.AddComplex(a, b, c))

	for i := range a {
		assert.Equal(t, complex(float32(20), float32(20)), c[i])
	}
}

func TestAddComplexMisalign(t *testing.T) {
	a := make(sdr.SamplesC64, 5)
	b := make(sdr.SamplesC64, 5)
	c := make(sdr.SamplesC64, 5)

	v := complex(float32(10), float32(10))

	for i := range a {
		a[i] = v
		b[i] = v
	}

	assert.NoError(t, simd.AddComplex(a, b, c))

	for i := range a {
		assert.Equal(t, complex(float32(20), float32(20)), c[i])
	}
}

func TestAddComplexSelf(t *testing.T) {
	a := make(sdr.SamplesC64, 1024)
	b := make(sdr.SamplesC64, 1024)

	v := complex(float32(10), float32(10))

	for i := range a {
		a[i] = v
		b[i] = v
	}

	assert.NoError(t, simd.AddComplex(a, b, a))

	for i := range a {
		assert.Equal(t, complex(float32(20), float32(20)), a[i])
	}
}

func TestAddComplexMisalignSub(t *testing.T) {
	a := make(sdr.SamplesC64, 300)
	b := make(sdr.SamplesC64, 300)
	c := make(sdr.SamplesC64, 300)

	v := complex(float32(10), float32(10))

	for i := range a {
		a[i] = v
		b[i] = v
	}

	assert.NoError(t, simd.AddComplex(a[100:201], b[100:201], c[100:201]))

	for _, el := range c[:100] {
		assert.Equal(t, complex(float32(0), float32(0)), el)
	}
	for _, el := range c[100:201] {
		assert.Equal(t, complex(float32(20), float32(20)), el)
	}
	for _, el := range c[201:] {
		assert.Equal(t, complex(float32(0), float32(0)), el)
	}
}

func BenchmarkAddComplex(b *testing.B) {
	bufa := make(sdr.SamplesC64, 1024*8)
	bufb := make(sdr.SamplesC64, 1024*8)
	bufc := make(sdr.SamplesC64, 1024*8)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		simd.AddComplex(bufa, bufb, bufc)
	}
}

// vim: foldmethod=marker
