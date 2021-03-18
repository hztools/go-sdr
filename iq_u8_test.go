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

func TestByteSliceRead(t *testing.T) {
	buf := make(sdr.SamplesU8, 10)
	buf[0] = [2]uint8{0x01, 0x02}
	bufBytes, err := sdr.UnsafeSamplesAsBytes(buf)
	assert.NoError(t, err)
	assert.Equal(t, uint8(0x01), bufBytes[0])
	assert.Equal(t, uint8(0x02), bufBytes[1])
}

func TestByteSliceWrite(t *testing.T) {
	buf := make(sdr.SamplesU8, 10)
	bufBytes, err := sdr.UnsafeSamplesAsBytes(buf)
	assert.NoError(t, err)
	bufBytes[0] = 0x03
	bufBytes[1] = 0x04
	assert.Equal(t, uint8(0x03), buf[0][0])
	assert.Equal(t, uint8(0x04), buf[0][1])
}

func TestConvertU8ToC64Long(t *testing.T) {
	u8Samples := make(sdr.SamplesU8, 1024)
	c64Samples := make(sdr.SamplesC64, 1024)
	for i := range u8Samples {
		u8Samples[i] = [2]uint8{0xFF, 0xFF}
	}
	_, err := u8Samples.ToC64(c64Samples)
	assert.NoError(t, err)
	for i := range c64Samples {
		assert.InEpsilon(t, 1, real(c64Samples[i]), epsilon)
		assert.InEpsilon(t, 1, imag(c64Samples[i]), epsilon)
	}
}

func TestConvertU8ToC64OverUnder(t *testing.T) {
	u8Samples := make(sdr.SamplesU8, 1024)
	c64Samples := make(sdr.SamplesC64, 1024)
	for i := range u8Samples {
		u8Samples[i] = [2]uint8{0xFF, 0xFF}
	}
	_, err := u8Samples[32:69].ToC64(c64Samples[32:69])
	assert.NoError(t, err)
	for i := 0; i < 32; i++ {
		assert.Equal(t, float32(0), real(c64Samples[i]))
		assert.Equal(t, float32(0), imag(c64Samples[i]))
	}
	for i := 32; i < 69; i++ {
		assert.InEpsilon(t, 1, real(c64Samples[i]), epsilon)
		assert.InEpsilon(t, 1, imag(c64Samples[i]), epsilon)
	}
	for i := 69; i < len(c64Samples); i++ {
		assert.Equal(t, float32(0), real(c64Samples[i]))
		assert.Equal(t, float32(0), imag(c64Samples[i]))
	}
}

func TestConvertU8ToC64Short(t *testing.T) {
	u8Samples := make(sdr.SamplesU8, 16)
	c64Samples := make(sdr.SamplesC64, 16)

	u8Samples[0] = [2]uint8{255, 255}
	_, err := u8Samples.ToC64(c64Samples)
	assert.NoError(t, err)

	assert.InEpsilon(t, 1, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, 1, imag(c64Samples[0]), epsilon)

	_, err = sdr.ConvertBuffer(c64Samples, u8Samples)
	assert.NoError(t, err)
	assert.InEpsilon(t, 1, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, 1, imag(c64Samples[0]), epsilon)

	u8Samples[0] = [2]uint8{0, 0}
	_, err = u8Samples.ToC64(c64Samples)
	assert.NoError(t, err)

	assert.InEpsilon(t, -1, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, -1, imag(c64Samples[0]), epsilon)
	_, err = sdr.ConvertBuffer(c64Samples, u8Samples)
	assert.NoError(t, err)
	assert.InEpsilon(t, -1, real(c64Samples[0]), epsilon)
	assert.InEpsilon(t, -1, imag(c64Samples[0]), epsilon)

	// 0 is 127.5 so, we should get 0 by taking two samples, one slightly
	// below 0 and 1 slightly over 0 (127 and 128) and checking how that stacks
	u8Samples[0] = [2]uint8{128, 128}
	u8Samples[1] = [2]uint8{127, 127}
	_, err = u8Samples.ToC64(c64Samples)
	assert.NoError(t, err)

	i := real(c64Samples[0]) + real(c64Samples[1])
	q := imag(c64Samples[0]) + imag(c64Samples[1])

	assert.InEpsilon(t, 1, i+1, epsilon)
	assert.InEpsilon(t, 1, q+1, epsilon)

	_, err = sdr.ConvertBuffer(c64Samples, u8Samples)
	assert.NoError(t, err)
	assert.InEpsilon(t, 1, i+1, epsilon)
	assert.InEpsilon(t, 1, q+1, epsilon)

}

func TestConvertU8ToI16(t *testing.T) {
	u8Samples := make(sdr.SamplesU8, 2)
	i16Samples := make(sdr.SamplesI16, 2)

	u8Samples[0] = [2]uint8{255, 255}
	_, err := u8Samples.ToI16(i16Samples)
	assert.NoError(t, err)

	// the max value here will be 255 shifted up. It's close enough.
	assert.Equal(t, i16Samples[0], [2]int16{
		math.MaxInt16 & 0xFF00,
		math.MaxInt16 & 0xFF00,
	})

	u8Samples[0] = [2]uint8{0, 0}
	_, err = u8Samples.ToI16(i16Samples)
	assert.NoError(t, err)
	assert.Equal(t, i16Samples[0], [2]int16{math.MinInt16, math.MinInt16})

	_, err = sdr.ConvertBuffer(i16Samples, u8Samples)
	assert.NoError(t, err)
	assert.Equal(t, i16Samples[0], [2]int16{math.MinInt16, math.MinInt16})
}

func BenchmarkConvertU8ToC64(b *testing.B) {
	in := make(sdr.SamplesU8, 1024*16)
	out := make(sdr.SamplesC64, 1024*16)

	for i := range in {
		in[i] = [2]uint8{0xFF, 0xFF}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		in.ToC64(out)
	}
}

// vim: foldmethod=marker
