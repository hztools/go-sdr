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
	"github.com/stretchr/testify/assert"
	"math"
	"testing"

	"hz.tools/sdr"
)

func TestConvertI16ToU8(t *testing.T) {
	t.Skip()
}

func TestConvertI16ToC64(t *testing.T) {
	i16Samples := make(sdr.SamplesI16, 1)
	c64Samples := make(sdr.SamplesC64, 1)

	i16Samples[0] = [2]int16{math.MaxInt16, math.MaxInt16}
	assert.NoError(t, i16Samples.ToC64(c64Samples))
	assert.InEpsilon(t, 1, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, 1, imag(c64Samples[0]), epsilon)

	assert.NoError(t, sdr.ConvertBuffer(c64Samples, i16Samples))
	assert.InEpsilon(t, 1, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, 1, imag(c64Samples[0]), epsilon)

	i16Samples[0] = [2]int16{math.MinInt16, math.MinInt16}
	assert.NoError(t, i16Samples.ToC64(c64Samples))
	assert.InEpsilon(t, -1, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, -1, imag(c64Samples[0]), epsilon)

	assert.NoError(t, sdr.ConvertBuffer(c64Samples, i16Samples))
	assert.InEpsilon(t, -1, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, -1, imag(c64Samples[0]), epsilon)

	i16Samples[0] = [2]int16{0, 0}
	assert.NoError(t, i16Samples.ToC64(c64Samples))
	assert.InEpsilon(t, 1, 1+real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, 1, 1+imag(c64Samples[0]), epsilon)
	assert.NoError(t, sdr.ConvertBuffer(c64Samples, i16Samples))
	assert.InEpsilon(t, 1, 1+real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, 1, 1+imag(c64Samples[0]), epsilon)

	i16Samples[0] = [2]int16{math.MaxInt16 / 2, math.MaxInt16 / 2}
	assert.NoError(t, i16Samples.ToC64(c64Samples))
	assert.InEpsilon(t, 0.5, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, 0.5, imag(c64Samples[0]), epsilon)
	assert.NoError(t, sdr.ConvertBuffer(c64Samples, i16Samples))
	assert.InEpsilon(t, 0.5, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, 0.5, imag(c64Samples[0]), epsilon)

	i16Samples[0] = [2]int16{math.MinInt16 / 2, math.MinInt16 / 2}
	assert.NoError(t, i16Samples.ToC64(c64Samples))
	assert.InEpsilon(t, -0.5, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, -0.5, imag(c64Samples[0]), epsilon)
	assert.NoError(t, sdr.ConvertBuffer(c64Samples, i16Samples))
	assert.InEpsilon(t, -0.5, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, -0.5, imag(c64Samples[0]), epsilon)
}

// vim: foldmethod=marker
