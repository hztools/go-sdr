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

func TestConvertI16ToU8(t *testing.T) {
	t.Skip()
}

func TestConvertI16ToI8(t *testing.T) {
	i16Samples := make(sdr.SamplesI16, 1)
	i8Samples := make(sdr.SamplesI8, 1)

	i16Samples[0] = [2]int16{math.MaxInt16, math.MinInt16}
	_, err := i16Samples.ToI8(i8Samples)
	assert.NoError(t, err)
	assert.Equal(t, int8(math.MaxInt8), i8Samples[0][0])
	assert.Equal(t, int8(math.MinInt8), i8Samples[0][1])
}

func TestConvertI16ToC64(t *testing.T) {
	i16Samples := make(sdr.SamplesI16, 1)
	c64Samples := make(sdr.SamplesC64, 1)

	i16Samples[0] = [2]int16{math.MaxInt16, math.MaxInt16}
	_, err := i16Samples.ToC64(c64Samples)
	assert.NoError(t, err)
	assert.InEpsilon(t, 1, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, 1, imag(c64Samples[0]), epsilon)

	_, err = sdr.ConvertBuffer(c64Samples, i16Samples)
	assert.NoError(t, err)
	assert.InEpsilon(t, 1, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, 1, imag(c64Samples[0]), epsilon)

	i16Samples[0] = [2]int16{math.MinInt16, math.MinInt16}
	_, err = i16Samples.ToC64(c64Samples)
	assert.NoError(t, err)
	assert.InEpsilon(t, -1, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, -1, imag(c64Samples[0]), epsilon)

	_, err = sdr.ConvertBuffer(c64Samples, i16Samples)
	assert.NoError(t, err)
	assert.InEpsilon(t, -1, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, -1, imag(c64Samples[0]), epsilon)

	i16Samples[0] = [2]int16{0, 0}
	i16Samples.ToC64(c64Samples)
	assert.NoError(t, err)
	assert.InEpsilon(t, 1, 1+real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, 1, 1+imag(c64Samples[0]), epsilon)
	_, err = sdr.ConvertBuffer(c64Samples, i16Samples)
	assert.NoError(t, err)
	assert.InEpsilon(t, 1, 1+real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, 1, 1+imag(c64Samples[0]), epsilon)

	i16Samples[0] = [2]int16{math.MaxInt16 / 2, math.MaxInt16 / 2}
	_, err = i16Samples.ToC64(c64Samples)
	assert.NoError(t, err)
	assert.InEpsilon(t, 0.5, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, 0.5, imag(c64Samples[0]), epsilon)
	_, err = sdr.ConvertBuffer(c64Samples, i16Samples)
	assert.NoError(t, err)
	assert.InEpsilon(t, 0.5, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, 0.5, imag(c64Samples[0]), epsilon)

	i16Samples[0] = [2]int16{math.MinInt16 / 2, math.MinInt16 / 2}
	_, err = i16Samples.ToC64(c64Samples)
	assert.NoError(t, err)
	assert.InEpsilon(t, -0.5, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, -0.5, imag(c64Samples[0]), epsilon)
	_, err = sdr.ConvertBuffer(c64Samples, i16Samples)
	assert.NoError(t, err)
	assert.InEpsilon(t, -0.5, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, -0.5, imag(c64Samples[0]), epsilon)
}

// vim: foldmethod=marker
