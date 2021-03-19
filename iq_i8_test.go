// {{{ Copyright (c) Paul R. Tagliamonte <paul@k3xec.com>, 2021
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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
)

func TestConvertI8ToC64(t *testing.T) {
	i8Samples := make(sdr.SamplesI8, 1)
	c64Samples := make(sdr.SamplesC64, 1)

	i8Samples[0] = [2]int8{math.MaxInt8, math.MinInt8}
	_, err := i8Samples.ToC64(c64Samples)
	assert.NoError(t, err)
	assert.InEpsilon(t, 1, real(c64Samples[0]), 0.008)
	assert.InEpsilon(t, -1, imag(c64Samples[0]), 0.008)
}

func TestConvertI8ToI16(t *testing.T) {
	i8Samples := make(sdr.SamplesI8, 1)
	i16Samples := make(sdr.SamplesI16, 1)

	i8Samples[0] = [2]int8{math.MaxInt8, math.MinInt8}
	_, err := i8Samples.ToI16(i16Samples)
	assert.NoError(t, err)
	// Let's check that a max int8 is a max int16 with the low bits lopped off.
	assert.Equal(t, int16((math.MaxInt16>>8)<<8), i16Samples[0][0])
	assert.Equal(t, int16((math.MinInt16>>8)<<8), i16Samples[0][1])
}

func TestConvertI8ToU8(t *testing.T) {
	i8Samples := make(sdr.SamplesI8, 1)
	u8Samples := make(sdr.SamplesU8, 1)

	i8Samples[0] = [2]int8{math.MaxInt8, math.MinInt8}
	_, err := i8Samples.ToU8(u8Samples)
	assert.NoError(t, err)

	assert.Equal(t, uint8(0xFF), u8Samples[0][0])
	assert.Equal(t, uint8(0x00), u8Samples[0][1])
}

// vim: foldmethod=marker
