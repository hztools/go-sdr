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

package stream_test

import (
	"math/rand"

	"github.com/stretchr/testify/assert"
	"testing"

	"hz.tools/sdr"
	"hz.tools/sdr/stream"
)

func nextComplex(r *rand.Rand, stdDev float64) complex64 {
	i := float32(r.NormFloat64() * stdDev)
	q := float32(r.NormFloat64() * stdDev)

	if i < -1 {
		i = -1
	}
	if q < -1 {
		q = -1
	}
	if i > 1 {
		i = 1
	}
	if q > 1 {
		q = 1
	}

	return complex(i, q)
}

func TestNoiseReader(t *testing.T) {
	randReader := stream.Noise(stream.NoiseConfig{
		Source:            rand.NewSource(1337),
		StandardDeviation: 1.0,
	})

	buf := make(sdr.SamplesC64, 1024*32)
	_, err := sdr.ReadFull(randReader, buf)
	assert.NoError(t, err)

	r := rand.New(rand.NewSource(1337))
	for i := range buf {
		refSample := nextComplex(r, 1.0)
		assert.Equal(t, refSample, buf[i])
	}
}

// vim: foldmethod=marker
