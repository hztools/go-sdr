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

package internal

import (
	"math"
	"math/cmplx"

	"hz.tools/sdr"
	"hz.tools/sdr/fft"
)

func conjMult(a, b complex64) complex64 {
	return a * complex(real(b), -imag(b))
}

// CrossCorrelater contains internal members to run CrossCorrelations.
type CrossCorrelater struct {
	fftLength              int
	bufIn1, bufIn2, bufOut sdr.SamplesC64
	bufIn1Freq, bufIn2Freq []complex64

	fftPlan1, fftPlan2 fft.Plan
	fftPlanOut         fft.Plan
}

func (cc CrossCorrelater) run() error {
	if err := cc.fftPlan1.Transform(); err != nil {
		return err
	}
	if err := cc.fftPlan2.Transform(); err != nil {
		return err
	}
	for i := range cc.bufIn1Freq {
		cc.bufIn1Freq[i] = conjMult(cc.bufIn1Freq[i], cc.bufIn2Freq[i])
	}
	return cc.fftPlanOut.Transform()
}

// Correlate will check the alignment of two buffers.
func (cc CrossCorrelater) Correlate(
	buf1, buf2 sdr.SamplesC64,
) ([]complex64, error) {
	if len(buf1) != cc.fftLength || len(buf2) != cc.fftLength {
		return nil, sdr.ErrDstTooSmall
	}

	copy(cc.bufIn1, buf1)
	copy(cc.bufIn2, buf2)

	if err := cc.run(); err != nil {
		return nil, err
	}

	bufOut := make([]complex64, cc.fftLength)
	copy(bufOut, cc.bufOut)
	return bufOut, nil
}

// NewCrossCorrelater will return a new CrossCorrelater ready to check buffers.
func NewCrossCorrelater(planner fft.Planner, fftLength int) (*CrossCorrelater, error) {

	cc := &CrossCorrelater{
		fftLength:  fftLength,
		bufIn1:     make(sdr.SamplesC64, fftLength),
		bufIn2:     make(sdr.SamplesC64, fftLength),
		bufOut:     make(sdr.SamplesC64, fftLength),
		bufIn1Freq: make([]complex64, fftLength),
		bufIn2Freq: make([]complex64, fftLength),
	}

	var err error

	cc.fftPlan1, err = planner(cc.bufIn1, cc.bufIn1Freq, fft.Forward)
	if err != nil {
		return nil, err
	}
	cc.fftPlan2, err = planner(cc.bufIn2, cc.bufIn2Freq, fft.Forward)
	if err != nil {
		return nil, err
	}
	cc.fftPlanOut, err = planner(cc.bufOut, cc.bufIn1Freq, fft.Backward)
	if err != nil {
		return nil, err
	}

	return cc, nil
}

// checkAlignment will cross-convolve each sdr.Reader against the 0th
// reader, and return the alignment offsets. Each offset is from the
// 0th buffer to the nth buffer. The 0th index will always be 0. It's
// included to make indexing easier.
func checkAlignment(planner fft.Planner, readers []sdr.Reader, bufs []sdr.SamplesC64) ([]int, error) {
	ret := make([]int, len(readers))
	ccr, err := NewCrossCorrelater(planner, len(bufs[0]))
	if err != nil {
		return nil, err
	}

	if err := ReadBuffers(readers, bufs); err != nil {
		return nil, err
	}

	for i := 1; i < len(bufs); i++ {
		cc, err := ccr.Correlate(bufs[0], bufs[1])
		if err != nil {
			return nil, err
		}

		var (
			maxPow  float64 = math.Inf(-1)
			maxPowI int     = -1
			leng    int     = bufs[0].Length()
		)

		for ci, el := range cc {
			if el == 0 {
				continue
			}
			pow := float64(real(el)*real(el) + imag(el)*imag(el))
			if pow > maxPow {
				maxPow = pow
				maxPowI = ci
			}
		}

		if maxPowI > (leng / 2) {
			maxPowI = maxPowI - leng
		}
		ret[i] = maxPowI
	}

	return ret, nil
}

func guessAlignment(in [][]int) ([]int, bool) {
	ret := in[0]
	for _, readings := range in {
		for j := range readings {
			if readings[j] != ret[j] {
				return nil, false
			}
		}
	}
	return ret, true
}

func alignReaders(alignments []int, readers []sdr.Reader) (bool, error) {
	// Any *positive* number here means that 0th reader is that many samples
	// behind the nth reader.
	//
	// Any *negative* number here means that the 0th reader is that many
	// samples ahead of the nth reader.
	//
	// Algorithm:
	//
	//   - Determine what the max number is in the slice. If that number is
	//     greater than 0, consume that many records from the 0th reader.
	//     Retry and mark not aligned.
	//
	//   - If the max number is less than 0, we can go from alignment
	//     index to alignment index, and consume that many records
	//     from every nth reader.
	//
	//   - If all numbers are 0, we're in lock.
	//

	var (
		min int
		max int
	)

	for _, alignment := range alignments {
		if alignment < min {
			min = alignment
		}
		if alignment > max {
			max = alignment
		}
	}

	if min == 0 && max == 0 {
		// we're in lock, all samples are aligned
		return true, nil
	}

	if max > 0 {
		// the 0th reader is ahead of at least one of the readers. We need
		// to consume the max number of samples from 0.
		buf := make(sdr.SamplesC64, max)
		_, err := sdr.ReadFull(readers[0], buf)
		if err != nil {
			return false, err
		}
		for i := 1; i < len(alignments); i++ {
			// We can avoid re-entering this function by sliding the window
			// up and cleaning up the slack rather than waiting to hear back
			// what the math here is.
			alignments[i] = alignments[i] - max
		}
	}

	for i, alignment := range alignments {
		if alignment == 0 {
			continue
		}

		// all the numbers should be <= 0. we need to consume that many samples
		// from each reader.

		buf := make(sdr.SamplesC64, -alignment)
		_, err := sdr.ReadFull(readers[i], buf)
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

func PhaseOffsets(readers []sdr.Reader) ([]complex64, error) {
	ret := make([]complex64, len(readers))

	bufs := make([]sdr.SamplesC64, len(readers))
	for i := range bufs {
		bufs[i] = make(sdr.SamplesC64, 1024*64)
	}

	if err := ReadBuffers(readers, bufs); err != nil {
		return nil, err
	}

	phases := make([]float64, len(readers))
	for i := range bufs[0] {
		// Everything is in ref to the 0th buffer.
		for j := 1; j < len(bufs); j++ {
			phases[j] += cmplx.Phase(complex128(conjMult(bufs[0][i], bufs[j][i])))
		}
	}

	for i := range phases {
		phases[i] /= float64(len(bufs[0]))
		ret[i] = complex64(cmplx.Rect(1, phases[i]))
	}

	return ret, nil
}

// AlignReaders will align multiple readers to be in sample lock.
func AlignReaders(planner fft.Planner, readers []sdr.Reader) error {
	lenr := len(readers)

	bufs := make([]sdr.SamplesC64, lenr)
	for i := range bufs {
		// TODO(paultag): Make this tuneable
		bufs[i] = make(sdr.SamplesC64, 1024*64)
	}

	alignments := make([][]int, 10)

	for {
		for i := range alignments {
			var err error
			alignments[i], err = checkAlignment(planner, readers, bufs)
			if err != nil {
				return err
			}
		}
		alignment, ok := guessAlignment(alignments)
		if !ok {
			continue
		}
		aligned, err := alignReaders(alignment, readers)
		if err != nil {
			return err
		}
		if aligned {
			return nil
		}
	}
}

// vim: foldmethod=marker
