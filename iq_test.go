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
	"testing"

	"hz.tools/sdr"
)

func TestMakeSamples(t *testing.T) {
	samples, err := sdr.MakeSamples(sdr.SampleFormatU8, 1024)
	assert.NoError(t, err)

	samplesU8, ok := samples.(sdr.SamplesU8)
	assert.True(t, ok)
	assert.Equal(t, 1024, len(samplesU8))

	samples, err = sdr.MakeSamples(sdr.SampleFormatI16, 1024)
	assert.True(t, ok)
	samplesI16, ok := samples.(sdr.SamplesI16)
	assert.True(t, ok)
	assert.Equal(t, 1024, len(samplesI16))

	samples, err = sdr.MakeSamples(sdr.SampleFormatC64, 1024)
	assert.True(t, ok)
	samplesC64, ok := samples.(sdr.SamplesC64)
	assert.True(t, ok)
	assert.Equal(t, 1024, len(samplesC64))

}

// vim: foldmethod=marker
