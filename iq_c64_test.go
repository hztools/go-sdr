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
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/sdr"
)

var (
	// epsilon is used in the test suite to determine if two floating point
	// numbers are the "same", due to floating point errors.
	epsilon float64 = 0.0001
)

func TestConvertC64ToU8(t *testing.T) {
	c64Samples := make(sdr.SamplesC64, 1)
	u8Samples := make(sdr.SamplesU8, 1)

	c64Samples[0] = complex(1, 1)
	_, err := c64Samples.ToU8(u8Samples)
	assert.NoError(t, err)
	assert.Equal(t, [2]uint8{255, 255}, u8Samples[0])
	_, err = sdr.ConvertBuffer(u8Samples, c64Samples)
	assert.NoError(t, err)
	assert.Equal(t, [2]uint8{255, 255}, u8Samples[0])

	c64Samples[0] = complex(-1, -1)
	_, err = c64Samples.ToU8(u8Samples)
	assert.NoError(t, err)
	assert.Equal(t, [2]uint8{0, 0}, u8Samples[0])
	_, err = sdr.ConvertBuffer(u8Samples, c64Samples)
	assert.NoError(t, err)
	assert.Equal(t, [2]uint8{0, 0}, u8Samples[0])

	c64Samples[0] = complex(0, 0)
	_, err = c64Samples.ToU8(u8Samples)
	assert.NoError(t, err)
	assert.Equal(t, [2]uint8{127, 127}, u8Samples[0])
	_, err = sdr.ConvertBuffer(u8Samples, c64Samples)
	assert.NoError(t, err)
	assert.Equal(t, [2]uint8{127, 127}, u8Samples[0])
}

func TestConvertC64ToI16(t *testing.T) {
	c64Samples := make(sdr.SamplesC64, 1)
	i16Samples := make(sdr.SamplesI16, 1)

	c64Samples[0] = complex(1, 1)
	_, err := c64Samples.ToI16(i16Samples)
	assert.NoError(t, err)
	assert.Equal(t, [2]int16{math.MaxInt16, math.MaxInt16}, i16Samples[0])
	_, err = sdr.ConvertBuffer(i16Samples, i16Samples)
	assert.NoError(t, err)
	assert.Equal(t, [2]int16{math.MaxInt16, math.MaxInt16}, i16Samples[0])

	c64Samples[0] = complex(-1, -1)
	_, err = c64Samples.ToI16(i16Samples)
	assert.NoError(t, err)
	assert.Equal(t, [2]int16{math.MinInt16 + 1, math.MinInt16 + 1}, i16Samples[0])
	_, err = sdr.ConvertBuffer(i16Samples, i16Samples)
	assert.NoError(t, err)
	assert.Equal(t, [2]int16{math.MinInt16 + 1, math.MinInt16 + 1}, i16Samples[0])

	c64Samples[0] = complex(0, 0)
	_, err = c64Samples.ToI16(i16Samples)
	assert.NoError(t, err)
	assert.Equal(t, [2]int16{0, 0}, i16Samples[0])
	_, err = sdr.ConvertBuffer(i16Samples, i16Samples)
	assert.NoError(t, err)
	assert.Equal(t, [2]int16{0, 0}, i16Samples[0])
}

func TestConvertC64ToI8(t *testing.T) {
	c64Samples := make(sdr.SamplesC64, 1)
	i8Samples := make(sdr.SamplesI8, 1)

	c64Samples[0] = complex(1, -1)
	_, err := c64Samples.ToI8(i8Samples)
	assert.NoError(t, err)

	// *sigh*
	//   This will clip a -1 to less than -1 in int8 due to max / min being
	//   different because of the signed bit.
	assert.Equal(t, [2]int8{127, -127}, i8Samples[0])
}

func TestC64Scale(t *testing.T) {
	c64Samples := make(sdr.SamplesC64, 31)
	for i := range c64Samples {
		c64Samples[i] = complex64(complex(10, 10))
	}
	c64Samples.Scale(0.5)
	for _, el := range c64Samples {
		assert.Equal(t, float32(5), real(el))
		assert.Equal(t, float32(5), imag(el))
	}
}

func TestC64Multiply(t *testing.T) {
	c64Samples := make(sdr.SamplesC64, 31)
	for i := range c64Samples {
		c64Samples[i] = complex64(complex(10, 10))
	}
	c64Samples.Multiply(complex(0.5, 0.5))
	for _, el := range c64Samples {
		assert.Equal(t, complex64(complex(0, 10)), el)
	}
}

func TestC64Add(t *testing.T) {
	c64Samples1 := make(sdr.SamplesC64, 31)
	c64Samples2 := make(sdr.SamplesC64, 31)
	for i := range c64Samples1 {
		c64Samples1[i] = complex64(complex(1, 3))
		c64Samples2[i] = complex64(complex(3, 1))
	}
	c64Samples1.Add(c64Samples2)
	for i, el := range c64Samples1 {
		assert.Equal(t, complex64(complex(4, 4)), el)
		assert.Equal(t, complex64(complex(3, 1)), c64Samples2[i])
	}
}

// vim: foldmethod=marker
