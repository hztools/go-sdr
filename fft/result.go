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

package fft

import (
	"fmt"

	"hz.tools/rf"
)

var (
	// ErrFrequencyOutOfSamplingRange is returned if the target frequency is too
	// far outside the sampling rate of the underlying sample rate.
	ErrFrequencyOutOfSamplingRange = fmt.Errorf("fft: target frequency is out of sampling rate")
)

// Order specifies what order the fft slice is in.
type Order bool

var (
	// ZeroFirst indicates that the fft data starts with 0, then increases
	// through frequencies to the positive nyquest frequency, then starts
	// at the negative nyquest frequency, back to 0.
	ZeroFirst Order = false

	// NegativeFirst represents what humans understand as an fft, where it
	// starts at the negative nyquest frequency through to the positive
	// nyquest frequency, with 0hz in the center.
	NegativeFirst Order = true
)

// FrequencySlice is the common struct we can use to make sense of the common
// data we need to pass around.
type FrequencySlice struct {
	// Frequency is a slice of frequency space.
	Frequency []complex64

	// SampleRate is the number of readings per second in the time domain
	// used to generate the data input into the FFT.
	SampleRate uint

	// Order is what order bins are in memory -- either ZeroFirst or
	// NegativeFirst. More orders may be added in future, so a switch ought
	// to be used, and default to returning an error case, even if this
	// is not possible given the current type.
	Order Order
}

// NewFrequencySlice will create a new fft.FrequencySlice - which is a struct that represents
// the results of a forward FFT in the frequency domain, *not* any time-domain
// samples. Those should be of type sdr.SamplesC64.
func NewFrequencySlice(frequency []complex64, sampleRate uint, order Order) FrequencySlice {
	return FrequencySlice{
		Frequency:  frequency,
		SampleRate: sampleRate,
		Order:      order,
	}
}

// BinBandwidth is the amount frequency each bin represents in a fft slice.
func (r FrequencySlice) BinBandwidth() rf.Hz {
	return BinBandwidth(len(r.Frequency), r.SampleRate)
}

// Shift will go from ZeroFirst to negativeFirst or vice versa.
func (r FrequencySlice) Shift() (FrequencySlice, error) {
	switch r.Order {
	case ZeroFirst, NegativeFirst:
	default:
		return r, fmt.Errorf("fft.FrequencySlice.Shift: Unknown fft layout")
	}

	zero := len(r.Frequency) / 2
	for i := 0; i < zero; i++ {
		r.Frequency[i], r.Frequency[i+zero] = r.Frequency[i+zero], r.Frequency[i]
	}
	r.Order = !r.Order
	return r, nil
}

// Nyquest is half the sampling rate.
func (r FrequencySlice) Nyquest() rf.Hz {
	return Nyquest(r.SampleRate)
}

// BinsByRange will return the bins representing the range provided.
func (r FrequencySlice) BinsByRange(rng rf.Range) ([]int, error) {
	return BinsByRange(len(r.Frequency), r.SampleRate, r.Order, rng)
}

// FreqByBin will return the center of the bin represented by an offset.
func (r FrequencySlice) FreqByBin(bin int) (rf.Hz, error) {
	return FreqByBin(len(r.Frequency), r.SampleRate, r.Order, bin)
}

// BinByFreq will return the bin index by a provided frequency.
func (r FrequencySlice) BinByFreq(freq rf.Hz) (int, error) {
	return BinByFreq(len(r.Frequency), r.SampleRate, r.Order, freq)
}

// BinBandwidth will return the bandwidth represented by a provided bin.
func BinBandwidth(frequencyLen int, sampleRate uint) rf.Hz {
	return rf.Hz(float32(sampleRate) / float32(frequencyLen))
}

func Nyquest(sampleRate uint) rf.Hz {
	return rf.Hz(sampleRate) / 2
}

func BinsByRange(frequencyLen int, sampleRate uint, order Order, rng rf.Range) ([]int, error) {
	nyquest := Nyquest(sampleRate)
	if rng[1] > nyquest || rng[1] < -nyquest {
		return nil, ErrFrequencyOutOfSamplingRange
	}

	lowFreq := rng[0]
	highFreq := rng[1]

	lowBin, err := BinByFreq(frequencyLen, sampleRate, order, lowFreq)
	if err != nil {
		return nil, err
	}
	highBin, err := BinByFreq(frequencyLen, sampleRate, order, highFreq)
	if err != nil {
		return nil, err
	}

	ret := []int{}

	// If the low end of the range is above 0 Hz, we know that this is purely
	// postivie frequency.
	if lowFreq >= 0 || highFreq < 0 {
		for i := lowBin; i <= highBin; i++ {
			ret = append(ret, i)
		}
		return ret, nil
	}

	switch order {
	case ZeroFirst:
		for i := lowBin; i < frequencyLen; i++ {
			ret = append(ret, i)
		}
		for i := 0; i <= highBin; i++ {
			ret = append(ret, i)
		}
	case NegativeFirst:
		for i := lowBin; i <= highBin; i++ {
			ret = append(ret, i)
		}
	default:
		return nil, fmt.Errorf("fft.FrequencySlice.BinsByRange: Unknown fft layout")
	}
	return ret, nil
}

func FreqByBin(frequencyLen int, sampleRate uint, order Order, bin int) (rf.Hz, error) {
	if bin < 0 || bin > frequencyLen {
		return rf.Hz(0), ErrFrequencyOutOfSamplingRange
	}

	midpoint := frequencyLen / 2
	bw := BinBandwidth(frequencyLen, sampleRate)

	switch order {
	case ZeroFirst:
		if bin > midpoint {
			bin = (bin - frequencyLen)
		}
		return bw * rf.Hz(bin), nil
	case NegativeFirst:
		bin = bin - midpoint
		return bw * rf.Hz(bin), nil
	default:
		return 0, fmt.Errorf("fft.FreqByBin: Unknown fft layout")
	}

	return rf.Hz(0), nil
}

// BinByFreq will return the bin index by a provided frequency.
func BinByFreq(frequencyLen int, sampleRate uint, order Order, freq rf.Hz) (int, error) {
	nyquest := Nyquest(sampleRate)
	if freq > nyquest || freq <= -nyquest {
		return 0, ErrFrequencyOutOfSamplingRange
	}

	// TODO(paultag): Something about this math seems... wrong. The fact that
	// it's not perfectly symmetric threw off the first version of this code
	// so i'm not convinced this is right at the edge between sign changes.

	binIdx := freq / BinBandwidth(frequencyLen, sampleRate)
	switch order {
	case ZeroFirst:
		if binIdx < 0 {
			return (frequencyLen + int(binIdx)), nil
		}
		return int(binIdx), nil
	case NegativeFirst:
		return (frequencyLen / 2) + int(binIdx), nil
	default:
		return 0, fmt.Errorf("fft.FrequencySlice.BinByFreq: Unknown fft layout")
	}
}

// Shift will shift a FFT window from the native 0-index being 0 hz, to
// 0hz being the center of the buffer.
func Shift(frequency []complex64) error {
	r := NewFrequencySlice(frequency, 0, ZeroFirst)
	_, err := r.Shift()
	return err
}

// vim: foldmethod=marker
