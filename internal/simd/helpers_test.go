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

//go:build sdr.simdtest
// +build sdr.simdtest

package simd_test

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
	"hz.tools/sdr/internal/simd"
)

func TestSimdTestHelpersAddr(t *testing.T) {
	buf := make(sdr.SamplesC64, 10)
	baseAddr := uintptr(unsafe.Pointer(&buf[0]))
	assert.Equal(t, baseAddr, simd.InternalBufferAddr(buf))
}

func TestSimdTestHelpersLen(t *testing.T) {
	buf := make(sdr.SamplesC64, 10)
	assert.Equal(t, 10, simd.InternalBufferLen(buf))
}

func TestSimdTestHelpersSize(t *testing.T) {
	buf := make(sdr.SamplesC64, 10)
	assert.Equal(t, 10*8, simd.InternalBufferSize(buf))
}

func TestSimdTestHelpersSizeSecond(t *testing.T) {
	buf := make(sdr.SamplesC64, 10)
	assert.Equal(t, 10*8, simd.InternalBufferSizeSecond([]complex64{}, buf))
}

// vim: foldmethod=marker
