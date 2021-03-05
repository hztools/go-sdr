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

package fft_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"hz.tools/rf"
	"hz.tools/sdr/fft"
)

func complexTestArray(dst []complex64) {
	nyquest := len(dst) / 2
	for i := 0; i < nyquest; i++ {
		dst[i] = complex(float32(i), 0)
	}
	for i := 0; i < nyquest; i++ {
		dst[nyquest+i] = complex(float32(i-(nyquest)), 0)
	}
}

func TestFreqBinOOR(t *testing.T) {
	freq := make([]complex64, 2048)

	_, err := fft.BinByFreq(freq, 2048, false, rf.MHz)
	assert.Equal(t, fft.ErrFrequencyOutOfSamplingRange, err)

	_, err = fft.BinByFreq(freq, 2048, false, -rf.MHz)
	assert.Equal(t, fft.ErrFrequencyOutOfSamplingRange, err)
}

func TestFreqBinRangeNyquestZero(t *testing.T) {
	freq := make([]complex64, 2048)
	bins, err := fft.BinsByRange(freq, 2048, fft.ZeroFirst, rf.Range{rf.Hz(-1023), rf.Hz(1023)})
	assert.Equal(t, 2047, len(bins))
	assert.NoError(t, err)
	myBins := map[int]bool{}
	for _, bi := range bins {
		assert.Equal(t, float32(0), real(freq[bi]))
		myBins[bi] = true
	}
	assert.Equal(t, 2047, len(myBins))
}

func TestFreqBinRangeNyquestNeg(t *testing.T) {
	freq := make([]complex64, 2048)
	bins, err := fft.BinsByRange(freq, 2048, fft.NegativeFirst, rf.Range{rf.Hz(-1023), rf.Hz(1023)})
	assert.Equal(t, 2047, len(bins))
	assert.NoError(t, err)
	myBins := map[int]bool{}
	for _, bi := range bins {
		assert.Equal(t, float32(0), real(freq[bi]))
		myBins[bi] = true
	}
	assert.Equal(t, 2047, len(myBins))
}

func TestFreqBinRange(t *testing.T) {
	freq := make([]complex64, 2048)
	complexTestArray(freq)

	// ZeroFirst

	bins, err := fft.BinsByRange(freq, 2048, fft.ZeroFirst, rf.Range{rf.Hz(0), rf.Hz(10)})
	assert.NoError(t, err)
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, bins)

	bins, err = fft.BinsByRange(freq, 2048, fft.ZeroFirst, rf.Range{rf.Hz(-10), rf.Hz(-1)})
	assert.NoError(t, err)
	assert.Equal(t, []int{2038, 2039, 2040, 2041, 2042, 2043, 2044, 2045, 2046, 2047}, bins)

	bins, err = fft.BinsByRange(freq, 2048, fft.ZeroFirst, rf.Range{rf.Hz(-5), rf.Hz(5)})
	assert.NoError(t, err)
	assert.Equal(t, []int{2043, 2044, 2045, 2046, 2047, 0, 1, 2, 3, 4, 5}, bins)

	bins, err = fft.BinsByRange(freq, 2048, fft.ZeroFirst, rf.Range{rf.Hz(-10), rf.Hz(0)})
	assert.NoError(t, err)
	assert.Equal(t, []int{2038, 2039, 2040, 2041, 2042, 2043, 2044, 2045, 2046, 2047, 0}, bins)

	// NegativeFirst

	bins, err = fft.BinsByRange(freq, 2048, fft.NegativeFirst, rf.Range{rf.Hz(0), rf.Hz(10)})
	assert.NoError(t, err)
	assert.Equal(t, []int{1024, 1025, 1026, 1027, 1028, 1029, 1030, 1031, 1032, 1033, 1034}, bins)

	bins, err = fft.BinsByRange(freq, 2048, fft.NegativeFirst, rf.Range{rf.Hz(-10), rf.Hz(-1)})
	assert.NoError(t, err)
	assert.Equal(t, []int{1014, 1015, 1016, 1017, 1018, 1019, 1020, 1021, 1022, 1023}, bins)

	bins, err = fft.BinsByRange(freq, 2048, fft.NegativeFirst, rf.Range{rf.Hz(-5), rf.Hz(5)})
	assert.NoError(t, err)
	assert.Equal(t, []int{1019, 1020, 1021, 1022, 1023, 1024, 1025, 1026, 1027, 1028, 1029}, bins)

	bins, err = fft.BinsByRange(freq, 2048, fft.NegativeFirst, rf.Range{rf.Hz(-10), rf.Hz(0)})
	assert.NoError(t, err)
	assert.Equal(t, []int{1014, 1015, 1016, 1017, 1018, 1019, 1020, 1021, 1022, 1023, 1024}, bins)
}

func TestFreqByBin(t *testing.T) {
	buf := make([]complex64, 2048)

	freq, err := fft.FreqByBin(buf, 2048, false, 10)
	assert.NoError(t, err)
	assert.Equal(t, rf.Hz(10), freq)

	idx, err := fft.BinByFreq(buf, 2048, false, freq)
	assert.NoError(t, err)
	assert.Equal(t, 10, idx)

	freq, err = fft.FreqByBin(buf, 2048, false, 2000)
	assert.NoError(t, err)
	assert.Equal(t, rf.Hz(-48), freq)

	idx, err = fft.BinByFreq(buf, 2048, false, freq)
	assert.NoError(t, err)
	assert.Equal(t, 2000, idx)

	freq, err = fft.FreqByBin(buf, 2048, true, 10)
	assert.NoError(t, err)
	assert.Equal(t, rf.Hz(-1014), freq)

	idx, err = fft.BinByFreq(buf, 2048, true, freq)
	assert.NoError(t, err)
	assert.Equal(t, 10, idx)

	freq, err = fft.FreqByBin(buf, 2048, true, 2000)
	assert.NoError(t, err)
	assert.Equal(t, rf.Hz(976), freq)

	idx, err = fft.BinByFreq(buf, 2048, true, freq)
	assert.NoError(t, err)
	assert.Equal(t, 2000, idx)
}

func TestBinByFreq(t *testing.T) {
	freq := make([]complex64, 2048)
	complexTestArray(freq)

	idx, err := fft.BinByFreq(freq, 2048, false, rf.KHz)
	assert.NoError(t, err)
	assert.Equal(t, complex(float32(1000), 0), freq[idx])

	idx, err = fft.BinByFreq(freq, 2048, false, rf.Hz(-1))
	assert.NoError(t, err)
	assert.Equal(t, complex(float32(-1), 0), freq[idx])

	idx, err = fft.BinByFreq(freq, 2048, false, -rf.KHz)
	assert.NoError(t, err)
	assert.Equal(t, complex(float32(-1000), 0), freq[idx])
}

func TestBinByFreqShifted(t *testing.T) {
	freq := make([]complex64, 2048)
	complexTestArray(freq)
	assert.NoError(t, fft.Shift(freq, 2048))

	idx, err := fft.BinByFreq(freq, 2048, true, rf.KHz)
	assert.NoError(t, err)
	assert.Equal(t, complex(float32(1000), 0), freq[idx])

	idx, err = fft.BinByFreq(freq, 2048, true, -rf.KHz)
	assert.NoError(t, err)
	assert.Equal(t, complex(float32(-1000), 0), freq[idx])
}

func TestFFTShift(t *testing.T) {
	frequency := make([]complex64, 2048)
	complexTestArray(frequency)

	assert.Equal(t, complex(float32(0), 0), frequency[0])
	assert.Equal(t, complex(float32(-1024), 0), frequency[1024])
	assert.Equal(t, complex(float32(1023), 0), frequency[1023])

	assert.NoError(t, fft.Shift(frequency, 2048))
	assert.Equal(t, complex(float32(-1024), 0), frequency[0])
	assert.Equal(t, complex(float32(0), 0), frequency[1024])
	assert.Equal(t, complex(float32(1023), 0), frequency[2047])

	assert.NoError(t, fft.Shift(frequency, 2048))

	assert.Equal(t, complex(float32(0), 0), frequency[0])
	assert.Equal(t, complex(float32(-1024), 0), frequency[1024])
	assert.Equal(t, complex(float32(1023), 0), frequency[1023])
}

// vim: foldmethod=marker
