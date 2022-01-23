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

//go:build arm64 && !sdr.nosimd
// +build arm64,!sdr.nosimd

package simd

func scaleComplex(v float32, buf []complex64) {
	neonScaleComplex([4]float32{v, v, v, v}, buf)

	lenBuf := len(buf)
	if rem := lenBuf % 2; rem != 0 {
		for i := 0; i < rem; i++ {
			idx := ((lenBuf - 1) - i)
			buf[idx] = scaleComplexByReal(v, buf[idx])
		}
	}
}

func rotateComplex(v complex64, buf []complex64) {
	var (
		vr float32 = real(v)
		vi float32 = imag(v)
	)
	neonRotateComplex(
		[4]float32{vr, vr, vr, vr},
		[4]float32{vi, vi, vi, vi},
		buf,
	)

	lenBuf := len(buf)
	if rem := lenBuf % 4; rem != 0 {
		rotateComplexNative(v, buf[lenBuf-rem:])
	}
}

func neonScaleComplex([4]float32, []complex64)

func neonRotateComplex([4]float32, [4]float32, []complex64)

// vim: foldmethod=marker
