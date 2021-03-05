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
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
	"hz.tools/sdr/internal/simd"
)

func TestScaleMisalign(t *testing.T) {
	buf := make(sdr.SamplesC64, 1024)
	for i := range buf {
		buf[i] = complex64(complex(5, 5))
	}
	simd.ScaleComplex(2, buf[:1023])

	assert.Equal(t, complex(
		float32(5),
		float32(5),
	), buf[1023])

	for i := 0; i < 1022; i++ {
		assert.Equal(t, complex(
			float32(10),
			float32(10),
		), buf[i])
	}
}

func TestScaleComplex(t *testing.T) {
	buf := make(sdr.SamplesC64, 1024)
	for i := range buf {
		buf[i] = complex64(complex(5, 5))
	}
	simd.ScaleComplex(2, buf)
	for i := range buf {
		assert.Equal(t, complex(
			float32(10),
			float32(10),
		), buf[i])
	}
}

func TestRotateComplex(t *testing.T) {
	buf := make(sdr.SamplesC64, 10)
	for i := range buf {
		buf[i] = complex64(complex(0, 5))
	}
	simd.RotateComplex(0+1i, buf)
	for i := range buf {
		assert.Equal(t, complex(
			float32(-5),
			float32(0),
		), buf[i])
	}
}

func BenchmarkScaleComplex(b *testing.B) {
	buf := make(sdr.SamplesC64, 1024*16)
	for i := range buf {
		buf[i] = complex64(complex(5, 5))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		simd.ScaleComplex(2, buf)
	}
}

func BenchmarkRotateComplex(b *testing.B) {
	buf := make(sdr.SamplesC64, 1024*16)
	for i := range buf {
		buf[i] = complex64(complex(5, 5))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		simd.RotateComplex(0+1i, buf)
	}
}

// vim: foldmethod=marker
